package routebuilder

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	pb "github.com/godnota/pkg/generated/provider/v1"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type candidate struct {
	id        string
	out, back *pb.TransportPart
	hotel     *pb.HotelPart
	total     *int64
	warnings  []string
}

func Build(ctx context.Context, trip providerkit.TripRequest, parts *pb.SearchPartsResponse) ([]*structpb.Struct, error) {
	if len(trip.Stops) > 0 {
		return buildItinerary(ctx, trip, parts, false)
	}
	if trip.Destination == nil || strings.TrimSpace(trip.Destination.City) == "" || trip.Dates.DepartureDate == "" {
		return nil, fmt.Errorf("destination and departure date are required")
	}
	start, err := time.Parse(time.DateOnly, trip.Dates.DepartureDate)
	if err != nil {
		return nil, fmt.Errorf("invalid departure date")
	}
	end := trip.Dates.ReturnDate
	if end == "" && trip.Dates.Nights > 0 {
		end = start.AddDate(0, 0, trip.Dates.Nights).Format(time.DateOnly)
	}
	round := trip.Transport == nil || trip.Transport.RoundTrip == nil || *trip.Transport.RoundTrip
	needHotel := trip.Hotel == nil || trip.Hotel.Needed == nil || *trip.Hotel.Needed
	if (round || needHotel) && end == "" {
		return nil, fmt.Errorf("return date or nights required")
	}
	if end != "" {
		d, e := time.Parse(time.DateOnly, end)
		if e != nil || !d.After(start) {
			return nil, fmt.Errorf("invalid return date")
		}
	}
	outbound, returns := []*pb.TransportPart{}, []*pb.TransportPart{}
	for _, p := range parts.GetTransportParts() {
		if p == nil || p.GetId() == "" || !validTimes(p.GetDepartureAt(), p.GetArrivalAt()) || !allowedTransport(p, trip) {
			continue
		}
		from, to := p.GetOriginCity(), p.GetDestinationCity()
		if from == "" {
			from = p.GetOrigin()
		}
		if to == "" {
			to = p.GetDestination()
		}
		if same(from, trip.Origin.City) && same(to, trip.Destination.City) && strings.HasPrefix(p.GetDepartureAt(), trip.Dates.DepartureDate) {
			outbound = append(outbound, p)
		}
		if same(from, trip.Destination.City) && same(to, trip.Origin.City) && end != "" && strings.HasPrefix(p.GetDepartureAt(), end) && p.GetReturnDepartureAt() == "" {
			returns = append(returns, p)
		}
	}
	sort.Slice(outbound, func(i, j int) bool { return outbound[i].GetId() < outbound[j].GetId() })
	sort.Slice(returns, func(i, j int) bool { return returns[i].GetId() < returns[j].GetId() })
	if len(outbound) > 40 {
		outbound = outbound[:40]
	}
	if len(returns) > 40 {
		returns = returns[:40]
	}
	hotels := []*pb.HotelPart{nil}
	if needHotel {
		hotels = nil
		for _, h := range parts.GetHotelParts() {
			if h == nil || h.GetId() == "" || !same(h.GetCity(), trip.Destination.City) || h.GetCheckIn() != trip.Dates.DepartureDate || h.GetCheckOut() != end {
				continue
			}
			if trip.Hotel != nil && trip.Hotel.StarRatingMin != nil && h.GetStars() > 0 && int(h.GetStars()) < *trip.Hotel.StarRatingMin {
				continue
			}
			hotels = append(hotels, h)
		}
		sort.Slice(hotels, func(i, j int) bool { return hotels[i].GetId() < hotels[j].GetId() })
		if len(hotels) > 40 {
			hotels = hotels[:40]
		}
	}
	items := []candidate{}
	seen := map[string]bool{}
	for _, out := range outbound {
		backs := []*pb.TransportPart{nil}
		if round {
			if out.GetReturnDepartureAt() != "" {
				if !strings.HasPrefix(out.GetReturnDepartureAt(), end) || !validTimes(out.GetReturnDepartureAt(), out.GetReturnArrivalAt()) || !after(out.GetReturnDepartureAt(), out.GetArrivalAt()) {
					continue
				}
			} else {
				backs = returns
			}
		} else if out.GetReturnDepartureAt() != "" {
			continue
		}
		for _, back := range backs {
			if back != nil && !after(back.GetDepartureAt(), out.GetArrivalAt()) {
				continue
			}
			for _, hotel := range hotels {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				if hotel != nil && (strings.Compare(out.GetArrivalAt()[:10], hotel.GetCheckIn()) > 0 || (back != nil && strings.Compare(back.GetDepartureAt()[:10], hotel.GetCheckOut()) < 0)) {
					continue
				}
				id := makeID(out, back, hotel)
				if seen[id] {
					continue
				}
				seen[id] = true
				c := candidate{id: id, out: out, back: back, hotel: hotel, warnings: []string{"Deterministic ordering; AI ranking has not been applied.", "Transfers, hotel check-in times, amenities and final availability require confirmation."}}
				c.warnings = append(c.warnings, out.GetWarnings()...)
				if back != nil {
					c.warnings = append(c.warnings, back.GetWarnings()...)
				}
				if hotel != nil {
					c.warnings = append(c.warnings, hotel.GetWarnings()...)
				}
				c.total = total(trip, out, back, hotel)
				if c.total == nil {
					c.warnings = append(c.warnings, "Total price is unknown or cannot be safely combined.")
				}
				if trip.Budget != nil && trip.Budget.Total > 0 {
					if c.total != nil && trip.Budget.Scope == "total" && int64(trip.Budget.Total) <= math.MaxInt64/100 && *c.total > int64(trip.Budget.Total)*100 {
						if trip.Budget.IsHard {
							continue
						}
						c.warnings = append(c.warnings, "Known total exceeds the stated budget.")
					} else if c.total == nil || trip.Budget.Scope != "total" {
						c.warnings = append(c.warnings, "Budget compatibility is unverified.")
					}
				}
				items = append(items, c)
			}
		}
	}
	sort.Slice(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if (a.total == nil) != (b.total == nil) {
			return a.total != nil
		}
		if a.total != nil && *a.total != *b.total {
			return *a.total < *b.total
		}
		return a.id < b.id
	})
	if len(items) > 20 {
		items = items[:20]
	}
	result := make([]*structpb.Struct, 0, len(items))
	for _, c := range items {
		transport := []any{object(c.out)}
		if c.back != nil {
			transport = append(transport, object(c.back))
		}
		warnings := make([]any, 0, len(c.warnings))
		for _, warning := range c.warnings {
			warnings = append(warnings, warning)
		}
		doc := map[string]any{"id": c.id, "transport": transport, "warnings": warnings, "price_status": "unknown", "ranking": "deterministic", "bookable": false}
		if c.hotel != nil {
			doc["hotel"] = object(c.hotel)
		}
		if c.total != nil {
			doc["price_status"] = "estimated"
			doc["total_price"] = map[string]any{"amount_minor": strconv.FormatInt(*c.total, 10), "currency": trip.Currency, "scope": "estimated_trip_total"}
		}
		v, e := structpb.NewStruct(doc)
		if e != nil {
			return nil, e
		}
		result = append(result, v)
	}
	if trip.Transport == nil || !trip.Transport.DirectOnly {
		connected, err := buildItinerary(ctx, trip, parts, true)
		if err != nil {
			return nil, err
		}
		ids := map[string]bool{}
		for _, candidate := range result {
			ids[candidate.GetFields()["id"].GetStringValue()] = true
		}
		for _, candidate := range connected {
			id := candidate.GetFields()["id"].GetStringValue()
			if !ids[id] {
				result = append(result, candidate)
				ids[id] = true
			}
		}
		if len(result) > 20 {
			result = result[:20]
		}
	}
	return result, nil
}

