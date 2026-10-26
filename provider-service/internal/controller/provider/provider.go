package provider

import (
	"context"

	pb "github.com/godnota/pkg/generated/provider/v1"
	"github.com/godnota/provider-service/internal/entity"
)

type providerService interface {
	SearchTransport(context.Context) ([]entity.Offer, error)
	SearchHotels(context.Context) ([]entity.Offer, error)
	SearchActivities(context.Context) ([]entity.Offer, error)
}

type Controller struct {
	service providerService
}

func New(service providerService) *Controller { return &Controller{service: service} }

var _ pb.ProviderServiceServer = (*Controller)(nil)

func (c *Controller) SearchTransport(ctx context.Context, _ *pb.SearchTransportRequest) (*pb.SearchTransportResponse, error) {
	_, _ = c.service.SearchTransport(ctx)
	return &pb.SearchTransportResponse{}, nil
}

func (c *Controller) SearchHotels(ctx context.Context, _ *pb.SearchHotelsRequest) (*pb.SearchHotelsResponse, error) {
	_, _ = c.service.SearchHotels(ctx)
	return &pb.SearchHotelsResponse{}, nil
}

func (c *Controller) SearchActivities(ctx context.Context, _ *pb.SearchActivitiesRequest) (*pb.SearchActivitiesResponse, error) {
	_, _ = c.service.SearchActivities(ctx)
	return &pb.SearchActivitiesResponse{}, nil
}
