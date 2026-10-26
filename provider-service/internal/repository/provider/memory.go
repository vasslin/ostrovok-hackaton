package provider

import (
	"context"

	"github.com/godnota/provider-service/internal/entity"
)

// Memory is an in-memory stub. External provider adapters will replace it later.
type Memory struct{}

func New() *Memory { return &Memory{} }

func (m *Memory) Search(_ context.Context, _ string) ([]entity.Offer, error) {
	return nil, nil
}
