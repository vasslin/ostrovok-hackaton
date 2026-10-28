package service

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"ai-service/internal/canonical"
)

var literalOrigin = regexp.MustCompile(`(?i)из\s+([а-яёa-z-]+)`)
var literalDestination = regexp.MustCompile(`[вВ]\s+([А-ЯЁ][а-яёА-ЯЁ-]+)`)
var literalRange = regexp.MustCompile(`(?i)с\s+(\d{1,2})\s+(?:по|до|[-–—])\s+(\d{1,2})\s+(января|февраля|марта|апреля|мая|июня|июля|августа|сентября|октября|ноября|декабря)\s+(\d{4})`)
var literalISORange = regexp.MustCompile(`(?i)с\s+(\d{4}-\d{2}-\d{2})\s+(?:по|до|[-–—])\s+(\d{4}-\d{2}-\d{2})`)
var literalAdults = regexp.MustCompile(`(?i)(\d+|один|одна|двое|два|две|трое|три|четверо|четыре|пять|шесть|семь|восемь|девять)\s+взросл`)
var literalChildAge = regexp.MustCompile(`(?i)ребен(?:ок|ком|ку|ка)[ ,]*(\d{1,2})\s*(?:лет|год)`)
var literalChildCount = regexp.MustCompile(`(?i)(\d+|двумя|двое|два|тремя|трое|три)\s+(?:дет|ребен)`)
var literalMoney = regexp.MustCompile(`(?i)(\d[\d \x{00A0}]*(?:[.,]\d+)?)\s*(тыс(?:яч[аиу]?)?\.?)?\s*(руб(?:лей|ля|ль)?|₽|доллар(?:ов|а)?|USD|RUB|EUR|евро|\$|€)`)

