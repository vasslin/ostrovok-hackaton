package service

import (
	"ai-service/internal/canonical"
	"bytes"
	"encoding/json"
	"math"
	"strings"
)

func cloneRequest(value canonical.TripRequest) canonical.TripRequest {
	data, _ := json.Marshal(value)
	var copy canonical.TripRequest
	_ = json.Unmarshal(data, &copy)
	return copy
}
func requestFromConfig(config PlanningConfig) (canonical.TripRequest, error) {
	if config.Version != "" && config.Version != "1.0" {
		return canonical.TripRequest{}, invalid("config.version must be 1.0")
	}
	if config.Request != nil && config.Trip != nil {
		return canonical.TripRequest{}, invalid("use request or trip, not both")
	}
	if config.Request != nil {
		return cloneRequest(*config.Request), nil
	}
	r := canonical.TripRequest{SchemaVersion: "1.0", Locale: "ru", Currency: "RUB"}
	if config.Version != "" && config.Version != "1.0" {
		return r, invalid("config.version must be 1.0")
	}
	t := config.Trip
	if t == nil {
		return r, nil
	}
	if t.Origin != nil {
		r.Origin.City = t.Origin.City
		if t.Origin.CountryCode != nil {
			r.Origin.CountryCode = *t.Origin.CountryCode
		}
	}
	if len(t.Destinations) > 1 {
		return r, invalid("canonical TripRequest supports one selected destination")
	}
	if len(t.Destinations) == 1 {
		r.Destination = &canonical.Destination{City: t.Destinations[0].City}
		if t.Destinations[0].CountryCode != nil {
			r.Destination.CountryCode = *t.Destinations[0].CountryCode
		}
	}
	if d := t.DateRange; d != nil {
		if d.StartDate != nil {
			r.Dates.DepartureDate = *d.StartDate
		}
		if d.EndDate != nil {
			r.Dates.ReturnDate = *d.EndDate
		}
		if d.FlexibleDates != nil && *d.FlexibleDates {
			return r, invalid("legacy flexible_dates has no day range; use config.request.dates.flexible_days")
		}
	}
	if v := t.Travelers; v != nil {
		if v.Adults != nil {
			r.Travelers.Adults = *v.Adults
		}
		for _, age := range v.ChildrenAges {
			r.Travelers.ChildrenAges = append(r.Travelers.ChildrenAges, canonical.Age(age))
		}
		if v.Children != nil && *v.Children != len(v.ChildrenAges) {
			r.AIContext = &canonical.AIContext{Notes: "Указаны дети, необходимо уточнить возраст каждого ребёнка."}
		}
	}
	if b := t.Budget; b != nil {
		if b.Currency != nil {
			r.Currency = *b.Currency
		}
		r.Budget = &canonical.Budget{}
		if b.Amount != nil {
			if *b.Amount >= float64(math.MaxInt) || *b.Amount <= float64(math.MinInt) || *b.Amount != float64(int(*b.Amount)) {
				return r, invalid("budget must use whole major units")
			}
			r.Budget.Total = int(*b.Amount)
		}
		if b.Scope != nil {
			r.Budget.Scope = *b.Scope
		}
		if b.IsHardLimit != nil {
			r.Budget.IsHard = *b.IsHardLimit
		}
	}
	if v := t.Transport; v != nil {
		r.Transport = &canonical.Transport{Modes: v.Modes, MaxTransfers: v.MaxTransfers}
		if v.MaxTransfers != nil && *v.MaxTransfers == 0 {
			r.Transport.DirectOnly = true
		}
		if v.MaxWaitHours != nil {
			return r, invalid("legacy max_wait_hours cannot be mapped losslessly; use config.request")
		}
	}
	if v := t.Accommodation; v != nil {
		r.Hotel = &canonical.Hotel{Needed: v.Required, StarRatingMin: v.MinStars, Kinds: v.PreferredTypes}
		if v.MaxPricePerNight != nil {
			r.Hotel.PricePerNightMax = int(*v.MaxPricePerNight)
		}
	}
	if v := t.Activities; v != nil {
		r.Leisure = &canonical.Leisure{Interests: v.Interests}
		if v.Pace != nil {
			appendRequestNote(&r, "Предпочтительный темп отдыха: "+*v.Pace)
		}
	}
	if v := t.Preferences; v != nil {
		if v.Language != nil {
			r.Locale = *v.Language
		}
		if v.AvoidOvernightTransfers != nil && *v.AvoidOvernightTransfers {
			if r.AIContext == nil {
				r.AIContext = &canonical.AIContext{}
			}
			r.AIContext.MustAvoid = append(r.AIContext.MustAvoid, "ночные пересадки")
		}
	}
	return r, nil
}
func appendRequestNote(r *canonical.TripRequest, note string) {
	if strings.TrimSpace(note) == "" {
		return
	}
	if r.AIContext == nil {
		r.AIContext = &canonical.AIContext{}
	}
	text := r.AIContext.Notes
	if text != "" {
		text += "; "
	}
	text += note
	runes := []rune(text)
	if len(runes) > 500 {
		runes = runes[len(runes)-500:]
	}
	r.AIContext.Notes = string(runes)
}
func requestMap(r canonical.TripRequest) map[string]any {
	data, _ := json.Marshal(r)
	out := map[string]any{}
	_ = json.Unmarshal(data, &out)
	return out
}
func mergeRequest(base canonical.TripRequest, patch map[string]any, cleared []string) (canonical.TripRequest, error) {
	root := requestMap(base)
	var merge func(map[string]any, map[string]any)
	merge = func(dst, src map[string]any) {
		for key, value := range src {
			if value == nil || value == "" {
				continue
			}
			if child, ok := value.(map[string]any); ok {
				prior, _ := dst[key].(map[string]any)
				if prior == nil {
					prior = map[string]any{}
				}
				merge(prior, child)
				dst[key] = prior
			} else {
				dst[key] = value
			}
		}
	}
	merge(root, patch)
	for _, field := range cleared {
		field = strings.TrimPrefix(field, "/")
		parts := strings.Split(strings.ReplaceAll(field, ".", "/"), "/")
		if len(parts) == 0 || parts[0] == "schema_version" || parts[0] == "locale" || parts[0] == "currency" || parts[0] == "completeness" {
			return base, invalid("unsupported cleared field")
		}
		current := root
		for _, key := range parts[:len(parts)-1] {
			child, _ := current[key].(map[string]any)
			current = child
			if current == nil {
				break
			}
		}
		if current != nil {
			delete(current, parts[len(parts)-1])
		}
	}
	data, err := json.Marshal(root)
	if err != nil {
		return base, err
	}
	var out canonical.TripRequest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&out); err != nil {
		return base, err
	}
	return out, nil
}

