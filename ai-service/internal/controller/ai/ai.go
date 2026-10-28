package ai

import (
	"context"
	"log/slog"

	usecase "github.com/godnota/ai-service/internal/usecase/ai"
	pb "github.com/godnota/pkg/generated/ai/v1"
)

type aiService interface {
	GenerateOrUpdateConfig(context.Context) (usecase.Result, error)
	UpdateConfig(context.Context) (usecase.Result, error)
	RefineConfig(context.Context) (usecase.Result, error)
	RankRoutes(context.Context) (usecase.Result, error)
}

type Controller struct {
	service aiService
}

func New(service aiService) *Controller { return &Controller{service: service} }

var _ pb.AIServiceServer = (*Controller)(nil)

func (c *Controller) GenerateOrUpdateConfig(ctx context.Context, _ *pb.GenerateOrUpdateConfigRequest) (*pb.GenerateOrUpdateConfigResponse, error) {
	_, _ = c.service.GenerateOrUpdateConfig(ctx)
	return &pb.GenerateOrUpdateConfigResponse{}, nil
}

func (c *Controller) UpdateConfig(ctx context.Context, _ *pb.UpdateConfigRequest) (*pb.UpdateConfigResponse, error) {
	_, _ = c.service.UpdateConfig(ctx)
	return &pb.UpdateConfigResponse{}, nil
}

func (c *Controller) RefineConfig(ctx context.Context, _ *pb.RefineConfigRequest) (*pb.RefineConfigResponse, error) {
	_, _ = c.service.RefineConfig(ctx)
	return &pb.RefineConfigResponse{}, nil
}

func (c *Controller) RankRoutes(ctx context.Context, _ *pb.RankRoutesRequest) (*pb.RankRoutesResponse, error) {
	_, _ = c.service.RankRoutes(ctx)
	return &pb.RankRoutesResponse{}, nil
}

// TravelAI is a stub of the external model contract.
type TravelAI struct{}

func NewTravelAI() *TravelAI { return &TravelAI{} }

var _ pb.TravelAIServer = (*TravelAI)(nil)

func (TravelAI) ExtractConfig(context.Context, *pb.ExtractConfigRequest) (*pb.ExtractConfigResponse, error) {
	slog.Info("ai-service: extract config")
	return &pb.ExtractConfigResponse{}, nil
}

func (TravelAI) RankRouteParts(_ context.Context, req *pb.RankRoutePartsRequest) (*pb.RankRoutePartsResponse, error) {
	slog.Info("ai-service: rank route parts", "session_id", req.GetSessionId(), "routes", len(req.GetRoutes()))
	return &pb.RankRoutePartsResponse{}, nil
}
