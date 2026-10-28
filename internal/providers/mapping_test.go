package providers_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ostrovok-hackathon-2026/godnota/internal/providers"
	"github.com/ostrovok-hackathon-2026/godnota/internal/trip"
)

func TestBuildTripPlanKazan(t *testing.T) {
	input := readTrip(t, "../../configs/examples/trip_request.kazan_weekend.json")
	var resolved providers.ResolvedTrip
	readJSONFile(t, "../../configs/examples/resolved_trip.kazan_weekend.json", &resolved)
	plan, err := providers.BuildTripPlan(input, resolved, time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Missing) != 0 || len(plan.Search["aviasales"]) != 1 || len(plan.Search["yandex_rasp"]) != 4 || len(plan.Search["wikipedia"]) != 1 {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	if !strings.Contains(strings.Join(plan.Limitations, " "), "поиск отелей выполняется отдельно") {
		t.Fatalf("hotel omission is not reported: %+v", plan.Limitations)
	}
	for _, requests := range plan.Search {
		for _, request := range requests {
			if err := providers.Validate(request.Params); err != nil {
				t.Fatalf("%s.%s: %v", request.Provider, request.Endpoint, err)
			}
		}
	}
}

func TestBuildTripPlanKeepsVagueDestinationUnresolved(t *testing.T) {
	input := readTrip(t, "../../configs/examples/trip_request.sea_november_vague.json")
	plan, err := providers.BuildTripPlan(input, providers.ResolvedTrip{}, time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Search) != 0 || len(plan.Missing) != 1 || plan.Missing[0] != "/destination/city" {
		t.Fatalf("vague destination should not produce a search: %+v", plan)
	}
}

func TestBuildTripPlanDoesNotInventHotelResults(t *testing.T) {
	input := readTrip(t, "../../configs/examples/trip_request.kazan_weekend.json")
	var resolved providers.ResolvedTrip
	readJSONFile(t, "../../configs/examples/resolved_trip.kazan_weekend.json", &resolved)
	plan, err := providers.BuildTripPlan(input, resolved, time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(plan.Limitations, " "), "партнёрского доступа") {
		t.Fatalf("hotel search should be omitted and reported: %+v", plan)
	}
}

func TestBuildTripPlanRejectsUntrustedCodes(t *testing.T) {
	input := readTrip(t, "../../configs/examples/trip_request.kazan_weekend.json")
	resolved := providers.ResolvedTrip{Origin: providers.ResolvedPlace{IATA: "mow"}}
	if _, err := providers.BuildTripPlan(input, resolved, time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("invalid resolver code accepted")
	}
}

func readTrip(t *testing.T, path string) trip.TripRequest {
	t.Helper()
	var input trip.TripRequest
	readJSONFile(t, path, &input)
	return input
}

func readJSONFile(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatal(err)
	}
}