func object(m proto.Message) map[string]any {
	data, _ := protojson.Marshal(m)
	var result map[string]any
	_ = json.Unmarshal(data, &result)
	return result
}
func same(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b)) && strings.TrimSpace(a) != ""
}
func validTimes(a, b string) bool {
	start, e1 := time.Parse(time.RFC3339, a)
	end, e2 := time.Parse(time.RFC3339, b)
	return e1 == nil && e2 == nil && end.After(start)
}
func after(a, b string) bool {
	x, e1 := time.Parse(time.RFC3339, a)
	y, e2 := time.Parse(time.RFC3339, b)
	return e1 == nil && e2 == nil && x.After(y)
}
func makeID(a, b *pb.TransportPart, h *pb.HotelPart) string {
	s := a.GetId() + "|" + b.GetId() + "|" + h.GetId()
	return fmt.Sprintf("route:%x", sha256.Sum256([]byte(s)))
}
func allowedTransport(p *pb.TransportPart, trip providerkit.TripRequest) bool {
	t := trip.Transport
	if p.GetTransfers() < 0 || p.GetReturnTransfers() < 0 || p.GetReturnTransfers() > 0 {
		return false
	}
	if p.GetTransfers() > 0 && !verifiedProviderLegs(p, trip) {
		return false
	}
	if t == nil {
		return true
	}
	if len(t.Modes) > 0 {
		found := false
		for _, mode := range t.Modes {
			if mode == p.GetMode() {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if t.DirectOnly && (p.GetTransfers() != 0 || p.GetReturnTransfers() != 0) {
		return false
	}
	if t.MaxTransfers != nil && (int(p.GetTransfers()) > *t.MaxTransfers || int(p.GetReturnTransfers()) > *t.MaxTransfers) {
		return false
	}
	for _, leg := range [][2]string{{p.GetDepartureAt(), p.GetArrivalAt()}, {p.GetReturnDepartureAt(), p.GetReturnArrivalAt()}} {
		if leg[0] == "" {
			continue
		}
		dep, e1 := time.Parse(time.RFC3339, leg[0])
		arr, e2 := time.Parse(time.RFC3339, leg[1])
		if e1 != nil || e2 != nil || !arr.After(dep) {
			return false
		}
		if t.MaxDurationHours > 0 && arr.Sub(dep).Hours() > t.MaxDurationHours {
			return false
		}
		clock := dep.Format("15:04")
		from, to := t.DepartureTimeFrom, t.DepartureTimeTo
		if from != "" && to != "" && from > to {
			if clock < from && clock > to {
				return false
			}
		} else {
			if (from != "" && clock < from) || (to != "" && clock > to) {
				return false
			}
		}
	}
	return true
}
func total(trip providerkit.TripRequest, out, back *pb.TransportPart, hotel *pb.HotelPart) *int64 {
	if trip.Travelers.Adults < 1 {
		return nil
	}
	if out == nil || out.GetPrice() == nil {
		return nil
	}
	if out.GetReturnDepartureAt() != "" {
		if back != nil || out.GetPrice().GetScope() != "per_person_round_trip" {
			return nil
		}
	} else if out.GetPrice().GetScope() != "per_person_one_way" {
		return nil
	}
	if back != nil && (back.GetPrice() == nil || back.GetPrice().GetScope() != "per_person_one_way") {
		return nil
	}
	var sum int64
	for _, p := range []*pb.TransportPart{out, back} {
		if p == nil {
			continue
		}
		v := p.GetPrice()
		if v == nil || v.GetCurrency() != trip.Currency || len(trip.Travelers.ChildrenAges) > 0 || (v.GetScope() != "per_person_one_way" && v.GetScope() != "per_person_round_trip") || v.GetAmountMinor() < 0 || v.GetAmountMinor() > math.MaxInt64/int64(trip.Travelers.Adults) {
			return nil
		}
		n := v.GetAmountMinor() * int64(trip.Travelers.Adults)
		if sum > math.MaxInt64-n {
			return nil
		}
		sum += n
	}
	if hotel != nil {
		p := hotel.GetPrice()
		if p == nil || p.GetCurrency() != trip.Currency || p.GetScope() != "stay_all_rooms" || p.GetAmountMinor() < 0 || sum > math.MaxInt64-p.GetAmountMinor() {
			return nil
		}
		sum += p.GetAmountMinor()
	}
	return &sum
}
