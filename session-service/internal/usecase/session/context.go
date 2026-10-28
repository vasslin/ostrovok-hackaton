package session

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	commonv1 "github.com/godnota/pkg/generated/common/v1"
	"github.com/godnota/session-service/internal/entity"
	"google.golang.org/protobuf/proto"
)

const (
	defaultPreviousChats = 4
	maxPreviousChats     = 20
)

var (
	ErrInvalid   = errors.New("invalid argument")
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("forbidden")
)

// HistoryRepository stores chat messages and per-chat summaries.
type HistoryRepository interface {
	AppendMessages(ctx context.Context, userID, chatID string, messages []entity.Message) (int, error)
	UpdateSummary(ctx context.Context, userID, chatID, summary string) error
	Find(ctx context.Context, userID, chatID string) (entity.Chat, error)
	PreviousSummaries(ctx context.Context, userID, chatID string, limit int) ([]entity.PreviousChat, error)
}

// ConfigRepository stores the single trip config of a chat.
type ConfigRepository interface {
	Save(ctx context.Context, userID, chatID string, cfg *commonv1.TripConfig) error
	Find(ctx context.Context, userID, chatID string) (*commonv1.TripConfig, error)
}

type AgentContext struct {
	UserID   string
	ChatID   string
	History  []entity.Message
	Summary  string
	Previous []entity.PreviousChat
	Config   *commonv1.TripConfig
}

type AgentContextRequest struct {
	UserID string
	ChatID string
	Limit  int
}

// SaveRequest updates only the parts whose flag is set.
type SaveRequest struct {
	UserID      string
	ChatID      string
	Messages    []entity.Message
	SetMessages bool
	Summary     string
	SetSummary  bool
	Config      *commonv1.TripConfig
	SetConfig   bool
}

func (s *Service) GetAgentContext(ctx context.Context, req AgentContextRequest) (AgentContext, error) {
	userID, chatID, err := ids(req.UserID, req.ChatID)
	if err != nil {
		return AgentContext{}, err
	}
	limit, err := s.limit(req.Limit)
	if err != nil {
		return AgentContext{}, err
	}
	slog.Info("session-service: get agent context", "user_id", userID, "chat_id", chatID, "previous_limit", limit)

	out := AgentContext{UserID: userID, ChatID: chatID}
	chat, err := s.history.Find(ctx, userID, chatID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return AgentContext{}, err
	}
	if err == nil {
		out.History = chat.Messages
		out.Summary = chat.Summary
	}
	cfg, err := s.configs.Find(ctx, userID, chatID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return AgentContext{}, err
	}
	out.Config = cfg
	previous, err := s.history.PreviousSummaries(ctx, userID, chatID, limit)
	if err != nil {
		return AgentContext{}, err
	}
	out.Previous = previous
	return out, nil
}

func (s *Service) SaveContext(ctx context.Context, req SaveRequest) error {
	userID, chatID, err := ids(req.UserID, req.ChatID)
	if err != nil {
		return err
	}
	var messages []entity.Message
	if req.SetMessages {
		messages, err = normalizeMessages(req.Messages)
		if err != nil {
			return err
		}
	}
	if req.SetConfig && req.Config == nil {
		return ErrInvalid
	}
	slog.Info("session-service: save context", "user_id", userID, "chat_id", chatID, "messages", req.SetMessages, "summary", req.SetSummary, "cfg", req.SetConfig)
	if req.SetMessages {
		if _, err = s.history.AppendMessages(ctx, userID, chatID, messages); err != nil {
			return err
		}
	}
	if req.SetSummary {
		if err = s.history.UpdateSummary(ctx, userID, chatID, req.Summary); err != nil {
			return err
		}
	}
	if req.SetConfig {
		if err = s.configs.Save(ctx, userID, chatID, proto.Clone(req.Config).(*commonv1.TripConfig)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) limit(n int) (int, error) {
	switch {
	case n < 0:
		return 0, ErrInvalid
	case n == 0:
		if s.previousLimit > 0 {
			return s.previousLimit, nil
		}
		return defaultPreviousChats, nil
	case n > maxPreviousChats:
		return maxPreviousChats, nil
	default:
		return n, nil
	}
}

func ids(userID, chatID string) (string, string, error) {
	userID = strings.TrimSpace(userID)
	chatID = strings.TrimSpace(chatID)
	if userID == "" || chatID == "" {
		return "", "", ErrInvalid
	}
	return userID, chatID, nil
}

func normalizeMessages(messages []entity.Message) ([]entity.Message, error) {
	if len(messages) == 0 {
		return nil, ErrInvalid
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	out := make([]entity.Message, len(messages))
	for i, message := range messages {
		message.Role = strings.TrimSpace(message.Role)
		message.Text = strings.TrimSpace(message.Text)
		if message.Role == "" || message.Text == "" {
			return nil, ErrInvalid
		}
		if strings.TrimSpace(message.CreatedAt) == "" {
			message.CreatedAt = now
		}
		out[i] = message
	}
	return out, nil
}
