package app

import (
	"github.com/godnota/pkg/generated/provider/v1"
	"github.com/godnota/pkg/grpcapp"
	"github.com/godnota/provider-service/internal/config"
	providerctrl "github.com/godnota/provider-service/internal/controller/provider"
	providerrepo "github.com/godnota/provider-service/internal/repository/provider"
	provideruc "github.com/godnota/provider-service/internal/usecase/provider"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
	"google.golang.org/grpc"
)

func New(cfg *config.Config) (*grpcapp.App, error) {
	var registry providerkit.Registry
	var err error
	if cfg.Mode == "live" {
		registry, err = providerkit.LoadRegistry(cfg.RegistryFile)
		if err != nil {
			return nil, err
		}
	}
	searcher, err := providerkit.NewSearcher(registry, cfg.Mode)
	if err != nil {
		return nil, err
	}
	controller := providerctrl.New(provideruc.New(providerrepo.New())).WithSearcher(searcher)
	return grpcapp.New(cfg.GRPC.Port, nil, func(server *grpc.Server) {
		providerv1.RegisterProviderServiceServer(server, controller)
	}), nil
}
