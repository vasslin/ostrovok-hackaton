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
	parts, err := c.searcher.SearchParts(ctx, input, resolved)
	if err != nil {
		if ctx.Err() != nil {
			return nil, status.FromContextError(ctx.Err()).Err()
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if c.searcher.Mode == "live" && c.searcher.Hotels == nil && c.service != nil && hotelProviderEnabled() && (input.Hotel == nil || input.Hotel.Needed == nil || *input.Hotel.Needed) {
		if len(input.Stops) > 0 {
			parts = c.addItineraryHotels(ctx, input, resolved, parts)
		} else {
			parts = c.addHotelParts(ctx, input, resolved, parts)
		}
		if ctx.Err() != nil {
			return nil, status.FromContextError(ctx.Err()).Err()
		}
	}
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
