package service

import (
	"strings"
	"time"
	"unicode/utf8"
)

func planningIssues(c PlanningConfig, today time.Time) []PlanningQuestion {
	issues := []PlanningQuestion{}
	seen := map[string]int{}
	add := func(field, text string, required bool, kind string) {
		if index, ok := seen[field]; ok {
			issues[index].Required = issues[index].Required || required
			return
		}
		seen[field] = len(issues)
		issues = append(issues, PlanningQuestion{ID: strings.ReplaceAll(field, ".", "_"), Field: field, Text: text, Required: required, InputType: kind})
	}
	trip := c.Trip
	if trip == nil {
		trip = &PlanningTrip{}
	}
	if trip.Origin == nil || strings.TrimSpace(trip.Origin.City) == "" {
		add("trip.origin.city", "Из какого города начнётся поездка?", true, "text")
	}
	if !hasDestination(c) {
		add("trip.destinations", "В какой город хотите поехать или какую идею выбираете?", true, "text")
	}
	dates := trip.DateRange
	if dates == nil {
		dates = &PlanningDates{}
	}
	start, end := time.Time{}, time.Time{}
	if dates.StartDate == nil || *dates.StartDate == "" {
		add("trip.date_range.start_date", "Когда хотите отправиться? Укажите дату начала поездки.", true, "date")
	} else {
		var err error
		start, err = time.Parse("2006-01-02", *dates.StartDate)
		if err != nil || start.Before(today) {
			add("trip.date_range.start_date", "Уточните дату отправления: нужна корректная дата не раньше текущей.", true, "date")
		}
	}
	if dates.EndDate == nil || *dates.EndDate == "" {
		add("trip.date_range.end_date", "Когда планируете вернуться?", true, "date")
	} else {
		var err error
		end, err = time.Parse("2006-01-02", *dates.EndDate)
		if err != nil || end.Before(today) || (!start.IsZero() && !end.After(start)) {
			add("trip.date_range.end_date", "Уточните дату возвращения: она должна быть позже отправления.", true, "date")
		}
	}
	travelers := trip.Travelers
	if travelers == nil {
		travelers = &PlanningTravelers{}
	}
	if travelers.Adults == nil || *travelers.Adults < 1 || *travelers.Adults > 9 {
		add("trip.travelers.adults", "Сколько взрослых поедет? Допустимо от 1 до 9.", true, "number")
	}
	badChildren := len(travelers.ChildrenAges) > 8
	for _, age := range travelers.ChildrenAges {
		if age < 0 || age > 17 {
			badChildren = true
		}
	}
	if travelers.Children != nil && (*travelers.Children < 0 || *travelers.Children > 8 || *travelers.Children != len(travelers.ChildrenAges)) {
		badChildren = true
	}
	if badChildren {
		add("trip.travelers.children_ages", "Укажите возраст каждого ребёнка в полных годах (0–17); число возрастов должно совпасть с числом детей.", true, "text")
	}
	checkPlace := func(p PlanningPlace, field string) {
		if utf8.RuneCountInString(p.City) > 100 || (p.CountryCode != nil && (len(*p.CountryCode) != 2 || strings.Trim(*p.CountryCode, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "")) {
			add(field, "Уточните город и двухбуквенный код страны.", false, "text")
		}
	}
	if trip.Origin != nil {
		checkPlace(*trip.Origin, "trip.origin")
	}
	if len(trip.Destinations) > 5 {
		add("trip.destinations", "Оставьте не более пяти городов назначения.", false, "text")
	}
	for _, p := range trip.Destinations {
		if strings.TrimSpace(p.City) == "" {
			add("trip.destinations", "Укажите название каждого города назначения.", true, "text")
			break
		}
		checkPlace(p, "trip.destinations")
	}
	if b := trip.Budget; b != nil {
		if b.Amount != nil && *b.Amount <= 0 {
			add("trip.budget.amount", "Какой положительный бюджет закладываем?", false, "number")
		}
		if b.Currency != nil && (len(*b.Currency) != 3 || strings.Trim(*b.Currency, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "") {
			add("trip.budget.currency", "Уточните валюту бюджета, например RUB, USD или EUR.", false, "text")
		}
		if b.Amount != nil && b.Currency == nil {
			add("trip.budget.currency", "В какой валюте указан бюджет?", true, "text")
		}
		if b.Scope != nil && *b.Scope != "total" && *b.Scope != "per_person" && *b.Scope != "per_night" {
			add("trip.budget.scope", "Бюджет на всю поездку, на человека или на ночь?", false, "text")
		}
	}
	if t := trip.Transport; t != nil {
		for _, mode := range t.Modes {
			if !containsString([]string{"plane", "train", "suburban", "bus", "water", "helicopter"}, mode) {
				add("trip.transport.modes", "Уточните допустимые виды транспорта.", false, "text")
				break
			}
		}
		if t.MaxTransfers != nil && (*t.MaxTransfers < 0 || *t.MaxTransfers > 3) {
			add("trip.transport.max_transfers", "Сколько пересадок допустимо (0–3)?", false, "number")
		}
		if t.MaxWaitHours != nil && (*t.MaxWaitHours < 0 || *t.MaxWaitHours > 48) {
			add("trip.transport.max_wait_hours", "Какое ожидание допустимо (0–48 часов)?", false, "number")
		}
	}
	if a := trip.Accommodation; a != nil {
		if a.MinStars != nil && (*a.MinStars < 0 || *a.MinStars > 5) {
			add("trip.accommodation.min_stars", "Уточните категорию отеля от 0 до 5 звёзд.", false, "number")
		}
		if a.MaxPricePerNight != nil && *a.MaxPricePerNight <= 0 {
			add("trip.accommodation.max_price_per_night", "Уточните положительную стоимость ночи.", false, "number")
		}
	}
	return issues
}
func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
func planningQuestions(issues, proposed []PlanningQuestion) []PlanningQuestion {
	out := []PlanningQuestion{}
	covered := map[string]bool{}
	for _, q := range proposed {
		if len(out) == 2 {
			break
		}
		if strings.TrimSpace(q.Text) == "" || utf8.RuneCountInString(q.Text) > 500 {
			continue
		}
		matches := []PlanningQuestion{}
		for _, issue := range issues {
			if issue.Field == q.Field || strings.HasPrefix(issue.Field, q.Field+".") {
				matches = append(matches, issue)
			}
		}
		if len(matches) == 0 || q.Field == "" || q.Field == "trip" {
			continue
		}
		duplicate := false
		required := false
		for _, match := range matches {
			if covered[match.Field] {
				duplicate = true
			}
			required = required || match.Required
		}
		if duplicate {
			continue
		}
		hasRequired := false
		for _, issue := range issues {
			hasRequired = hasRequired || issue.Required
		}
		if hasRequired && !required {
			continue
		}
		for _, match := range matches {
			covered[match.Field] = true
		}
		q.ID = strings.ReplaceAll(q.Field, ".", "_")
		q.Required = required
		if !containsString([]string{"text", "date", "number"}, q.InputType) {
			q.InputType = "text"
		}
		out = append(out, q)
	}
	for _, issue := range issues {
		if len(out) == 2 {
			break
		}
		if covered[issue.Field] {
			continue
		}
		out = append(out, issue)
		covered[issue.Field] = true
	}
	return out
}
