package app

import (
	"github.com/godnota/pkg/generated/provider/v1"
	"github.com/godnota/pkg/grpcapp"
	"github.com/godnota/provider-service/internal/config"
	providerctrl "github.com/godnota/provider-service/internal/controller/provider"
	providerrepo "github.com/godnota/provider-service/internal/repository/provider"
	provideruc "github.com/godnota/provider-service/internal/usecase/provider"
	"google.golang.org/grpc"
)

func New(cfg *config.Config) (*grpcapp.App, error) {
	controller := providerctrl.New(provideruc.New(providerrepo.New()))
	return grpcapp.New(cfg.GRPC.Port, nil, func(server *grpc.Server) {
		providerv1.RegisterProviderServiceServer(server, controller)
	}), nil
}
