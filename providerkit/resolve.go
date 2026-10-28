package providerkit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ostrovok-hackathon-2026/godnota/internal/providers"
)

type discoveryCache struct {
	Data    json.RawMessage
	Expires time.Time
}

var cityIATA = regexp.MustCompile(`^[A-Z]{3}$`)
var settlementCode = regexp.MustCompile(`^c[0-9]+$`)

func cityKey(city string) string {
	return strings.ReplaceAll(strings.ToLower(strings.Join(strings.Fields(city), " ")), "ё", "е")
}

func (s *Searcher) discovery(ctx context.Context, job Request, out any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	params, _ := json.Marshal(job.Params)
	key := job.Provider + ":" + job.Endpoint + ":" + string(params)
	s.mu.Lock()
	cached, ok := s.discoveryCache[key]
	s.mu.Unlock()
	if ok && s.Now().Before(cached.Expires) {
		return json.Unmarshal(cached.Data, out)
	}
	cfg := s.Registry[job.Provider]
	timeout := time.Duration(cfg.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = s.Timeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := s.waitSlot(ctx, job.Provider, cfg.RPS); err != nil {
		return fmt.Errorf("справочник не ответил вовремя")
	}
	req, err := s.Registry.NewHTTPRequest(ctx, job)
	if err != nil {
		return fmt.Errorf("справочник не настроен")
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("справочник временно недоступен")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("справочник вернул HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, (2<<20)+1))
	if err != nil || len(data) > 2<<20 || json.Unmarshal(data, out) != nil {
		return fmt.Errorf("справочник вернул некорректные данные")
	}
	if cfg.CacheTTLS > 0 {
		s.mu.Lock()
		if s.discoveryCache == nil {
			s.discoveryCache = map[string]discoveryCache{}
		}
		for k, v := range s.discoveryCache {
			if !s.Now().Before(v.Expires) {
				delete(s.discoveryCache, k)
			}
		}
		if len(s.discoveryCache) >= 2048 {
			s.discoveryCache = map[string]discoveryCache{}
		}
		s.discoveryCache[key] = discoveryCache{Data: data, Expires: s.Now().Add(time.Duration(cfg.CacheTTLS) * time.Second)}
		s.mu.Unlock()
	}
	return nil
}

func (s *Searcher) ResolveTrip(ctx context.Context, input TripRequest, resolved ResolvedTrip) (ResolvedTrip, []ProviderError) {
	if s.Mode != "live" || (os.Getenv(s.Registry["aviasales"].Auth.Env) == "" && os.Getenv(s.Registry["yandex_rasp"].Auth.Env) == "") {
		return resolved, nil
	}
	type place struct {
		city, country string
		target        *ResolvedPlace
	}
	targets := []place{{input.Origin.City, input.Origin.CountryCode, &resolved.Origin}}
	if input.Destination != nil {
		targets = append(targets, place{input.Destination.City, input.Destination.CountryCode, &resolved.Destination})
	}
	stops := make([]ResolvedPlace, len(input.Stops))
	copy(stops, resolved.Stops)
	resolved.Stops = stops
	for i, stop := range input.Stops {
		targets = append(targets, place{stop.City, "", &resolved.Stops[i]})
	}
	issues := make([]*ProviderError, len(targets))
	var wg sync.WaitGroup
	for i, p := range targets {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if p.target.City != "" && (cityKey(p.target.City) != cityKey(p.city) || p.target.CountryCode != p.country) {
				*p.target = ResolvedPlace{}
			}
			p.target.City, p.target.CountryCode = p.city, p.country
			if err := s.resolvePlace(ctx, input.Locale, p.city, p.country, p.target); err != nil {
				code := "resolution_failed"
				if strings.Contains(err.Error(), "Уточните") {
					code = "clarification_required"
				}
				issues[i] = &ProviderError{Provider: "places", Kind: "transport", Code: code, Message: p.city + ": " + err.Error(), Retryable: code == "resolution_failed"}
			}
		}()
	}
	wg.Wait()
	result := []ProviderError{}
	for _, e := range issues {
		if e != nil {
			result = append(result, *e)
		}
	}
	return resolved, result
}

func (s *Searcher) resolvePlace(ctx context.Context, locale, city, country string, target *ResolvedPlace) error {
	if target.IATA == "" || target.Latitude == nil || target.Longitude == nil {
		var places []struct {
			Type        string `json:"type"`
			Code        string `json:"code"`
			Name        string `json:"name"`
			Country     string `json:"country_code"`
			Coordinates struct {
				Lat *float64 `json:"lat"`
				Lon *float64 `json:"lon"`
			} `json:"coordinates"`
		}
		if err := s.discovery(ctx, Request{Provider: "travelpayouts_places", Endpoint: "places2", Params: providers.TravelpayoutsPlaces{Term: city, Locale: locale, Types: []string{"city"}}}, &places); err != nil {
			return err
		}
		matches := map[string]int{}
		for i, p := range places {
			if p.Type == "city" && cityKey(p.Name) == cityKey(city) && (country == "" || country == p.Country) {
				matches[p.Country+":"+p.Code] = i
			}
		}
		if len(matches) != 1 {
			return fmt.Errorf("Уточните город и код страны: найдено совпадений — %d", len(matches))
		}
		var index int
		for _, i := range matches {
			index = i
		}
		p := places[index]
		lat, lon := p.Coordinates.Lat, p.Coordinates.Lon
		if !cityIATA.MatchString(p.Code) || lat == nil || lon == nil || math.IsNaN(*lat) || math.IsNaN(*lon) || math.Abs(*lat) > 90 || math.Abs(*lon) > 180 {
			return fmt.Errorf("справочник вернул некорректный город")
		}
		if target.IATA != "" && target.IATA != p.Code {
			return fmt.Errorf("Уточните город: сохранённый IATA не совпадает со справочником")
		}
		target.IATA, target.Latitude, target.Longitude = p.Code, lat, lon
	}
	if target.YandexCode == "" && os.Getenv(s.Registry["yandex_rasp"].Auth.Env) != "" {
		var nearest struct {
			Code     string  `json:"code"`
			Distance float64 `json:"distance"`
		}
		if err := s.discovery(ctx, Request{Provider: "yandex_rasp", Endpoint: "nearest_settlement", Params: providers.RaspNearestSettlement{Lat: *target.Latitude, Lng: *target.Longitude, Distance: 5, Lang: "ru_RU", Format: "json"}}, &nearest); err != nil {
			return err
		}
		if !settlementCode.MatchString(nearest.Code) || nearest.Distance < 0 || nearest.Distance > 5 {
			return fmt.Errorf("не удалось подтвердить код города Яндекса")
		}
		target.YandexCode = nearest.Code
	}
	return ctx.Err()
}
