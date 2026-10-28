package session

import (
	"context"
	"log/slog"

	commonv1 "github.com/godnota/pkg/generated/common/v1"
	"github.com/godnota/session-service/internal/entity"
)

type Repository interface {
	Create(ctx context.Context, session entity.Session) (entity.Session, error)
	Find(ctx context.Context, sessionID string) (entity.Session, error)
	Save(ctx context.Context, session entity.Session) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service { return &Service{repo: repo} }

type CreateRequest struct {
	UserID string
}

type CreateResponse struct {
	SessionID string
}

type GetRequest struct {
	SessionID string
}

type GetResponse struct {
	Session entity.Session
}

type SaveRequest struct {
	SessionID string
	Summary   string
	Status    string
	Config    *commonv1.TripConfig
	Messages  []entity.Message
}

func (s *Service) CreateSession(ctx context.Context, req CreateRequest) (CreateResponse, error) {
	slog.Info("session-service: create session", "user_id", req.UserID)
	created, err := s.repo.Create(ctx, entity.Session{UserID: req.UserID})
	return CreateResponse{SessionID: created.ID}, err
}

func (s *Service) GetSession(ctx context.Context, req GetRequest) (GetResponse, error) {
	slog.Info("session-service: get session", "session_id", req.SessionID)
	found, err := s.repo.Find(ctx, req.SessionID)
	return GetResponse{Session: found}, err
}

func (s *Service) GetContext(ctx context.Context, req GetRequest) (GetResponse, error) {
	slog.Info("session-service: get context", "session_id", req.SessionID)
	return s.GetSession(ctx, req)
}

func (s *Service) SaveContext(ctx context.Context, req SaveRequest) error {
	return s.repo.Save(ctx, entity.Session{ID: req.SessionID, Status: req.Status, Summary: req.Summary, Config: req.Config, Messages: req.Messages})
}
