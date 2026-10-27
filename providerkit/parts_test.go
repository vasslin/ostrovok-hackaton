package providerkit

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type hotelFunc func(context.Context, HotelQuery) ([]HotelPart, error)

func (f hotelFunc) SearchHotels(ctx context.Context, q HotelQuery) ([]HotelPart, error) {
	return f(ctx, q)
}

func TestTransportAndHotelRunConcurrently(t *testing.T) {
	input, places, now := sample(t)
	input.Transport.Modes = []string{"plane"}
	reg, _ := LoadRegistry("../configs/providers.json")
	s, _ := NewSearcher(reg, "live")
	s.Now = func() time.Time { return now }
	s.Timeout = time.Second
	t.Setenv("TRAVELPAYOUTS_TOKEN", "test")
	t.Setenv("YANDEX_RASP_API_KEY", "test")
	transportStarted := make(chan struct{})
	hotelStarted := make(chan struct{})
	var once sync.Once
	s.Client = doFunc(func(r *http.Request) (*http.Response, error) {
		once.Do(func() { close(transportStarted) })
		select {
		case <-hotelStarted:
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"success":true,"currency":"rub","data":[]}`))}, nil
	})
	s.Hotels = hotelFunc(func(ctx context.Context, q HotelQuery) ([]HotelPart, error) {
		close(hotelStarted)
		select {
		case <-transportStarted:
			return []HotelPart{{ID: "test-hotel", City: q.City, Source: "synthetic"}}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})
	parts, e := s.SearchParts(context.Background(), input, places)
	if e != nil || len(parts.HotelParts) != 1 {
		t.Fatalf("provider calls were not concurrent: %+v %v", parts, e)
	}
}

func TestDepartureAndDurationFilters(t *testing.T) {
	input, places, now := sample(t)
	s, _ := NewSearcher(nil, "demo")
	s.Now = func() time.Time { return now }
	input.Transport.DepartureTimeFrom = "07:00"
	input.Transport.DepartureTimeTo = "09:00"
	input.Transport.MaxDurationHours = 3
	result, e := s.SearchParts(context.Background(), input, places)
	if e != nil {
		t.Fatal(e)
	}
	if len(result.TransportParts) != 1 || result.TransportParts[0].Mode != "plane" {
		t.Fatalf("constraints ignored: %+v", result.TransportParts)
	}
	input.Currency = "JPY"
	if _, e = s.SearchParts(context.Background(), input, places); e == nil {
		t.Fatal("unsupported minor unit silently converted")
	}
}

type doFunc func(*http.Request) (*http.Response, error)

func (f doFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }
func sample(t *testing.T) (TripRequest, ResolvedTrip, time.Time) {
	t.Helper()
	raw, e := os.ReadFile("../configs/examples/trip_request.kazan_weekend.json")
	if e != nil {
		t.Fatal(e)
	}
	input, e := DecodeTrip(raw)
	if e != nil {
		t.Fatal(e)
	}
	raw, e = os.ReadFile("../configs/examples/resolved_trip.kazan_weekend.json")
	if e != nil {
		t.Fatal(e)
	}
	var places ResolvedTrip
	if e = json.Unmarshal(raw, &places); e != nil {
		t.Fatal(e)
	}
	now, _ := time.Parse(time.DateOnly, "2026-10-28")
	return input, places, now
}
func TestDemoSearch(t *testing.T) {
	input, places, now := sample(t)
	s, _ := NewSearcher(nil, "demo")
	s.Now = func() time.Time { return now }
	s.Client = doFunc(func(*http.Request) (*http.Response, error) { t.Fatal("demo must not call network"); return nil, nil })
	parts, e := s.SearchParts(context.Background(), input, places)
	if e != nil {
		t.Fatal(e)
	}
	if len(parts.TransportParts) != 4 || len(parts.HotelParts) != 1 {
		t.Fatalf("unexpected demo parts: %+v", parts)
	}
	for _, p := range parts.TransportParts {
		if p.Source != "synthetic" || p.Price.AmountMinor <= 0 {
			t.Fatalf("unmarked demo: %+v", p)
		}
	}
	if parts.HotelParts[0].Price.AmountMinor != 1000000 {
		t.Fatal("stay total must include both nights")
	}
}
func TestValidationBeforeProvider(t *testing.T) {
	input, places, now := sample(t)
	s, _ := NewSearcher(nil, "demo")
	s.Now = func() time.Time { return now }
	input.Dates.ReturnDate = "2026-12-01"
	if _, e := s.SearchParts(context.Background(), input, places); e == nil {
		t.Fatal("invalid dates accepted")
	}
	input, places, now = sample(t)
	input.Destination = nil
	if _, e := s.SearchParts(context.Background(), input, places); e == nil {
		t.Fatal("missing destination accepted")
	}
	if _, e := DecodeTrip([]byte(`{"unknown":true}`)); e == nil {
		t.Fatal("unknown AI fields accepted")
	}
}
func TestHotelMappingPreservesConstraints(t *testing.T) {
	input, places, now := sample(t)
	mapped, e := MapConfig(input, places, now)
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, cfg := range mapped.Configs {
		if cfg.Provider == "hotel" {
			q := cfg.Params.(HotelQuery)
			found = true
			if q.Adults != 2 || q.Rooms != 1 || q.CheckOut != "2026-12-14" || q.Currency != "RUB" || q.MaxPricePerNight != 8000 || q.Meal != "breakfast" || !q.FreeCancellation {
				t.Fatalf("lost hotel constraints: %+v", q)
			}
		}
	}
	if !found {
		t.Fatal("no hotel mapping")
	}
}
func TestLivePartialFailureAndSecretRedaction(t *testing.T) {
	input, places, now := sample(t)
	reg, e := LoadRegistry("../configs/providers.json")
	if e != nil {
		t.Fatal(e)
	}
	s, _ := NewSearcher(reg, "live")
	s.Now = func() time.Time { return now }
	t.Setenv("TRAVELPAYOUTS_TOKEN", "private-flight-key")
	t.Setenv("YANDEX_RASP_API_KEY", "private-rasp-key")
	s.Client = doFunc(func(r *http.Request) (*http.Response, error) {
		if strings.Contains(r.URL.Host, "travelpayouts") {
			data, _ := os.ReadFile("../testdata/aviasales/prices_for_dates.json")
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(data)))}, nil
		}
		return nil, errors.New("request contains private-rasp-key")
	})
	parts, e := s.SearchParts(context.Background(), input, places)
	if e != nil {
		t.Fatal(e)
	}
	if len(parts.TransportParts) == 0 || len(parts.Errors) == 0 {
		t.Fatal("partial results lost")
	}
	raw, _ := json.Marshal(parts)
	if strings.Contains(string(raw), "private-") {
		t.Fatal("credential leaked")
	}
	if parts.TransportParts[0].Source != "cached" || parts.TransportParts[0].Price.AmountMinor != 966500 {
		t.Fatal("cached price normalized incorrectly")
	}
}
func TestTimeoutAndCancellation(t *testing.T) {
	input, places, now := sample(t)
	reg, _ := LoadRegistry("../configs/providers.json")
	s, _ := NewSearcher(reg, "live")
	s.Now = func() time.Time { return now }
	s.Timeout = 15 * time.Millisecond
	t.Setenv("TRAVELPAYOUTS_TOKEN", "test")
	t.Setenv("YANDEX_RASP_API_KEY", "test")
	s.Client = doFunc(func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() })
	parts, e := s.SearchParts(context.Background(), input, places)
	if e != nil || len(parts.Errors) == 0 {
		t.Fatalf("timeout was not represented: %v", e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e = s.SearchParts(ctx, input, places); !errors.Is(e, context.Canceled) {
		t.Fatal("cancellation lost")
	}
}
func TestYandexUnknownPriceAndEmptyResults(t *testing.T) {
	_, _, now := sample(t)
	input, places, _ := sample(t)
	mapped, _ := MapConfig(input, places, now)
	var job Request
	for _, r := range mapped.Configs {
		if r.Provider == "yandex_rasp" {
			job = r
			break
		}
	}
	raw, _ := os.ReadFile("../testdata/yandex_rasp/search.json")
	parts, e := normalize(job, raw, now)
	if e != nil || len(parts) == 0 {
		t.Fatalf("normalize schedules: %v", e)
	}
	if parts[0].Price != nil {
		t.Fatal("unknown schedule price must remain absent")
	}
	parts, e = normalize(job, []byte(`{"segments":[]}`), now)
	if e != nil || len(parts) != 0 {
		t.Fatal("valid empty result rejected")
	}
	if _, e = normalize(job, []byte(`{}`), now); e == nil {
		t.Fatal("invalid envelope accepted")
	}
}
