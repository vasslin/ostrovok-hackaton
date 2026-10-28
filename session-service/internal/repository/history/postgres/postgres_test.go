package postgres

import (
	"context"
	"errors"
	"os"
	"testing"

	commonv1 "github.com/godnota/pkg/generated/common/v1"
	"github.com/godnota/session-service/internal/entity"
	"github.com/godnota/session-service/internal/repository/pgdb"
	configpg "github.com/godnota/session-service/internal/repository/tripconfig/postgres"
	"github.com/godnota/session-service/internal/usecase/session"
)

func TestPostgresHistoryAndConfig(t *testing.T) {
	dsn := os.Getenv("SESSION_POSTGRES_DSN")
	if dsn == "" || os.Getenv("SESSION_POSTGRES_TEST") != "1" {
		t.Skip("set SESSION_POSTGRES_TEST=1 and SESSION_POSTGRES_DSN to run")
	}
	ctx := context.Background()
	pool, err := pgdb.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err = pool.Exec(ctx, `TRUNCATE agent_messages, agent_chats, trip_configs`); err != nil {
		t.Fatal(err)
	}
	history := New(pool)
	configs := configpg.New(pool)
	if _, err = history.AppendMessages(ctx, "user", "chat", []entity.Message{{Role: "user", Text: "hello", CreatedAt: "2026-10-28T12:00:00Z"}}); err != nil {
		t.Fatal(err)
	}
	if err = history.UpdateSummary(ctx, "user", "chat", "trip"); err != nil {
		t.Fatal(err)
	}
	if err = history.UpdateSummary(ctx, "user", "older", "before"); err != nil {
		t.Fatal(err)
	}
	chat, err := history.Find(ctx, "user", "chat")
	if err != nil || len(chat.Messages) != 1 || chat.Summary != "trip" {
		t.Fatalf("%+v %v", chat, err)
	}
	previous, err := history.PreviousSummaries(ctx, "user", "chat", 4)
	if err != nil || len(previous) != 1 || previous[0].ChatID != "older" {
		t.Fatalf("%+v %v", previous, err)
	}
	if _, err = history.Find(ctx, "other", "chat"); !errors.Is(err, session.ErrForbidden) {
		t.Fatal(err)
	}
	if err = configs.Save(ctx, "user", "chat", &commonv1.TripConfig{Version: "1"}); err != nil {
		t.Fatal(err)
	}
	if err = configs.Save(ctx, "user", "chat", &commonv1.TripConfig{Version: "2"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := configs.Find(ctx, "user", "chat")
	if err != nil || cfg.GetVersion() != "2" {
		t.Fatalf("%+v %v", cfg, err)
	}
	if err = configs.Save(ctx, "other", "chat", &commonv1.TripConfig{Version: "3"}); !errors.Is(err, session.ErrForbidden) {
		t.Fatal(err)
	}
}
