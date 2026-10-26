package ai

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

func (s *Service) GenerateOrUpdateConfig(ctx context.Context) (Result, error) {
	slog.Info("ai-service: generate or update config")
	_ = s.repo.Save(ctx, "GenerateOrUpdateConfig")
	return Result{}, nil
}

func (s *Service) UpdateConfig(ctx context.Context) (Result, error) {
	slog.Info("ai-service: update config")
	_ = s.repo.Save(ctx, "UpdateConfig")
	return Result{}, nil
}

func (s *Service) RefineConfig(ctx context.Context) (Result, error) {
	slog.Info("ai-service: refine config")
	_ = s.repo.Save(ctx, "RefineConfig")
	return Result{}, nil
}

func (s *Service) RankRoutes(ctx context.Context) (Result, error) {
	slog.Info("ai-service: rank routes")
	_ = s.repo.Save(ctx, "RankRoutes")
	return Result{}, nil
}
