package user

import (
	"context"
	"log/slog"

	"github.com/godnota/user-service/internal/entity"
)

type Repository interface {
	Find(ctx context.Context, userID string) (entity.User, error)
	Save(ctx context.Context, user entity.User) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service { return &Service{repo: repo} }

type GetRequest struct {
	UserID string
}

type PreferencesRequest struct {
	UserID   string
	Language string
	Currency string
}

func (s *Service) GetUser(ctx context.Context, req GetRequest) (entity.User, error) {
	slog.Info("user-service: get user", "user_id", req.UserID)
	_, _ = s.repo.Find(ctx, req.UserID)
	return entity.User{}, nil
}

func (s *Service) UpdatePreferences(ctx context.Context, req PreferencesRequest) error {
	slog.Info("user-service: update preferences", "user_id", req.UserID, "language", req.Language, "currency", req.Currency)
	return s.repo.Save(ctx, entity.User{ID: req.UserID, Language: req.Language, Currency: req.Currency})
}
