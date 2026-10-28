package memory

import (
	"context"
	"sync"
	"time"

	commonv1 "github.com/godnota/pkg/generated/common/v1"
	"github.com/godnota/session-service/internal/usecase/session"
	"google.golang.org/protobuf/proto"
)

type record struct {
	userID    string
	config    *commonv1.TripConfig
	updatedAt time.Time
}

// Config is an in-memory store of one trip config per chat.
type Config struct {
	mu   sync.Mutex
	rows map[string]record
}

func New() *Config { return &Config{rows: map[string]record{}} }

var _ session.ConfigRepository = (*Config)(nil)

func (c *Config) Save(ctx context.Context, userID, chatID string, cfg *commonv1.TripConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if current, ok := c.rows[chatID]; ok && current.userID != userID {
		return session.ErrForbidden
	}
	c.rows[chatID] = record{
		userID:    userID,
		config:    proto.Clone(cfg).(*commonv1.TripConfig),
		updatedAt: time.Now().UTC(),
	}
	return nil
}

func (c *Config) Find(ctx context.Context, userID, chatID string) (*commonv1.TripConfig, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	current, ok := c.rows[chatID]
	if !ok {
		return nil, session.ErrNotFound
	}
	if current.userID != userID {
		return nil, session.ErrForbidden
	}
	return proto.Clone(current.config).(*commonv1.TripConfig), nil
}
