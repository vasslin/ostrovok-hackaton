package providerkit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/ostrovok-hackathon-2026/godnota/internal/providers"
)

func id(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:12])
}

func normalize(job Request, data []byte, now time.Time) ([]TransportPart, error) {
	result := []TransportPart{}
	if job.Provider == "aviasales" {
		var response struct {
			Success  bool   `json:"success"`
			Currency string `json:"currency"`
			Data     []struct {
				Origin          string `json:"origin"`
				Destination     string `json:"destination"`
				DepartureAt     string `json:"departure_at"`
				ReturnAt        string `json:"return_at"`
				DurationTo      int    `json:"duration_to"`
				Duration        int    `json:"duration"`
				DurationBack    int    `json:"duration_back"`
				Price           int64  `json:"price"`
				Link            string `json:"link"`
				Transfers       int    `json:"transfers"`
				ReturnTransfers int    `json:"return_transfers"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, err
		}
		if !response.Success || response.Data == nil {
			return nil, errors.New("unsuccessful response")
		}
		q := job.Params.(*providers.AviasalesPricesForDates)
		for _, item := range response.Data {
			dep, err := time.Parse(time.RFC3339, item.DepartureAt)
			if err != nil || item.Origin != q.Origin || item.Destination != q.Destination || !strings.HasPrefix(item.DepartureAt, q.DepartureAt) || item.Price <= 0 || item.Price > 1e12 {
				return nil, errors.New("invalid flight")
			}
			if q.Direct && (item.Transfers != 0 || item.ReturnTransfers != 0) {
				continue
			}
			duration := item.DurationTo
			if duration == 0 && q.OneWay {
				duration = item.Duration
			}
			if duration <= 0 || duration > 2880 {
				return nil, errors.New("invalid duration")
			}
			currency := strings.ToUpper(response.Currency)
			if currency == "" {
				currency = strings.ToUpper(q.Currency)
			}
			if currency != strings.ToUpper(q.Currency) {
				return nil, errors.New("unexpected currency")
			}
			scope := "per_person_one_way"
			returnDeparture, returnArrival := "", ""
			warnings := []string{"Cached price; availability and traveler-specific total require recheck"}
			if !q.OneWay {
				back, err := time.Parse(time.RFC3339, item.ReturnAt)
				if err != nil || !strings.HasPrefix(item.ReturnAt, q.ReturnAt) || !back.After(dep) || item.DurationBack <= 0 || item.DurationBack > 2880 {
					return nil, errors.New("invalid return date")
				}
				scope = "per_person_round_trip"
				returnDeparture = item.ReturnAt
				returnArrival = back.Add(time.Duration(item.DurationBack) * time.Minute).Format(time.RFC3339)
				warnings = append(warnings, "Return departure: "+item.ReturnAt+"; round-trip price covers both directions")
			}
			booking := ""
			u, err := url.Parse(item.Link)
			if err == nil && u.Host == "" && strings.HasPrefix(u.Path, "/search/") {
				booking = "https://www.aviasales.ru" + item.Link
			}
			result = append(result, TransportPart{ID: id(job.Provider, item.Origin, item.Destination, item.DepartureAt, item.ReturnAt, fmt.Sprint(item.Price)), Provider: job.Provider, Source: "cached", FetchedAt: now.UTC().Format(time.RFC3339), Mode: "plane", Origin: item.Origin, Destination: item.Destination, DepartureAt: item.DepartureAt, ArrivalAt: dep.Add(time.Duration(duration) * time.Minute).Format(time.RFC3339), Price: &Money{item.Price * 100, currency, scope}, BookingURL: booking, Warnings: warnings, ReturnDepartureAt: returnDeparture, ReturnArrivalAt: returnArrival, Transfers: item.Transfers, ReturnTransfers: item.ReturnTransfers})
		}
	} else if job.Provider == "yandex_rasp" {
		var response struct {
			Segments []yandexSegment `json:"segments"`
		}
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, err
		}
		if response.Segments == nil {
			return nil, errors.New("missing segments")
		}
		q := job.Params.(*providers.RaspSearch)
		for _, item := range response.Segments {
			if item.HasTransfers {
				if q.Transfers {
					if transfer, ok := normalizeYandexTransfer(item, q.TransportTypes, q.Date, now); ok {
						result = append(result, transfer)
					}
				}
				continue
			}
			dep, e1 := time.Parse(time.RFC3339, item.Departure)
			arr, e2 := time.Parse(time.RFC3339, item.Arrival)
			if e1 != nil || e2 != nil || !arr.After(dep) || item.Thread.UID == "" || item.From.Code == "" || item.To.Code == "" {
				return nil, errors.New("invalid segment")
			}
			if !strings.HasPrefix(item.Departure, q.Date) || item.Thread.TransportType != q.TransportTypes {
				continue
			}
			result = append(result, TransportPart{ID: id(job.Provider, item.Thread.UID, item.Departure, item.From.Code, item.To.Code), Provider: job.Provider, Source: "live", FetchedAt: now.UTC().Format(time.RFC3339), Mode: item.Thread.TransportType, Origin: item.From.Code, Destination: item.To.Code, DepartureAt: item.Departure, ArrivalAt: item.Arrival, Warnings: []string{"Schedule only: price and availability unknown", "Данные: Яндекс.Расписания - https://rasp.yandex.ru/"}})
		}
	}
	return result, nil
}

func demoParts(ctx context.Context, input TripRequest, now time.Time) Parts {
	if len(input.Stops) > 0 {
		return demoItinerary(ctx, input, now)
	}
	result := Parts{TransportParts: []TransportPart{}, HotelParts: []HotelPart{}, Errors: []ProviderError{}, Limitations: []string{"SYNTHETIC DEMO DATA: no real availability, fares or booking", "Hotel check-in uses trip departure date; arrival/check-in compatibility belongs to route-service"}}
	if ctx.Err() != nil {
		return result
	}
	stamp := now.UTC().Format(time.RFC3339)
	start, _ := time.Parse(time.DateOnly, input.Dates.DepartureDate)
	modes := []string{"plane", "train"}
	if input.Transport != nil && len(input.Transport.Modes) > 0 {
		modes = input.Transport.Modes
	}
	for _, mode := range modes {
		duration := 2 * time.Hour
		if mode != "plane" {
			duration = 8 * time.Hour
		}
		if input.Transport != nil && input.Transport.MaxDurationHours > 0 && duration.Hours() > input.Transport.MaxDurationHours {
			continue
		}
		dates := []string{input.Dates.DepartureDate}
		if (input.Transport == nil || input.Transport.RoundTrip == nil || *input.Transport.RoundTrip) && checkout(input) != "" {
			dates = append(dates, checkout(input))
		}
		for leg, date := range dates {
			dep, _ := time.Parse(time.DateOnly, date)
			dep = dep.Add(8 * time.Hour)
			origin, dest := input.Origin.City, input.Destination.City
			if leg == 1 {
				origin, dest = dest, origin
				dep = dep.Add(10 * time.Hour)
			}
			result.TransportParts = append(result.TransportParts, TransportPart{ID: id("demo", mode, origin, dest, date), Provider: "demo_transport", Source: "synthetic", FetchedAt: stamp, Mode: mode, Origin: origin, Destination: dest, DepartureAt: dep.Format(time.RFC3339), ArrivalAt: dep.Add(duration).Format(time.RFC3339), Price: &Money{250000, input.Currency, "per_person_one_way"}, Warnings: []string{"Синтетические данные; UTC times are illustrative"}})
		}
	}
	if input.Hotel == nil || input.Hotel.Needed == nil || *input.Hotel.Needed {
		end, _ := time.Parse(time.DateOnly, checkout(input))
		nights := int64(end.Sub(start).Hours() / 24)
		price := int64(5000)
		stars := 3
		warnings := []string{"Синтетические данные; hotel and conditions are illustrative"}
		if h := input.Hotel; h != nil {
			if h.StarRatingMin != nil {
				stars = max(stars, *h.StarRatingMin)
			}
			if h.PricePerNightMax > 0 {
				price = min(price, int64(h.PricePerNightMax))
			}
			if h.Meal != "" || len(h.Amenities) > 0 || len(h.Location) > 0 || h.NearPlace != "" || len(h.Kinds) > 0 || h.EarlyCheckin != "" || h.LateCheckout != "" {
				warnings = append(warnings, "Detailed accommodation preferences are not verified by the demo adapter")
			}
		}
		if len(input.Travelers.Pets) > 0 {
			warnings = append(warnings, "Pet acceptance requires confirmation")
		}
		free := input.Hotel != nil && input.Hotel.FreeCancellation
		rooms := 1
		if input.Hotel != nil && input.Hotel.Rooms > 0 {
			rooms = input.Hotel.Rooms
		}
		result.HotelParts = append(result.HotelParts, HotelPart{ID: id("demo-hotel", input.Destination.City, input.Dates.DepartureDate, checkout(input), fmt.Sprint(rooms)), Provider: "demo_hotel", Source: "synthetic", FetchedAt: stamp, City: input.Destination.City, Name: "Демо-отель (синтетические данные)", CheckIn: input.Dates.DepartureDate, CheckOut: checkout(input), Stars: stars, Price: &Money{price * 100 * nights * int64(rooms), input.Currency, "stay_all_rooms"}, FreeCancellation: free, Warnings: warnings})
	}
	// Stable order makes integration and demos reproducible.
	slices.SortFunc(result.TransportParts, func(a, b TransportPart) int { return strings.Compare(a.ID, b.ID) })
	filtered := result.TransportParts[:0]
	for _, part := range result.TransportParts {
		part.OriginCity, part.DestinationCity = part.Origin, part.Destination
		if matchesTransport(part, input) {
			filtered = append(filtered, part)
		}
	}
	result.TransportParts = filtered
	return result
}
