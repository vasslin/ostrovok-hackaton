package provider

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/godnota/provider-service/internal/entity"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func fixtureRepo(t *testing.T) (*Repository, providerkit.TripRequest, providerkit.ResolvedTrip) {
	t.Helper()
	root := filepath.Join("..", "..", "..", "..")
	reg, err := providerkit.LoadRegistry(filepath.Join(root, "configs", "providers.json"))
	if err != nil {
		t.Fatal(err)
	}
	repo, err := New(reg, filepath.Join(root, "configs", "poi", "kazan.json"))
	if err != nil {
		t.Fatal(err)
	}
	var trip providerkit.TripRequest
	data, err := os.ReadFile(filepath.Join(root, "configs", "examples", "trip_request.kazan_weekend.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &trip); err != nil {
		t.Fatal(err)
	}
	day := time.Now().UTC().AddDate(0, 0, 10)
	trip.Dates.DepartureDate = day.Format(time.DateOnly)
	trip.Dates.ReturnDate = day.AddDate(0, 0, 2).Format(time.DateOnly)
	lat, lon := 55.7987, 49.1059
	resolved := providerkit.ResolvedTrip{
		Origin:       providerkit.ResolvedPlace{IATA: "MOW", YandexCode: "c213"},
		Destination:  providerkit.ResolvedPlace{IATA: "KZN", YandexCode: "c43", Latitude: &lat, Longitude: &lon},
		HotelCheckin: trip.Dates.DepartureDate, HotelCheckout: trip.Dates.ReturnDate, HotelbedsSourceMarket: "RU",
	}
	return repo, trip, resolved
}

