package provider

import (
	"bytes"
	"context"
	"crypto/sha256"
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

func (r *Repository) searchHotels(ctx context.Context, trip providerkit.TripRequest, resolved providerkit.ResolvedTrip, result *entity.Result) {
	if trip.Hotel != nil && trip.Hotel.Needed != nil && !*trip.Hotel.Needed {
		return
	}
	issue := func(code, message string) {
		result.Issues = append(result.Issues, &pb.ProviderIssue{Provider: "ratehawk", Code: code, Message: message})
	}
	mode := os.Getenv("RATEHAWK_MODE")
	keyID, key := os.Getenv("RATEHAWK_KEY_ID"), os.Getenv("RATEHAWK_API_KEY")
	if (mode != "sandbox" && mode != "production") || keyID == "" || key == "" {
		issue("credentials_missing", "RateHawk requires partner credentials and RATEHAWK_MODE=sandbox or production.")
		return
	}
	if resolved.Destination.Latitude == nil || resolved.Destination.Longitude == nil {
		result.Plan.Missing = append(result.Plan.Missing, "/resolved/destination/coordinates")
		return
	}
	checkin, err1 := time.Parse(time.DateOnly, resolved.HotelCheckin)
	checkout, err2 := time.Parse(time.DateOnly, resolved.HotelCheckout)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if err1 != nil || err2 != nil || !checkout.After(checkin) || checkout.Sub(checkin) > 30*24*time.Hour || checkin.Before(today) || checkin.After(today.AddDate(0, 0, 730)) {
		result.Plan.Missing = append(result.Plan.Missing, "/resolved/hotel_checkin|hotel_checkout")
		issue("invalid_stay", "Supply confirmed check-in and checkout dates, at most 30 nights apart.")
		return
	}
	if trip.Travelers.Residency == "" {
		result.Plan.Missing = append(result.Plan.Missing, "/travelers/residency")
		return
	}
	if trip.Hotel == nil || trip.Hotel.Rooms != 1 {
		result.Plan.Missing = append(result.Plan.Missing, "/hotel/rooms")
		issue("room_assignment_required", "Hotel search currently requires an explicit single room; multi-room guest assignment is not supported.")
		return
	}
	if trip.Travelers.Adults > 6 || len(trip.Travelers.ChildrenAges) > 4 || trip.Travelers.Adults+len(trip.Travelers.ChildrenAges) > 10 {
		issue("unsupported_guest_count", "The guest count exceeds the supported single-room limits.")
		return
	}
	children := make([]int, 0, len(trip.Travelers.ChildrenAges))
	for _, age := range trip.Travelers.ChildrenAges {
		children = append(children, int(age))
	}
	body := struct {
		Checkin   string  `json:"checkin"`
		Checkout  string  `json:"checkout"`
		Residency string  `json:"residency"`
		Language  string  `json:"language"`
		Currency  string  `json:"currency"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Radius    int     `json:"radius"`
		Guests    []struct {
			Adults   int   `json:"adults"`
			Children []int `json:"children"`
		} `json:"guests"`
	}{Checkin: resolved.HotelCheckin, Checkout: resolved.HotelCheckout, Residency: trip.Travelers.Residency,
		Language: trip.Locale, Currency: trip.Currency, Latitude: *resolved.Destination.Latitude, Longitude: *resolved.Destination.Longitude, Radius: 3000}
	body.Guests = append(body.Guests, struct {
		Adults   int   `json:"adults"`
		Children []int `json:"children"`
	}{trip.Travelers.Adults, children})
	data, err := json.Marshal(body)
	if err != nil {
		issue("invalid_request", "Cannot encode hotel request.")
		return
	}
	base := "https://api-sandbox.ratehawk.com"
	if mode == "production" {
		base = "https://api.ratehawk.com"
	}
	if err := r.wait(ctx, "ratehawk", 1); err != nil {
		return
	}
	requestCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, base+"/api/b2b/v3/search/serp/geo/", bytes.NewReader(data))
	if err != nil {
		issue("invalid_request", "Cannot create hotel request.")
		return
	}
	req.SetBasicAuth(keyID, key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := r.client.Do(req)
	if err != nil {
		issue("provider_error", "RateHawk request failed.")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		issue(statusCode(resp.StatusCode), fmt.Sprintf("RateHawk returned HTTP %d.", resp.StatusCode))
		return
	}
	content, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20+1))
	if err != nil || len(content) > 2<<20 {
		issue("invalid_response", "RateHawk response is unreadable or exceeds 2 MiB.")
		return
	}
	var response struct {
		Status string `json:"status"`
		Data   struct {
			Hotels []struct {
				HID   int64  `json:"hid"`
				ID    string `json:"id"`
				Rates []struct {
					SearchHash     string `json:"search_hash"`
					PaymentOptions struct {
						PaymentTypes []struct {
							ShowAmount       string `json:"show_amount"`
							ShowCurrencyCode string `json:"show_currency_code"`
						} `json:"payment_types"`
					} `json:"payment_options"`
				} `json:"rates"`
			} `json:"hotels"`
		} `json:"data"`
	}
	if json.Unmarshal(content, &response) != nil || response.Status != "ok" || response.Data.Hotels == nil {
		issue("invalid_response", "RateHawk did not return a successful hotel search.")
		return
	}
	state := "live"
	if mode == "sandbox" {
		state = "sandbox_synthetic"
	}
	for _, h := range response.Data.Hotels {
		if h.HID == 0 || strings.TrimSpace(h.ID) == "" {
			continue
		}
		for _, rate := range h.Rates {
			if rate.SearchHash == "" {
				continue
			}
			for index, payment := range rate.PaymentOptions.PaymentTypes {
				price := moneyMinor(json.Number(payment.ShowAmount))
				currency := strings.ToUpper(payment.ShowCurrencyCode)
				if price == nil || currency == "" {
					continue
				}
				ref := sha256.Sum256([]byte(rate.SearchHash))
				result.Offers = append(result.Offers, &pb.Offer{Id: fmt.Sprintf("ratehawk:%d:%x:%d", h.HID, ref, index),
					Provider: "ratehawk", Kind: "hotel", Title: h.ID, PriceMinor: price, Currency: currency,
					PriceScope:   "provider quote for stay; taxes, restrictions and final price require confirmation",
					SourceStatus: state, ObservedAt: time.Now().UTC().Format(time.RFC3339)})
			}
		}
	}
	issue("content_and_prebook_required", "Hotel details, guest restrictions and rate availability require content lookup and prebook before use.")
	issue("filters_unverified", "Requested hotel amenities, location, cancellation and other filters must be checked before presenting a match.")
}
