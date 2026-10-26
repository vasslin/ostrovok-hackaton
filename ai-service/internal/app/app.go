package app

import (
	"github.com/godnota/ai-service/internal/config"
	aictrl "github.com/godnota/ai-service/internal/controller/ai"
	airepo "github.com/godnota/ai-service/internal/repository/ai"
	aiuc "github.com/godnota/ai-service/internal/usecase/ai"
	pb "github.com/godnota/pkg/generated/ai/v1"
	"github.com/godnota/pkg/grpcapp"
	"google.golang.org/grpc"
)

func New(cfg *config.Config) (*grpcapp.App, error) {
	controller := aictrl.New(aiuc.New(airepo.New()))
	return grpcapp.New(cfg.GRPC.Port, nil, func(server *grpc.Server) {
		pb.RegisterAIServiceServer(server, controller)
	}), nil
}
