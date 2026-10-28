package provider

import (
	"context"
	"encoding/json"
	pb "github.com/godnota/pkg/generated/provider/v1"
	"github.com/godnota/pkg/tripconfig"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func (c *Controller) SearchParts(ctx context.Context, req *pb.SearchPartsRequest) (*pb.SearchPartsResponse, error) {
	if c.searcher == nil {
		return nil, status.Error(codes.Unavailable, "provider search not configured")
	}
	input, resolved, err := tripconfig.Decode(req.GetConfig(), req.GetResolved())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := providerkit.ValidateSearch(input, c.searcher.Now()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resolved, resolutionErrors := c.searcher.ResolveTrip(ctx, input, resolved)
	parts, err := c.searcher.SearchParts(ctx, input, resolved)
	if err != nil {
		if ctx.Err() != nil {
			return nil, status.FromContextError(ctx.Err()).Err()
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if c.searcher.Mode == "live" && c.searcher.Hotels == nil && c.service != nil && hotelProviderEnabled() && (input.Hotel == nil || input.Hotel.Needed == nil || *input.Hotel.Needed) {
		if len(input.Stops) > 0 {
			parts.Errors = withoutHotelPlaceholder(parts.Errors)
			parts = c.addItineraryHotels(ctx, input, resolved, parts)
		} else {
			parts.Errors = withoutHotelPlaceholder(parts.Errors)
			parts = c.addHotelParts(ctx, input, resolved, parts)
		}
		if ctx.Err() != nil {
			return nil, status.FromContextError(ctx.Err()).Err()
		}
	}
	if (input.Hotel == nil || input.Hotel.Needed == nil || *input.Hotel.Needed) && len(parts.HotelParts) == 0 {
		hasHotelError := false
		for _, issue := range parts.Errors {
			if issue.Kind == "hotel" {
				hasHotelError = true
			}
		}
		if !hasHotelError {
			parts.Errors = append(parts.Errors, providerkit.ProviderError{Provider: "hotel", Kind: "hotel", Code: "no_results", Message: "Подходящие отели не найдены. Измените даты или требования."})
		}
	}
	parts.Resolved = &resolved
	parts.Errors = append(resolutionErrors, parts.Errors...)
	raw, err := json.Marshal(parts)
	if err != nil {
		return nil, status.Error(codes.Internal, "encode parts")
	}
	out := &pb.SearchPartsResponse{}
	if err = protojson.Unmarshal(raw, out); err != nil {
		return nil, status.Error(codes.Internal, "encode parts")
	}
	return out, nil
}
func (c *Controller) WithSearcher(s *providerkit.Searcher) *Controller { c.searcher = s; return c }

func withoutHotelPlaceholder(issues []providerkit.ProviderError) []providerkit.ProviderError {
	out := make([]providerkit.ProviderError, 0, len(issues))
	for _, issue := range issues {
		if issue.Provider == "hotel" && issue.Code == "not_configured" {
			continue
		}
		out = append(out, issue)
	}
	return out
}
