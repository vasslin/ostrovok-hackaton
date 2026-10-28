package routebuilder

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	pb "github.com/godnota/pkg/generated/provider/v1"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestMultiCityDemoItinerary(t *testing.T) {
	var trip providerkit.TripRequest
	err := json.Unmarshal([]byte(`{"schema_version":"1.0","locale":"ru","currency":"RUB","origin":{"city":"Москва"},"destination":{"city":"Самара"},"dates":{"departure_date":"2026-12-12","return_date":"2026-12-14"},"stops":[{"city":"Казань","check_in":"2026-12-12","check_out":"2026-12-13"},{"city":"Самара","check_in":"2026-12-13","check_out":"2026-12-14"}],"travelers":{"adults":2},"transport":{"modes":["plane","train"],"round_trip":true},"hotel":{"needed":true,"rooms":1}}`), &trip)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := providerkit.NewSearcher(nil, "demo")
	s.Now = func() time.Time { return time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC) }
	demo, err := s.SearchParts(context.Background(), trip, providerkit.ResolvedTrip{})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(demo)
	parts := &pb.SearchPartsResponse{}
	if err = protojson.Unmarshal(b, parts); err != nil {
		t.Fatal(err)
	}
	got, err := Build(context.Background(), trip, parts)
	if err != nil || len(got) == 0 {
		t.Fatalf("missing itinerary %v %v", got, err)
	}
	route := got[0].AsMap()
	if len(route["transport"].([]any)) != 3 || len(route["hotels"].([]any)) != 2 || route["bookable"] != false {
		t.Fatalf("wrong itinerary %+v", route)
	}
}
func TestTransferRules(t *testing.T) {
	trip, parts := fixture(t)
	a := &pb.TransportPart{Id: "first", Mode: "plane", OriginCity: "Москва", DestinationCity: "Самара", Origin: "MOW", Destination: "KUF", Source: "synthetic", DepartureAt: "2026-12-12T06:00:00+03:00", ArrivalAt: "2026-12-12T08:00:00+03:00"}
	b := &pb.TransportPart{Id: "second", Mode: "plane", OriginCity: "Самара", DestinationCity: "Казань", Origin: "KUF", Destination: "KZN", Source: "synthetic", DepartureAt: "2026-12-12T10:00:00+03:00", ArrivalAt: "2026-12-12T12:00:00+03:00"}
	parts.TransportParts = []*pb.TransportPart{a, b, parts.TransportParts[1]}
	got, err := Build(context.Background(), trip, parts)
	if err != nil || len(got) != 1 || len(got[0].AsMap()["transport"].([]any)) != 3 || got[0].AsMap()["price_status"] != "unknown" {
		t.Fatalf("valid connection lost: %v %v", got, err)
	}
	b.DepartureAt = "2026-12-12T09:30:00+03:00"
	got, err = Build(context.Background(), trip, parts)
	if err != nil || len(got) != 1 {
		t.Fatalf("minimum wait boundary rejected: %v %v", got, err)
	}
	b.DepartureAt = "2026-12-12T08:30:00+03:00"
	got, err = Build(context.Background(), trip, parts)
	if err != nil || len(got) != 0 {
		t.Fatalf("short layover accepted: %v %v", got, err)
	}
	b.DepartureAt = "2026-12-12T21:00:00+03:00"
	b.ArrivalAt = "2026-12-12T23:00:00+03:00"
	got, err = Build(context.Background(), trip, parts)
	if err != nil || len(got) != 0 {
		t.Fatalf("excessive layover accepted: %v %v", got, err)
	}
	b.DepartureAt = "2026-12-12T10:00:00+03:00"
	b.ArrivalAt = "2026-12-12T12:00:00+03:00"
	b.Origin = "OTHER"
	got, err = Build(context.Background(), trip, parts)
	if err != nil || len(got) != 0 {
		t.Fatalf("station transfer invented: %v %v", got, err)
	}
	b.Origin = "KUF"
	a.Source = "cached"
	b.Source = "cached"
	got, err = Build(context.Background(), trip, parts)
	if err != nil || len(got) != 0 {
		t.Fatalf("ambiguous airport codes accepted: %v %v", got, err)
	}
	a.Source = "synthetic"
	b.Source = "synthetic"
	trip.Transport.DirectOnly = true
	got, err = Build(context.Background(), trip, parts)
	if err != nil || len(got) != 0 {
		t.Fatalf("direct-only ignored: %v %v", got, err)
	}
}

func TestProviderTransferEnvelopeSafety(t *testing.T) {
	trip, parts := fixture(t)
	outbound := parts.TransportParts[0]
	outbound.Provider = "yandex_rasp"
	outbound.Source = "live"
	outbound.Transfers = 1
	outbound.DepartureAt = "2026-12-12T08:00:00+03:00"
	outbound.ArrivalAt = "2026-12-12T14:00:00+03:00"
	outbound.Legs = []*pb.TransferLeg{{OriginStation: "s10", DestinationStation: "s20", DepartureAt: outbound.DepartureAt, ArrivalAt: "2026-12-12T10:00:00+03:00", Mode: "plane", ServiceId: "a"}, {OriginStation: "s20", DestinationStation: "s30", DepartureAt: "2026-12-12T12:00:00+03:00", ArrivalAt: outbound.ArrivalAt, Mode: "plane", ServiceId: "b"}}
	check := func(expected int) {
		t.Helper()
		got, err := Build(context.Background(), trip, parts)
		if err != nil || len(got) != expected {
			t.Fatalf("routes=%d, expected %d, error %v", len(got), expected, err)
		}
	}
	check(1)
	outbound.Legs[1].OriginStation = "s21"
	check(0)
	outbound.Legs[1].OriginStation = "s20"
	outbound.Legs[1].DepartureAt = "2026-12-12T11:29:00+03:00"
	check(0)
	outbound.Legs[1].DepartureAt = "2026-12-12T11:30:00+03:00"
	check(1)
	outbound.Legs[1].Mode = "train"
	check(0)
	outbound.Legs[1].Mode = "plane"
	outbound.Legs = nil
	check(0)
}
