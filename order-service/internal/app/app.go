package app

import (
	"github.com/godnota/order-service/internal/config"
	orderctrl "github.com/godnota/order-service/internal/controller/order"
	orderrepo "github.com/godnota/order-service/internal/repository/order"
	orderuc "github.com/godnota/order-service/internal/usecase/order"
	pb "github.com/godnota/pkg/generated/order/v1"
	"github.com/godnota/pkg/grpcapp"
	"google.golang.org/grpc"
)

func New(cfg *config.Config) (*grpcapp.App, error) {
	service := orderuc.New(orderrepo.New())
	controller := orderctrl.New(service)
	return grpcapp.New(cfg.GRPC.Port, nil, func(server *grpc.Server) {
		pb.RegisterOrderServiceServer(server, controller)
	}), nil
}
