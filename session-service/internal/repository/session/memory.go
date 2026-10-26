package session

import (
	"context"

	"github.com/godnota/session-service/internal/entity"
)

// Memory is an in-memory stub. Persistence will replace it later.
type Memory struct{}

func New() *Memory { return &Memory{} }

func (m *Memory) Create(_ context.Context, _ entity.Session) (entity.Session, error) {
	return entity.Session{}, nil
}

func (m *Memory) Find(_ context.Context, _ string) (entity.Session, error) {
	return entity.Session{}, nil
}

func (m *Memory) Save(_ context.Context, _ entity.Session) error {
	return nil
}
