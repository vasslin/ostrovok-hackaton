package join_test

import (
	"strings"
	"testing"

	providerv1 "github.com/godnota/pkg/generated/provider/v1"
	"github.com/godnota/route-service/internal/join"
)

func TestMaterializeKeepsMRoutesPerJoin(t *testing.T) {
	parts := &providerv1.SearchPartsResponse{
		TransportParts: []*providerv1.TransportPart{
			priced(road("out", "Москва", "Казань", "2026-12-12T08:00:00Z", "2026-12-12T10:00:00Z"), 500),
			priced(road("back-noon", "Казань", "Москва", "2026-12-12T12:00:00Z", "2026-12-12T14:00:00Z"), 100),
			priced(road("back-night", "Казань", "Москва", "2026-12-12T21:00:00Z", "2026-12-12T23:00:00Z"), 100),
			priced(road("train", "Москва", "Казань", "2026-12-12T06:00:00Z", "2026-12-12T20:00:00Z"), 900),
		},
		HotelParts: []*providerv1.HotelPart{
			pricedHotel(stay("cheap", "Казань", "2026-12-12", "2026-12-12"), 1000),
		},
	}
	joins, err := join.Assemble(parts, "Москва", 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(joins) != 1 {
		t.Fatalf("joins: %d", len(joins))
	}
	routes, err := join.Materialize(joins, 3)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, route := range routes {
		got[blockIDs(route)] = true
	}
	if !got["out,cheap,back-noon"] || !got["out,cheap,back-night"] || !got["train,cheap,back-night"] || got["train,cheap,back-noon"] {
		t.Fatalf("combinations: %v", got)
	}
	capped, err := join.Materialize(joins, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(capped) != 2 || capped[0].AmountMinor != 1600 || capped[1].AmountMinor != 1600 {
		t.Fatalf("cap: %+v", capped)
	}
}

func TestMaterializeCountsRoundTripOnce(t *testing.T) {
	ticket := priced(roundTrip("rt", "Москва", "Казань", "2026-12-12T08:00:00Z", "2026-12-12T10:00:00Z", "2026-12-14T18:00:00Z", "2026-12-14T20:00:00Z"), 10000)
	parts := &providerv1.SearchPartsResponse{
		TransportParts: []*providerv1.TransportPart{ticket},
		HotelParts:     []*providerv1.HotelPart{pricedHotel(stay("kazan", "Казань", "2026-12-12", "2026-12-14"), 5000)},
	}
	joins, err := join.Assemble(parts, "Москва", 2)
	if err != nil {
		t.Fatal(err)
	}
	routes, err := join.Materialize(joins, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(routes) != 1 {
		t.Fatalf("routes: %d", len(routes))
	}
	if routes[0].AmountMinor != 15000 || routes[0].Partial || routes[0].Currency != "RUB" {
		t.Fatalf("price counted twice or incomplete: %+v", routes[0])
	}
	if blockIDs(routes[0]) != "rt,kazan,rt" {
		t.Fatalf("blocks: %s", blockIDs(routes[0]))
	}
	doc, err := routes[0].Proto()
	if err != nil {
		t.Fatal(err)
	}
	if doc.GetFields()["price"].GetStructValue().GetFields()["amount_minor"].GetNumberValue() != 15000 {
		t.Fatalf("document price: %v", doc.GetFields()["price"])
	}
}

func TestMaterializeMultiCityRoute(t *testing.T) {
	parts := &providerv1.SearchPartsResponse{
		TransportParts: []*providerv1.TransportPart{
			road("mk", "Москва", "Казань", "2026-12-01T08:00:00Z", "2026-12-01T18:00:00Z"),
			road("ks", "Казань", "Санкт-Петербург", "2026-12-03T09:00:00Z", "2026-12-03T20:00:00Z"),
			road("sm", "Санкт-Петербург", "Москва", "2026-12-06T10:00:00Z", "2026-12-06T12:00:00Z"),
		},
		HotelParts: []*providerv1.HotelPart{
			stay("kazan", "Казань", "2026-12-01", "2026-12-03"),
			stay("spb", "Санкт-Петербург", "2026-12-03", "2026-12-06"),
		},
	}
	joins, err := join.Assemble(parts, "Москва", 4)
	if err != nil {
		t.Fatal(err)
	}
	routes, err := join.Materialize(joins, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(routes) != 1 || blockIDs(routes[0]) != "mk,kazan,ks,spb,sm" {
		t.Fatalf("route: %+v", routes)
	}
	if routes[0].Partial != true {
		t.Fatal("missing prices were treated as zero")
	}
}

func priced(part *providerv1.TransportPart, amount int64) *providerv1.TransportPart {
	part.Price = &providerv1.Money{AmountMinor: amount, Currency: "RUB", Scope: "per_person_one_way"}
	if part.GetReturnDepartureAt() != "" {
		part.Price.Scope = "per_person_round_trip"
	}
	return part
}

func pricedHotel(part *providerv1.HotelPart, amount int64) *providerv1.HotelPart {
	part.Price = &providerv1.Money{AmountMinor: amount, Currency: "RUB", Scope: "stay_all_rooms"}
	part.Stars = 4
	return part
}

func blockIDs(route join.Route) string {
	ids := make([]string, len(route.Blocks))
	for i, block := range route.Blocks {
		ids[i] = block.PartID
	}
	return strings.Join(ids, ",")
}
