package postgres

import (
	"context"
	"errors"

	commonv1 "github.com/godnota/pkg/generated/common/v1"
	"github.com/godnota/session-service/internal/usecase/session"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Config stores one trip config per chat in PostgreSQL.
type Config struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Config { return &Config{pool: pool} }

var _ session.ConfigRepository = (*Config)(nil)

func (c *Config) Save(ctx context.Context, userID, chatID string, cfg *commonv1.TripConfig) error {
	raw, err := protojson.Marshal(cfg)
	if err != nil {
		return err
	}
	var saved string
	err = c.pool.QueryRow(ctx, `
		INSERT INTO trip_configs (chat_id, user_id, config, updated_at)
		VALUES ($1, $2, $3::jsonb, now())
		ON CONFLICT (chat_id) DO UPDATE
		SET config = EXCLUDED.config, updated_at = now()
		WHERE trip_configs.user_id = EXCLUDED.user_id
		RETURNING chat_id`, chatID, userID, string(raw)).Scan(&saved)
	if errors.Is(err, pgx.ErrNoRows) {
		return session.ErrForbidden
	}
	return err
}

func (c *Config) Find(ctx context.Context, userID, chatID string) (*commonv1.TripConfig, error) {
	var owner string
	var raw []byte
	err := c.pool.QueryRow(ctx, `SELECT user_id, config FROM trip_configs WHERE chat_id = $1`, chatID).Scan(&owner, &raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, session.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if owner != userID {
		return nil, session.ErrForbidden
	}
	cfg := &commonv1.TripConfig{}
	if err = protojson.Unmarshal(raw, cfg); err != nil {
		return nil, err
	}
	return proto.Clone(cfg).(*commonv1.TripConfig), nil
}
