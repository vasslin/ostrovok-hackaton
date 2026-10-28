package app

import (
	"context"
	"fmt"

	pb "github.com/godnota/pkg/generated/session/v1"
	"github.com/godnota/pkg/grpcapp"
	"github.com/godnota/session-service/internal/adapter/storage"
	"github.com/godnota/session-service/internal/config"
	sessionctrl "github.com/godnota/session-service/internal/controller/session"
	sessionrepo "github.com/godnota/session-service/internal/repository/session"
	sessionuc "github.com/godnota/session-service/internal/usecase/session"
	"google.golang.org/grpc"
)

func New(cfg *config.Config) (*grpcapp.App, error) {
	history, configs, closeStorage, err := storage.Open(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("open session storage: %w", err)
	}
	service := sessionuc.New(sessionrepo.New(), history, configs).WithPreviousLimit(cfg.Context.PreviousChatsLimit)
	controller := sessionctrl.New(service)
	return grpcapp.New(cfg.GRPC.Port, closeStorage, func(server *grpc.Server) {
		pb.RegisterSessionServiceServer(server, controller)
	}), nil
}
