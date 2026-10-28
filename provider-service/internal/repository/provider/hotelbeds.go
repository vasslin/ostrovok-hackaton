package provider

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	pb "github.com/godnota/pkg/generated/provider/v1"
	"github.com/godnota/provider-service/internal/entity"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
)

type hotelbedsStay struct {
	CheckIn  string `json:"checkIn"`
	CheckOut string `json:"checkOut"`
}

type hotelbedsPax struct {
	Type string `json:"type"`
	Age  int    `json:"age"`
}

type hotelbedsOccupancy struct {
	Rooms    int            `json:"rooms"`
	Adults   int            `json:"adults"`
	Children int            `json:"children"`
	Paxes    []hotelbedsPax `json:"paxes,omitempty"`
}

type hotelbedsGeo struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    int     `json:"radius"`
	Unit      string  `json:"unit"`
}

type hotelbedsAvailability struct {
	Stay         hotelbedsStay        `json:"stay"`
	Occupancies  []hotelbedsOccupancy `json:"occupancies"`
	Geolocation  hotelbedsGeo         `json:"geolocation"`
	SourceMarket string               `json:"sourceMarket"`
}

func hotelbedsSignature(key, secret string, now time.Time) string {
	sum := sha256.Sum256([]byte(key + secret + fmt.Sprint(now.Unix())))
	return hex.EncodeToString(sum[:])
}

func (r *Repository) searchHotelbeds(ctx context.Context, trip providerkit.TripRequest, resolved providerkit.ResolvedTrip, result *entity.Result) {
	if trip.Hotel != nil && trip.Hotel.Needed != nil && !*trip.Hotel.Needed {
		return
	}
	issue := func(code, message string) {
		result.Issues = append(result.Issues, &pb.ProviderIssue{Provider: "hotelbeds", Code: code, Message: message})
	}
	mode, key, secret := os.Getenv("HOTELBEDS_MODE"), os.Getenv("HOTELBEDS_API_KEY"), os.Getenv("HOTELBEDS_SECRET")
	if mode != "evaluation" || key == "" || secret == "" {
		issue("credentials_missing", "Set HOTELBEDS_MODE=evaluation and supply the evaluation key and secret.")
		return
	}
	if resolved.Destination.Latitude == nil || resolved.Destination.Longitude == nil {
		result.Plan.Missing = append(result.Plan.Missing, "/resolved/destination/coordinates")
		return
	}
	checkin, e1 := time.Parse(time.DateOnly, resolved.HotelCheckin)
	checkout, e2 := time.Parse(time.DateOnly, resolved.HotelCheckout)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if e1 != nil || e2 != nil || checkin.Before(today) || !checkout.After(checkin) || checkout.Sub(checkin) > 30*24*time.Hour {
		result.Plan.Missing = append(result.Plan.Missing, "/resolved/hotel_checkin|hotel_checkout")
		issue("invalid_stay", "Supply confirmed future check-in and checkout dates, at most 30 nights apart.")
		return
	}
	market := resolved.HotelbedsSourceMarket
	if len(market) != 2 || market[0] < 'A' || market[0] > 'Z' || market[1] < 'A' || market[1] > 'Z' {
		result.Plan.Missing = append(result.Plan.Missing, "/resolved/hotelbeds_source_market")
		return
	}
	if trip.Hotel == nil || trip.Hotel.Rooms != 1 {
		result.Plan.Missing = append(result.Plan.Missing, "/hotel/rooms")
		issue("room_assignment_required", "Hotelbeds search currently requires one explicitly specified room.")
		return
	}
	occupancy := hotelbedsOccupancy{Rooms: 1, Adults: trip.Travelers.Adults, Children: len(trip.Travelers.ChildrenAges)}
	for _, age := range trip.Travelers.ChildrenAges {
		occupancy.Paxes = append(occupancy.Paxes, hotelbedsPax{Type: "CH", Age: int(age)})
	}
	body := hotelbedsAvailability{
		Stay:         hotelbedsStay{CheckIn: resolved.HotelCheckin, CheckOut: resolved.HotelCheckout},
		Occupancies:  []hotelbedsOccupancy{occupancy},
		Geolocation:  hotelbedsGeo{Latitude: *resolved.Destination.Latitude, Longitude: *resolved.Destination.Longitude, Radius: 3, Unit: "km"},
		SourceMarket: market,
	}
	data, err := json.Marshal(body)
	if err != nil {
		issue("invalid_request", "Cannot encode Hotelbeds request.")
		return
	}
	if err := r.wait(ctx, "hotelbeds", 1); err != nil {
		return
	}
	requestCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, "https://api.test.hotelbeds.com/hotel-api/1.0/hotels", bytes.NewReader(data))
	if err != nil {
		issue("invalid_request", "Cannot create Hotelbeds request.")
		return
	}
	req.Header.Set("Api-key", key)
	req.Header.Set("X-Signature", hotelbedsSignature(key, secret, time.Now()))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := r.client.Do(req)
	if err != nil {
		issue("provider_error", "Hotelbeds request failed.")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		issue(statusCode(resp.StatusCode), fmt.Sprintf("Hotelbeds returned HTTP %d.", resp.StatusCode))
		return
	}
	content, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20+1))
	if err != nil || len(content) > 2<<20 {
		issue("invalid_response", "Hotelbeds response is unreadable or exceeds 2 MiB.")
		return
	}
	var response struct {
		Hotels *struct {
			Currency string `json:"currency"`
			Hotels   []struct {
				Code     int64  `json:"code"`
				Name     string `json:"name"`
				Currency string `json:"currency"`
				Rooms    []struct {
					Name  string `json:"name"`
					Rates []struct {
						RateKey  string      `json:"rateKey"`
						RateType string      `json:"rateType"`
						Net      json.Number `json:"net"`
					} `json:"rates"`
				} `json:"rooms"`
			} `json:"hotels"`
		} `json:"hotels"`
	}
	if json.Unmarshal(content, &response) != nil || response.Hotels == nil {
		issue("invalid_response", "Hotelbeds did not return an availability response.")
		return
	}
	for _, hotel := range response.Hotels.Hotels {
		if hotel.Code == 0 || strings.TrimSpace(hotel.Name) == "" {
			continue
		}
		currency := strings.ToUpper(hotel.Currency)
		if currency == "" {
			currency = strings.ToUpper(response.Hotels.Currency)
		}
		for _, room := range hotel.Rooms {
			for _, rate := range room.Rates {
				if rate.RateKey == "" {
					continue
				}
				var price *int64
				if currency == "RUB" || currency == "EUR" || currency == "USD" {
					price = moneyMinor(rate.Net)
				}
				if price == nil {
					continue
				}
				ref := sha256.Sum256([]byte(rate.RateKey))
				result.Offers = append(result.Offers, &pb.Offer{
					Id:       fmt.Sprintf("hotelbeds:%d:%x", hotel.Code, ref),
					Provider: "hotelbeds", Kind: "hotel", Title: hotel.Name,
					PriceMinor: price, Currency: currency, SourceStatus: "sandbox_evaluation",
					PriceScope:  "evaluation net rate; final consumer price and availability are not confirmed",
					Description: room.Name + " (" + rate.RateType + ")", ObservedAt: time.Now().UTC().Format(time.RFC3339),
				})
			}
		}
	}
	issue("sandbox_only", "Evaluation data are not production offers; review availability, fees and restrictions before presenting or booking.")
	issue("filters_unverified", "Requested hotel amenities, location, cancellation and other filters must be checked before presenting a match.")
}
