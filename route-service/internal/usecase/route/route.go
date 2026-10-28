package route

import (
	"context"
	"log/slog"

	"github.com/godnota/route-service/internal/entity"
)

type Repository interface {
	Save(ctx context.Context, operation string) error
}

type ProviderPort interface {
	SearchTransport(ctx context.Context) error
	SearchHotels(ctx context.Context) error
	SearchActivities(ctx context.Context) error
}

type Service struct {
	repo      Repository
	providers ProviderPort
}

func New(repo Repository, providers ProviderPort) *Service {
	return &Service{repo: repo, providers: providers}
}

func (s *Service) SearchRoutes(ctx context.Context) ([]entity.Variant, error) {
	slog.Info("route-service: search routes")
	_ = s.repo.Save(ctx, "SearchRoutes")
	s.call(ctx, "provider.SearchTransport", s.providers.SearchTransport)
	s.call(ctx, "provider.SearchHotels", s.providers.SearchHotels)
	s.call(ctx, "provider.SearchActivities", s.providers.SearchActivities)
	return nil, nil
}

func (s *Service) call(ctx context.Context, name string, fn func(context.Context) error) {
	if err := fn(ctx); err != nil {
		slog.Error("route-service: downstream call failed", "call", name, "error", err)
	}
}
