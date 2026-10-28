package service

import (
	"ai-service/internal/canonical"
	"regexp"
	"strconv"
	"strings"
)

var nightsLiteral = regexp.MustCompile(`(?:на|ночей[: ]+)\s*([0-9]{1,2})\s*(?:ноч|$)`)
var roomsLiteral = regexp.MustCompile(`([0-9])\s*(?:номер|комнат)`)
var petWeightLiteral = regexp.MustCompile(`([0-9]+(?:[.,][0-9]+)?)\s*кг`)

func extendedLiterals(r canonical.TripRequest, in PlanningInput) canonical.TripRequest {
	texts := []string{in.Text}
	for _, a := range in.Answers {
		texts = append(texts, a.Text)
	}
	for _, value := range texts {
		text := norm(value)
		if match := nightsLiteral.FindStringSubmatch(text); match != nil {
			r.Dates.Nights, _ = strconv.Atoi(match[1])
		}
		if hasAny(text, "отель не нужен", "без отеля", "жилье не нужно") {
			if r.Hotel == nil {
				r.Hotel = &canonical.Hotel{}
			}
			r.Hotel.Needed = valuePointer(false)
		}
		if hasAny(text, "нужен отель", "нужна гостиница", "забронировать отель") {
			if r.Hotel == nil {
				r.Hotel = &canonical.Hotel{}
			}
			r.Hotel.Needed = valuePointer(true)
		}
		if hasAny(text, "бесплатной отмен", "бесплатная отмен") {
			if r.Hotel == nil {
				r.Hotel = &canonical.Hotel{}
			}
			r.Hotel.FreeCancellation = !hasAny(text, "без бесплатной отмен", "бесплатная отмена не нужна", "бесплатной отмены не нужно")
		}
		if match := roomsLiteral.FindStringSubmatch(text); match != nil {
			if r.Hotel == nil {
				r.Hotel = &canonical.Hotel{}
			}
			r.Hotel.Rooms, _ = strconv.Atoi(match[1])
		}
		for _, pair := range [][2]string{{"бассейн", "pool"}, {"парковк", "parking"}, {"детский клуб", "kids_club"}, {"вайфай", "wifi"}, {"wi-fi", "wifi"}, {"с питомцами", "pet_friendly"}} {
			if strings.Contains(text, pair[0]) && !hasAny(text, "без "+pair[0], "не нужен "+pair[0]) {
				if r.Hotel == nil {
					r.Hotel = &canonical.Hotel{}
				}
				found := false
				for _, v := range r.Hotel.Amenities {
					found = found || v == pair[1]
				}
				if !found {
					r.Hotel.Amenities = append(r.Hotel.Amenities, pair[1])
				}
			}
		}
		if hasAny(text, "только в одну сторону", "обратный билет не нужен", "без обратного билета") {
			if r.Transport == nil {
				r.Transport = &canonical.Transport{}
			}
			r.Transport.RoundTrip = valuePointer(false)
		}
		if hasAny(text, "туда и обратно", "нужен обратный билет") {
			if r.Transport == nil {
				r.Transport = &canonical.Transport{}
			}
			r.Transport.RoundTrip = valuePointer(true)
		}
		if len(r.Travelers.Pets) == 1 {
			if match := petWeightLiteral.FindStringSubmatch(text); match != nil {
				r.Travelers.Pets[0].WeightKg, _ = strconv.ParseFloat(strings.ReplaceAll(match[1], ",", "."), 64)
			}
		}
	}
	return r
}

func valuePointer[T any](v T) *T { return &v }