func TestSchedulePricesAndUnknownValues(t *testing.T) {
	data := []byte(`{"segments":[{"thread":{"uid":"train","title":"Train","transport_type":"train"},"from":{"title":"A"},"to":{"title":"B"},"departure":"2026-12-12T10:00:00+03:00","arrival":"2026-12-12T12:00:00+03:00","tickets_info":{"places":[{"name":"seat","currency":"RUB","price":{"whole":123,"cents":45}},{"name":"other","currency":"JPY","price":{"whole":50,"cents":0}}]}},{"thread":{"uid":"ferry","title":"Ferry","transport_type":"water"},"departure":"2026-12-12T10:00:00+03:00","arrival":"2026-12-12T12:00:00+03:00","tickets_info":{"places":[]}}]}`)
	offers, err := scheduleOffers(data, time.Now(), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(offers) != 3 || offers[0].PriceMinor == nil || *offers[0].PriceMinor != 12345 || offers[0].Currency != "RUB" {
		t.Fatalf("unexpected prices: %+v", offers)
	}
	if offers[1].PriceMinor != nil || offers[2].PriceMinor != nil || offers[2].PriceScope != "unknown" {
		t.Fatalf("unknown price became a quote: %+v", offers)
	}
}

func TestWikipediaBusyIsClassified(t *testing.T) {
	_, err := wikipediaOffers([]byte(`{"error":{"code":"cirrussearch-too-busy-error"}}`), time.Now(), false)
	var apiErr providerAPIError
	if !errors.As(err, &apiErr) || apiErr.code != "upstream_busy" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHotelbedsEvaluationRequestAndResponse(t *testing.T) {
	t.Setenv("HOTELBEDS_MODE", "evaluation")
	t.Setenv("HOTELBEDS_API_KEY", "test-key")
	t.Setenv("HOTELBEDS_SECRET", "test-secret")
	t.Setenv("RATEHAWK_MODE", "")
	repo, trip, resolved := fixtureRepo(t)
	if err := json.Unmarshal([]byte(`{"adults":2,"residency":"ru","children_ages":[7]}`), &trip.Travelers); err != nil {
		t.Fatal(err)
	}
	called := false
	repo.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		called = true
		if req.Method != "POST" || req.URL.String() != "https://api.test.hotelbeds.com/hotel-api/1.0/hotels" || req.Header.Get("Api-key") != "test-key" {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL)
		}
		sig := req.Header.Get("X-Signature")
		if sig != hotelbedsSignature("test-key", "test-secret", time.Now()) && sig != hotelbedsSignature("test-key", "test-secret", time.Now().Add(-time.Second)) {
			t.Fatal("invalid signature")
		}
		var body hotelbedsAvailability
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Stay.CheckIn != resolved.HotelCheckin || body.SourceMarket != "RU" || len(body.Occupancies) != 1 || body.Occupancies[0].Rooms != 1 {
			t.Fatalf("unexpected body: %+v", body)
		}
		if len(body.Occupancies[0].Paxes) != 1 || body.Occupancies[0].Paxes[0].Type != "CH" || body.Occupancies[0].Paxes[0].Age != 7 {
			t.Fatalf("unexpected child occupancy: %+v", body.Occupancies[0])
		}
		payload := `{"hotels":{"hotels":[{"code":123,"name":"Example hotel","currency":"EUR","rooms":[{"name":"Double","rates":[{"rateKey":"example-rate","rateType":"RECHECK","net":"123.45"}]}]}]}}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	})
	result, err := repo.Search(context.Background(), trip, resolved, "hotel")
	if err != nil {
		t.Fatal(err)
	}
	if !called || len(result.Offers) != 1 || result.Offers[0].PriceMinor == nil || *result.Offers[0].PriceMinor != 12345 || result.Offers[0].SourceStatus != "sandbox_evaluation" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if strings.Contains(result.Offers[0].Id, "example-rate") {
		t.Fatal("rate key leaked in offer ID")
	}
}

func TestHotelsWithoutCredentialsDoNotInventOffers(t *testing.T) {
	t.Setenv("HOTELBEDS_MODE", "")
	t.Setenv("RATEHAWK_MODE", "")
	repo, trip, resolved := fixtureRepo(t)
	result, err := repo.Search(context.Background(), trip, resolved, "hotel")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Offers) != 0 || len(result.Issues) == 0 || result.Issues[0].Code != "credentials_missing" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestHotelSearchDoesNotRequireTransportCodes(t *testing.T) {
	t.Setenv("HOTELBEDS_MODE", "")
	t.Setenv("RATEHAWK_MODE", "")
	repo, trip, resolved := fixtureRepo(t)
	resolved.Origin.IATA, resolved.Origin.YandexCode = "", ""
	resolved.Destination.IATA, resolved.Destination.YandexCode = "", ""
	result, err := repo.Search(context.Background(), trip, resolved, "hotel")
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range result.Plan.Missing {
		if strings.Contains(field, "iata") || strings.Contains(field, "yandex_code") {
			t.Fatalf("hotel search requires transport code: %s", field)
		}
	}
}

func TestOSMIndexIncludesVenueTypesAndAttribution(t *testing.T) {
	repo, trip, resolved := fixtureRepo(t)
	trip.Leisure.Interests = []string{"museums", "cinemas", "theatres"}
	result := entity.Result{}
	repo.searchPOI(trip, resolved, &result)
	kinds := map[string]bool{}
	for _, offer := range result.Offers {
		kinds[offer.Kind] = true
		if offer.Attribution == "" || offer.PriceMinor != nil || offer.SourceStatus != "offline_snapshot" {
			t.Fatalf("invalid OSM offer: %+v", offer)
		}
	}
	if !kinds["museum"] || !kinds["cinema"] || !kinds["theatre"] {
		t.Fatalf("missing venue types: %v", kinds)
	}
}

func TestUpstreamErrorsDoNotLeakCredentialsOrFollowRedirects(t *testing.T) {
	t.Setenv("TRAVELPAYOUTS_TOKEN", "")
	t.Setenv("YANDEX_RASP_API_KEY", "private-test-token")
	repo, trip, resolved := fixtureRepo(t)
	entry := repo.registry["yandex_rasp"]
	entry.RPS = 1000
	repo.registry["yandex_rasp"] = entry
	requests := 0
	repo.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		header := make(http.Header)
		header.Set("Location", "https://untrusted.example/collect")
		return &http.Response{StatusCode: http.StatusFound, Body: io.NopCloser(strings.NewReader("")), Header: header, Request: req}, nil
	})
	result, err := repo.Search(context.Background(), trip, resolved, "transport")
	if err != nil {
		t.Fatal(err)
	}
	if requests != 4 {
		t.Fatalf("expected four original requests, got %d", requests)
	}
	for _, issue := range result.Issues {
		if strings.Contains(issue.Message, "private-test-token") || strings.Contains(issue.Message, "untrusted.example") {
			t.Fatalf("leaked provider data: %+v", issue)
		}
	}
	if statusCode(http.StatusUnauthorized) != "access_denied" || statusCode(http.StatusTooManyRequests) != "rate_limited" {
		t.Fatal("wrong upstream status classification")
	}
}

func TestUpstreamHTTPStatusAndTimeout(t *testing.T) {
	for _, tc := range []struct {
		status int
		code   string
	}{{401, "access_denied"}, {429, "rate_limited"}} {
		t.Run(tc.code, func(t *testing.T) {
			t.Setenv("TRAVELPAYOUTS_TOKEN", "")
			t.Setenv("YANDEX_RASP_API_KEY", "private-test-token")
			repo, trip, resolved := fixtureRepo(t)
			trip.Transport.Modes = []string{"train"}
			entry := repo.registry["yandex_rasp"]
			entry.RPS = 1000
			repo.registry["yandex_rasp"] = entry
			repo.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header), Request: req}, nil
			})
			result, err := repo.Search(context.Background(), trip, resolved, "transport")
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Issues) == 0 || result.Issues[0].Code != tc.code || strings.Contains(result.Issues[0].Message, "private-test-token") {
				t.Fatalf("unexpected issue: %+v", result.Issues)
			}
		})
	}
	t.Setenv("TRAVELPAYOUTS_TOKEN", "")
	t.Setenv("YANDEX_RASP_API_KEY", "private-test-token")
	repo, trip, resolved := fixtureRepo(t)
	trip.Transport.Modes = []string{"train"}
	repo.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := repo.Search(ctx, trip, resolved, "transport"); err == nil {
		t.Fatal("expected cancellation")
	}
}

func TestProviderTimeoutIsReportedWithoutRequestDetails(t *testing.T) {
	t.Setenv("YANDEX_RASP_API_KEY", "private-test-token")
	repo, trip, resolved := fixtureRepo(t)
	trip.Transport.Modes = []string{"train"}
	entry := repo.registry["yandex_rasp"]
	entry.TimeoutMS = 10
	entry.RPS = 1000
	repo.registry["yandex_rasp"] = entry
	repo.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	})
	result, err := repo.Search(context.Background(), trip, resolved, "transport")
	if err != nil || len(result.Issues) == 0 || result.Issues[0].Code != "timeout" || strings.Contains(result.Issues[0].Message, "private-test-token") {
		t.Fatalf("unexpected timeout result: %+v, %v", result.Issues, err)
	}
}

func TestRateHawkSandboxDoesNotExposeRateHash(t *testing.T) {
	t.Setenv("HOTELBEDS_MODE", "")
	t.Setenv("RATEHAWK_MODE", "sandbox")
	t.Setenv("RATEHAWK_KEY_ID", "test-id")
	t.Setenv("RATEHAWK_API_KEY", "test-key")
	repo, trip, resolved := fixtureRepo(t)
	repo.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		user, password, ok := req.BasicAuth()
		if !ok || user != "test-id" || password != "test-key" || req.URL.Hostname() != "api-sandbox.ratehawk.com" {
			t.Fatal("invalid sandbox request")
		}
		body := `{"status":"ok","data":{"hotels":[{"hid":123,"id":"example","rates":[{"search_hash":"private-rate-hash","payment_options":{"payment_types":[{"show_amount":"123.45","show_currency_code":"RUB"}]}}]}]}}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: req}, nil
	})
	result, err := repo.Search(context.Background(), trip, resolved, "hotel")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Offers) != 1 || result.Offers[0].SourceStatus != "sandbox_synthetic" || strings.Contains(result.Offers[0].Id, "private-rate-hash") {
		t.Fatalf("unexpected ratehawk result: %+v", result)
	}
}
