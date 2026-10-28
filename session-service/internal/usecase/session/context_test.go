package session

import (
	"context"
	"errors"
	"testing"
	"time"

	commonv1 "github.com/godnota/pkg/generated/common/v1"
	"github.com/godnota/session-service/internal/entity"
)

type fakeHistory struct {
	chats map[string]entity.Chat
}

func (f *fakeHistory) AppendMessages(_ context.Context, userID, chatID string, messages []entity.Message) (int, error) {
	chat := f.chats[chatID]
	if chat.ID != "" && chat.UserID != userID {
		return 0, ErrForbidden
	}
	if chat.ID == "" {
		chat = entity.Chat{ID: chatID, UserID: userID}
	}
	chat.Messages = append(chat.Messages, messages...)
	f.chats[chatID] = chat
	return len(chat.Messages), nil
}

func (f *fakeHistory) UpdateSummary(_ context.Context, userID, chatID, summary string) error {
	chat := f.chats[chatID]
	if chat.ID != "" && chat.UserID != userID {
		return ErrForbidden
	}
	if chat.ID == "" {
		chat = entity.Chat{ID: chatID, UserID: userID}
	}
	chat.Summary = summary
	f.chats[chatID] = chat
	return nil
}

func (f *fakeHistory) Find(_ context.Context, userID, chatID string) (entity.Chat, error) {
	chat, ok := f.chats[chatID]
	if !ok {
		return entity.Chat{}, ErrNotFound
	}
	if chat.UserID != userID {
		return entity.Chat{}, ErrForbidden
	}
	return chat, nil
}

func (f *fakeHistory) PreviousSummaries(_ context.Context, userID, chatID string, limit int) ([]entity.PreviousChat, error) {
	var found []entity.PreviousChat
	for id, chat := range f.chats {
		if chat.UserID == userID && id != chatID && chat.Summary != "" {
			found = append(found, entity.PreviousChat{ChatID: id, Summary: chat.Summary, UpdatedAt: chat.UpdatedAt})
		}
	}
	if len(found) > limit {
		found = found[:limit]
	}
	return found, nil
}

type fakeConfig struct {
	rows map[string]struct {
		userID string
		cfg    *commonv1.TripConfig
	}
}

func (f *fakeConfig) Save(_ context.Context, userID, chatID string, cfg *commonv1.TripConfig) error {
	if current, ok := f.rows[chatID]; ok && current.userID != userID {
		return ErrForbidden
	}
	f.rows[chatID] = struct {
		userID string
		cfg    *commonv1.TripConfig
	}{userID: userID, cfg: cfg}
	return nil
}

func (f *fakeConfig) Find(_ context.Context, userID, chatID string) (*commonv1.TripConfig, error) {
	current, ok := f.rows[chatID]
	if !ok {
		return nil, ErrNotFound
	}
	if current.userID != userID {
		return nil, ErrForbidden
	}
	return current.cfg, nil
}

func TestAgentContextUsesDefaultLimitAndSeparateConfig(t *testing.T) {
	history := &fakeHistory{chats: map[string]entity.Chat{}}
	configs := &fakeConfig{rows: map[string]struct {
		userID string
		cfg    *commonv1.TripConfig
	}{}}
	service := New(nil, history, configs)
	ctx := context.Background()
	for i, summary := range []string{"one", "two", "three", "four", "five"} {
		history.chats["old-"+summary] = entity.Chat{ID: "old-" + summary, UserID: "u", Summary: summary, UpdatedAt: time.Unix(int64(i), 0)}
	}
	if err := service.SaveContext(ctx, SaveRequest{UserID: "u", ChatID: "current", SetMessages: true, Messages: []entity.Message{{Role: "user", Text: "к морю"}}}); err != nil {
		t.Fatal(err)
	}
	if err := service.SaveContext(ctx, SaveRequest{UserID: "u", ChatID: "current", SetMessages: true, Messages: []entity.Message{{Role: "assistant", Text: "когда?"}}}); err != nil {
		t.Fatal(err)
	}
	if err := service.SaveContext(ctx, SaveRequest{UserID: "u", ChatID: "current", SetSummary: true, Summary: "море"}); err != nil {
		t.Fatal(err)
	}
	if err := service.SaveContext(ctx, SaveRequest{UserID: "u", ChatID: "current", SetConfig: true, Config: &commonv1.TripConfig{Version: "1"}}); err != nil {
		t.Fatal(err)
	}
	if err := service.SaveContext(ctx, SaveRequest{UserID: "u", ChatID: "current", SetConfig: true, Config: &commonv1.TripConfig{Version: "2"}}); err != nil {
		t.Fatal(err)
	}
	got, err := service.GetAgentContext(ctx, AgentContextRequest{UserID: "u", ChatID: "current"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.History) != 2 || got.Summary != "море" || got.Config.GetVersion() != "2" || len(got.Previous) != defaultPreviousChats {
		t.Fatalf("%+v", got)
	}
	if err = service.SaveContext(ctx, SaveRequest{UserID: "u", ChatID: "current", SetSummary: true, Summary: "другое"}); err != nil {
		t.Fatal(err)
	}
	kept, err := service.GetAgentContext(ctx, AgentContextRequest{UserID: "u", ChatID: "current"})
	if err != nil || kept.Summary != "другое" || len(kept.History) != 2 || kept.Config.GetVersion() != "2" {
		t.Fatalf("partial save changed other fields: %+v %v", kept, err)
	}
	if err = service.SaveContext(ctx, SaveRequest{UserID: "other", ChatID: "current", SetMessages: true, Messages: []entity.Message{{Role: "user", Text: "нет"}}}); !errors.Is(err, ErrForbidden) {
		t.Fatal(err)
	}
	if err = service.SaveContext(ctx, SaveRequest{UserID: "other", ChatID: "current", SetConfig: true, Config: &commonv1.TripConfig{Version: "3"}}); !errors.Is(err, ErrForbidden) {
		t.Fatal(err)
	}
	if _, err = service.GetAgentContext(ctx, AgentContextRequest{UserID: "", ChatID: "current"}); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if _, err = service.GetAgentContext(ctx, AgentContextRequest{UserID: "u", ChatID: "current", Limit: -1}); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	empty, err := service.GetAgentContext(ctx, AgentContextRequest{UserID: "u", ChatID: "fresh"})
	if err != nil || len(empty.History) != 0 || empty.Config != nil || empty.Summary != "" {
		t.Fatalf("%+v %v", empty, err)
	}
}
