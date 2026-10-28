package mapping

import (
	"context"
	"encoding/json"
	providerv1 "github.com/godnota/pkg/generated/provider/v1"
	"github.com/godnota/pkg/tripconfig"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"time"

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

func (c *Controller) MapToProviderConfigs(ctx context.Context, req *pb.MapToProviderConfigsRequest) (*pb.MapToProviderConfigsResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}
	input, resolved, err := tripconfig.Decode(req.GetConfig(), req.GetResolved())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	mapped, err := providerkit.MapConfig(input, resolved, time.Now())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	out := &pb.MapToProviderConfigsResponse{Missing: mapped.Missing, Limitations: mapped.Limitations}
	for _, cfg := range mapped.Configs {
		raw, e := json.Marshal(cfg.Params)
		if e != nil {
			return nil, status.Error(codes.Internal, "encode parameters")
		}
		var plain map[string]any
		if json.Unmarshal(raw, &plain) != nil {
			return nil, status.Error(codes.Internal, "encode parameters")
		}
		params, e := structpb.NewStruct(plain)
		if e != nil {
			return nil, status.Error(codes.Internal, "encode parameters")
		}
		out.Configs = append(out.Configs, &providerv1.ProviderConfig{Provider: cfg.Provider, Endpoint: cfg.Endpoint, Params: params})
	}
	return out, nil
}

func (c *Controller) MapRoutesToFrontend(ctx context.Context, _ *pb.MapRoutesToFrontendRequest) (*pb.MapRoutesToFrontendResponse, error) {
	_, _ = c.service.MapRoutesToFrontend(ctx)
	return &pb.MapRoutesToFrontendResponse{}, nil
}