func cityLiteral(city string) string {
	aliases := map[string]string{"москвы": "Москва", "москва": "Москва", "москве": "Москва", "казани": "Казань", "казань": "Казань", "сочи": "Сочи", "петербурга": "Санкт-Петербург", "петербург": "Санкт-Петербург", "санкт-петербурга": "Санкт-Петербург", "краснодара": "Краснодар"}
	if name, ok := aliases[norm(city)]; ok {
		return name
	}
	runes := []rune(city)
	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}
	return string(runes)
}
func literalPlanningPatch(text string) PlanningConfig {
	trip := &PlanningTrip{}
	if match := literalOrigin.FindStringSubmatch(text); len(match) > 0 {
		if literalCityAllowed(match[1]) {
			trip.Origin = &PlanningPlace{City: cityLiteral(match[1])}
		}
	}
	if match := literalDestination.FindStringSubmatch(text); len(match) > 0 {
		if literalCityAllowed(match[1]) {
			trip.Destinations = []PlanningPlace{{City: cityLiteral(match[1])}}
		}
	}
	if match := literalRange.FindStringSubmatch(text); len(match) > 0 {
		monthNames := []string{"января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря"}
		month := 1
		for i, name := range monthNames {
			if norm(match[3]) == name {
				month = i + 1
			}
		}
		year, _ := strconv.Atoi(match[4])
		start, _ := strconv.Atoi(match[1])
		end, _ := strconv.Atoi(match[2])
		date := func(day int) string { return strconv.Itoa(year) + "-" + twoDigits(month) + "-" + twoDigits(day) }
		a, b := date(start), date(end)
		trip.DateRange = &PlanningDates{StartDate: &a, EndDate: &b}
	}
	if match := literalISORange.FindStringSubmatch(text); len(match) > 0 {
		trip.DateRange = &PlanningDates{StartDate: &match[1], EndDate: &match[2]}
	}
	if match := literalAdults.FindStringSubmatch(text); len(match) > 0 {
		number, err := strconv.Atoi(match[1])
		if err != nil {
			words := map[string]int{"один": 1, "одна": 1, "двое": 2, "два": 2, "две": 2, "трое": 3, "три": 3, "четверо": 4, "четыре": 4, "пять": 5, "шесть": 6, "семь": 7, "восемь": 8, "девять": 9}
			number = words[norm(match[1])]
		}
		trip.Travelers = &PlanningTravelers{Adults: &number}
	}
	if hasAny(norm(text), "бюджет", "на всю поездку") {
		if match := literalMoney.FindStringSubmatch(text); len(match) > 0 {
			raw := strings.ReplaceAll(strings.ReplaceAll(match[1], " ", ""), "\u00a0", "")
			amount, err := strconv.ParseFloat(strings.ReplaceAll(raw, ",", "."), 64)
			if err == nil {
				if match[2] != "" {
					amount *= 1000
				}
				currency := "RUB"
				if hasAny(norm(match[3]), "доллар", "usd", "$") {
					currency = "USD"
				}
				if hasAny(norm(match[3]), "евро", "eur", "€") {
					currency = "EUR"
				}
				trip.Budget = &PlanningBudget{Amount: &amount, Currency: &currency}
				if hasAny(norm(text), "всю поездку", "всей поездки", "общий бюджет") {
					scope := "total"
					trip.Budget.Scope = &scope
				}
			}
		}
	}
	textNorm := norm(text)
	if hasAny(textNorm, "без пересадок", "прямым рейсом", "прямой рейс") {
		zero := 0
		trip.Transport = &PlanningTransport{MaxTransfers: &zero}
	}
	if hasAny(textNorm, "без ночных", "избегать ночных", "не хочу ночных") {
		yes := true
		trip.Preferences = &PlanningPreferences{AvoidOvernightTransfers: &yes}
	}
	childCount := -1
	if hasAny(textNorm, "с ребенком", "один ребенок", "одного ребенка") {
		childCount = 1
	}
	if hasAny(textNorm, "без детей", "дети не едут") {
		childCount = 0
	}
	if match := literalChildCount.FindStringSubmatch(textNorm); len(match) > 0 {
		n, err := strconv.Atoi(match[1])
		if err != nil {
			n = map[string]int{"двумя": 2, "двое": 2, "два": 2, "тремя": 3, "трое": 3, "три": 3}[match[1]]
		}
		childCount = n
	}
	var ages []int
	if match := literalChildAge.FindStringSubmatch(textNorm); len(match) > 0 {
		age, _ := strconv.Atoi(match[1])
		ages = []int{age}
		if childCount < 0 {
			childCount = 1
		}
	}
	if childCount >= 0 {
		if trip.Travelers == nil {
			trip.Travelers = &PlanningTravelers{}
		}
		trip.Travelers.Children = &childCount
		trip.Travelers.ChildrenAges = ages
	}
	return PlanningConfig{Version: "1.0", Trip: trip}
}
func literalCityAllowed(value string) bool {
	if containsString([]string{"этой", "этого", "этом", "любой", "любом", "январе", "феврале", "марте", "апреле", "мае", "июне", "июле", "августе", "сентябре", "октябре", "ноябре", "декабре", "новый", "новом", "новую"}, norm(value)) {
		return false
	}
	runes := []rune(value)
	return len(runes) > 1 && (unicode.IsUpper(runes[0]) || containsString([]string{"москва", "москвы", "москве", "казань", "казани", "сочи", "петербург", "петербурга", "санкт-петербург", "санкт-петербурга", "краснодар", "краснодара"}, norm(value)))
}
func twoDigits(n int) string {
	value := strconv.Itoa(n)
	if n >= 0 && n < 10 {
		return "0" + value
	}
	return value
}
func applyLiteralPlanning(base PlanningConfig, in PlanningInput) (PlanningConfig, error) {
	questions := map[string]string{}
	for _, q := range in.Questions {
		questions[q.ID] = q.Field
	}
	for _, answer := range in.Answers {
		patch := literalPlanningPatch(answer.Text)
		text := strings.TrimSpace(answer.Text)
		if field := questions[answer.QuestionID]; field == "trip.travelers.adults" {
			if n, err := strconv.Atoi(text); err == nil {
				patch.Trip.Travelers = &PlanningTravelers{Adults: &n}
			}
		}
		if _, err := time.Parse("2006-01-02", text); err == nil {
			if field := questions[answer.QuestionID]; field == "trip.date_range.start_date" {
				patch.Trip.DateRange = &PlanningDates{StartDate: &text}
			} else if field == "trip.date_range.end_date" {
				patch.Trip.DateRange = &PlanningDates{EndDate: &text}
			}
		}
		var err error
		base, err = mergePlanning(base, patch, nil)
		if err != nil {
			return base, err
		}
	}
	base, err := mergePlanning(base, literalPlanningPatch(in.Text), nil)
	if err != nil {
		return base, err
	}
	return mergeTypoDates(base, collectedText(in), in.CurrentDate)
}

func collectedText(in PlanningInput) string {
	parts := []string{in.Text}
	for _, answer := range in.Answers {
		parts = append(parts, answer.Text)
	}
	return strings.Join(parts, "\n")
}

