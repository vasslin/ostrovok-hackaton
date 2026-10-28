package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
)

func hotelProviderEnabled() bool {
	return os.Getenv("HOTELBEDS_MODE") != "" || os.Getenv("RATEHAWK_MODE") != ""
}

func (c *Controller) addItineraryHotels(ctx context.Context, input providerkit.TripRequest, resolved providerkit.ResolvedTrip, parts providerkit.Parts) providerkit.Parts {
	if len(resolved.Stops) != len(input.Stops) {
		parts.Errors = append(parts.Errors, providerkit.ProviderError{Provider: "hotel", Kind: "hotel", Code: "missing_field", Message: "/resolved/stops must contain one verified place per requested stop"})
		return parts
	}
	out := make([]providerkit.Parts, len(input.Stops))
	var wg sync.WaitGroup
	for i, stop := range input.Stops {
		wg.Add(1)
		go func() {
			defer wg.Done()
			raw, err := json.Marshal(input)
			var one providerkit.TripRequest
			if err != nil || json.Unmarshal(raw, &one) != nil {
				out[i].Errors = append(out[i].Errors, providerkit.ProviderError{Provider: "hotel", Kind: "hotel", Code: "invalid_request", Message: "Hotel search could not be prepared"})
				return
			}
			one.Stops = nil
			one.Destination.City = stop.City
			one.Destination.CountryCode = ""
			one.Destination.Candidates = nil
			one.Dates.DepartureDate = stop.CheckIn
			one.Dates.ReturnDate = stop.CheckOut
			one.Dates.DepartureMonth = ""
			one.Dates.Nights = 0
			place := providerkit.ResolvedTrip{Origin: resolved.Origin, Destination: resolved.Stops[i], HotelCheckin: stop.CheckIn, HotelCheckout: stop.CheckOut, HotelbedsSourceMarket: resolved.HotelbedsSourceMarket}
			out[i] = c.addHotelParts(ctx, one, place, providerkit.Parts{})
			for j := range out[i].Errors {
				out[i].Errors[j].Message = fmt.Sprintf("stop %d: %s", i+1, out[i].Errors[j].Message)
			}
		}()
	}
	wg.Wait()
	for _, item := range out {
		parts.HotelParts = append(parts.HotelParts, item.HotelParts...)
		parts.Errors = append(parts.Errors, item.Errors...)
		parts.Limitations = append(parts.Limitations, item.Limitations...)
	}
	return parts
}

func (c *Controller) addHotelParts(ctx context.Context, input providerkit.TripRequest, resolved providerkit.ResolvedTrip, parts providerkit.Parts) providerkit.Parts {
	filtered := parts.Errors[:0]
	for _, issue := range parts.Errors {
		if issue.Provider != "hotel" || issue.Code != "not_configured" {
			filtered = append(filtered, issue)
		}
	}
	parts.Errors = filtered
	result, err := c.service.Search(ctx, input, resolved, "hotel")
	if err != nil {
		parts.Errors = append(parts.Errors, providerkit.ProviderError{Provider: "hotel", Kind: "hotel", Code: "search_failed", Message: "Hotel search failed", Retryable: true})
		return parts
	}
	for _, missing := range result.Plan.Missing {
		parts.Errors = append(parts.Errors, providerkit.ProviderError{Provider: "hotel", Kind: "hotel", Code: "missing_field", Message: missing})
	}
	warnings := []string{"Hotel filters, cancellation policy and final availability are not verified."}
	for _, issue := range result.Issues {
		switch issue.Code {
		case "sandbox_only", "filters_unverified", "content_and_prebook_required":
			warnings = append(warnings, issue.Message)
			parts.Limitations = append(parts.Limitations, issue.Message)
		default:
			parts.Errors = append(parts.Errors, providerkit.ProviderError{Provider: issue.Provider, Kind: "hotel", Code: issue.Code, Message: issue.Message, Retryable: issue.Code == "timeout" || issue.Code == "provider_error" || issue.Code == "rate_limited"})
		}
	}
	city := ""
	if input.Destination != nil {
		city = input.Destination.City
	}
	for _, offer := range result.Offers {
		if offer == nil || offer.Kind != "hotel" {
			continue
		}
		part := providerkit.HotelPart{ID: offer.Id, Provider: offer.Provider, Source: offer.SourceStatus, FetchedAt: offer.ObservedAt, City: city, Name: offer.Title, CheckIn: resolved.HotelCheckin, CheckOut: resolved.HotelCheckout, Warnings: append([]string(nil), warnings...)}
		if offer.PriceMinor != nil {
			part.Price = &providerkit.Money{AmountMinor: *offer.PriceMinor, Currency: offer.Currency, Scope: offer.PriceScope}
		}
		parts.HotelParts = append(parts.HotelParts, part)
	}
	return parts
}
