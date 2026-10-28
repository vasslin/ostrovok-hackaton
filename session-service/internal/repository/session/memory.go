package session

import (
	"context"
	"crypto/rand"
	commonv1 "github.com/godnota/pkg/generated/common/v1"
	"github.com/godnota/session-service/internal/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

type Memory struct {
	mu       sync.RWMutex
	sessions map[string]entity.Session
}

func New() *Memory { return &Memory{sessions: map[string]entity.Session{}} }
func clone(s entity.Session) entity.Session {
	if s.Config != nil {
		s.Config = proto.Clone(s.Config).(*commonv1.TripConfig)
	}
	s.Messages = append([]entity.Message(nil), s.Messages...)
	return s
}
func (m *Memory) Create(ctx context.Context, s entity.Session) (entity.Session, error) {
	if err := ctx.Err(); err != nil {
		return entity.Session{}, err
	}
	if s.UserID == "" {
		return entity.Session{}, status.Error(codes.InvalidArgument, "user_id required")
	}
	s.ID = rand.Text()
	s.Status = "draft"
	s.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	s.UpdatedAt = s.CreatedAt
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[s.ID] = clone(s)
	return clone(s), nil
}
func (m *Memory) Find(ctx context.Context, id string) (entity.Session, error) {
	if err := ctx.Err(); err != nil {
		return entity.Session{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	if !ok {
		return s, status.Error(codes.NotFound, "session not found")
	}
	return clone(s), nil
}
func (m *Memory) Save(ctx context.Context, s entity.Session) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	old, ok := m.sessions[s.ID]
	if !ok {
		return status.Error(codes.NotFound, "session not found")
	}
	old.Summary = s.Summary
	if s.Status != "" {
		old.Status = s.Status
	}
	if s.Config != nil {
		old.Config = s.Config
	}
	if s.Messages != nil {
		old.Messages = s.Messages
	}
	old.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	m.sessions[s.ID] = clone(old)
	return nil
}
