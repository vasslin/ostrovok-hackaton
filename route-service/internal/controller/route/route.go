package route

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/godnota/pkg/generated/route/v1"
	"github.com/godnota/pkg/tripconfig"
	"github.com/godnota/route-service/internal/entity"
	"github.com/godnota/route-service/internal/routebuilder"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
	"time"
)

type routeService interface {
	SearchRoutes(context.Context) ([]entity.Variant, error)
}

type Controller struct {
	service routeService
}

func New(service routeService) *Controller { return &Controller{service: service} }

var _ pb.RouteServiceServer = (*Controller)(nil)

func (c *Controller) BuildRoutes(ctx context.Context, req *pb.BuildRoutesRequest) (*pb.BuildRoutesResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}
	if req.GetSessionId() == "" || req.GetConfig() == nil || req.GetParts() == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, config and parts required")
	}
	trip, _, err := tripconfig.Decode(req.GetConfig(), nil)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid trip config")
	}
	if err := providerkit.ValidateSearch(trip, time.Now()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid trip config: %v", err)
	}
	candidates, err := routebuilder.Build(ctx, trip, req.GetParts())
	if err != nil {
		if ctx.Err() != nil {
			return nil, status.FromContextError(ctx.Err()).Err()
		}
		return nil, status.Error(codes.InvalidArgument, "cannot assemble routes")
	}
	return &pb.BuildRoutesResponse{Candidates: candidates}, nil
}

func (c *Controller) SearchRoutes(ctx context.Context, _ *pb.SearchRoutesRequest) (*pb.SearchRoutesResponse, error) {
	_, _ = c.service.SearchRoutes(ctx)
	return &pb.SearchRoutesResponse{}, nil
}
