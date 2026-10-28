package storage

import (
	"context"
	"fmt"

	"github.com/godnota/session-service/internal/config"
	historymem "github.com/godnota/session-service/internal/repository/history/memory"
	historypg "github.com/godnota/session-service/internal/repository/history/postgres"
	"github.com/godnota/session-service/internal/repository/pgdb"
	configmem "github.com/godnota/session-service/internal/repository/tripconfig/memory"
	configpg "github.com/godnota/session-service/internal/repository/tripconfig/postgres"
	"github.com/godnota/session-service/internal/usecase/session"
)

// Open selects the chat-history and trip-config adapters.
func Open(ctx context.Context, cfg *config.Config) (session.HistoryRepository, session.ConfigRepository, func(), error) {
	switch cfg.Repository.Driver {
	case "", "memory":
		return historymem.New(), configmem.New(), func() {}, nil
	case "postgres":
		pool, err := pgdb.Connect(ctx, cfg.Postgres.DSN)
		if err != nil {
			return nil, nil, nil, err
		}
		return historypg.New(pool), configpg.New(pool), pool.Close, nil
	default:
		return nil, nil, nil, fmt.Errorf("unknown SESSION_REPOSITORY %q", cfg.Repository.Driver)
	}
}
