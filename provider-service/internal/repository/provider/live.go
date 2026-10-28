package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	pb "github.com/godnota/pkg/generated/provider/v1"
	"github.com/godnota/provider-service/internal/entity"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
)

type cacheEntry struct {
	data    []byte
	at      time.Time
	expires time.Time
}

type upstreamStatusError struct{ status int }

func (e upstreamStatusError) Error() string {
	return fmt.Sprintf("provider returned HTTP %d", e.status)
}

func statusCode(status int) string {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return "access_denied"
	case http.StatusTooManyRequests:
		return "rate_limited"
	default:
		return "provider_error"
	}
}

type Repository struct {
	registry providerkit.Registry
	client   *http.Client
	poi      poiIndex
	mu       sync.Mutex
	next     map[string]time.Time
	cache    map[string]cacheEntry
}

func New(registry providerkit.Registry, poiPath string) (*Repository, error) {
	data, err := os.ReadFile(poiPath)
	if err != nil {
		return nil, fmt.Errorf("load POI index: %w", err)
	}
	var poi poiIndex
	if err := json.Unmarshal(data, &poi); err != nil {
		return nil, fmt.Errorf("parse POI index: %w", err)
	}
	return &Repository{
		registry: registry,
		client:   &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
		poi:      poi,
		next:     make(map[string]time.Time),
		cache:    make(map[string]cacheEntry),
	}, nil
}

func (r *Repository) Search(ctx context.Context, trip providerkit.TripRequest, resolved providerkit.ResolvedTrip, kind string) (entity.Result, error) {
	plan, err := providerkit.BuildTripPlan(trip, resolved, time.Now())
	if err != nil {
		return entity.Result{}, err
	}
	result := entity.Result{Plan: plan}
	result.Plan.Limitations = append(result.Plan.Limitations, "Provider offers are preliminary: traveler constraints, group totals, connections and final availability must be verified before recommending or booking.")
	if kind == "hotel" {
		missing := result.Plan.Missing[:0]
		for _, field := range result.Plan.Missing {
			switch field {
			case "/resolved/origin/iata", "/resolved/destination/iata", "/resolved/origin/yandex_code", "/resolved/destination/yandex_code":
			default:
				missing = append(missing, field)
			}
		}
		result.Plan.Missing = missing
		hb, rh := os.Getenv("HOTELBEDS_MODE"), os.Getenv("RATEHAWK_MODE")
		if hb != "" {
			r.searchHotelbeds(ctx, trip, resolved, &result)
		}
		if rh != "" {
			r.searchHotels(ctx, trip, resolved, &result)
		}
		if hb == "" && rh == "" && (trip.Hotel == nil || trip.Hotel.Needed == nil || *trip.Hotel.Needed) {
			result.Issues = append(result.Issues, &pb.ProviderIssue{Provider: "hotels", Code: "credentials_missing", Message: "Configure Hotelbeds evaluation or RateHawk partner access."})
		}
		return result, ctx.Err()
	}
	providers := []string{"aviasales", "yandex_rasp"}
	if kind == "activity" {
		providers = []string{"wikipedia"}
		r.searchPOI(trip, resolved, &result)
	}
	for _, name := range providers {
		for _, req := range plan.Search[name] {
			data, cached, at, err := r.fetch(ctx, req)
			if err != nil {
				if ctx.Err() != nil {
					return result, ctx.Err()
				}
				code := "provider_error"
				if errors.Is(err, context.DeadlineExceeded) {
					code = "timeout"
				}
				var upstream upstreamStatusError
				if errors.As(err, &upstream) {
					code = statusCode(upstream.status)
				}
				if strings.Contains(err.Error(), "is not set") {
					code = "credentials_missing"
				}
				result.Issues = append(result.Issues, &pb.ProviderIssue{Provider: name, Code: code, Message: err.Error()})
				if code == "credentials_missing" {
					break
				}
				continue
			}
			var offers []*pb.Offer
			switch name {
			case "aviasales":
				offers, err = flightOffers(data, trip.Currency, at)
			case "yandex_rasp":
				offers, err = scheduleOffers(data, at, cached)
			case "wikipedia":
				offers, err = wikipediaOffers(data, at, cached)
			}
			if err != nil {
				key, _ := json.Marshal(req)
				r.mu.Lock()
				delete(r.cache, string(key))
				r.mu.Unlock()
				code, message := "invalid_response", "provider returned an invalid response"
				var apiErr providerAPIError
				if errors.As(err, &apiErr) {
					code, message = apiErr.code, "provider reported an API error"
				}
				result.Issues = append(result.Issues, &pb.ProviderIssue{Provider: name, Code: code, Message: message})
				continue
			}
			result.Offers = append(result.Offers, offers...)
		}
	}
	if kind == "activity" {
		limit := 5
		if trip.Leisure != nil && trip.Leisure.MaxItemsPerStop > 0 {
			limit = trip.Leisure.MaxItemsPerStop
		}
		if len(result.Offers) > limit {
			result.Offers = fairLimit(result.Offers, limit)
			result.Issues = append(result.Issues, &pb.ProviderIssue{Provider: "activity", Code: "results_limited", Message: "Results were limited by max_items_per_stop."})
		}
	}
	return result, nil
}

