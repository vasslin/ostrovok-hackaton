package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ai-service/internal/canonical"
)

type PlanningEvidence struct {
	Field string `json:"field"`
	Quote string `json:"quote"`
}

var planningNumbers = regexp.MustCompile(`[0-9]+(?:[.,][0-9]+)?`)

func norm(s string) string { return strings.ToLower(strings.ReplaceAll(s, "ё", "е")) }
func hasNumber(quote string, want float64) bool {
	text := norm(quote)
	for _, v := range planningNumbers.FindAllString(text, -1) {
		n, _ := strconv.ParseFloat(strings.ReplaceAll(v, ",", "."), 64)
		if n == want {
			return true
		}
		if strings.Contains(text, "тыс") && n*1000 == want {
			return true
		}
	}
	words := map[float64][]string{0: {"ноль", "без детей", "без пересад"}, 1: {"один", "одна", "одного", "одной", "одним"}, 2: {"двое", "двоих", "два", "двух", "две", "вдвоем"}, 3: {"трое", "троих", "три", "трех", "втроем"}, 4: {"четыре", "четверо", "четырех"}, 5: {"пять", "пятеро"}, 6: {"шесть", "шестеро"}, 7: {"семь", "семеро"}, 8: {"восемь"}, 9: {"девять"}}
	for _, word := range words[want] {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}
func quoteSupports(path string, value any, quote string) bool {
	q := norm(quote)
	switch v := value.(type) {
	case float64:
		if strings.HasSuffix(path, ".adults") && !hasAny(q, "взросл", "дво", "тро", "четвер", "нас ", "человек", "поеду один", "поеду одна") {
			return false
		}
		if strings.HasSuffix(path, ".children") && !hasAny(q, "ребен", "дет", "без детей") {
			return false
		}
		if strings.HasSuffix(path, ".max_wait_hours") && !hasAny(q, "час") {
			return false
		}
		if strings.HasSuffix(path, ".max_transfers") && v != 0 && hasAny(q, "без пересад", "прямой", "прямым") {
			return false
		}
		if strings.HasSuffix(path, ".amount") && !hasAny(q, "бюджет", "руб", "доллар", "евро", "стоим", "цен", "₽", "$", "€") {
			return false
		}
		if strings.HasSuffix(path, ".min_stars") && !hasAny(q, "звезд", "звезд") {
			return false
		}
		if strings.HasSuffix(path, ".max_price_per_night") && !hasAny(q, "ноч", "сутк") {
			return false
		}
		if strings.HasSuffix(path, ".max_transfers") && v == 0 && hasAny(q, "без пересад", "прямой", "прямым") {
			return true
		}
		if strings.HasSuffix(path, ".children") && v == 1 && hasAny(q, "с ребенком", "один ребенок", "одного ребенка") {
			return true
		}
		return hasNumber(q, v)
	case string:
		if strings.HasSuffix(path, ".city") {
			return cityQuoteMatches(v, q)
		}
		if strings.HasSuffix(path, ".country_code") {
			return regexp.MustCompile(`\b`+regexp.QuoteMeta(norm(v))+`\b`).MatchString(q) || (v == "RU" && hasAny(q, "росси"))
		}
		if strings.HasSuffix(path, ".currency") {
			return strings.Contains(q, norm(v)) || (v == "RUB" && hasAny(q, "руб", "₽")) || (v == "USD" && hasAny(q, "доллар", "$")) || (v == "EUR" && hasAny(q, "евро", "€"))
		}
		if strings.HasSuffix(path, ".start_date") || strings.HasSuffix(path, ".end_date") || strings.HasSuffix(path, ".departure_date") || strings.HasSuffix(path, ".return_date") {
			d, err := time.Parse("2006-01-02", v)
			if err != nil {
				return false
			}
			for _, number := range planningNumbers.FindAllString(q, -1) {
				year, _ := strconv.Atoi(number)
				if year >= 1900 && year <= 2100 && year != d.Year() {
					return false
				}
			}
			if hasAny(q, "завтра", "послезавтра", "через ", "на неделю", "на две недели") {
				return true
			}
			if month, ok := quoteMonth(q); ok {
				if int(d.Month()) != month {
					return false
				}
				if hasNumber(q, float64(d.Day())) {
					return true
				}
			}
			months := []string{"январ", "феврал", "март", "апрел", "мая", "июн", "июл", "август", "сентябр", "октябр", "ноябр", "декабр"}
			seenMonth, correctMonth := false, false
			for index, month := range months {
				if strings.Contains(q, month) {
					seenMonth = true
					correctMonth = correctMonth || int(d.Month()) == index+1
				}
			}
			if seenMonth && !correctMonth {
				return false
			}
			return hasNumber(q, float64(d.Day())) && (hasAny(q, "январ", "феврал", "март", "апрел", "мая", "май", "июн", "июл", "август", "сентябр", "октябр", "ноябр", "декабр") || strings.Contains(q, "-") || strings.Contains(q, "."))
		}
		if strings.HasSuffix(path, ".scope") {
			return v == "total" && hasAny(q, "всю поездку", "общий", "всего") || v == "per_person" && hasAny(q, "на человека", "на каждого") || v == "per_night" && hasAny(q, "ноч", "сутк")
		}
		if strings.HasSuffix(path, ".language") && !hasAny(q, "язык", "русск", "english") {
			return false
		}
		return strings.Contains(q, norm(v))
	case bool:
		switch {
		case strings.HasSuffix(path, ".avoid_overnight_transfers"):
			return hasAny(q, "ноч")
		case strings.HasSuffix(path, ".flexible_dates"):
			return hasAny(q, "дат", "числ", "дней") && hasAny(q, "гибк", "позже", "раньше", "плюс", "минус", "точно", "строго")
		case strings.HasSuffix(path, ".required"):
			return hasAny(q, "отел", "гостиниц", "жиль", "прожив")
		case strings.HasSuffix(path, ".is_hard_limit"):
			return hasAny(q, "не больше", "не более", "строго", "максимум", "до ", "жестк")
		}
		return false
	case []any:
		if len(v) == 0 {
			return false
		}
		for _, item := range v {
			if place, ok := item.(map[string]any); ok {
				city, _ := place["city"].(string)
				if !quoteSupports(path+".city", city, q) {
					return false
				}
				continue
			}
			if n, ok := item.(float64); ok {
				if !hasAny(q, "ребен", "дет", "лет", "год") || !hasNumber(q, n) {
					return false
				}
				continue
			}
			str, ok := item.(string)
			if !ok {
				return false
			}
			aliases := map[string][]string{"plane": {"самолет", "авиа", "вылет", "перелет"}, "train": {"поездом", "поезде", "железнодорож"}, "bus": {"автобус"}, "water": {"паром", "корабл", "водн"}, "suburban": {"электрич"}, "helicopter": {"вертолет"}, "museums": {"музе"}, "nature": {"природ"}, "beach": {"море", "пляж"}, "hotel": {"отел", "гостиниц"}, "apartment": {"апартамент", "квартир"}}
			if !strings.Contains(q, norm(str)) && !hasAny(q, aliases[str]...) {
				return false
			}
		}
		return true
	}
	return false
}
func cityQuoteMatches(city, quote string) bool {
	name := norm(city)
	stem := []rune(name)
	if len(stem) > 4 {
		stem = stem[:len(stem)-2]
	}
	if len(stem) > 1 && strings.Contains(quote, string(stem)) {
		return true
	}
	aliases := map[string][]string{
		"санкт-петербург": {"питер", "спб", "петербург", "ленинград"},
		"москва":          {"мск"},
	}
	for official, words := range aliases {
		if name == official || strings.Contains(name, official) {
			if hasAny(quote, words...) {
				return true
			}
		}
	}
	return false
}
func datePath(path string) bool {
	return strings.HasSuffix(path, ".start_date") || strings.HasSuffix(path, ".end_date") || strings.HasSuffix(path, ".departure_date") || strings.HasSuffix(path, ".return_date")
}
func quotedFieldAccepts(path string, value any, quote string) bool {
	switch item := value.(type) {
	case string:
		if strings.HasSuffix(path, ".city") {
			return strings.TrimSpace(item) != ""
		}
		if datePath(path) {
			parsed, err := time.Parse("2006-01-02", item)
			if err != nil || monthWithoutDay(norm(quote), parsed) {
				return false
			}
			if month, ok := quoteMonth(quote); ok && (int(parsed.Month()) != month || !hasNumber(quote, float64(parsed.Day()))) {
				return false
			}
			return true
		}
	case []any:
		if !strings.HasSuffix(path, ".destinations") || len(item) == 0 {
			return false
		}
		for _, place := range item {
			fields, ok := place.(map[string]any)
			city, _ := fields["city"].(string)
			if !ok || strings.TrimSpace(city) == "" {
				return false
			}
		}
		return true
	}
	return false
}
func monthWithoutDay(quote string, date time.Time) bool {
	for _, number := range planningNumbers.FindAllString(quote, -1) {
		year, _ := strconv.Atoi(number)
		if year >= 1900 && year <= 2100 && year != date.Year() {
			return true
		}
	}
	if hasAny(quote, "завтра", "послезавтра", "через ", "на неделю", "на две недели", "выходн", "понедельник", "вторник", "сред", "четверг", "пятниц", "суббот", "воскрес") {
		return false
	}
	if month, ok := quoteMonth(quote); ok {
		if int(date.Month()) != month {
			return true
		}
		if hasNumber(quote, float64(date.Day())) || hasAny(quote, "перв", "второ", "треть", "четверт", "пятнадцат", "двадцат", "тридцат") {
			return false
		}
		return true
	}
	months := []string{"январ", "феврал", "март", "апрел", "мая", "июн", "июл", "август", "сентябр", "октябр", "ноябр", "декабр"}
	seenMonth, correctMonth := false, false
	for index, month := range months {
		if strings.Contains(quote, month) {
			seenMonth = true
			correctMonth = correctMonth || int(date.Month()) == index+1
		}
	}
	if seenMonth && !correctMonth {
		return true
	}
	if strings.Contains(quote, "-") || strings.Contains(quote, ".") || hasNumber(quote, float64(date.Day())) || hasAny(quote, "перв", "второ", "треть", "четверт", "пятнадцат", "двадцат", "тридцат") {
		return false
	}
	return seenMonth
}
func hasAny(text string, parts ...string) bool {
	for _, part := range parts {
		if strings.Contains(text, part) {
			return true
		}
	}
	return false
}
func groundedPlanningPatch(in PlanningInput, c planningCompletion) (PlanningConfig, []string, []PlanningAssumption) {
	sources := []string{in.Text, in.Context.Summary}
	for _, m := range in.Context.Messages {
		if m.Role == "user" {
			sources = append(sources, m.Content)
		}
	}
	for _, a := range in.Answers {
		sources = append(sources, a.Text)
	}
	quotes := map[string]string{}
	assumptions := []PlanningAssumption{}
	for _, e := range c.Evidence {
		field := strings.TrimPrefix(strings.ReplaceAll(strings.TrimPrefix(e.Field, "/"), "/", "."), "config.")
		if len(e.Quote) < 2 || len(e.Quote) > 2000 {
			continue
		}
		for _, source := range sources {
			if strings.Contains(norm(source), norm(e.Quote)) {
				quotes[field] = e.Quote
				break
			}
		}
	}
	data, _ := json.Marshal(c.Config)
	var patch map[string]any
	_ = json.Unmarshal(data, &patch)
	var clean func(map[string]any, string) map[string]any
	clean = func(value map[string]any, prefix string) map[string]any {
		result := map[string]any{}
		for key, item := range value {
			path := prefix + key
			if path == "version" {
				continue
			}
			if child, ok := item.(map[string]any); ok {
				kept := clean(child, path+".")
				if len(kept) > 0 {
					result[key] = kept
				}
				continue
			}
			if item != nil && quotes[path] != "" && quotedFieldAccepts(path, item, quotes[path]) {
				result[key] = item
				if date, ok := item.(string); ok && datePath(path) && len(date) >= 4 && !strings.Contains(quotes[path], date[:4]) {
					assumptions = append(assumptions, PlanningAssumption{Field: path, Value: date, Reason: "Дата истолкована относительно " + in.CurrentDate + "; исходная цитата: " + quotes[path]})
				}
				continue
			}
			candidates := []string{quotes[path]}
			parent := path
			for strings.Contains(parent, ".") {
				parent = parent[:strings.LastIndex(parent, ".")]
				candidates = append(candidates, quotes[parent])
			}
			for field, quote := range quotes {
				if strings.HasPrefix(field, path+".") {
					candidates = append(candidates, quote)
				}
			}
			if strings.Contains(path, ".") {
				family := path[:strings.LastIndex(path, ".")+1]
				for field, quote := range quotes {
					if strings.HasPrefix(field, family) {
						candidates = append(candidates, quote)
					}
				}
			}
			candidates = append(candidates, sources...)
			for _, quote := range candidates {
				if item != nil && quote != "" && quoteSupports(path, item, quote) {
					result[key] = item
					if date, ok := item.(string); ok && (strings.HasSuffix(path, ".start_date") || strings.HasSuffix(path, ".end_date")) && len(date) >= 4 && !strings.Contains(quote, date[:4]) {
						assumptions = append(assumptions, PlanningAssumption{Field: path, Value: date, Reason: "Дата истолкована относительно " + in.CurrentDate + "; исходная цитата: " + quote})
					}
					break
				}
			}

		}
		return result
	}
	kept := clean(patch, "")
	kept["version"] = "1.0"
	data, _ = json.Marshal(kept)
	var out PlanningConfig
	_ = json.Unmarshal(data, &out)
	cleared := []string{}
	for _, field := range c.ClearedFields {
		if quotes[field] != "" {
			cleared = append(cleared, field)
		}
	}
	return out, cleared, assumptions
}
func notesMemory(summary string) bool {
	var memory struct {
		Notes []json.RawMessage `json:"notes"`
	}
	return json.Unmarshal([]byte(strings.TrimSpace(summary)), &memory) == nil && memory.Notes != nil
}
func describeTrip(trip canonical.TripRequest) string {
	origin := strings.TrimSpace(trip.Origin.City)
	destination := ""
	if trip.Destination != nil {
		destination = strings.TrimSpace(trip.Destination.City)
	}
	sentence := "Вот описание вашей поездки: направление пока не собрано"
	switch {
	case origin != "" && destination != "":
		sentence = "Вот описание вашей поездки: вы едете из " + origin + " в " + destination
	case destination != "":
		sentence = "Вот описание вашей поездки: вы едете в " + destination
	case origin != "":
		sentence = "Вот описание вашей поездки: вы выезжаете из " + origin
	}
	when := ""
	switch {
	case trip.Dates.DepartureDate != "" && trip.Dates.ReturnDate != "":
		when = " с " + trip.Dates.DepartureDate + " по " + trip.Dates.ReturnDate
	case trip.Dates.DepartureDate != "" && trip.Dates.Nights > 0:
		when = fmt.Sprintf(" с %s, %d ночей", trip.Dates.DepartureDate, trip.Dates.Nights)
	case trip.Dates.DepartureDate != "":
		when = " с выездом " + trip.Dates.DepartureDate
	case trip.Dates.DepartureMonth != "":
		when = " в " + trip.Dates.DepartureMonth
	}
	who := ""
	if trip.Travelers.Adults > 0 {
		who = fmt.Sprintf(", %d взрослых", trip.Travelers.Adults)
		if n := len(trip.Travelers.ChildrenAges); n > 0 {
			who += fmt.Sprintf(" и %d детей", n)
		}
	}
	budget := ""
	if trip.Budget != nil && trip.Budget.Total > 0 {
		currency := trip.Currency
		if currency == "" {
			currency = "RUB"
		}
		budget = fmt.Sprintf(", бюджет %d %s", trip.Budget.Total, currency)
	}
	sentence += when + who + budget + "."
	missing := []string{}
	if origin == "" {
		missing = append(missing, "город отправления")
	}
	if destination == "" {
		missing = append(missing, "направление")
	}
	if trip.Dates.DepartureDate == "" && trip.Dates.DepartureMonth == "" {
		missing = append(missing, "дата выезда")
	}
	if trip.Dates.ReturnDate == "" && trip.Dates.Nights == 0 {
		missing = append(missing, "дата возвращения")
	}
	if trip.Travelers.Adults == 0 {
		missing = append(missing, "число взрослых")
	}
	if len(missing) > 0 {
		sentence += " Вам ещё нужно назвать: " + strings.Join(missing, ", ") + "."
	}
	return sentence
}
func planningSummary(in PlanningInput) string {
	type memory struct {
		Notes []string `json:"notes"`
	}
	previous := memory{}
	if strings.TrimSpace(in.Context.Summary) != "" {
		if json.Unmarshal([]byte(in.Context.Summary), &previous) != nil || len(previous.Notes) == 0 {
			previous.Notes = []string{in.Context.Summary}
		}
	}
	for _, m := range in.Context.Messages {
		if m.Role == "user" {
			previous.Notes = append(previous.Notes, m.Content)
		}
	}
	previous.Notes = append(previous.Notes, in.Text)
	for _, answer := range in.Answers {
		previous.Notes = append(previous.Notes, answer.Text)
	}
	notes := []string{}
	seen := map[string]bool{}
	for _, note := range previous.Notes {
		note = strings.TrimSpace(note)
		if note == "" || seen[note] {
			continue
		}
		seen[note] = true
		runes := []rune(note)
		if len(runes) > 1000 {
			note = string(runes[:1000])
		}
		notes = append(notes, note)
	}
	for len(notes) > 6 {
		notes = notes[1:]
	}
	for {
		data, _ := json.Marshal(memory{Notes: notes})
		if len([]rune(string(data))) <= 3000 || len(notes) == 0 {
			return string(data)
		}
		notes = notes[1:]
	}
}
