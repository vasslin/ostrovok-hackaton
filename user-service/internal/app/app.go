package app

import (
	pb "github.com/godnota/pkg/generated/user/v1"
	"github.com/godnota/pkg/grpcapp"
	"github.com/godnota/user-service/internal/config"
	userctrl "github.com/godnota/user-service/internal/controller/user"
	userrepo "github.com/godnota/user-service/internal/repository/user"
	useruc "github.com/godnota/user-service/internal/usecase/user"
	"google.golang.org/grpc"
)

func New(cfg *config.Config) (*grpcapp.App, error) {
	service := useruc.New(userrepo.New())
	controller := userctrl.New(service)
	return grpcapp.New(cfg.GRPC.Port, nil, func(server *grpc.Server) {
		pb.RegisterUserServiceServer(server, controller)
	}), nil
}
