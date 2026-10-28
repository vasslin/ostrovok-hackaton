package join_test

import (
	"strings"
	"testing"
	"time"

	providerv1 "github.com/godnota/pkg/generated/provider/v1"
	"github.com/godnota/route-service/internal/join"
)

func TestMultiCitySpreadsDateCombinations(t *testing.T) {
	parts := &providerv1.SearchPartsResponse{
		TransportParts: []*providerv1.TransportPart{
			road("mk-01a", "Москва", "Казань", "2026-12-01T08:00:00Z", "2026-12-01T12:00:00Z"),
			road("mk-01b", "Москва", "Казань", "2026-12-01T09:00:00Z", "2026-12-01T18:00:00Z"),
			road("mk-10", "Москва", "Казань", "2026-12-10T08:00:00Z", "2026-12-10T18:00:00Z"),
			road("ks-03", "Казань", "Санкт-Петербург", "2026-12-03T09:00:00Z", "2026-12-03T20:00:00Z"),
			road("ks-12", "Казань", "Санкт-Петербург", "2026-12-12T09:00:00Z", "2026-12-12T20:00:00Z"),
			road("ks-bad", "Казань", "Санкт-Петербург", "2026-12-01T12:00:00Z", "2026-12-01T20:00:00Z"),
			road("sm-06", "Санкт-Петербург", "Москва", "2026-12-06T10:00:00Z", "2026-12-06T12:00:00Z"),
			road("sm-16", "Санкт-Петербург", "Москва", "2026-12-16T10:00:00Z", "2026-12-16T12:00:00Z"),
		},
		HotelParts: []*providerv1.HotelPart{
			stay("kazan-wide", "Казань", "2026-12-01", "2026-12-16"),
			stay("kazan-short", "Казань", "2026-12-01", "2026-12-03"),
			stay("kazan-late", "Казань", "2026-12-20", "2026-12-22"),
			stay("spb-wide", "Санкт-Петербург", "2026-12-01", "2026-12-16"),
		},
	}
	joins, err := join.Assemble(parts, "Москва", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(joins) != 3 {
		t.Fatalf("joins: got %d, want 3", len(joins))
	}
	departures := map[string]bool{}
	for _, item := range joins {
		if strings.Join(item.Cities, ">") != "Москва>Казань>Санкт-Петербург>Москва" {
			t.Fatalf("cities: %v", item.Cities)
		}
		departures[item.Slots[0].DepartureDate] = true
		if ids(item).contains("ks-bad") || ids(item).contains("kazan-late") {
			t.Fatalf("incompatible part kept: %s", item.ID)
		}
		if item.Slots[0].DepartureDate == "2026-12-01" && len(item.Slots[0].Transport) != 2 {
			t.Fatalf("departure 1 Dec variants: %d", len(item.Slots[0].Transport))
		}
	}
	if !departures["2026-12-01"] || !departures["2026-12-10"] {
		t.Fatalf("date combinations collapsed: %v", departures)
	}
	short := findJoin(t, joins, "2026-12-03")
	if hotelIDs(short, "Казань") != "kazan-short,kazan-wide" {
		t.Fatalf("kazan hotels for 3 Dec: %s", hotelIDs(short, "Казань"))
	}
	long := findJoin(t, joins, "2026-12-12")
	if hotelIDs(long, "Казань") != "kazan-wide" {
		t.Fatalf("kazan hotels for 12 Dec: %s", hotelIDs(long, "Казань"))
	}
	if hotelIDs(long, "Санкт-Петербург") != "spb-wide" {
		t.Fatalf("spb hotel: %s", hotelIDs(long, "Санкт-Петербург"))
	}
}

func TestAlternativeCitiesBothAppear(t *testing.T) {
	parts := &providerv1.SearchPartsResponse{
		TransportParts: []*providerv1.TransportPart{
			road("mk-10", "Москва", "Казань", "2026-12-10T08:00:00Z", "2026-12-10T18:00:00Z"),
			road("mk-11", "Москва", "Казань", "2026-12-11T08:00:00Z", "2026-12-11T18:00:00Z"),
			road("km-14", "Казань", "Москва", "2026-12-14T10:00:00Z", "2026-12-14T20:00:00Z"),
			road("km-15", "Казань", "Москва", "2026-12-15T10:00:00Z", "2026-12-15T20:00:00Z"),
			road("ms-10", "Москва", "Санкт-Петербург", "2026-12-10T08:00:00Z", "2026-12-10T10:00:00Z"),
			road("ms-11", "Москва", "Санкт-Петербург", "2026-12-11T08:00:00Z", "2026-12-11T10:00:00Z"),
			road("sm-14", "Санкт-Петербург", "Москва", "2026-12-14T19:00:00Z", "2026-12-14T21:00:00Z"),
			road("sm-15", "Санкт-Петербург", "Москва", "2026-12-15T19:00:00Z", "2026-12-15T21:00:00Z"),
		},
		HotelParts: []*providerv1.HotelPart{
			stay("kazan", "Казань", "2026-12-10", "2026-12-15"),
			stay("spb", "Санкт-Петербург", "2026-12-10", "2026-12-15"),
		},
	}
	joins, err := join.Assemble(parts, "москва", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(joins) != 2 {
		t.Fatalf("joins: got %d, want 2", len(joins))
	}
	got := map[string]bool{}
	for _, item := range joins {
		if item.Cities[0] != "москва" {
			t.Fatalf("origin spelling: %v", item.Cities)
		}
		got[item.Cities[1]] = true
	}
	if !got["Казань"] || !got["Санкт-Петербург"] {
		t.Fatalf("city combinations collapsed: %v", got)
	}
}

func TestRoundTripPartStaysOnMatchingDates(t *testing.T) {
	parts := &providerv1.SearchPartsResponse{
		TransportParts: []*providerv1.TransportPart{
			roundTrip("rt", "Москва", "Казань", "2026-12-12T08:00:00Z", "2026-12-12T10:00:00Z", "2026-12-14T18:00:00Z", "2026-12-14T20:00:00Z"),
			road("ow-12", "Москва", "Казань", "2026-12-12T07:00:00Z", "2026-12-12T15:00:00Z"),
			road("ow-13", "Москва", "Казань", "2026-12-13T07:00:00Z", "2026-12-13T15:00:00Z"),
			road("back-14", "Казань", "Москва", "2026-12-14T12:00:00Z", "2026-12-14T22:00:00Z"),
			road("back-15", "Казань", "Москва", "2026-12-15T12:00:00Z", "2026-12-15T22:00:00Z"),
		},
		HotelParts: []*providerv1.HotelPart{
			stay("kazan", "Казань", "2026-12-12", "2026-12-15"),
		},
	}
	joins, err := join.Assemble(parts, "Москва", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(joins) != 4 {
		t.Fatalf("joins: got %d, want 4", len(joins))
	}
	matched := findByDates(t, joins, "2026-12-12", "2026-12-14")
	if transportIDs(matched, "Москва", "Казань") != "ow-12,rt" || transportIDs(matched, "Казань", "Москва") != "back-14,rt" {
		t.Fatalf("round trip split: %+v", matched.Slots)
	}
	if strings.Join(matched.RoundTripParts, ",") != "rt" {
		t.Fatalf("round trip marker: %v", matched.RoundTripParts)
	}
	other := findByDates(t, joins, "2026-12-12", "2026-12-15")
	if strings.Contains(transportIDs(other, "Москва", "Казань")+transportIDs(other, "Казань", "Москва"), "rt") {
		t.Fatalf("round trip leaked onto another return date: %s", other.ID)
	}
}

func TestDropsIncompatibleConnectionsAndHotels(t *testing.T) {
	parts := &providerv1.SearchPartsResponse{
		TransportParts: []*providerv1.TransportPart{
			road("flight", "Москва", "Казань", "2026-12-13T05:00:00Z", "2026-12-13T08:00:00Z"),
			road("train", "Москва", "Казань", "2026-12-13T06:00:00Z", "2026-12-13T20:00:00Z"),
			road("back", "Казань", "Москва", "2026-12-13T12:00:00Z", "2026-12-13T14:00:00Z"),
		},
		HotelParts: []*providerv1.HotelPart{
			stay("good", "Казань", "2026-12-12", "2026-12-13"),
			stay("early", "Казань", "2026-12-10", "2026-12-14"),
			stay("late-checkin", "Казань", "2026-12-14", "2026-12-16"),
			stay("short", "Казань", "2026-12-12", "2026-12-12"),
		},
	}
	joins, err := join.Assemble(parts, "Москва", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(joins) != 1 {
		t.Fatalf("joins: got %d, want 1", len(joins))
	}
	if transportIDs(joins[0], "Москва", "Казань") != "flight" {
		t.Fatalf("train leaves after the only return: %s", transportIDs(joins[0], "Москва", "Казань"))
	}
	if hotelIDs(joins[0], "Казань") != "early,good" {
		t.Fatalf("hotels: %s", hotelIDs(joins[0], "Казань"))
	}
	onlyTrain := &providerv1.SearchPartsResponse{
		TransportParts: []*providerv1.TransportPart{parts.TransportParts[1], parts.TransportParts[2]},
		HotelParts:     parts.HotelParts,
	}
	none, err := join.Assemble(onlyTrain, "Москва", 5)
	if err != nil || len(none) != 0 {
		t.Fatalf("impossible connection became a join: %v %v", none, err)
	}
}

func TestOneWayStayAndKBounds(t *testing.T) {
	parts := &providerv1.SearchPartsResponse{
		TransportParts: []*providerv1.TransportPart{
			road("d1", "Москва", "Казань", "2026-12-01T08:00:00Z", "2026-12-01T12:00:00Z"),
			road("d5", "Москва", "Казань", "2026-12-05T08:00:00Z", "2026-12-05T12:00:00Z"),
		},
		HotelParts: []*providerv1.HotelPart{
			stay("nights", "Казань", "2026-12-01", "2026-12-08"),
			stay("same-day", "Казань", "2026-12-01", "2026-12-01"),
		},
	}
	joins, err := join.Assemble(parts, "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(joins) != 2 {
		t.Fatalf("joins: got %d, want 2", len(joins))
	}
	for i := 1; i < len(joins); i++ {
		if joins[i-1].ID > joins[i].ID {
			t.Fatalf("not sorted: %s then %s", joins[i-1].ID, joins[i].ID)
		}
	}
	if joins[0].Cities[0] != "Москва" || len(joins[0].Slots) != 2 || joins[0].Slots[1].Kind != join.KindHotel {
		t.Fatalf("one-way shape: %+v", joins[0])
	}
	if hotelIDs(joins[0], "Казань") != "nights" {
		t.Fatalf("same-day checkout kept: %s", hotelIDs(joins[0], "Казань"))
	}
	limited, err := join.Assemble(parts, "Москва", 1)
	if err != nil || len(limited) != 1 {
		t.Fatalf("k=1: %v %v", limited, err)
	}
	if _, err = join.Assemble(parts, "Москва", 0); err == nil {
		t.Fatal("k=0 accepted")
	}
	empty, err := join.Assemble(nil, "Москва", 3)
	if err != nil || len(empty) != 0 {
		t.Fatalf("nil parts: %v %v", empty, err)
	}
	nowhere, err := join.Assemble(parts, "Сочи", 3)
	if err != nil || len(nowhere) != 0 {
		t.Fatalf("unknown origin: %v %v", nowhere, err)
	}
}

func road(id, from, to, dep, arr string) *providerv1.TransportPart {
	return &providerv1.TransportPart{
		Id: id, Provider: "test", Source: "synthetic", FetchedAt: time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Mode: "train", Origin: from, Destination: to, OriginCity: from, DestinationCity: to,
		DepartureAt: dep, ArrivalAt: arr,
	}
}

func roundTrip(id, from, to, dep, arr, retDep, retArr string) *providerv1.TransportPart {
	part := road(id, from, to, dep, arr)
	part.Mode = "plane"
	part.ReturnDepartureAt = retDep
	part.ReturnArrivalAt = retArr
	return part
}

func stay(id, city, in, out string) *providerv1.HotelPart {
	return &providerv1.HotelPart{Id: id, Provider: "test", Source: "synthetic", City: city, Name: id, CheckIn: in, CheckOut: out}
}

type idSet map[string]bool

func ids(item join.Join) idSet {
	set := idSet{}
	for _, slot := range item.Slots {
		for _, part := range slot.Transport {
			set[part.GetId()] = true
		}
		for _, part := range slot.Hotels {
			set[part.GetId()] = true
		}
	}
	return set
}

func (s idSet) contains(id string) bool { return s[id] }

func findJoin(t *testing.T, joins []join.Join, hopDeparture string) join.Join {
	t.Helper()
	for _, item := range joins {
		for _, slot := range item.Slots {
			if slot.Kind == join.KindTransport && slot.DepartureDate == hopDeparture && slot.FromCity != item.Cities[0] {
				return item
			}
		}
	}
	t.Fatalf("no join with hop %s", hopDeparture)
	return join.Join{}
}

func findByDates(t *testing.T, joins []join.Join, outbound, back string) join.Join {
	t.Helper()
	for _, item := range joins {
		var deps []string
		for _, slot := range item.Slots {
			if slot.Kind == join.KindTransport {
				deps = append(deps, slot.DepartureDate)
			}
		}
		if len(deps) == 2 && deps[0] == outbound && deps[1] == back {
			return item
		}
	}
	t.Fatalf("no join %s -> %s", outbound, back)
	return join.Join{}
}

func transportIDs(item join.Join, from, to string) string {
	var list []string
	for _, slot := range item.Slots {
		if slot.Kind == join.KindTransport && slot.FromCity == from && slot.ToCity == to {
			for _, part := range slot.Transport {
				list = append(list, part.GetId())
			}
		}
	}
	return strings.Join(list, ",")
}

func hotelIDs(item join.Join, city string) string {
	var list []string
	for _, slot := range item.Slots {
		if slot.Kind == join.KindHotel && slot.City == city {
			for _, part := range slot.Hotels {
				list = append(list, part.GetId())
			}
		}
	}
	return strings.Join(list, ",")
}
