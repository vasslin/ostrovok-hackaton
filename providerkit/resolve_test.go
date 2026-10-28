package providerkit

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ostrovok-hackathon-2026/godnota/internal/trip"
)

func resolverFixture(t *testing.T) *Searcher {
	t.Helper()
	reg, err := LoadRegistry("../configs/providers.json")
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range reg {
		v.RPS = 10000
		reg[k] = v
	}
	s, _ := NewSearcher(reg, "live")
	t.Setenv("TRAVELPAYOUTS_TOKEN", "test")
	t.Setenv("YANDEX_RASP_API_KEY", "test")
	return s
}
func TestResolveTripCachePartialAndInvalidation(t *testing.T) {
	s := resolverFixture(t)
	var places, nearest atomic.Int32
	fail := false
	s.Client = doFunc(func(r *http.Request) (*http.Response, error) {
		body := `{"code":"c213","distance":0.5}`
		if r.URL.Path == "/places2" {
			places.Add(1)
			code, name := "MOW", "Москва"
			if r.URL.Query().Get("term") == "Париж" {
				code, name = "PAR", "Париж"
			}
			body = `[{"type":"city","code":"` + code + `","name":"` + name + `","country_code":"RU","coordinates":{"lat":55,"lon":37}}]`
		} else {
			nearest.Add(1)
			if fail {
				return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader(""))}, nil
			}
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	input := TripRequest{Locale: "ru", Origin: trip.Place{City: "Москва"}}
	resolved, issues := s.ResolveTrip(context.Background(), input, ResolvedTrip{})
	if len(issues) != 0 || resolved.Origin.IATA != "MOW" || resolved.Origin.YandexCode != "c213" {
		t.Fatalf("%+v %+v", resolved, issues)
	}
	_, issues = s.ResolveTrip(context.Background(), input, ResolvedTrip{})
	if len(issues) != 0 || places.Load() != 1 || nearest.Load() != 1 {
		t.Fatal("cache not used")
	}
	input.Origin.City = "Париж"
	resolved, issues = s.ResolveTrip(context.Background(), input, resolved)
	if len(issues) != 0 || resolved.Origin.IATA != "PAR" {
		t.Fatalf("stale city: %+v %+v", resolved, issues)
	}
	s.Now = func() time.Time { return time.Now().Add(8 * 24 * time.Hour) }
	fail = true
	resolved, issues = s.ResolveTrip(context.Background(), input, ResolvedTrip{})
	if len(issues) != 1 || resolved.Origin.IATA != "PAR" || resolved.Origin.YandexCode != "" {
		t.Fatal("partial result lost")
	}
}
func TestResolveAmbiguousAndDemo(t *testing.T) {
	s := resolverFixture(t)
	var calls atomic.Int32
	s.Client = doFunc(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`[{"type":"city","code":"PAR","name":"Париж","country_code":"FR","coordinates":{"lat":48,"lon":2}},{"type":"city","code":"PRX","name":"Париж","country_code":"US","coordinates":{"lat":33,"lon":-95}}]`))}, nil
	})
	input := TripRequest{Locale: "ru", Origin: trip.Place{City: "Париж"}}
	got, issues := s.ResolveTrip(context.Background(), input, ResolvedTrip{})
	if len(issues) != 1 || issues[0].Code != "clarification_required" || got.Origin.IATA != "" {
		t.Fatal("ambiguous city chosen")
	}
	input.Origin.CountryCode = "FR"
	t.Setenv("YANDEX_RASP_API_KEY", "")
	got, issues = s.ResolveTrip(context.Background(), input, ResolvedTrip{})
	if len(issues) != 0 || got.Origin.IATA != "PAR" {
		t.Fatalf("country ignored: %+v", issues)
	}
	count := calls.Load()
	s.Mode = "demo"
	_, _ = s.ResolveTrip(context.Background(), input, ResolvedTrip{})
	if count != calls.Load() {
		t.Fatal("demo made external requests")
	}
	s.Mode = "live"
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, issues = s.ResolveTrip(ctx, TripRequest{Locale: "en", Origin: trip.Place{City: "Unknown"}}, ResolvedTrip{})
	if len(issues) != 1 || count != calls.Load() {
		t.Fatal("cancellation ignored")
	}
}
