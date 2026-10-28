package providerkit

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLiveItineraryMapsEveryLegWithoutInventingCodes(t *testing.T) {
	input, places, now := sample(t)
	if err := json.Unmarshal([]byte(`[{"city":"Казань","check_in":"2026-12-12","check_out":"2026-12-13"},{"city":"Самара","check_in":"2026-12-13","check_out":"2026-12-14"}]`), &input.Stops); err != nil {
		t.Fatal(err)
	}
	input.Destination.City = "Самара"
	places.Stops = []ResolvedPlace{{IATA: "KZN", YandexCode: "c43"}, {IATA: "KUF", YandexCode: "c51"}}
	reg, err := LoadRegistry("../configs/providers.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"aviasales", "yandex_rasp"} {
		p := reg[name]
		p.RPS = 1000
		reg[name] = p
	}
	s, _ := NewSearcher(reg, "live")
	s.Now = func() time.Time { return now }
	t.Setenv("TRAVELPAYOUTS_TOKEN", "test")
	t.Setenv("YANDEX_RASP_API_KEY", "test")
	var mu sync.Mutex
	calls := map[string]bool{}
	s.Client = doFunc(func(req *http.Request) (*http.Response, error) {
		q := req.URL.Query()
		body := `{"segments":[]}`
		if strings.Contains(req.URL.Path, "prices_for_dates") {
			mu.Lock()
			calls[q.Get("origin")+"-"+q.Get("destination")+"-"+q.Get("departure_at")] = true
			mu.Unlock()
			body = `{"success":true,"currency":"rub","data":[]}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	result, err := s.SearchParts(context.Background(), input, places)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"MOW-KZN-2026-12-12", "KZN-KUF-2026-12-13", "KUF-MOW-2026-12-14"} {
		if !calls[key] {
			t.Fatalf("missing leg %s, calls=%v errors=%v", key, calls, result.Errors)
		}
	}
	places.Stops = nil
	calls = map[string]bool{}
	result, err = s.SearchParts(context.Background(), input, places)
	if err != nil || len(calls) != 0 || len(result.Errors) == 0 || result.Errors[0].Code != "missing_field" {
		t.Fatalf("unresolved stops used: %+v %v", result.Errors, err)
	}
}
