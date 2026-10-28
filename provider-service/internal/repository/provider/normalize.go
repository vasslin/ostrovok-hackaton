package provider

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	pb "github.com/godnota/pkg/generated/provider/v1"
	"google.golang.org/protobuf/proto"
)

type providerAPIError struct{ code string }

func (e providerAPIError) Error() string { return "provider reported an API error" }

func moneyMinor(value json.Number) *int64 {
	amount, ok := new(big.Rat).SetString(value.String())
	if !ok || amount.Sign() < 0 {
		return nil
	}
	amount.Mul(amount, big.NewRat(100, 1))
	if !amount.IsInt() || !amount.Num().IsInt64() {
		return nil
	}
	n := amount.Num().Int64()
	return &n
}

func flightOffers(data []byte, currency string, at time.Time) ([]*pb.Offer, error) {
	var response struct {
		Data []struct {
			FlightNumber string      `json:"flight_number"`
			Airline      string      `json:"airline"`
			Origin       string      `json:"origin"`
			Destination  string      `json:"destination"`
			Departure    string      `json:"departure_at"`
			Price        json.Number `json:"price"`
			Link         string      `json:"link"`
		} `json:"data"`
		Success *bool `json:"success"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&response); err != nil {
		return nil, err
	}
	if response.Success != nil && !*response.Success {
		return nil, fmt.Errorf("flight provider reported failure")
	}
	if response.Data == nil {
		return nil, fmt.Errorf("missing data")
	}
	var out []*pb.Offer
	for _, f := range response.Data {
		if f.Origin == "" || f.Destination == "" || f.Departure == "" {
			continue
		}
		link := ""
		if strings.HasPrefix(f.Link, "/search/") {
			link = "https://www.aviasales.ru" + f.Link
		}
		key := sha256.Sum256([]byte(f.Link + "|" + f.Airline + "|" + f.FlightNumber + "|" + f.Departure + "|" + f.Price.String()))
		out = append(out, &pb.Offer{
			Id:       fmt.Sprintf("aviasales:%x", key),
			Provider: "aviasales", Kind: "plane", Title: strings.TrimSpace(f.Airline + " " + f.FlightNumber),
			Origin: f.Origin, Destination: f.Destination, Departure: f.Departure,
			PriceMinor: moneyMinor(f.Price), Currency: currency, PriceScope: "cached quote; passenger count and final availability unconfirmed",
			SourceStatus: "provider_cached", SourceUrl: link, ObservedAt: at.Format(time.RFC3339),
		})
	}
	return out, nil
}

func scheduleOffers(data []byte, at time.Time, cached bool) ([]*pb.Offer, error) {
	var response struct {
		Segments []struct {
			Thread struct {
				UID           string `json:"uid"`
				Title         string `json:"title"`
				TransportType string `json:"transport_type"`
			} `json:"thread"`
			From struct {
				Title string `json:"title"`
			} `json:"from"`
			To struct {
				Title string `json:"title"`
			} `json:"to"`
			Departure   string `json:"departure"`
			Arrival     string `json:"arrival"`
			TicketsInfo struct {
				Places []struct {
					Name     string `json:"name"`
					Currency string `json:"currency"`
					Price    *struct {
						Whole *int64 `json:"whole"`
						Cents *int64 `json:"cents"`
					} `json:"price"`
				} `json:"places"`
			} `json:"tickets_info"`
		} `json:"segments"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}
	if response.Segments == nil {
		return nil, fmt.Errorf("missing segments")
	}
	state := "live"
	if cached {
		state = "service_cached"
	}
	var out []*pb.Offer
	for _, s := range response.Segments {
		if s.Thread.UID == "" || s.Departure == "" || s.Arrival == "" {
			continue
		}
		base := &pb.Offer{
			Id: s.Thread.UID + ":" + s.Departure, Provider: "yandex_rasp", Kind: s.Thread.TransportType,
			Title: s.Thread.Title, Origin: s.From.Title, Destination: s.To.Title,
			Departure: s.Departure, Arrival: s.Arrival, PriceScope: "unknown",
			SourceStatus: state, SourceUrl: "https://rasp.yandex.ru/",
			Attribution: "Данные предоставлены сервисом Яндекс.Расписания", ObservedAt: at.Format(time.RFC3339),
		}
		if len(s.TicketsInfo.Places) == 0 {
			out = append(out, base)
			continue
		}
		for index, place := range s.TicketsInfo.Places {
			offer := proto.Clone(base).(*pb.Offer)
			offer.Id = fmt.Sprintf("%s:%d", base.Id, index)
			if place.Name != "" {
				offer.Title = strings.TrimSpace(base.Title + " — " + place.Name)
			}
			currency := strings.ToUpper(place.Currency)
			if place.Price != nil && place.Price.Whole != nil && place.Price.Cents != nil &&
				(currency == "RUB" || currency == "USD" || currency == "EUR") &&
				*place.Price.Whole >= 0 && *place.Price.Whole <= (1<<63-1-99)/100 &&
				*place.Price.Cents >= 0 && *place.Price.Cents < 100 {
				amount := *place.Price.Whole*100 + *place.Price.Cents
				offer.PriceMinor = &amount
				offer.Currency = currency
				offer.PriceScope = "provider ticket quote; passenger count and availability unconfirmed"
			}
			out = append(out, offer)
		}
	}
	return out, nil
}

func wikipediaOffers(data []byte, at time.Time, cached bool) ([]*pb.Offer, error) {
	var response struct {
		Error *struct {
			Code string `json:"code"`
		} `json:"error"`
		Query struct {
			Pages []struct {
				ID          int64  `json:"pageid"`
				Title       string `json:"title"`
				Description string `json:"description"`
				URL         string `json:"fullurl"`
				Coordinates []struct {
					Lat float64 `json:"lat"`
					Lon float64 `json:"lon"`
				} `json:"coordinates"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}
	if response.Error != nil {
		if response.Error.Code == "cirrussearch-too-busy-error" {
			return nil, providerAPIError{code: "upstream_busy"}
		}
		return nil, providerAPIError{code: "provider_error"}
	}
	if response.Query.Pages == nil {
		return nil, fmt.Errorf("missing pages")
	}
	state := "live"
	if cached {
		state = "service_cached"
	}
	var out []*pb.Offer
	for _, p := range response.Query.Pages {
		if p.ID == 0 || p.Title == "" || len(p.Coordinates) == 0 {
			continue
		}
		source := ""
		if u, err := url.Parse(p.URL); err == nil && u.Scheme == "https" && u.Hostname() == "ru.wikipedia.org" {
			source = p.URL
		}
		lat, lon := p.Coordinates[0].Lat, p.Coordinates[0].Lon
		out = append(out, &pb.Offer{
			Id: fmt.Sprintf("wiki:%d", p.ID), Provider: "wikipedia", Kind: "attraction", Title: p.Title,
			Latitude: &lat, Longitude: &lon, Description: p.Description, SourceUrl: source,
			SourceStatus: state, PriceScope: "unknown", Attribution: "Wikipedia, CC BY-SA; see article for authors",
			ObservedAt: at.Format(time.RFC3339),
		})
	}
	return out, nil
}