func (r *Repository) fetch(ctx context.Context, request providerkit.Request) ([]byte, bool, time.Time, error) {
	provider, ok := r.registry[request.Provider]
	if !ok {
		return nil, false, time.Time{}, fmt.Errorf("provider is not configured")
	}
	keyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, false, time.Time{}, fmt.Errorf("invalid provider request")
	}
	key := string(keyBytes)
	httpRequest, err := r.registry.NewHTTPRequest(ctx, request)
	if err != nil {
		return nil, false, time.Time{}, err
	}
	now := time.Now()
	r.mu.Lock()
	if item, ok := r.cache[key]; ok && now.Before(item.expires) {
		r.mu.Unlock()
		return item.data, true, item.at, nil
	}
	r.mu.Unlock()
	if err := r.wait(ctx, request.Provider, provider.RPS); err != nil {
		return nil, false, time.Time{}, err
	}
	requestCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.TimeoutMS)*time.Millisecond)
	defer cancel()
	httpRequest = httpRequest.WithContext(requestCtx)
	response, err := r.client.Do(httpRequest)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, false, time.Time{}, fmt.Errorf("provider request timed out: %w", context.DeadlineExceeded)
		}
		return nil, false, time.Time{}, fmt.Errorf("provider request failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, false, time.Time{}, upstreamStatusError{status: response.StatusCode}
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 2<<20+1))
	if err != nil || len(data) > 2<<20 {
		return nil, false, time.Time{}, fmt.Errorf("provider response is unreadable or exceeds 2 MiB")
	}
	if !json.Valid(data) {
		return nil, false, time.Time{}, fmt.Errorf("provider response is invalid JSON")
	}
	at := time.Now().UTC()
	if provider.CacheTTLS > 0 {
		r.mu.Lock()
		if len(r.cache) >= 256 {
			keys := make([]string, 0, len(r.cache))
			for k := range r.cache {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			delete(r.cache, keys[0])
		}
		r.cache[key] = cacheEntry{data: data, at: at, expires: at.Add(time.Duration(provider.CacheTTLS) * time.Second)}
		r.mu.Unlock()
	}
	return data, false, at, nil
}

func (r *Repository) wait(ctx context.Context, name string, rps float64) error {
	r.mu.Lock()
	start := time.Now()
	if next := r.next[name]; next.After(start) {
		start = next
	}
	r.next[name] = start.Add(time.Duration(float64(time.Second) / rps))
	r.mu.Unlock()
	if delay := time.Until(start); delay > 0 {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func fairLimit(offers []*pb.Offer, limit int) []*pb.Offer {
	var providers []string
	groups := map[string][]*pb.Offer{}
	for _, offer := range offers {
		group := offer.Provider + ":" + offer.Kind
		if _, exists := groups[group]; !exists {
			providers = append(providers, group)
		}
		groups[group] = append(groups[group], offer)
	}
	selected := make([]*pb.Offer, 0, limit)
	for len(selected) < limit {
		for _, name := range providers {
			if len(groups[name]) == 0 {
				continue
			}
			selected = append(selected, groups[name][0])
			groups[name] = groups[name][1:]
			if len(selected) == limit {
				break
			}
		}
	}
	return selected
}
