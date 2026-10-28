package app

import (
	"fmt"

	providerv1 "github.com/godnota/pkg/generated/provider/v1"
	routepb "github.com/godnota/pkg/generated/route/v1"
	"github.com/godnota/pkg/grpcapp"
	providergrpc "github.com/godnota/route-service/internal/adapter/provider/grpc"
	"github.com/godnota/route-service/internal/config"
	routectrl "github.com/godnota/route-service/internal/controller/route"
	routerepo "github.com/godnota/route-service/internal/repository/route"
	routeuc "github.com/godnota/route-service/internal/usecase/route"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func New(cfg *config.Config) (*grpcapp.App, error) {
	conn, err := grpc.NewClient(cfg.Clients.ProviderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial provider-service: %w", err)
	}

	service := routeuc.New(routerepo.New(), providergrpc.New(providerv1.NewProviderServiceClient(conn)))
	controller := routectrl.New(service)
	return grpcapp.New(cfg.GRPC.Port, func() { _ = conn.Close() }, func(server *grpc.Server) {
		routepb.RegisterRouteServiceServer(server, controller)
	}), nil
}
