package providerkit

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ostrovok-hackathon-2026/godnota/internal/trip"
)

type itineraryTimeoutKey struct{}

func (s *Searcher) searchItinerary(ctx context.Context, input TripRequest, resolved ResolvedTrip) (Parts, error) {
	result := Parts{TransportParts: []TransportPart{}, HotelParts: []HotelPart{}, Errors: []ProviderError{}, Limitations: []string{"Multi-city legs are independent provider searches; connection and availability must be verified."}}
	if len(resolved.Stops) != len(input.Stops) {
		result.Errors = append(result.Errors, ProviderError{Provider: "mapping", Kind: "transport", Code: "missing_field", Message: "/resolved/stops must contain one verified place per requested stop"})
		return result, nil
	}
	type segment struct {
		from, to           string
		fromPlace, toPlace ResolvedPlace
		date               string
	}
	segments := []segment{}
	from := input.Origin.City
	fromPlace := resolved.Origin
	date := input.Dates.DepartureDate
	for i, stop := range input.Stops {
		segments = append(segments, segment{from, stop.City, fromPlace, resolved.Stops[i], date})
		from = stop.City
		fromPlace = resolved.Stops[i]
		date = stop.CheckOut
	}
	if input.Transport == nil || input.Transport.RoundTrip == nil || *input.Transport.RoundTrip {
		segments = append(segments, segment{from, input.Origin.City, fromPlace, resolved.Origin, date})
	}
	type outcome struct {
		parts Parts
		err   error
	}
	out := make([]outcome, len(segments))
	var wg sync.WaitGroup
	for i, segment := range segments {
		wg.Add(1)
		go func() {
			defer wg.Done()
			raw, err := json.Marshal(input)
			if err != nil {
				out[i].err = err
				return
			}
			var single TripRequest
			if err = json.Unmarshal(raw, &single); err != nil {
				out[i].err = err
				return
			}
			single.Stops = nil
			single.Origin.City = segment.from
			single.Origin.CountryCode = ""
			single.Destination.City = segment.to
			single.Destination.CountryCode = ""
			single.Destination.Candidates = nil
			single.Dates.DepartureDate = segment.date
			single.Dates.DepartureMonth = ""
			single.Dates.ReturnDate = ""
			single.Dates.Nights = 0
			if single.Transport == nil {
				single.Transport = &trip.Transport{}
			}
			no := false
			single.Transport.RoundTrip = &no
			if single.Hotel == nil {
				single.Hotel = &trip.Hotel{}
			}
			single.Hotel.Needed = &no
			legCtx := context.WithValue(ctx, itineraryTimeoutKey{}, 8*time.Second)
			out[i].parts, out[i].err = s.SearchParts(legCtx, single, ResolvedTrip{Origin: segment.fromPlace, Destination: segment.toPlace})
		}()
	}
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return result, err
	}
	for i, item := range out {
		if item.err != nil {
			result.Errors = append(result.Errors, ProviderError{Provider: "mapping", Kind: "transport", Code: "segment_failed", Message: fmt.Sprintf("segment %d could not be searched", i+1)})
			continue
		}
		result.TransportParts = append(result.TransportParts, item.parts.TransportParts...)
		for _, issue := range item.parts.Errors {
			issue.Message = fmt.Sprintf("segment %d: %s", i+1, issue.Message)
			result.Errors = append(result.Errors, issue)
		}
		result.Limitations = append(result.Limitations, item.parts.Limitations...)
	}
	return result, nil
}
