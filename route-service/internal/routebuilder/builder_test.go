package routebuilder

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	pb "github.com/godnota/pkg/generated/provider/v1"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
)

func fixture(t *testing.T) (providerkit.TripRequest, *pb.SearchPartsResponse) {
	t.Helper()
	var trip providerkit.TripRequest
	err := json.Unmarshal([]byte(`{"schema_version":"1.0","locale":"ru","currency":"RUB","origin":{"city":"Москва"},"destination":{"city":"Казань"},"dates":{"departure_date":"2026-12-12","return_date":"2026-12-14"},"travelers":{"adults":2},"transport":{"modes":["plane"],"round_trip":true},"hotel":{"needed":true,"rooms":1},"budget":{"total":20000,"scope":"total","is_hard":true}}`), &trip)
	if err != nil {
		t.Fatal(err)
	}
	fare := func() *pb.Money { return &pb.Money{AmountMinor: 250000, Currency: "RUB", Scope: "per_person_one_way"} }
	out := &pb.TransportPart{Id: "out", Mode: "plane", OriginCity: "Москва", DestinationCity: "Казань", DepartureAt: "2026-12-12T08:00:00+03:00", ArrivalAt: "2026-12-12T10:00:00+03:00", Price: fare(), Source: "synthetic"}
	back := &pb.TransportPart{Id: "back", Mode: "plane", OriginCity: "Казань", DestinationCity: "Москва", DepartureAt: "2026-12-14T18:00:00+03:00", ArrivalAt: "2026-12-14T20:00:00+03:00", Price: fare(), Source: "synthetic"}
	hotel := &pb.HotelPart{Id: "hotel", City: "Казань", Name: "Demo", CheckIn: "2026-12-12", CheckOut: "2026-12-14", Price: &pb.Money{AmountMinor: 1000000, Currency: "RUB", Scope: "stay_all_rooms"}, Source: "synthetic"}
	return trip, &pb.SearchPartsResponse{TransportParts: []*pb.TransportPart{out, back}, HotelParts: []*pb.HotelPart{hotel}}
}
func TestAssembleRoundTripAndPrice(t *testing.T) {
	trip, parts := fixture(t)
	got, err := Build(context.Background(), trip, parts)
	if err != nil || len(got) != 1 {
		t.Fatalf("result %v %v", got, err)
	}
	route := got[0].AsMap()
	total := route["total_price"].(map[string]any)
	if total["amount_minor"] != "2000000" || route["bookable"] != false || len(route["transport"].([]any)) != 2 {
		t.Fatalf("incorrect route %+v", route)
	}
}

func TestRoundTripFareIsNotChargedTwice(t *testing.T) {
	trip, parts := fixture(t)
	parts.TransportParts[0].ReturnDepartureAt = parts.TransportParts[1].DepartureAt
	parts.TransportParts[0].ReturnArrivalAt = parts.TransportParts[1].ArrivalAt
	parts.TransportParts[0].Price.Scope = "per_person_round_trip"
	parts.TransportParts = parts.TransportParts[:1]
	got, err := Build(context.Background(), trip, parts)
	if err != nil || len(got) != 1 || got[0].AsMap()["total_price"].(map[string]any)["amount_minor"] != "1500000" {
		t.Fatalf("round trip charged incorrectly: %v, %v", got, err)
	}
}
func TestUnknownAndConstraints(t *testing.T) {
	t.Run("child price unknown", func(t *testing.T) {
		trip, parts := fixture(t)
		_ = json.Unmarshal([]byte(`{"adults":2,"children_ages":[7]}`), &trip.Travelers)
		got, err := Build(context.Background(), trip, parts)
		if err != nil || len(got) != 1 || got[0].AsMap()["price_status"] != "unknown" {
			t.Fatalf("%v %v", got, err)
		}
	})
	t.Run("unknown currency", func(t *testing.T) {
		trip, parts := fixture(t)
		parts.HotelParts[0].Price.Currency = "EUR"
		got, err := Build(context.Background(), trip, parts)
		if err != nil || len(got) != 1 || got[0].AsMap()["price_status"] != "unknown" {
			t.Fatalf("%v %v", got, err)
		}
	})
	t.Run("hard budget", func(t *testing.T) {
		trip, parts := fixture(t)
		trip.Budget.Total = 19000
		got, err := Build(context.Background(), trip, parts)
		if err != nil || len(got) != 0 {
			t.Fatalf("%v %v", got, err)
		}
	})
	t.Run("invalid hotel dates", func(t *testing.T) {
		trip, parts := fixture(t)
		parts.HotelParts[0].CheckOut = "2026-12-15"
		got, err := Build(context.Background(), trip, parts)
		if err != nil || len(got) != 0 {
			t.Fatalf("%v %v", got, err)
		}
	})
	t.Run("invalid return", func(t *testing.T) {
		trip, parts := fixture(t)
		parts.TransportParts[1].ArrivalAt = "2026-12-14T17:00:00+03:00"
		got, err := Build(context.Background(), trip, parts)
		if err != nil || len(got) != 0 {
			t.Fatalf("%v %v", got, err)
		}
	})
	t.Run("sandbox warnings", func(t *testing.T) {
		trip, parts := fixture(t)
		parts.HotelParts[0].Source = "sandbox_evaluation"
		parts.HotelParts[0].Warnings = []string{"Sandbox data"}
		parts.HotelParts[0].Price.Scope = "evaluation net rate"
		got, err := Build(context.Background(), trip, parts)
		if err != nil || len(got) != 1 || got[0].AsMap()["price_status"] != "unknown" || !strings.Contains(got[0].String(), "Sandbox data") {
			t.Fatalf("%v %v", got, err)
		}
	})
}
