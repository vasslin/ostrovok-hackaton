package app

import (
	"fmt"

	aiv1 "github.com/godnota/pkg/generated/ai/v1"
	mapv1 "github.com/godnota/pkg/generated/map/v1"
	routev1 "github.com/godnota/pkg/generated/route/v1"
	sessionv1 "github.com/godnota/pkg/generated/session/v1"
	travelv1 "github.com/godnota/pkg/generated/travel/v1"
	"github.com/godnota/pkg/grpcapp"
	aigrpc "github.com/godnota/travel-service/internal/adapter/ai/grpc"
	mapgrpc "github.com/godnota/travel-service/internal/adapter/map/grpc"
	routegrpc "github.com/godnota/travel-service/internal/adapter/route/grpc"
	sessiongrpc "github.com/godnota/travel-service/internal/adapter/session/grpc"
	"github.com/godnota/travel-service/internal/config"
	travelctrl "github.com/godnota/travel-service/internal/controller/travel"
	travelrepo "github.com/godnota/travel-service/internal/repository/travel"
	traveluc "github.com/godnota/travel-service/internal/usecase/travel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func New(cfg *config.Config) (*grpcapp.App, error) {
	sessionConn, err := dial(cfg.Clients.SessionAddr)
	if err != nil {
		return nil, fmt.Errorf("dial session-service: %w", err)
	}
	aiConn, err := dial(cfg.Clients.AIAddr)
	if err != nil {
		_ = sessionConn.Close()
		return nil, fmt.Errorf("dial ai-service: %w", err)
	}
	mapConn, err := dial(cfg.Clients.MapAddr)
	if err != nil {
		_ = sessionConn.Close()
		_ = aiConn.Close()
		return nil, fmt.Errorf("dial map-service: %w", err)
	}
	routeConn, err := dial(cfg.Clients.RouteAddr)
	if err != nil {
		_ = sessionConn.Close()
		_ = aiConn.Close()
		_ = mapConn.Close()
		return nil, fmt.Errorf("dial route-service: %w", err)
	}

	service := traveluc.New(
		travelrepo.New(),
		sessiongrpc.New(sessionv1.NewSessionServiceClient(sessionConn)),
		aigrpc.New(aiv1.NewAIServiceClient(aiConn)),
		mapgrpc.New(mapv1.NewMapServiceClient(mapConn)),
		routegrpc.New(routev1.NewRouteServiceClient(routeConn)),
	)

	closeFn := func() {
		_ = sessionConn.Close()
		_ = aiConn.Close()
		_ = mapConn.Close()
		_ = routeConn.Close()
	}

	return grpcapp.New(cfg.GRPC.Port, closeFn, func(server *grpc.Server) {
		travelv1.RegisterTravelServiceServer(server, travelctrl.New(service))
	}), nil
}

func dial(target string) (*grpc.ClientConn, error) {
	return grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
