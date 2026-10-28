package providerkit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ostrovok-hackathon-2026/godnota/internal/trip"
)

// DecodeTrip rejects unknown fields and trailing JSON, including invalid AI output.
func DecodeTrip(raw []byte) (TripRequest, error) {
	var input TripRequest
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&input); err != nil {
		return input, fmt.Errorf("invalid trip config: %w", err)
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return input, errors.New("expected exactly one trip config")
	}
	return input, nil
}

func ValidateSearch(input TripRequest, now time.Time) error {
	if err := input.ValidateAt(now); err != nil {
		return err
	}
	if input.Currency != "RUB" && input.Currency != "USD" && input.Currency != "EUR" {
		return errors.New("stage-2 prices currently support RUB, USD and EUR (two decimal minor units)")
	}
	if strings.TrimSpace(input.Origin.City) == "" || input.Destination == nil || strings.TrimSpace(input.Destination.City) == "" {
		return errors.New("origin and destination cities are required for search")
	}
	if strings.EqualFold(strings.TrimSpace(input.Origin.City), strings.TrimSpace(input.Destination.City)) {
		return errors.New("origin and destination must differ")
	}
	if input.Dates.DepartureDate == "" {
		return errors.New("exact departure date is required for search")
	}
	roundTrip := input.Transport == nil || input.Transport.RoundTrip == nil || *input.Transport.RoundTrip
	hotel := input.Hotel == nil || input.Hotel.Needed == nil || *input.Hotel.Needed
	if (roundTrip || hotel) && input.Dates.ReturnDate == "" && input.Dates.Nights == 0 {
		return errors.New("return date or nights is required")
	}
	return nil
}

type HotelQuery struct {
	City             string     `json:"city"`
	CheckIn          string     `json:"check_in"`
	CheckOut         string     `json:"check_out"`
	Adults           int        `json:"adults"`
	ChildrenAges     []trip.Age `json:"children_ages,omitempty"`
	Pets             []trip.Pet `json:"pets,omitempty"`
	Rooms            int        `json:"rooms"`
	MinStars         *int       `json:"min_stars,omitempty"`
	MaxPricePerNight int        `json:"max_price_per_night,omitempty"`
	Currency         string     `json:"currency"`
	Meal             string     `json:"meal,omitempty"`
	FreeCancellation bool       `json:"free_cancellation"`
	Kinds            []string   `json:"kinds,omitempty"`
	Amenities        []string   `json:"amenities,omitempty"`
	Location         []string   `json:"location,omitempty"`
	NearPlace        string     `json:"near_place,omitempty"`
	EarlyCheckin     string     `json:"early_checkin,omitempty"`
	LateCheckout     string     `json:"late_checkout,omitempty"`
}

type Mapping struct {
	Configs     []Request `json:"configs"`
	Missing     []string  `json:"missing"`
	Limitations []string  `json:"limitations"`
}

func MapConfig(input TripRequest, resolved ResolvedTrip, now time.Time) (Mapping, error) {
	plan, err := BuildTripPlan(input, resolved, now)
	if err != nil {
		return Mapping{}, err
	}
	result := Mapping{Configs: []Request{}, Missing: plan.Missing, Limitations: plan.Limitations}
	keys := make([]string, 0, len(plan.Search))
	for key := range plan.Search {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		result.Configs = append(result.Configs, plan.Search[key]...)
	}
	if input.Hotel == nil || input.Hotel.Needed == nil || *input.Hotel.Needed {
		end := checkout(input)
		if input.Destination == nil || input.Destination.City == "" || input.Dates.DepartureDate == "" || end == "" {
			result.Missing = append(result.Missing, "/hotel/city|check_in|check_out")
		} else {
			q := HotelQuery{City: input.Destination.City, CheckIn: input.Dates.DepartureDate, CheckOut: end, Adults: input.Travelers.Adults, ChildrenAges: input.Travelers.ChildrenAges, Pets: input.Travelers.Pets, Rooms: 1, Currency: input.Currency}
			if h := input.Hotel; h != nil {
				if h.Rooms > 0 {
					q.Rooms = h.Rooms
				}
				q.MinStars = h.StarRatingMin
				q.MaxPricePerNight = h.PricePerNightMax
				q.Meal = h.Meal
				q.FreeCancellation = h.FreeCancellation
				q.Kinds = h.Kinds
				q.Amenities = h.Amenities
				q.Location = h.Location
				q.NearPlace = h.NearPlace
				q.EarlyCheckin = h.EarlyCheckin
				q.LateCheckout = h.LateCheckout
			}
			result.Configs = append(result.Configs, Request{Provider: "hotel", Endpoint: "search", Params: q})
		}
	}
	return result, nil
}

