package app

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/godnota/pkg/generated/provider/v1"
	"github.com/godnota/provider-service/internal/config"
	providerctrl "github.com/godnota/provider-service/internal/controller/provider"
	providerrepo "github.com/godnota/provider-service/internal/repository/provider"
	provideruc "github.com/godnota/provider-service/internal/usecase/provider"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
	"google.golang.org/grpc"
)

type App struct {
	server *grpc.Server
	host   string
	port   string
}

func New(cfg *config.Config) (*App, error) {
	registry, err := providerkit.LoadRegistry(cfg.RegistryPath)
	if err != nil {
		return nil, err
	}
	repo, err := providerrepo.New(registry, cfg.POIPath)
	if err != nil {
		return nil, err
	}
	controller := providerctrl.New(provideruc.New(repo))
	server := grpc.NewServer()
	providerv1.RegisterProviderServiceServer(server, controller)
	return &App{server: server, host: cfg.GRPC.Host, port: cfg.GRPC.Port}, nil
}

func (a *App) Run() error {
	listener, err := net.Listen("tcp", net.JoinHostPort(a.host, a.port))
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}
	slog.Info("provider grpc server started", "port", a.port)
	return a.server.Serve(listener)
}

func (a *App) Shutdown() { a.server.GracefulStop() }
