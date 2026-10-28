package route

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/godnota/pkg/generated/route/v1"
	"github.com/godnota/route-service/internal/entity"
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
	return nil, status.Error(codes.Unimplemented, "stage-3 route assembly is not implemented")
}

func (c *Controller) SearchRoutes(ctx context.Context, _ *pb.SearchRoutesRequest) (*pb.SearchRoutesResponse, error) {
	_, _ = c.service.SearchRoutes(ctx)
	return &pb.SearchRoutesResponse{}, nil
}