func mergeTypoDates(base PlanningConfig, text, current string) (PlanningConfig, error) {
	start, end := typoDates(text, current)
	if start == "" && end == "" {
		return base, nil
	}
	existingStart, existingEnd := "", ""
	if base.Trip != nil && base.Trip.DateRange != nil {
		if base.Trip.DateRange.StartDate != nil {
			existingStart = *base.Trip.DateRange.StartDate
		}
		if base.Trip.DateRange.EndDate != nil {
			existingEnd = *base.Trip.DateRange.EndDate
		}
	}
	if start != "" && end == "" && existingStart != "" && existingStart != start && existingEnd == "" {
		end = start
		start = ""
	}
	dates := &PlanningDates{}
	if start != "" && existingStart == "" {
		dates.StartDate = &start
	}
	if end != "" && existingEnd == "" {
		dates.EndDate = &end
	}
	if dates.StartDate == nil && dates.EndDate == nil {
		return base, nil
	}
	return mergePlanning(base, PlanningConfig{Version: "1.0", Trip: &PlanningTrip{DateRange: dates}}, nil)
}

func shiftReplyDate(base, literal canonical.TripRequest, text, current string) canonical.TripRequest {
	start, end := typoDates(text, current)
	if start != "" && end == "" && base.Dates.DepartureDate != "" && base.Dates.DepartureDate != start && base.Dates.ReturnDate == "" && literal.Dates.ReturnDate == "" {
		literal.Dates.ReturnDate = start
		if literal.Dates.DepartureDate == start {
			literal.Dates.DepartureDate = ""
		}
	}
	return literal
}

var dayMonthPattern = regexp.MustCompile(`(\d{1,2})\s+([а-яе]{3,})`)

func typoDates(text, current string) (string, string) {
	found := []string{}
	for _, match := range dayMonthPattern.FindAllStringSubmatch(norm(text), -1) {
		day, err := strconv.Atoi(match[1])
		month, ok := fuzzyMonth(match[2])
		if err != nil || !ok {
			continue
		}
		iso, ok := isoFromDayMonth(day, month, current)
		if !ok {
			continue
		}
		found = append(found, iso)
	}
	if len(found) == 0 {
		return "", ""
	}
	if len(found) >= 2 {
		return found[0], found[1]
	}
	if hasAny(norm(text), "возвращ", "обратно", "окончан") {
		return "", found[0]
	}
	return found[0], ""
}

