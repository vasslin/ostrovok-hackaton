package route

import (
	"context"

	pb "github.com/godnota/pkg/generated/route/v1"
	"github.com/godnota/route-service/internal/entity"
	"github.com/godnota/route-service/internal/join"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type routeService interface {
	SearchRoutes(context.Context) ([]entity.Variant, error)
}

// Ranker sends the assembled routes to TravelAI.RankRouteParts.
type Ranker interface {
	RankRouteParts(context.Context, string, *structpb.Struct, []*structpb.Struct) ([]*structpb.Struct, error)
}

type Controller struct {
	service routeService
	rank    Ranker
	k       int
	m       int
}

func New(service routeService) *Controller {
	return &Controller{service: service, k: 4, m: 3}
}

func (c *Controller) WithRanker(rank Ranker, k, m int) *Controller {
	c.rank = rank
	if k > 0 {
		c.k = k
	}
	if m > 0 {
		c.m = m
	}
	return c
}

var _ pb.RouteServiceServer = (*Controller)(nil)

func (c *Controller) BuildRoutes(ctx context.Context, req *pb.BuildRoutesRequest) (*pb.BuildRoutesResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}
	if req.GetSessionId() == "" || req.GetConfig() == nil || req.GetParts() == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, config and parts required")
	}
	joins, err := join.Assemble(req.GetParts(), originCity(req.GetConfig()), c.k)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	routes, err := join.Materialize(joins, c.m)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	payload := make([]*structpb.Struct, 0, len(routes))
	for _, route := range routes {
		document, err := route.Proto()
		if err != nil {
			return nil, status.Error(codes.Internal, "route document is invalid")
		}
		payload = append(payload, document)
	}
	if len(payload) == 0 {
		return &pb.BuildRoutesResponse{}, nil
	}
	if c.rank == nil {
		return nil, status.Error(codes.FailedPrecondition, "ai-service is not configured")
	}
	ranked, err := c.rank.RankRouteParts(ctx, req.GetSessionId(), req.GetConfig(), payload)
	if err != nil {
		return nil, err
	}
	return &pb.BuildRoutesResponse{Candidates: ranked}, nil
}

func originCity(config *structpb.Struct) string {
	origin := config.GetFields()["origin"].GetStructValue()
	if origin == nil {
		return ""
	}
	return origin.GetFields()["city"].GetStringValue()
}

func (c *Controller) SearchRoutes(ctx context.Context, _ *pb.SearchRoutesRequest) (*pb.SearchRoutesResponse, error) {
	_, _ = c.service.SearchRoutes(ctx)
	return &pb.SearchRoutesResponse{}, nil
}
