package route

import (
	"context"

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

func (c *Controller) SearchRoutes(ctx context.Context, _ *pb.SearchRoutesRequest) (*pb.SearchRoutesResponse, error) {
	_, _ = c.service.SearchRoutes(ctx)
	return &pb.SearchRoutesResponse{}, nil
}
