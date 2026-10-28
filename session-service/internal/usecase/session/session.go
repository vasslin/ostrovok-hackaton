package session

import (
	"context"
	"log/slog"

	"github.com/godnota/session-service/internal/entity"
)

type Repository interface {
	Create(ctx context.Context, session entity.Session) (entity.Session, error)
	Find(ctx context.Context, sessionID string) (entity.Session, error)
	Save(ctx context.Context, session entity.Session) error
}

type Service struct {
	repo          Repository
	history       HistoryRepository
	configs       ConfigRepository
	previousLimit int
}

func New(repo Repository, history HistoryRepository, configs ConfigRepository) *Service {
	return &Service{repo: repo, history: history, configs: configs, previousLimit: defaultPreviousChats}
}

// WithPreviousLimit sets how many earlier chat summaries GetAgentContext returns when the request limit is 0.
func (s *Service) WithPreviousLimit(n int) *Service {
	if n > 0 {
		s.previousLimit = n
	}
	return s
}

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
