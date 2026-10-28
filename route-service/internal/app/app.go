package app

import (
	"fmt"

	aiv1 "github.com/godnota/pkg/generated/ai/v1"
	providerv1 "github.com/godnota/pkg/generated/provider/v1"
	routepb "github.com/godnota/pkg/generated/route/v1"
	"github.com/godnota/pkg/grpcapp"
	aigrpc "github.com/godnota/route-service/internal/adapter/ai/grpc"
	providergrpc "github.com/godnota/route-service/internal/adapter/provider/grpc"
	"github.com/godnota/route-service/internal/config"
	routectrl "github.com/godnota/route-service/internal/controller/route"
	routerepo "github.com/godnota/route-service/internal/repository/route"
	routeuc "github.com/godnota/route-service/internal/usecase/route"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func New(cfg *config.Config) (*grpcapp.App, error) {
	providerConn, err := grpc.NewClient(cfg.Clients.ProviderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial provider-service: %w", err)
	}
	aiConn, err := grpc.NewClient(cfg.Clients.AIAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = providerConn.Close()
		return nil, fmt.Errorf("dial ai-service: %w", err)
	}

	service := routeuc.New(routerepo.New(), providergrpc.New(providerv1.NewProviderServiceClient(providerConn)))
	controller := routectrl.New(service).WithRanker(aigrpc.New(aiv1.NewTravelAIClient(aiConn)), cfg.Join.MaxJoins, cfg.Join.RoutesPerJoin)
	return grpcapp.New(cfg.GRPC.Port, func() {
		_ = providerConn.Close()
		_ = aiConn.Close()
	}, func(server *grpc.Server) {
		routepb.RegisterRouteServiceServer(server, controller)
	}), nil
}
