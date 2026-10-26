package order

import (
	"context"

	"github.com/godnota/order-service/internal/entity"
)

// Memory is an in-memory stub. Persistence will replace it later.
type Memory struct{}

func New() *Memory { return &Memory{} }

func (m *Memory) Create(_ context.Context, _ entity.Order) (entity.Order, error) {
	return entity.Order{}, nil
}

func (m *Memory) Find(_ context.Context, _ string) (entity.Order, error) {
	return entity.Order{}, nil
}

func (m *Memory) Save(_ context.Context, _ entity.Order) error {
	return nil
}
