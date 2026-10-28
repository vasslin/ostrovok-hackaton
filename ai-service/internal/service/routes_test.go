package service

import (
	"ai-service/internal/canonical"
	"context"
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"
)

func rankingInput(t *testing.T) RankRoutesInput {
	t.Helper()
	data, err := os.ReadFile("../../../examples/ranking/routes.json")
	if err != nil {
		t.Fatal(err)
	}
	var in RankRoutesInput
	if err = json.Unmarshal(data, &in); err != nil {
		t.Fatal(err)
	}
	return in
}
func TestRoutePriceAndProvenance(t *testing.T) {
	in := rankingInput(t)
	svc, calls := planningService(t, routeScores{Results: []routeScore{{ID: "route-1", Score: ptr(90.0)}}}, nil)
	in.Parts.Errors = []*ProviderError{{Provider: "second", Code: "timeout", Retryable: true}}
	out, err := svc.RankRoutes(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if *calls != 1 || len(out.Results) != 1 || out.Results[0].TotalPrice == nil || out.Results[0].TotalPrice.AmountMinor != 6500000 || !out.Results[0].PriceComplete || !out.Results[0].RequiresVerification || len(out.Parts.Errors) != 1 || out.Parts.TransportParts[0].Source != "synthetic" {
		t.Fatalf("incorrect total or provenance: %+v", out)
	}
}
func TestRouteMissingPriceCurrencyChildrenAndScope(t *testing.T) {
	for _, test := range []string{"unknown", "currency", "children", "scope", "zero", "missing return", "hotel gap"} {
		t.Run(test, func(t *testing.T) {
			in := rankingInput(t)
			switch test {
			case "unknown":
				in.Parts.HotelParts[0].Price = nil
			case "currency":
				in.Parts.HotelParts[0].Price.Currency = "USD"
			case "children":
				in.Config.Request.Travelers.ChildrenAges = append(in.Config.Request.Travelers.ChildrenAges, 5)
			case "scope":
				in.Parts.TransportParts[0].Price.Scope = "total"
			case "zero":
				in.Parts.HotelParts[0].Price.AmountMinor = 0
			case "missing return":
				in.Parts.TransportParts[0].ReturnDepartureAt = ""
				in.Parts.TransportParts[0].ReturnArrivalAt = ""
				in.Parts.TransportParts[0].Price.Scope = "per_person_one_way"
			case "hotel gap":
				in.Parts.HotelParts[0].CheckIn = "2026-11-11"
			}
			svc, _ := planningService(t, routeScores{Results: []routeScore{{ID: "route-1", Score: ptr(80.0)}}}, nil)
			out, err := svc.RankRoutes(context.Background(), in)
			if err != nil {
				t.Fatal(err)
			}
			if len(out.Results) != 1 || out.Results[0].TotalPrice != nil || out.Results[0].PriceComplete || !out.Results[0].RequiresVerification {
				t.Fatalf("unconfirmed total: %+v", out)
			}
		})
	}
}
func TestRoutesRejectHardViolationsWithoutModel(t *testing.T) {
	for _, test := range []string{"budget", "chronology", "stars", "cancellation", "overflow", "wrong date", "wrong city"} {
		t.Run(test, func(t *testing.T) {
			in := rankingInput(t)
			switch test {
			case "budget":
				in.Config.Request.Budget.Total = 100
			case "chronology":
				in.Parts.TransportParts[0].ArrivalAt = "2026-11-10T06:00:00+03:00"
			case "stars":
				in.Parts.HotelParts[0].Stars = 2
			case "cancellation":
				in.Config.Request.Hotel.FreeCancellation = true
				in.Parts.HotelParts[0].FreeCancellation = false
			case "overflow":
				in.Parts.TransportParts[0].Price.AmountMinor = Minor(math.MaxInt64)
			case "wrong date":
				in.Parts.TransportParts[0].DepartureAt = "2026-11-09T08:00:00+03:00"
			case "wrong city":
				in.Parts.TransportParts[0].OriginCity = "Казань"
			}
			svc, calls := planningService(t, routeScores{}, nil)
			out, err := svc.RankRoutes(context.Background(), in)
			if err != nil {
				t.Fatal(err)
			}
			if len(out.Results) != 0 || len(out.Rejected) != 1 || *calls != 0 {
				t.Fatalf("bad candidate survived: %+v calls=%d", out, *calls)
			}
		})
	}
}
func TestRankingIDsAndMoney(t *testing.T) {
	for _, value := range []string{`-1`, `"9223372036854775808"`, `1.5`, `1e3`, `null`, `"1.0"`} {
		var m Minor
		if json.Unmarshal([]byte(value), &m) == nil {
			t.Fatalf("invalid minor accepted: %s", value)
		}
	}
	for _, value := range []string{`9223372036854775807`, `"9223372036854775807"`} {
		var m Minor
		if err := json.Unmarshal([]byte(value), &m); err != nil || m != Minor(math.MaxInt64) {
			t.Fatal("int64 precision lost")
		}
	}
	in := rankingInput(t)
	in.Candidates[0].HotelPartIDs = []string{"unknown"}
	svc, calls := planningService(t, routeScores{}, nil)
	if _, err := svc.RankRoutes(context.Background(), in); err == nil || *calls != 0 {
		t.Fatal("unknown reference accepted")
	}
	in = rankingInput(t)
	bad, _ := planningService(t, routeScores{Results: []routeScore{{ID: "invented", Score: ptr(80.0)}}}, nil)
	_, err := bad.RankRoutes(context.Background(), in)
	if err == nil || !strings.Contains(err.Error(), "invalid data") {
		t.Fatal("invented score ID accepted", err)
	}
}

func TestDirectConstraintCountsConnectionsBetweenParts(t *testing.T) {
	in := rankingInput(t)
	request := *in.Config.Request
	request.Transport = &canonical.Transport{DirectOnly: true}
	outbound := *in.Parts.TransportParts[0]
	outbound.ID = "leg-1"
	outbound.Destination = "KZN"
	outbound.DestinationCity = "Казань"
	outbound.ReturnDepartureAt = ""
	outbound.ReturnArrivalAt = ""
	next := outbound
	next.ID = "leg-2"
	next.Origin = "KZN"
	next.OriginCity = "Казань"
	next.Destination = "AER"
	next.DestinationCity = "Сочи"
	next.DepartureAt = "2026-11-10T13:00:00+03:00"
	next.ArrivalAt = "2026-11-10T15:00:00+03:00"
	candidate := RouteCandidate{ID: "multi", TransportPartIDs: []string{"leg-1", "leg-2"}}
	_, reason := evaluateRoute(request, candidate, map[string]*TransportPart{"leg-1": &outbound, "leg-2": &next}, nil)
	if reason == "" {
		t.Fatal("two non-stop parts were treated as a direct flight")
	}
}