func isoFromDayMonth(day, month int, current string) (string, bool) {
	if day < 1 || day > 31 || month < 1 || month > 12 {
		return "", false
	}
	today, err := time.Parse("2006-01-02", current)
	if err != nil {
		return "", false
	}
	candidate := time.Date(today.Year(), time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if candidate.Day() != day {
		return "", false
	}
	if candidate.Before(today) {
		candidate = time.Date(today.Year()+1, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		if candidate.Day() != day {
			return "", false
		}
	}
	return candidate.Format("2006-01-02"), true
}

var replyCityAliases = []struct {
	phrase string
	city   string
}{
	{"санкт-петербург", "Санкт-Петербург"},
	{"северная столица", "Санкт-Петербург"},
	{"ленинград", "Санкт-Петербург"},
	{"петербург", "Санкт-Петербург"},
	{"питер", "Санкт-Петербург"},
	{"спб", "Санкт-Петербург"},
	{"москва", "Москва"},
	{"мск", "Москва"},
}

func replyCities(text string) []string {
	runes := []rune(" " + norm(text) + " ")
	used := make([]bool, len(runes))
	type found struct {
		pos  int
		city string
	}
	founds := []found{}
	for _, alias := range replyCityAliases {
		phrase := []rune(alias.phrase)
		for i := 1; i+len(phrase) < len(runes); i++ {
			if !matchPhrase(runes, used, i, phrase) {
				continue
			}
			for j := i; j < i+len(phrase); j++ {
				used[j] = true
			}
			founds = append(founds, found{pos: i, city: alias.city})
		}
	}
	for i := 1; i < len(founds); i++ {
		for j := i; j > 0 && founds[j].pos < founds[j-1].pos; j-- {
			founds[j], founds[j-1] = founds[j-1], founds[j]
		}
	}
	cities := []string{}
	seen := map[string]bool{}
	for _, item := range founds {
		if seen[item.city] {
			continue
		}
		seen[item.city] = true
		cities = append(cities, item.city)
	}
	return cities
}

func matchPhrase(runes []rune, used []bool, start int, phrase []rune) bool {
	if unicode.IsLetter(runes[start-1]) || unicode.IsLetter(runes[start+len(phrase)]) {
		return false
	}
	for j := 0; j < len(phrase); j++ {
		if used[start+j] || runes[start+j] != phrase[j] {
			return false
		}
	}
	return true
}

func fillMissingFromReply(base canonical.TripRequest, text, current string) canonical.TripRequest {
	cities := replyCities(text)
	if strings.TrimSpace(base.Origin.City) == "" && len(cities) > 0 {
		base.Origin.City = cities[0]
		cities = cities[1:]
	}
	if (base.Destination == nil || strings.TrimSpace(base.Destination.City) == "") && len(cities) > 0 {
		base.Destination = &canonical.Destination{City: cities[0]}
	}
	start, end := typoDates(text, current)
	if start != "" && end == "" && base.Dates.DepartureDate != "" && base.Dates.DepartureDate != start && base.Dates.ReturnDate == "" {
		end = start
		start = ""
	}
	if base.Dates.DepartureDate == "" && start != "" {
		base.Dates.DepartureDate = start
	}
	if base.Dates.ReturnDate == "" && end != "" {
		base.Dates.ReturnDate = end
	}
	return base
}

var monthForms = [][]string{
	{"январь", "января", "январ"},
	{"февраль", "февраля", "феврал"},
	{"март", "марта"},
	{"апрель", "апреля", "апрел"},
	{"май", "мая"},
	{"июнь", "июня", "июн"},
	{"июль", "июля", "июл"},
	{"август", "августа"},
	{"сентябрь", "сентября", "сентябр"},
	{"октябрь", "октября", "октябр"},
	{"ноябрь", "ноября", "ноябр"},
	{"декабрь", "декабря", "декабр"},
}

func fuzzyMonth(word string) (int, bool) {
	word = norm(word)
	length := len([]rune(word))
	if length < 3 || length > 12 {
		return 0, false
	}
	bestByMonth := make([]int, len(monthForms))
	for i := range bestByMonth {
		bestByMonth[i] = 99
	}
	for i, forms := range monthForms {
		for _, form := range forms {
			if distance := levenshtein(word, form); distance < bestByMonth[i] {
				bestByMonth[i] = distance
			}
		}
	}
	best, second, month := 99, 99, 0
	for i, distance := range bestByMonth {
		if distance < best {
			second = best
			best = distance
			month = i + 1
		} else if distance < second {
			second = distance
		}
	}
	limit := 0
	if length == 4 {
		limit = 1
	}
	if length >= 5 {
		limit = 2
	}
	if best <= limit && best+1 <= second {
		return month, true
	}
	return 0, false
}

func levenshtein(left, right string) int {
	a, b := []rune(left), []rune(right)
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del, ins, sub := prev[j]+1, cur[j-1]+1, prev[j-1]+cost
			cur[j] = del
			if ins < cur[j] {
				cur[j] = ins
			}
			if sub < cur[j] {
				cur[j] = sub
			}
		}
		prev = cur
	}
	return prev[len(b)]
}

func quoteMonth(quote string) (int, bool) {
	best, count := 0, 0
	for _, match := range regexp.MustCompile(`[а-яе]{3,}`).FindAllString(norm(quote), -1) {
		month, ok := fuzzyMonth(match)
		if !ok {
			continue
		}
		if count == 0 {
			best = month
		}
		count++
	}
	return best, count == 1
}

func mentionsChildren(in PlanningInput) bool {
	state := false
	if in.Config.Trip != nil && in.Config.Trip.Travelers != nil && in.Config.Trip.Travelers.Children != nil {
		state = *in.Config.Trip.Travelers.Children > 0
	}
	sources := []string{in.Context.Summary}
	for _, m := range in.Context.Messages {
		if m.Role == "user" {
			sources = append(sources, m.Content)
		}
	}
	for _, a := range in.Answers {
		sources = append(sources, a.Text)
	}
	sources = append(sources, in.Text)
	for _, source := range sources {
		text := norm(source)
		if hasAny(text, "с ребенком", "с детьми", "ребенку", "ребенок", "детям") {
			state = true
		}
		if hasAny(text, "без детей", "дети не едут") {
			state = false
		}
	}
	return state
}
