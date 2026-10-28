package join

import (
	"fmt"
	"sort"
	"strings"
	"time"

	providerv1 "github.com/godnota/pkg/generated/provider/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

const maxExpanded = 256

// Route is one concrete choice of a single option from every slot of a join.
type Route struct {
	ID          string
	JoinID      string
	Cities      []string
	Blocks      []Block
	AmountMinor int64
	Currency    string
	Partial     bool
	Duration    time.Duration
	Stars       int
	Warnings    []string
}

// Block is one chosen road or hotel inside a route.
type Block struct {
	Kind        string
	PartID      string
	From        string
	To          string
	City        string
	Name        string
	Mode        string
	DepartureAt string
	ArrivalAt   string
	CheckIn     string
	CheckOut    string
	Stars       int
	AmountMinor int64
	Currency    string
	PriceScope  string
	Priced      bool
	Warnings    []string
}

// Expand returns every valid route of one join, cheapest complete price first.
// The walk stops at maxExpanded routes. The per-join cap is applied after ranking.
func Expand(item Join) []Route {
	return expand(item, maxExpanded)
}

// Materialize keeps at most m valid routes for each join, in the local price order.
func Materialize(joins []Join, m int) ([]Route, error) {
	if m < 1 {
		return nil, fmt.Errorf("m must be positive")
	}
	var routes []Route
	for _, item := range joins {
		expanded := Expand(item)
		if len(expanded) > m {
			expanded = expanded[:m]
		}
		routes = append(routes, expanded...)
	}
	return routes, nil
}

// Proto returns the document sent to ai-service and returned as a candidate.
func (r Route) Proto() (*structpb.Struct, error) {
	blocks := make([]any, len(r.Blocks))
	for i, block := range r.Blocks {
		item := map[string]any{"kind": block.Kind, "part_id": block.PartID}
		if block.Kind == KindTransport {
			item["from"] = block.From
			item["to"] = block.To
			item["mode"] = block.Mode
			item["departure_at"] = block.DepartureAt
			item["arrival_at"] = block.ArrivalAt
		} else {
			item["city"] = block.City
			item["name"] = block.Name
			item["check_in"] = block.CheckIn
			item["check_out"] = block.CheckOut
			item["stars"] = float64(block.Stars)
		}
		if block.Priced {
			item["price"] = map[string]any{
				"amount_minor": float64(block.AmountMinor),
				"currency":     block.Currency,
				"scope":        block.PriceScope,
			}
		}
		if len(block.Warnings) > 0 {
			item["warnings"] = stringsToAny(block.Warnings)
		}
		blocks[i] = item
	}
	payload := map[string]any{
		"id":               r.ID,
		"join_id":          r.JoinID,
		"cities":           stringsToAny(r.Cities),
		"blocks":           blocks,
		"duration_minutes": float64(r.Duration / time.Minute),
		"stars":            float64(r.Stars),
		"price": map[string]any{
			"amount_minor": float64(r.AmountMinor),
			"currency":     r.Currency,
			"partial":      r.Partial,
		},
	}
	if len(r.Warnings) > 0 {
		payload["warnings"] = stringsToAny(r.Warnings)
	}
	return structpb.NewStruct(payload)
}

type option struct {
	slot      Slot
	transport *providerv1.TransportPart
	hotel     *providerv1.HotelPart
}

func expand(item Join, m int) []Route {
	if len(item.Slots) == 0 {
		return nil
	}
	options := make([][]option, len(item.Slots))
	for i, slot := range item.Slots {
		if slot.Kind == KindTransport {
			for _, part := range slot.Transport {
				if part != nil {
					options[i] = append(options[i], option{slot: slot, transport: part})
				}
			}
		} else {
			for _, part := range slot.Hotels {
				if part != nil {
					options[i] = append(options[i], option{slot: slot, hotel: part})
				}
			}
		}
		if len(options[i]) == 0 {
			return nil
		}
	}
	var found []Route
	seen := map[string]bool{}
	var walk func(int, []option, time.Time, bool, map[string]int)
	walk = func(index int, chosen []option, last time.Time, hasLast bool, roundTrips map[string]int) {
		if len(found) >= maxExpanded {
			return
		}
		if index == len(options) {
			if !roundTripClosed(roundTrips) {
				return
			}
			route, ok := newRoute(item, chosen)
			if ok && !seen[route.ID] {
				seen[route.ID] = true
				found = append(found, route)
			}
			return
		}
		for _, choice := range options[index] {
			nextTrips := roundTrips
			departed := last
			known := hasLast
			if choice.transport != nil {
				second := roundTrips[choice.transport.GetId()] > 0
				dep, arr, ok := legTimes(choice.transport, choice.slot, second)
				if !ok || (hasLast && !dep.After(last)) {
					continue
				}
				if choice.transport.GetReturnDepartureAt() != "" {
					nextTrips = copyCounts(roundTrips)
					nextTrips[choice.transport.GetId()]++
				}
				departed = arr
				known = true
			}
			next := make([]option, len(chosen)+1)
			copy(next, chosen)
			next[len(chosen)] = choice
			walk(index+1, next, departed, known, nextTrips)
		}
	}
	walk(0, nil, time.Time{}, false, map[string]int{})
	sort.Slice(found, func(i, j int) bool {
		if found[i].Partial != found[j].Partial {
			return !found[i].Partial
		}
		if found[i].AmountMinor != found[j].AmountMinor {
			return found[i].AmountMinor < found[j].AmountMinor
		}
		return found[i].ID < found[j].ID
	})
	if len(found) > m {
		found = found[:m]
	}
	return found
}

func roundTripClosed(counts map[string]int) bool {
	for _, n := range counts {
		if n != 2 {
			return false
		}
	}
	return true
}

func copyCounts(counts map[string]int) map[string]int {
	out := make(map[string]int, len(counts)+1)
	for id, n := range counts {
		out[id] = n
	}
	return out
}

func legTimes(part *providerv1.TransportPart, slot Slot, secondHalf bool) (time.Time, time.Time, bool) {
	depRaw, arrRaw := part.GetDepartureAt(), part.GetArrivalAt()
	if secondHalf {
		depRaw, arrRaw = part.GetReturnDepartureAt(), part.GetReturnArrivalAt()
	}
	depDate, dep, okDep := stamp(depRaw)
	arrDate, arr, okArr := stamp(arrRaw)
	if !okDep || !okArr || !arr.After(dep) {
		return time.Time{}, time.Time{}, false
	}
	if slot.DepartureDate != depDate || slot.ArrivalDate != arrDate {
		return time.Time{}, time.Time{}, false
	}
	return dep, arr, true
}

func newRoute(item Join, chosen []option) (Route, bool) {
	blocks := make([]Block, 0, len(chosen))
	ids := make([]string, 0, len(chosen))
	var warnings []string
	seenPrice := map[string]bool{}
	var amount int64
	currency := ""
	partial := false
	var duration time.Duration
	stars := 0
	for _, choice := range chosen {
		if choice.hotel != nil {
			part := choice.hotel
			block := Block{
				Kind: KindHotel, PartID: part.GetId(), City: part.GetCity(), Name: part.GetName(),
				CheckIn: part.GetCheckIn(), CheckOut: part.GetCheckOut(), Stars: int(part.GetStars()),
				Warnings: append([]string{}, part.GetWarnings()...),
			}
			stars += block.Stars
			applyPrice(&block, part.GetPrice())
			blocks = append(blocks, block)
			ids = append(ids, part.GetId())
			warnings = append(warnings, block.Warnings...)
			amount, currency, partial = addPrice(amount, currency, partial, block, false)
			continue
		}
		part := choice.transport
		second := false
		if part.GetReturnDepartureAt() != "" && seenPrice[part.GetId()] {
			second = true
		}
		dep, arr, ok := legTimes(part, choice.slot, second)
		if !ok {
			return Route{}, false
		}
		block := Block{
			Kind: KindTransport, PartID: part.GetId(), From: choice.slot.FromCity, To: choice.slot.ToCity,
			Mode: part.GetMode(), DepartureAt: dep.Format(time.RFC3339), ArrivalAt: arr.Format(time.RFC3339),
			Warnings: append([]string{}, part.GetWarnings()...),
		}
		duration += arr.Sub(dep)
		priced := false
		if part.GetReturnDepartureAt() == "" || !seenPrice[part.GetId()] {
			applyPrice(&block, part.GetPrice())
			priced = true
			seenPrice[part.GetId()] = true
		}
		blocks = append(blocks, block)
		ids = append(ids, part.GetId())
		warnings = append(warnings, block.Warnings...)
		amount, currency, partial = addPrice(amount, currency, partial, block, !priced && part.GetReturnDepartureAt() != "")
	}
	return Route{
		ID: strings.Join(append([]string{item.ID}, ids...), "#"), JoinID: item.ID, Cities: append([]string{}, item.Cities...),
		Blocks: blocks, AmountMinor: amount, Currency: currency, Partial: partial, Duration: duration, Stars: stars, Warnings: warnings,
	}, true
}

func applyPrice(block *Block, price *providerv1.Money) {
	if price == nil || price.GetCurrency() == "" {
		return
	}
	block.Priced = true
	block.AmountMinor = price.GetAmountMinor()
	block.Currency = price.GetCurrency()
	block.PriceScope = price.GetScope()
}

func addPrice(amount int64, currency string, partial bool, block Block, skip bool) (int64, string, bool) {
	if skip {
		return amount, currency, partial
	}
	if !block.Priced {
		return amount, currency, true
	}
	if currency == "" {
		currency = block.Currency
	} else if currency != block.Currency {
		return amount, currency, true
	}
	return amount + block.AmountMinor, currency, partial
}

func stringsToAny(values []string) []any {
	out := make([]any, len(values))
	for i, value := range values {
		out[i] = value
	}
	return out
}
