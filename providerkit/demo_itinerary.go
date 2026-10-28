package providerkit

import (
	"context"
	"fmt"
	"time"
)

func demoItinerary(ctx context.Context, input TripRequest, now time.Time) Parts {
	r := Parts{TransportParts: []TransportPart{}, HotelParts: []HotelPart{}, Errors: []ProviderError{}, Limitations: []string{"SYNTHETIC DEMO DATA: no real availability, fares or booking", "Stop dates and transport times require confirmation."}}
	if ctx.Err() != nil {
		return r
	}
	modes := []string{"plane", "train"}
	if input.Transport != nil && len(input.Transport.Modes) > 0 {
		modes = input.Transport.Modes
	}
	stamp := now.UTC().Format(time.RFC3339)
	from := input.Origin.City
	fromDate := input.Dates.DepartureDate
	for _, stop := range input.Stops {
		for _, mode := range modes {
			duration := 2 * time.Hour
			if mode != "plane" {
				duration = 8 * time.Hour
			}
			if input.Transport != nil && input.Transport.MaxDurationHours > 0 && duration.Hours() > input.Transport.MaxDurationHours {
				continue
			}
			departure, _ := time.Parse(time.DateOnly, fromDate)
			departure = departure.Add(8 * time.Hour)
			p := TransportPart{ID: id("demo", mode, from, stop.City, stop.CheckIn), Provider: "demo_transport", Source: "synthetic", FetchedAt: stamp, Mode: mode, Origin: from, Destination: stop.City, OriginCity: from, DestinationCity: stop.City, DepartureAt: departure.Format(time.RFC3339), ArrivalAt: departure.Add(duration).Format(time.RFC3339), Price: &Money{250000, input.Currency, "per_person_one_way"}, Warnings: []string{"Synthetic demonstration transport."}}
			if matchesTransport(p, input) {
				r.TransportParts = append(r.TransportParts, p)
			}
		}
		if input.Hotel == nil || input.Hotel.Needed == nil || *input.Hotel.Needed {
			checkin, _ := time.Parse(time.DateOnly, stop.CheckIn)
			checkout, _ := time.Parse(time.DateOnly, stop.CheckOut)
			nights := int64(checkout.Sub(checkin).Hours() / 24)
			rooms := 1
			stars := 3
			price := int64(5000)
			free := false
			if h := input.Hotel; h != nil {
				if h.Rooms > 0 {
					rooms = h.Rooms
				}
				if h.StarRatingMin != nil {
					stars = max(stars, *h.StarRatingMin)
				}
				if h.PricePerNightMax > 0 {
					price = min(price, int64(h.PricePerNightMax))
				}
				free = h.FreeCancellation
			}
			r.HotelParts = append(r.HotelParts, HotelPart{ID: id("demo-hotel", stop.City, stop.CheckIn, stop.CheckOut, fmt.Sprint(rooms)), Provider: "demo_hotel", Source: "synthetic", FetchedAt: stamp, City: stop.City, Name: "Демо-отель (синтетические данные)", CheckIn: stop.CheckIn, CheckOut: stop.CheckOut, Stars: stars, Price: &Money{price * 100 * nights * int64(rooms), input.Currency, "stay_all_rooms"}, FreeCancellation: free, Warnings: []string{"Synthetic demonstration hotel; amenities and availability require confirmation."}})
		}
		from = stop.City
		fromDate = stop.CheckOut
	}
	if input.Transport == nil || input.Transport.RoundTrip == nil || *input.Transport.RoundTrip {
		for _, mode := range modes {
			duration := 2 * time.Hour
			if mode != "plane" {
				duration = 8 * time.Hour
			}
			if input.Transport != nil && input.Transport.MaxDurationHours > 0 && duration.Hours() > input.Transport.MaxDurationHours {
				continue
			}
			departure, _ := time.Parse(time.DateOnly, input.Dates.ReturnDate)
			departure = departure.Add(18 * time.Hour)
			p := TransportPart{ID: id("demo", mode, from, input.Origin.City, input.Dates.ReturnDate), Provider: "demo_transport", Source: "synthetic", FetchedAt: stamp, Mode: mode, Origin: from, Destination: input.Origin.City, OriginCity: from, DestinationCity: input.Origin.City, DepartureAt: departure.Format(time.RFC3339), ArrivalAt: departure.Add(duration).Format(time.RFC3339), Price: &Money{250000, input.Currency, "per_person_one_way"}, Warnings: []string{"Synthetic demonstration transport."}}
			if matchesTransport(p, input) {
				r.TransportParts = append(r.TransportParts, p)
			}
		}
	}
	return r
}
