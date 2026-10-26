package mapping

import (
	"context"

	usecase "github.com/godnota/map-service/internal/usecase/mapping"
	pb "github.com/godnota/pkg/generated/map/v1"
)

type mapService interface {
	MapToProviderConfigs(context.Context) (usecase.Result, error)
	MapRoutesToFrontend(context.Context) (usecase.Result, error)
}

type Controller struct {
	service mapService
}

func New(service mapService) *Controller { return &Controller{service: service} }

var _ pb.MapServiceServer = (*Controller)(nil)

func (c *Controller) MapToProviderConfigs(ctx context.Context, _ *pb.MapToProviderConfigsRequest) (*pb.MapToProviderConfigsResponse, error) {
	_, _ = c.service.MapToProviderConfigs(ctx)
	return &pb.MapToProviderConfigsResponse{}, nil
}

func (c *Controller) MapRoutesToFrontend(ctx context.Context, _ *pb.MapRoutesToFrontendRequest) (*pb.MapRoutesToFrontendResponse, error) {
	_, _ = c.service.MapRoutesToFrontend(ctx)
	return &pb.MapRoutesToFrontendResponse{}, nil
}
