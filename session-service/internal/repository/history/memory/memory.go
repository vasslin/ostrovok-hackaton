package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/godnota/session-service/internal/entity"
	"github.com/godnota/session-service/internal/usecase/session"
)

// History is an in-memory chat history store.
type History struct {
	mu   sync.Mutex
	now  func() time.Time
	chat map[string]entity.Chat
}

func New() *History {
	return &History{chat: map[string]entity.Chat{}}
}

var _ session.HistoryRepository = (*History)(nil)

func (h *History) AppendMessages(ctx context.Context, userID, chatID string, messages []entity.Message) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	chat, err := h.ensure(userID, chatID)
	if err != nil {
		return 0, err
	}
	copied := append([]entity.Message(nil), messages...)
	for i := range copied {
		if strings.TrimSpace(copied[i].CreatedAt) == "" {
			copied[i].CreatedAt = h.timestamp().Format(time.RFC3339Nano)
		}
	}
	chat.Messages = append(chat.Messages, copied...)
	chat.UpdatedAt = h.timestamp()
	h.chat[chatID] = chat
	return len(chat.Messages), nil
}

func (h *History) UpdateSummary(ctx context.Context, userID, chatID, summary string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	chat, err := h.ensure(userID, chatID)
	if err != nil {
		return err
	}
	chat.Summary = summary
	chat.UpdatedAt = h.timestamp()
	h.chat[chatID] = chat
	return nil
}

func (h *History) Find(ctx context.Context, userID, chatID string) (entity.Chat, error) {
	if err := ctx.Err(); err != nil {
		return entity.Chat{}, err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	chat, ok := h.chat[chatID]
	if !ok {
		return entity.Chat{}, session.ErrNotFound
	}
	if chat.UserID != userID {
		return entity.Chat{}, session.ErrForbidden
	}
	chat.Messages = append([]entity.Message(nil), chat.Messages...)
	return chat, nil
}

func (h *History) PreviousSummaries(ctx context.Context, userID, chatID string, limit int) ([]entity.PreviousChat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		return nil, nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	found := make([]entity.PreviousChat, 0)
	for id, chat := range h.chat {
		if chat.UserID != userID || id == chatID || chat.Summary == "" {
			continue
		}
		found = append(found, entity.PreviousChat{ChatID: id, Summary: chat.Summary, UpdatedAt: chat.UpdatedAt})
	}
	sort.Slice(found, func(i, j int) bool {
		if found[i].UpdatedAt.Equal(found[j].UpdatedAt) {
			return found[i].ChatID < found[j].ChatID
		}
		return found[i].UpdatedAt.After(found[j].UpdatedAt)
	})
	if len(found) > limit {
		found = found[:limit]
	}
	return found, nil
}

func (h *History) ensure(userID, chatID string) (entity.Chat, error) {
	chat, ok := h.chat[chatID]
	if !ok {
		now := h.timestamp()
		return entity.Chat{ID: chatID, UserID: userID, CreatedAt: now, UpdatedAt: now}, nil
	}
	if chat.UserID != userID {
		return entity.Chat{}, session.ErrForbidden
	}
	chat.Messages = append([]entity.Message(nil), chat.Messages...)
	return chat, nil
}

func (h *History) timestamp() time.Time {
	if h.now != nil {
		return h.now().UTC()
	}
	return time.Now().UTC()
}
