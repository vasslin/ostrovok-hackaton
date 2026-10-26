package user

import (
	"context"

	"github.com/godnota/user-service/internal/entity"
)

// Memory is an in-memory stub. Persistence will replace it later.
type Memory struct{}

func New() *Memory { return &Memory{} }

func (m *Memory) Find(_ context.Context, _ string) (entity.User, error) {
	return entity.User{}, nil
}

func (m *Memory) Save(_ context.Context, _ entity.User) error {
	return nil
}
