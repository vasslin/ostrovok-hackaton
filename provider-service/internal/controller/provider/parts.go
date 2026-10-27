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
