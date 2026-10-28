package routebuilder

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	pb "github.com/godnota/pkg/generated/provider/v1"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
	"google.golang.org/protobuf/types/known/structpb"
)

type stop struct{ city, in, out string }
type pathState struct {
	legs      []*pb.TransportPart
	hotels    []*pb.HotelPart
	transfers int
}

func buildItinerary(ctx context.Context, trip providerkit.TripRequest, parts *pb.SearchPartsResponse, transfersOnly bool) ([]*structpb.Struct, error) {
	if trip.Destination == nil || trip.Destination.City == "" {
		return nil, fmt.Errorf("destination required")
	}
	stops := []stop{}
	for _, s := range trip.Stops {
		stops = append(stops, stop{s.City, s.CheckIn, s.CheckOut})
	}
	if len(stops) == 0 {
		end := trip.Dates.ReturnDate
		if end == "" && trip.Dates.Nights > 0 {
			d, e := time.Parse(time.DateOnly, trip.Dates.DepartureDate)
			if e != nil {
				return nil, e
			}
			end = d.AddDate(0, 0, trip.Dates.Nights).Format(time.DateOnly)
		}
		stops = append(stops, stop{trip.Destination.City, trip.Dates.DepartureDate, end})
	}
	maxTransfers := 2
	minWait := 90 * time.Minute
	maxWait := 12 * time.Hour
	if trip.Transport != nil {
		if trip.Transport.MaxTransfers != nil {
			maxTransfers = *trip.Transport.MaxTransfers
		}
		if trip.Transport.DirectOnly {
			maxTransfers = 0
		}
		if trip.Transport.MinTransferMinutes > 0 {
			minWait = time.Duration(trip.Transport.MinTransferMinutes) * time.Minute
		}
		if trip.Transport.MaxWaitHours > 0 {
			maxWait = time.Duration(trip.Transport.MaxWaitHours) * time.Hour
		}
	}
	edges := []*pb.TransportPart{}
	for _, p := range parts.GetTransportParts() {
		if p != nil && p.GetId() != "" && p.GetReturnDepartureAt() == "" && validTimes(p.GetDepartureAt(), p.GetArrivalAt()) && allowedTransport(p, trip) {
			edges = append(edges, p)
		}
	}
	sort.Slice(edges, func(i, j int) bool { return edges[i].GetId() < edges[j].GetId() })
	if len(edges) > 120 {
		edges = edges[:120]
	}
	states := []pathState{{}}
	from := trip.Origin.City
	date := trip.Dates.DepartureDate
	needHotel := trip.Hotel == nil || trip.Hotel.Needed == nil || *trip.Hotel.Needed
	for _, s := range stops {
		paths := findPaths(edges, from, s.city, date, s.in, minWait, maxWait, maxTransfers, trip)
		if len(paths) == 0 {
			return []*structpb.Struct{}, nil
		}
		hotels := []*pb.HotelPart{nil}
		if needHotel {
			hotels = nil
			for _, h := range parts.GetHotelParts() {
				if h != nil && h.GetId() != "" && same(h.GetCity(), s.city) && h.GetCheckIn() == s.in && h.GetCheckOut() == s.out {
					if trip.Hotel != nil && trip.Hotel.StarRatingMin != nil && h.GetStars() > 0 && int(h.GetStars()) < *trip.Hotel.StarRatingMin {
						continue
					}
					hotels = append(hotels, h)
				}
			}
			sort.Slice(hotels, func(i, j int) bool { return hotels[i].GetId() < hotels[j].GetId() })
			if len(hotels) > 8 {
				hotels = hotels[:8]
			}
		}
		states = combine(ctx, states, paths, hotels)
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(states) == 0 {
			return []*structpb.Struct{}, nil
		}
		from = s.city
		date = s.out
	}
	round := trip.Transport == nil || trip.Transport.RoundTrip == nil || *trip.Transport.RoundTrip
	if round {
		paths := findPaths(edges, from, trip.Origin.City, date, "", minWait, maxWait, maxTransfers, trip)
		states = combine(ctx, states, paths, []*pb.HotelPart{nil})
	}
	result := []*structpb.Struct{}
	seen := map[string]bool{}
	for _, s := range states {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if transfersOnly && s.transfers == 0 {
			continue
		}
		ids := []string{}
		legs := []any{}
		hotels := []any{}
		warnings := []any{"Детерминированная сборка, не AI-ранжирование.", "Время до отеля, бронирование и условия требуют проверки."}
		for _, p := range s.legs {
			ids = append(ids, p.GetId())
			legs = append(legs, object(p))
			for _, w := range p.GetWarnings() {
				warnings = append(warnings, w)
			}
		}
		for _, h := range s.hotels {
			ids = append(ids, h.GetId())
			hotels = append(hotels, object(h))
			for _, w := range h.GetWarnings() {
				warnings = append(warnings, w)
			}
		}
		key := fmt.Sprintf("route:%x", sha256.Sum256([]byte(strings.Join(ids, "\x00"))))
		if len(trip.Stops) == 0 && len(s.legs) <= 2 && len(s.hotels) <= 1 {
			var back *pb.TransportPart
			var hotel *pb.HotelPart
			if len(s.legs) == 2 {
				back = s.legs[1]
			}
			if len(s.hotels) == 1 {
				hotel = s.hotels[0]
			}
			key = makeID(s.legs[0], back, hotel)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		price := itineraryTotal(trip, s)
		if price == nil {
			warnings = append(warnings, "Итоговая цена неизвестна; бюджет и цена группы не подтверждены.")
		}
		if trip.Budget != nil && trip.Budget.Total > 0 && price != nil && trip.Budget.Scope == "total" && int64(trip.Budget.Total) <= math.MaxInt64/100 && *price > int64(trip.Budget.Total)*100 {
			if trip.Budget.IsHard {
				continue
			}
			warnings = append(warnings, "Известная оценка выше бюджета.")
		}
		if trip.Budget != nil && trip.Budget.Total > 0 && (price == nil || trip.Budget.Scope != "total") {
			warnings = append(warnings, "Соответствие бюджету не подтверждено.")
		}
		doc := map[string]any{"id": key, "transport": legs, "hotels": hotels, "warnings": warnings, "price_status": "unknown", "ranking": "deterministic", "bookable": false, "transfers": s.transfers}
		if len(hotels) > 0 {
			doc["hotel"] = hotels[0]
		}
		if price != nil {
			doc["price_status"] = "estimated"
			doc["total_price"] = map[string]any{"amount_minor": strconv.FormatInt(*price, 10), "currency": trip.Currency, "scope": "estimated_trip_total"}
		}
		resultItem, err := structpb.NewStruct(doc)
		if err != nil {
			return nil, err
		}
		result = append(result, resultItem)
		if len(result) >= 20 {
			break
		}
	}
	return result, nil
}

func combine(ctx context.Context, states []pathState, paths [][]*pb.TransportPart, hotels []*pb.HotelPart) []pathState {
	out := []pathState{}
	for _, s := range states {
		for _, path := range paths {
			for _, hotel := range hotels {
				if ctx.Err() != nil {
					return out
				}
				next := pathState{legs: append(append([]*pb.TransportPart{}, s.legs...), path...), hotels: append([]*pb.HotelPart{}, s.hotels...), transfers: s.transfers + len(path) - 1}
				for _, p := range path {
					next.transfers += int(p.GetTransfers())
				}
				if hotel != nil {
					next.hotels = append(next.hotels, hotel)
				}
				out = append(out, next)
				if len(out) >= 128 {
					return out
				}
			}
		}
	}
	return out
}

func findPaths(edges []*pb.TransportPart, from, to, date, arrivalDate string, minWait, maxWait time.Duration, maxTransfers int, trip providerkit.TripRequest) [][]*pb.TransportPart {
	results := [][]*pb.TransportPart{}
	seen := map[string]bool{}
	visited := map[string]bool{strings.ToLower(strings.TrimSpace(from)): true}
	var visit func(string, []*pb.TransportPart, int)
	visit = func(city string, path []*pb.TransportPart, n int) {
		if len(results) >= 12 || len(path) > 3 {
			return
		}
		if same(city, to) && len(path) > 0 {
			last := path[len(path)-1]
			if arrivalDate != "" && !strings.HasPrefix(last.GetArrivalAt(), arrivalDate) {
				return
			}
			if trip.Transport != nil && trip.Transport.MaxDurationHours > 0 {
				a, _ := time.Parse(time.RFC3339, path[0].GetDepartureAt())
				b, _ := time.Parse(time.RFC3339, last.GetArrivalAt())
				if b.Sub(a).Hours() > trip.Transport.MaxDurationHours {
					return
				}
			}
			results = append(results, append([]*pb.TransportPart{}, path...))
			return
		}
		for _, p := range edges {
			if seen[p.GetId()] || !same(edgeFrom(p), city) {
				continue
			}
			nextCity := strings.ToLower(strings.TrimSpace(edgeTo(p)))
			if nextCity == "" || visited[nextCity] {
				continue
			}
			next := n + int(p.GetTransfers())
			if len(path) > 0 {
				next++
			}
			if next > maxTransfers {
				continue
			}
			if len(path) == 0 && !strings.HasPrefix(p.GetDepartureAt(), date) {
				continue
			}
			if len(path) > 0 {
				prev := path[len(path)-1]
				if !sameTransferStation(prev, p) {
					continue
				}
				a, _ := time.Parse(time.RFC3339, prev.GetArrivalAt())
				b, _ := time.Parse(time.RFC3339, p.GetDepartureAt())
				wait := b.Sub(a)
				if wait < minWait || wait > maxWait {
					continue
				}
			}
			seen[p.GetId()] = true
			visited[nextCity] = true
			visit(edgeTo(p), append(path, p), next)
			delete(seen, p.GetId())
			delete(visited, nextCity)
		}
	}
	visit(from, nil, 0)
	return results
}

func sameTransferStation(a, b *pb.TransportPart) bool {
	if a.GetDestination() == "" || b.GetOrigin() == "" || !same(a.GetDestination(), b.GetOrigin()) {
		return false
	}
	if a.GetSource() == "synthetic" && b.GetSource() == "synthetic" {
		return true
	}
	code := strings.ToLower(a.GetDestination())
	if a.GetProvider() != "yandex_rasp" || b.GetProvider() != "yandex_rasp" || !strings.HasPrefix(code, "s") || len(code) < 2 {
		return false
	}
	_, err := strconv.ParseUint(code[1:], 10, 64)
	return err == nil
}

func verifiedProviderLegs(p *pb.TransportPart, trip providerkit.TripRequest) bool {
	legs := p.GetLegs()
	if p.GetProvider() != "yandex_rasp" || len(legs) != int(p.GetTransfers())+1 || len(legs) < 2 || len(legs) > 4 {
		return false
	}
	minWait, maxWait := 90*time.Minute, 12*time.Hour
	if trip.Transport != nil {
		if trip.Transport.MinTransferMinutes > 0 {
			minWait = time.Duration(trip.Transport.MinTransferMinutes) * time.Minute
		}
		if trip.Transport.MaxWaitHours > 0 {
			maxWait = time.Duration(trip.Transport.MaxWaitHours) * time.Hour
		}
	}
	if legs[0].GetDepartureAt() != p.GetDepartureAt() || legs[len(legs)-1].GetArrivalAt() != p.GetArrivalAt() {
		return false
	}
	for i, leg := range legs {
		if leg == nil || leg.GetServiceId() == "" || leg.GetMode() != p.GetMode() || !validTimes(leg.GetDepartureAt(), leg.GetArrivalAt()) {
			return false
		}
		if i == 0 {
			continue
		}
		previous := legs[i-1]
		code := previous.GetDestinationStation()
		if len(code) < 2 || code[0] != 's' || code != leg.GetOriginStation() {
			return false
		}
		if _, err := strconv.ParseUint(code[1:], 10, 64); err != nil {
			return false
		}
		arr, _ := time.Parse(time.RFC3339, previous.GetArrivalAt())
		dep, _ := time.Parse(time.RFC3339, leg.GetDepartureAt())
		wait := dep.Sub(arr)
		if wait < minWait || wait > maxWait {
			return false
		}
	}
	return true
}
func edgeFrom(p *pb.TransportPart) string {
	if p.GetOriginCity() != "" {
		return p.GetOriginCity()
	}
	return p.GetOrigin()
}
func edgeTo(p *pb.TransportPart) string {
	if p.GetDestinationCity() != "" {
		return p.GetDestinationCity()
	}
	return p.GetDestination()
}
func itineraryTotal(trip providerkit.TripRequest, s pathState) *int64 {
	if trip.Travelers.Adults < 1 || len(trip.Travelers.ChildrenAges) > 0 {
		return nil
	}
	var sum int64
	for _, leg := range s.legs {
		p := leg.GetPrice()
		if p == nil || p.GetCurrency() != trip.Currency || p.GetScope() != "per_person_one_way" || p.GetAmountMinor() < 0 || p.GetAmountMinor() > math.MaxInt64/int64(trip.Travelers.Adults) {
			return nil
		}
		n := p.GetAmountMinor() * int64(trip.Travelers.Adults)
		if sum > math.MaxInt64-n {
			return nil
		}
		sum += n
	}
	for _, hotel := range s.hotels {
		p := hotel.GetPrice()
		if p == nil || p.GetCurrency() != trip.Currency || p.GetScope() != "stay_all_rooms" || p.GetAmountMinor() < 0 || sum > math.MaxInt64-p.GetAmountMinor() {
			return nil
		}
		sum += p.GetAmountMinor()
	}
	return &sum
}
