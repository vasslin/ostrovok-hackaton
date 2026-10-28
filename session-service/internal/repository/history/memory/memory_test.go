package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/godnota/session-service/internal/entity"
	"github.com/godnota/session-service/internal/usecase/session"
)

func TestHistoryKeepsOrderAndLastSummaries(t *testing.T) {
	store := New()
	base := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	step := 0
	store.now = func() time.Time {
		step++
		return base.Add(time.Duration(step) * time.Minute)
	}
	ctx := context.Background()
	for _, chatID := range []string{"a", "b", "c", "d", "e"} {
		if err := store.UpdateSummary(ctx, "user", chatID, "summary "+chatID); err != nil {
			t.Fatal(err)
		}
	}
	count, err := store.AppendMessages(ctx, "user", "current", []entity.Message{{Role: "user", Text: "hello", CreatedAt: "2026-10-28T12:00:00Z"}})
	if err != nil || count != 1 {
		t.Fatal(err, count)
	}
	count, err = store.AppendMessages(ctx, "user", "current", []entity.Message{{Role: "assistant", Text: "hi"}})
	if err != nil || count != 2 {
		t.Fatal(err, count)
	}
	chat, err := store.Find(ctx, "user", "current")
	if err != nil || chat.Messages[0].Text != "hello" || chat.Messages[1].Text != "hi" || chat.Messages[1].CreatedAt == "" {
		t.Fatalf("%+v %v", chat, err)
	}
	chat.Messages[0].Text = "mutated"
	chat, _ = store.Find(ctx, "user", "current")
	if chat.Messages[0].Text != "hello" {
		t.Fatal("history shares the stored slice")
	}
	previous, err := store.PreviousSummaries(ctx, "user", "current", 4)
	if err != nil || len(previous) != 4 || previous[0].ChatID != "e" || previous[3].ChatID != "b" {
		t.Fatalf("%+v %v", previous, err)
	}
	if _, err = store.AppendMessages(ctx, "stranger", "current", []entity.Message{{Role: "user", Text: "no"}}); !errors.Is(err, session.ErrForbidden) {
		t.Fatal(err)
	}
	if _, err = store.Find(ctx, "user", "missing"); !errors.Is(err, session.ErrNotFound) {
		t.Fatal(err)
	}
}
