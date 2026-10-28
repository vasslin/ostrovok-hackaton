package travel

import (
	"context"

	"github.com/godnota/travel-service/internal/entity"
)

// Memory is an in-memory stub. It keeps the port in place until persistence is added.
type Memory struct{}

func New() *Memory { return &Memory{} }

func (m *Memory) Save(_ context.Context, _ entity.Plan) error { return nil }
