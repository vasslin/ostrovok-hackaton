// Package join builds date-and-city skeletons from provider parts.
// One join is a city path with a single departure and arrival date on each leg.
// Every transport and hotel that fits that interval stays in the join as an alternative.
// Assemble returns at most k joins, preferring combinations that differ by city or date.
package join

import (
	"fmt"
	"sort"
	"strings"
	"time"

	providerv1 "github.com/godnota/pkg/generated/provider/v1"
)

const (
	KindTransport = "transport"
	KindHotel     = "hotel"

	maxCities = 6
	maxPaths  = 64
	maxCombos = 256
)

// Slot is one interval inside a join: a group of roads, or a group of hotels.
type Slot struct {
	Kind          string
	FromCity      string
	ToCity        string
	City          string
	DepartureDate string
	ArrivalDate   string
	// StayUntil is the next leg's departure date. Hotel checkout must be on or after it.
	// Empty for an open stay at the end of a one-way trip and for transport slots.
	StayUntil string
	Transport []*providerv1.TransportPart
	Hotels    []*providerv1.HotelPart
}

// Join is one compatible combination of cities and dates.
// RoundTripParts are Aviasales-style fares that occupy two transport slots and must be kept together.
type Join struct {
	ID             string
	Cities         []string
	Slots          []Slot
	RoundTripParts []string
}

type dateKey struct {
	dep string
	arr string
}

type leg struct {
	part     *providerv1.TransportPart
	from     string
	to       string
	fromName string
	toName   string
	dep      time.Time
	arr      time.Time
	depDate  string
	arrDate  string
}

type stay struct {
	part *providerv1.HotelPart
	city string
	name string
	in   string
	out  string
}

// Assemble groups parts into at most k joins.
// origin is the trip start city. An empty origin uses the city of the earliest departure.
func Assemble(parts *providerv1.SearchPartsResponse, origin string, k int) ([]Join, error) {
	if k < 1 {
		return nil, fmt.Errorf("k must be positive")
	}
	if parts == nil {
		return nil, nil
	}
	names := map[string]string{}
	originNorm := remember(names, origin, true)
	legs := transportLegs(parts.GetTransportParts(), names)
	hotels := hotelStays(parts.GetHotelParts(), names)
	if originNorm == "" {
		originNorm = earliestOrigin(legs)
	}
	if originNorm == "" || len(legs) == 0 {
		return nil, nil
	}
	edges := map[string]map[string][]leg{}
	for _, item := range legs {
		if edges[item.from] == nil {
			edges[item.from] = map[string][]leg{}
		}
		edges[item.from][item.to] = append(edges[item.from][item.to], item)
	}
	hotelCities := map[string]bool{}
	for _, item := range hotels {
		hotelCities[item.city] = true
	}
	requireHotel := len(hotels) > 0
	paths := cityPaths(originNorm, edges, hotelCities, requireHotel)
	var joins []Join
	for _, path := range paths {
		joins = append(joins, joinsForPath(path, names, edges, hotels, requireHotel)...)
	}
	if len(joins) == 0 {
		return nil, nil
	}
	return pick(joins, k), nil
}

func transportLegs(parts []*providerv1.TransportPart, names map[string]string) []leg {
	ordered := append([]*providerv1.TransportPart{}, parts...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].GetId() < ordered[j].GetId() })
	var legs []leg
	for _, part := range ordered {
		if part == nil || part.GetId() == "" {
			continue
		}
		fromName := endpoint(part.GetOriginCity(), part.GetOrigin())
		toName := endpoint(part.GetDestinationCity(), part.GetDestination())
		depDate, dep, okDep := stamp(part.GetDepartureAt())
		arrDate, arr, okArr := stamp(part.GetArrivalAt())
		if !okDep || !okArr || !arr.After(dep) {
			continue
		}
		from := remember(names, fromName, false)
		to := remember(names, toName, false)
		if from == "" || to == "" || from == to {
			continue
		}
		item := leg{part: part, from: from, to: to, fromName: names[from], toName: names[to], dep: dep, arr: arr, depDate: depDate, arrDate: arrDate}
		if part.GetReturnDepartureAt() != "" || part.GetReturnArrivalAt() != "" {
			retDepDate, retDep, okRetDep := stamp(part.GetReturnDepartureAt())
			retArrDate, retArr, okRetArr := stamp(part.GetReturnArrivalAt())
			if !okRetDep || !okRetArr || !retDep.After(arr) || !retArr.After(retDep) {
				continue
			}
			legs = append(legs, item)
			legs = append(legs, leg{
				part: part, from: to, to: from, fromName: names[to], toName: names[from],
				dep: retDep, arr: retArr, depDate: retDepDate, arrDate: retArrDate,
			})
			continue
		}
		legs = append(legs, item)
	}
	return legs
}

