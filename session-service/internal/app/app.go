package app

import (
	pb "github.com/godnota/pkg/generated/session/v1"
	"github.com/godnota/pkg/grpcapp"
	"github.com/godnota/session-service/internal/config"
	sessionctrl "github.com/godnota/session-service/internal/controller/session"
	sessionrepo "github.com/godnota/session-service/internal/repository/session"
	sessionuc "github.com/godnota/session-service/internal/usecase/session"
	"google.golang.org/grpc"
)

func New(cfg *config.Config) (*grpcapp.App, error) {
	service := sessionuc.New(sessionrepo.New())
	controller := sessionctrl.New(service)
	return grpcapp.New(cfg.GRPC.Port, nil, func(server *grpc.Server) {
		pb.RegisterSessionServiceServer(server, controller)
	}), nil
}
