package provider

import (
	"math"
	"strings"

	pb "github.com/godnota/pkg/generated/provider/v1"
	"github.com/godnota/provider-service/internal/entity"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
)

type poiIndex struct {
	Attribution string `json:"attribution"`
	ObservedAt  string `json:"observed_at"`
	Coverage    string `json:"coverage"`
	Items       []struct {
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		Kind       string  `json:"kind"`
		Latitude   float64 `json:"latitude"`
		Longitude  float64 `json:"longitude"`
		SourceURL  string  `json:"source_url"`
		ObservedAt string  `json:"observed_at,omitempty"`
	} `json:"items"`
}

func (r *Repository) searchPOI(trip providerkit.TripRequest, resolved providerkit.ResolvedTrip, result *entity.Result) {
	lat, lon := resolved.Destination.Latitude, resolved.Destination.Longitude
	if lat == nil || lon == nil {
		return
	}
	wanted := map[string]bool{}
	if trip.Leisure != nil {
		for _, interest := range trip.Leisure.Interests {
			switch interest {
			case "museums":
				wanted["museum"] = true
			case "theatres":
				wanted["theatre"] = true
			case "cinemas":
				wanted["cinema"] = true
			}
		}
	}
	if len(wanted) == 0 {
		return
	}
	for _, item := range r.poi.Items {
		if !wanted[item.Kind] || item.Name == "" || !strings.HasPrefix(item.SourceURL, "https://www.openstreetmap.org/") || distanceMeters(*lat, *lon, item.Latitude, item.Longitude) > 3000 {
			continue
		}
		placeLat, placeLon := item.Latitude, item.Longitude
		observed := item.ObservedAt
		if observed == "" {
			observed = r.poi.ObservedAt
		}
		result.Offers = append(result.Offers, &pb.Offer{
			Id: item.ID, Provider: "osm_index", Kind: item.Kind, Title: item.Name,
			Latitude: &placeLat, Longitude: &placeLon, SourceUrl: item.SourceURL,
			Attribution: r.poi.Attribution, SourceStatus: "offline_snapshot", ObservedAt: observed,
			PriceScope: "unknown", Description: "Opening hours, tickets and availability are not confirmed.",
		})
	}
	result.Issues = append(result.Issues, &pb.ProviderIssue{Provider: "osm_index", Code: "partial_coverage", Message: r.poi.Coverage})
}

func distanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const radians = math.Pi / 180
	dlat, dlon := (lat2-lat1)*radians, (lon2-lon1)*radians
	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1*radians)*math.Cos(lat2*radians)*math.Sin(dlon/2)*math.Sin(dlon/2)
	return 6371000 * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
