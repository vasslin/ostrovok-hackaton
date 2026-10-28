package memory

import (
	"context"
	"errors"
	"testing"

	commonv1 "github.com/godnota/pkg/generated/common/v1"
	"github.com/godnota/session-service/internal/usecase/session"
)

func TestConfigIsUniquePerChatAndCopied(t *testing.T) {
	store := New()
	ctx := context.Background()
	cfg := &commonv1.TripConfig{Version: "1"}
	if err := store.Save(ctx, "user", "chat", cfg); err != nil {
		t.Fatal(err)
	}
	cfg.Version = "mutated"
	if err := store.Save(ctx, "user", "chat", &commonv1.TripConfig{Version: "2"}); err != nil {
		t.Fatal(err)
	}
	got, err := store.Find(ctx, "user", "chat")
	if err != nil || got.GetVersion() != "2" {
		t.Fatalf("%+v %v", got, err)
	}
	got.Version = "local"
	got, _ = store.Find(ctx, "user", "chat")
	if got.GetVersion() != "2" {
		t.Fatal("config shares stored state")
	}
	if _, err = store.Find(ctx, "other", "chat"); !errors.Is(err, session.ErrForbidden) {
		t.Fatal(err)
	}
	if err = store.Save(ctx, "other", "chat", &commonv1.TripConfig{Version: "3"}); !errors.Is(err, session.ErrForbidden) {
		t.Fatal(err)
	}
	if _, err = store.Find(ctx, "user", "missing"); !errors.Is(err, session.ErrNotFound) {
		t.Fatal(err)
	}
}