func hotelStays(parts []*providerv1.HotelPart, names map[string]string) []stay {
	ordered := append([]*providerv1.HotelPart{}, parts...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].GetId() < ordered[j].GetId() })
	var stays []stay
	for _, part := range ordered {
		if part == nil || part.GetId() == "" {
			continue
		}
		in := strings.TrimSpace(part.GetCheckIn())
		out := strings.TrimSpace(part.GetCheckOut())
		if !isoDate(in) || !isoDate(out) || out < in {
			continue
		}
		city := remember(names, part.GetCity(), false)
		if city == "" {
			continue
		}
		stays = append(stays, stay{part: part, city: city, name: names[city], in: in, out: out})
	}
	return stays
}

func earliestOrigin(legs []leg) string {
	if len(legs) == 0 {
		return ""
	}
	best := legs[0]
	for _, item := range legs[1:] {
		if item.dep.Before(best.dep) || (item.dep.Equal(best.dep) && item.from < best.from) {
			best = item
		}
	}
	return best.from
}

func cityPaths(origin string, edges map[string]map[string][]leg, hotelCities map[string]bool, requireHotel bool) [][]string {
	seen := map[string]bool{origin: true}
	found := map[string][]string{}
	var walk func([]string)
	walk = func(path []string) {
		if len(found) >= maxPaths {
			return
		}
		here := path[len(path)-1]
		extended := false
		for _, to := range sortedKeys(edges[here]) {
			if to == origin {
				if len(path) >= 2 {
					cycle := append(append([]string{}, path...), origin)
					if pathOK(cycle, true, hotelCities, requireHotel) {
						found[strings.Join(cycle, ">")] = cycle
					}
				}
				extended = true
				continue
			}
			if seen[to] || len(path) >= maxCities {
				continue
			}
			extended = true
			seen[to] = true
			next := make([]string, len(path)+1)
			copy(next, path)
			next[len(path)] = to
			walk(next)
			delete(seen, to)
			if len(found) >= maxPaths {
				return
			}
		}
		if !extended && len(path) >= 2 && pathOK(path, false, hotelCities, requireHotel) {
			found[strings.Join(path, ">")] = append([]string{}, path...)
		}
	}
	walk([]string{origin})
	keys := make([]string, 0, len(found))
	for key := range found {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	paths := make([][]string, 0, len(keys))
	for _, key := range keys {
		paths = append(paths, found[key])
	}
	return paths
}

func pathOK(path []string, cycle bool, hotelCities map[string]bool, requireHotel bool) bool {
	if len(path) < 2 {
		return false
	}
	if !requireHotel {
		return true
	}
	stays := path[1:]
	if cycle {
		if len(path) < 3 {
			return false
		}
		stays = path[1 : len(path)-1]
	}
	if len(stays) == 0 {
		return false
	}
	for _, city := range stays {
		if !hotelCities[city] {
			return false
		}
	}
	return true
}

func joinsForPath(path []string, names map[string]string, edges map[string]map[string][]leg, hotels []stay, requireHotel bool) []Join {
	type hop struct {
		keys  []dateKey
		byKey map[dateKey][]leg
	}
	hops := make([]hop, 0, len(path)-1)
	for i := 0; i < len(path)-1; i++ {
		grouped := map[dateKey][]leg{}
		for _, item := range edges[path[i]][path[i+1]] {
			key := dateKey{item.depDate, item.arrDate}
			grouped[key] = append(grouped[key], item)
		}
		if len(grouped) == 0 {
			return nil
		}
		keys := make([]dateKey, 0, len(grouped))
		for key := range grouped {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(a, b int) bool {
			if keys[a].dep != keys[b].dep {
				return keys[a].dep < keys[b].dep
			}
			return keys[a].arr < keys[b].arr
		})
		hops = append(hops, hop{keys: keys, byKey: grouped})
	}
	keySets := make([][]dateKey, len(hops))
	for i, item := range hops {
		keySets[i] = item.keys
	}
	var joins []Join
	for _, combo := range combinations(keySets) {
		groups := make([][]leg, len(combo))
		for i, key := range combo {
			groups[i] = append([]leg{}, hops[i].byKey[key]...)
		}
		built, ok := buildJoin(path, names, combo, groups, hotels, requireHotel)
		if ok {
			joins = append(joins, built)
		}
	}
	return joins
}

func combinations(keys [][]dateKey) [][]dateKey {
	var out [][]dateKey
	var walk func(int, []dateKey)
	walk = func(i int, cur []dateKey) {
		if len(out) >= maxCombos {
			return
		}
		if i == len(keys) {
			out = append(out, append([]dateKey{}, cur...))
			return
		}
		for _, key := range keys[i] {
			walk(i+1, append(cur, key))
			if len(out) >= maxCombos {
				return
			}
		}
	}
	walk(0, nil)
	return out
}

func buildJoin(path []string, names map[string]string, keys []dateKey, groups [][]leg, hotels []stay, requireHotel bool) (Join, bool) {
	groups = connected(anchorRoundTrips(groups))
	for _, group := range groups {
		if len(group) == 0 {
			return Join{}, false
		}
	}
	show := make([]string, len(path))
	for i, city := range path {
		show[i] = names[city]
	}
	cycle := path[0] == path[len(path)-1] && len(path) > 2
	slots := make([]Slot, 0, len(groups)*2)
	for i, group := range groups {
		slots = append(slots, Slot{
			Kind:          KindTransport,
			FromCity:      names[path[i]],
			ToCity:        names[path[i+1]],
			DepartureDate: keys[i].dep,
			ArrivalDate:   keys[i].arr,
			Transport:     dedupeTransport(group),
		})
		last := i == len(groups)-1
		if last && cycle {
			continue
		}
		stayCity := path[i+1]
		until := ""
		if !last {
			until = keys[i+1].dep
		}
		matched := matchHotels(hotels, stayCity, keys[i].arr, until)
		if len(matched) == 0 {
			if requireHotel {
				return Join{}, false
			}
			continue
		}
		slots = append(slots, Slot{
			Kind:        KindHotel,
			City:        names[stayCity],
			ArrivalDate: keys[i].arr,
			StayUntil:   until,
			Hotels:      matched,
		})
	}
	idParts := make([]string, len(keys))
	for i, key := range keys {
		idParts[i] = key.dep + "/" + key.arr
	}
	return Join{
		ID:             strings.Join(show, ">") + "|" + strings.Join(idParts, "|"),
		Cities:         show,
		Slots:          slots,
		RoundTripParts: sharedParts(slots),
	}, true
}

// anchorRoundTrips drops a round-trip fare unless both of its halves sit in this date combination.
func anchorRoundTrips(groups [][]leg) [][]leg {
	hops := map[string]int{}
	for _, group := range groups {
		seen := map[string]bool{}
		for _, item := range group {
			id := item.part.GetId()
			if item.part.GetReturnDepartureAt() == "" || seen[id] {
				continue
			}
			seen[id] = true
			hops[id]++
		}
	}
	out := make([][]leg, len(groups))
	for i, group := range groups {
		for _, item := range group {
			id := item.part.GetId()
			if item.part.GetReturnDepartureAt() != "" && hops[id] < 2 {
				continue
			}
			out[i] = append(out[i], item)
		}
	}
	return out
}

// connected keeps options that belong to at least one chain whose times increase.
func connected(groups [][]leg) [][]leg {
	if len(groups) == 0 {
		return groups
	}
	reach := make([][]bool, len(groups))
	for i := range groups {
		reach[i] = make([]bool, len(groups[i]))
	}
	for i := range groups[0] {
		reach[0][i] = true
	}
	for i := 1; i < len(groups); i++ {
		for j, item := range groups[i] {
			for p, prev := range groups[i-1] {
				if reach[i-1][p] && prev.arr.Before(item.dep) {
					reach[i][j] = true
					break
				}
			}
		}
	}
	keep := make([][]bool, len(groups))
	for i := range groups {
		keep[i] = make([]bool, len(groups[i]))
	}
	last := len(groups) - 1
	for i := range groups[last] {
		keep[last][i] = reach[last][i]
	}
	for i := last - 1; i >= 0; i-- {
		for j, item := range groups[i] {
			if !reach[i][j] {
				continue
			}
			for q, next := range groups[i+1] {
				if keep[i+1][q] && item.arr.Before(next.dep) {
					keep[i][j] = true
					break
				}
			}
		}
	}
	out := make([][]leg, len(groups))
	for i, group := range groups {
		for j, item := range group {
			if keep[i][j] {
				out[i] = append(out[i], item)
			}
		}
	}
	return out
}

func matchHotels(hotels []stay, city, arrival, until string) []*providerv1.HotelPart {
	var matched []stay
	seen := map[string]bool{}
	for _, item := range hotels {
		if item.city != city || seen[item.part.GetId()] || item.in > arrival {
			continue
		}
		if until != "" {
			if item.out < until {
				continue
			}
		} else if item.out <= arrival {
			continue
		}
		seen[item.part.GetId()] = true
		matched = append(matched, item)
	}
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].in != matched[j].in {
			return matched[i].in < matched[j].in
		}
		return matched[i].part.GetId() < matched[j].part.GetId()
	})
	out := make([]*providerv1.HotelPart, len(matched))
	for i, item := range matched {
		out[i] = item.part
	}
	return out
}

