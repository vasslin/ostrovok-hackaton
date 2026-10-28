package mapping

import (
	"context"
	"log/slog"
)

type Repository interface {
	Save(ctx context.Context, operation string) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service { return &Service{repo: repo} }

type Result struct{}

func (s *Service) MapToProviderConfigs(ctx context.Context) (Result, error) {
	slog.Info("map-service: map trip config to provider configs")
	_ = s.repo.Save(ctx, "MapToProviderConfigs")
	return Result{}, nil
}

func (s *Service) MapRoutesToFrontend(ctx context.Context) (Result, error) {
	slog.Info("map-service: map routes to frontend dto")
	_ = s.repo.Save(ctx, "MapRoutesToFrontend")
	return Result{}, nil
}
