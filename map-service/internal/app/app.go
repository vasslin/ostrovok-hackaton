package app

import (
	"github.com/godnota/map-service/internal/config"
	mapctrl "github.com/godnota/map-service/internal/controller/mapping"
	maprepo "github.com/godnota/map-service/internal/repository/mapping"
	mapuc "github.com/godnota/map-service/internal/usecase/mapping"
	pb "github.com/godnota/pkg/generated/map/v1"
	"github.com/godnota/pkg/grpcapp"
	"google.golang.org/grpc"
)

func New(cfg *config.Config) (*grpcapp.App, error) {
	controller := mapctrl.New(mapuc.New(maprepo.New()))
	return grpcapp.New(cfg.GRPC.Port, nil, func(server *grpc.Server) {
		pb.RegisterMapServiceServer(server, controller)
	}), nil
}