func dedupeTransport(group []leg) []*providerv1.TransportPart {
	sort.Slice(group, func(i, j int) bool {
		if !group[i].dep.Equal(group[j].dep) {
			return group[i].dep.Before(group[j].dep)
		}
		return group[i].part.GetId() < group[j].part.GetId()
	})
	var out []*providerv1.TransportPart
	seen := map[string]bool{}
	for _, item := range group {
		if seen[item.part.GetId()] {
			continue
		}
		seen[item.part.GetId()] = true
		out = append(out, item.part)
	}
	return out
}

func sharedParts(slots []Slot) []string {
	count := map[string]int{}
	for _, slot := range slots {
		seen := map[string]bool{}
		for _, part := range slot.Transport {
			if seen[part.GetId()] {
				continue
			}
			seen[part.GetId()] = true
			count[part.GetId()]++
		}
	}
	var ids []string
	for id, n := range count {
		if n > 1 {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

func pick(joins []Join, k int) []Join {
	if k >= len(joins) {
		sort.Slice(joins, func(i, j int) bool { return joins[i].ID < joins[j].ID })
		return joins
	}
	rest := append([]Join{}, joins...)
	sort.Slice(rest, func(i, j int) bool {
		if len(rest[i].Cities) != len(rest[j].Cities) {
			return len(rest[i].Cities) > len(rest[j].Cities)
		}
		return rest[i].ID < rest[j].ID
	})
	picked := []Join{rest[0]}
	rest = rest[1:]
	for len(picked) < k && len(rest) > 0 {
		best := 0
		bestScore := -1
		for i, cand := range rest {
			score := distance(cand, picked[0])
			for _, have := range picked[1:] {
				if d := distance(cand, have); d < score {
					score = d
				}
			}
			if score > bestScore || (score == bestScore && cand.ID < rest[best].ID) {
				best = i
				bestScore = score
			}
		}
		picked = append(picked, rest[best])
		rest = append(rest[:best], rest[best+1:]...)
	}
	sort.Slice(picked, func(i, j int) bool { return picked[i].ID < picked[j].ID })
	return picked
}

func distance(a, b Join) int {
	score := 0
	same := len(a.Cities) == len(b.Cities)
	n := len(a.Cities)
	if len(b.Cities) > n {
		n = len(b.Cities)
	}
	for i := 0; i < n; i++ {
		left, right := "", ""
		if i < len(a.Cities) {
			left = norm(a.Cities[i])
		}
		if i < len(b.Cities) {
			right = norm(b.Cities[i])
		}
		if left != right {
			same = false
			score += 10
		}
	}
	if !same {
		score += 100
	}
	leftDates, rightDates := dateSig(a), dateSig(b)
	m := len(leftDates)
	if len(rightDates) > m {
		m = len(rightDates)
	}
	for i := 0; i < m; i++ {
		var left, right string
		if i < len(leftDates) {
			left = leftDates[i]
		}
		if i < len(rightDates) {
			right = rightDates[i]
		}
		if left != right {
			score++
		}
	}
	return score
}

func dateSig(j Join) []string {
	var sig []string
	for _, slot := range j.Slots {
		if slot.Kind == KindTransport {
			sig = append(sig, slot.DepartureDate+"/"+slot.ArrivalDate)
		}
	}
	return sig
}

func endpoint(city, code string) string {
	if strings.TrimSpace(city) != "" {
		return city
	}
	return code
}

func remember(names map[string]string, city string, force bool) string {
	city = strings.TrimSpace(city)
	key := norm(city)
	if key == "" {
		return ""
	}
	if force || names[key] == "" {
		names[key] = city
	}
	return key
}

func norm(city string) string {
	return strings.ToLower(strings.TrimSpace(city))
}

func stamp(value string) (string, time.Time, bool) {
	for _, layout := range []string{time.RFC3339, time.RFC3339Nano} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.Format(time.DateOnly), parsed, true
		}
	}
	return "", time.Time{}, false
}

func isoDate(value string) bool {
	_, err := time.Parse(time.DateOnly, value)
	return err == nil
}

func sortedKeys(values map[string][]leg) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
