package provider

import (
	"context"
	"log/slog"

	"github.com/godnota/provider-service/internal/entity"
)

type Repository interface {
	Search(ctx context.Context, kind string) ([]entity.Offer, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) SearchTransport(ctx context.Context) ([]entity.Offer, error) {
	slog.Info("provider-service: search transport")
	offers, _ := s.repo.Search(ctx, "transport")
	return offers, nil
}

func (s *Service) SearchHotels(ctx context.Context) ([]entity.Offer, error) {
	slog.Info("provider-service: search hotels")
	offers, _ := s.repo.Search(ctx, "hotel")
	return offers, nil
}

func (s *Service) SearchActivities(ctx context.Context) ([]entity.Offer, error) {
	slog.Info("provider-service: search activities")
	offers, _ := s.repo.Search(ctx, "activity")
	return offers, nil
}
