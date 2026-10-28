package ai

import "context"

// Memory is an in-memory stub. Model calls will replace it later.
type Memory struct{}

func New() *Memory { return &Memory{} }

func (m *Memory) Save(_ context.Context, _ string) error { return nil }
