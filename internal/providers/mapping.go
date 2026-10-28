package providers

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ostrovok-hackathon-2026/godnota/internal/trip"
)

type ResolvedPlace struct {
	City        string   `json:"city,omitempty"`
	CountryCode string   `json:"country_code,omitempty"`
	IATA        string   `json:"iata,omitempty"`
	YandexCode  string   `json:"yandex_code,omitempty"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
}

type ResolvedTrip struct {
	Origin                ResolvedPlace   `json:"origin"`
	Destination           ResolvedPlace   `json:"destination"`
	Stops                 []ResolvedPlace `json:"stops,omitempty"`
	HotelCheckin          string          `json:"hotel_checkin,omitempty"`
	HotelCheckout         string          `json:"hotel_checkout,omitempty"`
	HotelbedsSourceMarket string          `json:"hotelbeds_source_market,omitempty"`
}

type TripPlan struct {
	Discovery   RequestSet `json:"discovery"`
	Search      RequestSet `json:"search"`
	Missing     []string   `json:"missing"`
	Limitations []string   `json:"limitations"`
}

func BuildTripPlan(input trip.TripRequest, resolved ResolvedTrip, now time.Time) (TripPlan, error) {
	plan := TripPlan{
		Discovery:   RequestSet{},
		Search:      RequestSet{},
		Missing:     []string{},
		Limitations: []string{},
	}
	if err := input.ValidateAt(now); err != nil {
		return plan, err
	}
	if err := validateResolvedPlace("origin", resolved.Origin); err != nil {
		return plan, err
	}
	if err := validateResolvedPlace("destination", resolved.Destination); err != nil {
		return plan, err
	}

	addDiscovery(&plan, input, resolved)
	if len(plan.Discovery["nominatim"]) > 0 {
		plan.Limitations = append(plan.Limitations, "Nominatim: только точечный геокодинг по запросу пользователя; результаты кэшировать и соблюдать лимит 1 запрос/с")
	}
	if input.Destination == nil || strings.TrimSpace(input.Destination.City) == "" {
		plan.Missing = append(plan.Missing, "/destination/city")
		return plan, nil
	}

	departure := input.Dates.DepartureDate
	if departure == "" {
		plan.Missing = append(plan.Missing, "/dates/departure_date")
	}
	returnDate := input.Dates.ReturnDate
	if returnDate == "" && departure != "" && input.Dates.Nights > 0 {
		date, _ := time.Parse(time.DateOnly, departure)
		returnDate = date.AddDate(0, 0, input.Dates.Nights).Format(time.DateOnly)
	}

	modes := []string{"plane", "train", "suburban", "bus", "water", "helicopter"}
	roundTrip := true
	directOnly := false
	if input.Transport != nil {
		if len(input.Transport.Modes) > 0 {
			modes = input.Transport.Modes
		}
		if input.Transport.RoundTrip != nil {
			roundTrip = *input.Transport.RoundTrip
		}
		directOnly = input.Transport.DirectOnly
	}
	if roundTrip && returnDate == "" && input.Dates.Nights == 0 {
		plan.Missing = append(plan.Missing, "/dates/return_date|nights")
	}

	if slices.Contains(modes, "plane") {
		if resolved.Origin.IATA == "" {
			plan.Missing = append(plan.Missing, "/resolved/origin/iata")
		}
		if resolved.Destination.IATA == "" {
			plan.Missing = append(plan.Missing, "/resolved/destination/iata")
		}
		if resolved.Origin.IATA != "" && resolved.Destination.IATA != "" {
			outboundDate := departure
			if outboundDate == "" {
				outboundDate = input.Dates.DepartureMonth
			}
			if outboundDate != "" {
				params := AviasalesPricesForDates{
					Origin: resolved.Origin.IATA, Destination: resolved.Destination.IATA,
					DepartureAt: outboundDate, OneWay: !roundTrip || returnDate == "",
					Direct: directOnly, Currency: strings.ToLower(input.Currency),
					Sorting: "price", Limit: 30, Page: 1,
				}
				if roundTrip {
					params.ReturnAt = returnDate
				}
				addRequest(plan.Search, "aviasales", "prices_for_dates", &params)
			}
		}
		plan.Limitations = append(plan.Limitations, "aviasales: цена из кэша, наличие и итоговую цену нужно проверить перед выбором")
	}

	if departure != "" {
		if resolved.Origin.YandexCode == "" {
			plan.Missing = append(plan.Missing, "/resolved/origin/yandex_code")
		}
		if resolved.Destination.YandexCode == "" {
			plan.Missing = append(plan.Missing, "/resolved/destination/yandex_code")
		}
		if resolved.Origin.YandexCode != "" && resolved.Destination.YandexCode != "" {
			for _, mode := range modes {
				params := RaspSearch{
					From: resolved.Origin.YandexCode, To: resolved.Destination.YandexCode,
					Date: departure, TransportTypes: mode, Transfers: !directOnly,
					Limit: 100, Lang: "ru_RU", Format: "json",
				}
				addRequest(plan.Search, "yandex_rasp", "search", &params)
				if roundTrip && returnDate != "" {
					back := params
					back.From, back.To = params.To, params.From
					back.Date = returnDate
					addRequest(plan.Search, "yandex_rasp", "search", &back)
				}
			}
		}
		plan.Limitations = append(plan.Limitations, "yandex_rasp: расписание может не содержать цену и наличие билетов")
	}

	if input.Hotel == nil || input.Hotel.Needed == nil || *input.Hotel.Needed {
		plan.Limitations = append(plan.Limitations, "поиск отелей выполняется отдельно и требует подтверждённых дат, гражданства гостей и партнёрского доступа")
	}

	if hasCoordinates(resolved.Destination) {
		limit := 20
		if input.Leisure != nil && input.Leisure.MaxItemsPerStop > 0 {
			limit = min(50, max(20, input.Leisure.MaxItemsPerStop*4))
		}
		params := WikipediaGeosearch{
			Action: "query", Generator: "geosearch",
			Coord:  strconv.FormatFloat(*resolved.Destination.Latitude, 'f', 6, 64) + "|" + strconv.FormatFloat(*resolved.Destination.Longitude, 'f', 6, 64),
			Radius: 3000, Limit: limit,
			Prop: "coordinates|pageimages|description|info", InProp: "url",
			PIProp: "thumbnail", PIThumbSize: 400, Format: "json", FormatVersion: 2,
		}
		addRequest(plan.Search, "wikipedia", "geosearch", &params)
	} else if input.Leisure != nil && len(input.Leisure.Interests) > 0 {
		plan.Missing = append(plan.Missing, "/resolved/destination/coordinates")
	}
	if input.Leisure != nil && input.Leisure.DuringLayovers != nil && *input.Leisure.DuringLayovers {
		plan.Limitations = append(plan.Limitations, "досуг на пересадке ищется после получения конкретного сегмента, координат остановки и свободного времени")
	}
	if input.Leisure != nil && len(input.Leisure.Interests) > 0 {
		plan.Limitations = append(plan.Limitations, "Wikimedia возвращает достопримечательности; афишу и наличие билетов на события этот запрос не подтверждает")
	}
	for _, requests := range []RequestSet{plan.Discovery, plan.Search} {
		for _, group := range requests {
			for _, request := range group {
				if err := Validate(request.Params); err != nil {
					return plan, fmt.Errorf("%s.%s: %w", request.Provider, request.Endpoint, err)
				}
			}
		}
	}
	return plan, nil
}

func addDiscovery(plan *TripPlan, input trip.TripRequest, resolved ResolvedTrip) {
	for _, place := range []struct {
		city     string
		resolved ResolvedPlace
	}{
		{input.Origin.City, resolved.Origin},
		{destinationCity(input), resolved.Destination},
	} {
		if strings.TrimSpace(place.city) == "" {
			continue
		}
		if place.resolved.IATA == "" {
			params := TravelpayoutsPlaces{Term: place.city, Locale: input.Locale, Types: []string{"city"}}
			addRequest(plan.Discovery, "travelpayouts_places", "places2", &params)
		}
		if !hasCoordinates(place.resolved) {
			params := NominatimSearch{Q: place.city, Format: "jsonv2", Limit: 5, AcceptLanguage: input.Locale}
			addRequest(plan.Discovery, "nominatim", "search", &params)
		} else if place.resolved.YandexCode == "" {
			params := RaspNearestSettlement{Lat: *place.resolved.Latitude, Lng: *place.resolved.Longitude, Distance: 50, Lang: "ru_RU", Format: "json"}
			addRequest(plan.Discovery, "yandex_rasp", "nearest_settlement", &params)
		}
	}
}

func addRequest(set RequestSet, provider, endpoint string, params any) {
	set[provider] = append(set[provider], Request{Provider: provider, Endpoint: endpoint, Params: params})
}

func destinationCity(input trip.TripRequest) string {
	if input.Destination == nil {
		return ""
	}
	return input.Destination.City
}

func hasCoordinates(place ResolvedPlace) bool {
	return place.Latitude != nil && place.Longitude != nil
}

func validateResolvedPlace(name string, place ResolvedPlace) error {
	if place.IATA != "" {
		if len(place.IATA) != 3 {
			return fmt.Errorf("/resolved/%s/iata: must be three uppercase letters", name)
		}
		for _, char := range place.IATA {
			if char < 'A' || char > 'Z' {
				return fmt.Errorf("/resolved/%s/iata: must be three uppercase letters", name)
			}
		}
	}
	if place.YandexCode != "" {
		if len(place.YandexCode) < 2 || (place.YandexCode[0] != 'c' && place.YandexCode[0] != 's') {
			return fmt.Errorf("/resolved/%s/yandex_code: invalid Yandex code", name)
		}
		for _, char := range place.YandexCode[1:] {
			if char < '0' || char > '9' {
				return fmt.Errorf("/resolved/%s/yandex_code: invalid Yandex code", name)
			}
		}
	}
	if (place.Latitude == nil) != (place.Longitude == nil) {
		return fmt.Errorf("/resolved/%s/coordinates: latitude and longitude must be set together", name)
	}
	if hasCoordinates(place) && !validCoordinates(*place.Latitude, *place.Longitude) {
		return fmt.Errorf("/resolved/%s/coordinates: invalid latitude/longitude", name)
	}
	return nil
}
