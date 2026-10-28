package provider

import (
	"context"

	"github.com/godnota/provider-service/internal/entity"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
)

type Repository interface {
	Search(context.Context, providerkit.TripRequest, providerkit.ResolvedTrip, string) (entity.Result, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Search(ctx context.Context, trip providerkit.TripRequest, resolved providerkit.ResolvedTrip, kind string) (entity.Result, error) {
	return s.repo.Search(ctx, trip, resolved, kind)
}