var requestPaths = map[string]string{
	"trip.origin": "/origin", "trip.origin.city": "/origin/city", "trip.origin.country_code": "/origin/country_code", "trip.destinations": "/destination/city", "trip.date_range": "/dates", "trip.date_range.start_date": "/dates/departure_date", "trip.date_range.end_date": "/dates/return_date", "trip.travelers.adults": "/travelers/adults", "trip.travelers.children_ages": "/travelers/children_ages", "trip.budget": "/budget", "trip.budget.amount": "/budget/total", "trip.budget.currency": "/currency", "trip.budget.scope": "/budget/scope", "trip.transport.max_transfers": "/transport/max_transfers", "trip.accommodation.min_stars": "/hotel/star_rating_min"}

func canonicalPath(field string) string {
	if value, ok := requestPaths[field]; ok {
		return value
	}
	if strings.HasPrefix(field, "/") {
		return field
	}
	return "/" + strings.ReplaceAll(field, ".", "/")
}
func legacyPath(field string) string {
	for legacy, path := range requestPaths {
		if path == field {
			return legacy
		}
	}
	return field
}
func canonicalLiteral(base canonical.TripRequest, in PlanningInput) (canonical.TripRequest, error) {
	mapped := in
	mapped.Questions = append([]PlanningQuestion(nil), in.Questions...)
	for i, q := range mapped.Questions {
		mapped.Questions[i].Field = legacyPath(q.Field)
	}
	legacy, err := applyLiteralPlanning(PlanningConfig{}, mapped)
	if err != nil {
		return base, err
	}
	literal, err := requestFromConfig(legacy)
	if err != nil {
		return base, err
	}
	patch := requestMap(literal)
	if legacy.Trip == nil || legacy.Trip.Travelers == nil || legacy.Trip.Travelers.Adults == nil {
		if guests, ok := patch["travelers"].(map[string]any); ok {
			delete(guests, "adults")
		}
	}
	delete(patch, "locale")
	delete(patch, "currency")
	delete(patch, "schema_version")
	if legacy.Trip != nil && legacy.Trip.Budget != nil && legacy.Trip.Budget.Currency != nil {
		patch["currency"] = literal.Currency
	}
	out, err := mergeRequest(base, patch, nil)
	if err != nil {
		return base, err
	}
	text := norm(in.Text)
	for _, a := range in.Answers {
		text += " " + norm(a.Text)
	}
	if hasAny(text, "с собакой", "собаку", "собака") {
		if len(out.Travelers.Pets) != 1 || out.Travelers.Pets[0].Kind != "dog" {
			out.Travelers.Pets = []canonical.Pet{{Kind: "dog"}}
		}
	}
	if hasAny(text, "с кошкой", "кошку", "котом") {
		if len(out.Travelers.Pets) != 1 || out.Travelers.Pets[0].Kind != "cat" {
			out.Travelers.Pets = []canonical.Pet{{Kind: "cat"}}
		}
	}
	if hasAny(text, "без питомцев", "без животных") {
		out.Travelers.Pets = nil
	}
	if hasAny(text, "без детей", "дети не едут") {
		out.Travelers.ChildrenAges = nil
	}
	if out.Budget != nil {
		if hasAny(text, "на человека", "с человека") {
			out.Budget.Scope = "per_person"
		}
		if hasAny(text, "за ночь", "на ночь", "в сутки") {
			out.Budget.Scope = "per_night"
		}
	}
	return extendedLiterals(out, in), nil
}