func checkout(input TripRequest) string {
	if input.Dates.ReturnDate != "" {
		return input.Dates.ReturnDate
	}
	start, err := time.Parse(time.DateOnly, input.Dates.DepartureDate)
	if err != nil || input.Dates.Nights == 0 {
		return ""
	}
	return start.AddDate(0, 0, input.Dates.Nights).Format(time.DateOnly)
}

type Money struct {
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
	Scope       string `json:"scope"`
}
type TransportPart struct {
	ID                string   `json:"id"`
	Provider          string   `json:"provider"`
	Source            string   `json:"source"`
	FetchedAt         string   `json:"fetched_at"`
	Mode              string   `json:"mode"`
	Origin            string   `json:"origin"`
	Destination       string   `json:"destination"`
	DepartureAt       string   `json:"departure_at"`
	ArrivalAt         string   `json:"arrival_at"`
	Price             *Money   `json:"price,omitempty"`
	BookingURL        string   `json:"booking_url,omitempty"`
	Warnings          []string `json:"warnings,omitempty"`
	ReturnDepartureAt string   `json:"return_departure_at,omitempty"`
	ReturnArrivalAt   string   `json:"return_arrival_at,omitempty"`
	OriginCity        string   `json:"origin_city,omitempty"`
	DestinationCity   string   `json:"destination_city,omitempty"`
	Transfers         int      `json:"transfers"`
	ReturnTransfers   int      `json:"return_transfers"`
}
type HotelPart struct {
	ID               string   `json:"id"`
	Provider         string   `json:"provider"`
	Source           string   `json:"source"`
	FetchedAt        string   `json:"fetched_at"`
	City             string   `json:"city"`
	Name             string   `json:"name"`
	CheckIn          string   `json:"check_in"`
	CheckOut         string   `json:"check_out"`
	Stars            int      `json:"stars"`
	Price            *Money   `json:"price,omitempty"`
	FreeCancellation bool     `json:"free_cancellation"`
	Warnings         []string `json:"warnings,omitempty"`
}
type ProviderError struct {
	Provider  string `json:"provider"`
	Kind      string `json:"kind"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}
type Parts struct {
	TransportParts []TransportPart `json:"transport_parts"`
	HotelParts     []HotelPart     `json:"hotel_parts"`
	Errors         []ProviderError `json:"errors"`
	Limitations    []string        `json:"limitations"`
}

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}
type HotelAdapter interface {
	SearchHotels(context.Context, HotelQuery) ([]HotelPart, error)
}
type Searcher struct {
	Registry Registry
	Client   HTTPDoer
	Mode     string // demo or live; no automatic fallback
	Timeout  time.Duration
	Now      func() time.Time
	Hotels   HotelAdapter
	mu       sync.Mutex
	next     map[string]time.Time
}

func NewSearcher(reg Registry, mode string) (*Searcher, error) {
	if mode != "demo" && mode != "live" {
		return nil, errors.New("PROVIDER_MODE must be demo or live")
	}
	return &Searcher{Registry: reg, Mode: mode, Timeout: 4 * time.Second, Now: time.Now, Client: &http.Client{Timeout: 4 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, next: map[string]time.Time{}}, nil
}

func (s *Searcher) SearchParts(ctx context.Context, input TripRequest, resolved ResolvedTrip) (Parts, error) {
	now := s.Now()
	result := Parts{TransportParts: []TransportPart{}, HotelParts: []HotelPart{}, Errors: []ProviderError{}, Limitations: []string{}}
	if err := ValidateSearch(input, now); err != nil {
		return result, err
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	mapping, err := MapConfig(input, resolved, now)
	if err != nil {
		return result, err
	}
	if s.Mode == "demo" {
		return demoParts(ctx, input, now), nil
	}
	result.Limitations = mapping.Limitations
	for _, field := range mapping.Missing {
		result.Errors = append(result.Errors, ProviderError{Provider: "mapping", Kind: "transport", Code: "missing_field", Message: field})
	}
	type outcome struct {
		parts  []TransportPart
		hotels []HotelPart
		err    *ProviderError
	}
	jobs := []Request{}
	for _, cfg := range mapping.Configs {
		if cfg.Provider == "aviasales" || cfg.Provider == "yandex_rasp" || cfg.Provider == "hotel" {
			jobs = append(jobs, cfg)
		}
	}
	outcomes := make([]outcome, len(jobs))
	var wg sync.WaitGroup
	for i, job := range jobs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if job.Provider == "hotel" {
				hotels, e := s.searchHotels(ctx, job.Params.(HotelQuery))
				outcomes[i] = outcome{hotels: hotels, err: e}
				return
			}
			parts, e := s.searchTransport(ctx, job, now)
			outcomes[i] = outcome{parts: parts, err: e}
		}()
	}
	wg.Wait()
	seen := map[string]bool{}
	for i, o := range outcomes {
		for _, part := range o.parts {
			part.OriginCity, part.DestinationCity = input.Origin.City, input.Destination.City
			if jobs[i].Provider == "yandex_rasp" && checkout(input) != "" && strings.HasPrefix(part.DepartureAt, checkout(input)) {
				part.OriginCity, part.DestinationCity = part.DestinationCity, part.OriginCity
			}
			if !seen[part.ID] && matchesTransport(part, input) {
				result.TransportParts = append(result.TransportParts, part)
				seen[part.ID] = true
			}
		}
		result.HotelParts = append(result.HotelParts, o.hotels...)
		if o.err != nil {
			result.Errors = append(result.Errors, *o.err)
		}
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	return result, nil
}

func (s *Searcher) searchHotels(ctx context.Context, q HotelQuery) ([]HotelPart, *ProviderError) {
	failure := func(code, message string, retry bool) ([]HotelPart, *ProviderError) {
		return nil, &ProviderError{Provider: "hotel", Kind: "hotel", Code: code, Message: message, Retryable: retry}
	}
	if s.Hotels == nil {
		return failure("not_configured", "No live hotel API configured; explicit demo mode provides synthetic hotels", false)
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	parts, err := s.Hotels.SearchHotels(ctx, q)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return failure("timeout", "Hotel provider timed out", true)
		}
		return failure("unavailable", "Hotel provider request failed", true)
	}
	return parts, nil
}

func matchesTransport(part TransportPart, input TripRequest) bool {
	p := input.Transport
	if p == nil {
		return true
	}
	if p.DirectOnly && (part.Transfers != 0 || part.ReturnTransfers != 0) {
		return false
	}
	if p.MaxTransfers != nil && (part.Transfers > *p.MaxTransfers || part.ReturnTransfers > *p.MaxTransfers) {
		return false
	}
	for _, leg := range [][2]string{{part.DepartureAt, part.ArrivalAt}, {part.ReturnDepartureAt, part.ReturnArrivalAt}} {
		if leg[0] == "" {
			continue
		}
		dep, e1 := time.Parse(time.RFC3339, leg[0])
		arr, e2 := time.Parse(time.RFC3339, leg[1])
		if e1 != nil || e2 != nil {
			return false
		}
		if p.MaxDurationHours > 0 && arr.Sub(dep).Hours() > p.MaxDurationHours {
			return false
		}
		clock := dep.Format("15:04")
		from, to := p.DepartureTimeFrom, p.DepartureTimeTo
		if from != "" && to != "" && from > to {
			if clock < from && clock > to {
				return false
			}
		} else {
			if from != "" && clock < from {
				return false
			}
			if to != "" && clock > to {
				return false
			}
		}
	}
	return true
}

func (s *Searcher) searchTransport(ctx context.Context, job Request, now time.Time) ([]TransportPart, *ProviderError) {
	failure := func(code, message string, retry bool) ([]TransportPart, *ProviderError) {
		return nil, &ProviderError{Provider: job.Provider, Kind: "transport", Code: code, Message: message, Retryable: retry}
	}
	cfg, ok := s.Registry[job.Provider]
	if !ok {
		return failure("not_configured", "Provider is not configured", false)
	}
	timeout := s.Timeout
	if cfg.TimeoutMS > 0 && time.Duration(cfg.TimeoutMS)*time.Millisecond < timeout {
		timeout = time.Duration(cfg.TimeoutMS) * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := s.waitSlot(ctx, job.Provider, cfg.RPS); err != nil {
		return failure("timeout", "Provider rate limit wait exceeded deadline", true)
	}
	req, err := s.Registry.NewHTTPRequest(ctx, job)
	if err != nil {
		return failure("configuration", "Missing credentials or invalid provider configuration", false)
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return failure("timeout", "Provider timed out", true)
		}
		return failure("unavailable", "Provider request failed", true)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return failure(fmt.Sprintf("http_%d", resp.StatusCode), "Provider returned an HTTP error", resp.StatusCode == 429 || resp.StatusCode >= 500)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, (4<<20)+1))
	if err != nil || len(data) > 4<<20 {
		return failure("invalid_response", "Provider response unreadable or too large", false)
	}
	parts, err := normalize(job, data, now)
	if err != nil {
		return failure("invalid_response", "Provider returned invalid data", false)
	}
	return parts, nil
}

func (s *Searcher) waitSlot(ctx context.Context, provider string, rps float64) error {
	if rps <= 0 {
		rps = 1
	}
	s.mu.Lock()
	if s.next == nil {
		s.next = map[string]time.Time{}
	}
	at := s.next[provider]
	if at.Before(time.Now()) {
		at = time.Now()
	}
	s.next[provider] = at.Add(time.Duration(float64(time.Second) / rps))
	s.mu.Unlock()
	timer := time.NewTimer(time.Until(at))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
