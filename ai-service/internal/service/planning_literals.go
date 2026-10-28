package service

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
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
	return mergePlanning(base, literalPlanningPatch(in.Text), nil)
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
