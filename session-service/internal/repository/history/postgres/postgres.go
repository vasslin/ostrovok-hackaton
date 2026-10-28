package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/godnota/session-service/internal/entity"
	"github.com/godnota/session-service/internal/usecase/session"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// History stores chat messages and summaries in PostgreSQL.
type History struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *History { return &History{pool: pool} }

var _ session.HistoryRepository = (*History)(nil)

func (h *History) AppendMessages(ctx context.Context, userID, chatID string, messages []entity.Message) (int, error) {
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err = ensureChat(ctx, tx, userID, chatID); err != nil {
		return 0, err
	}
	for _, message := range messages {
		createdAt, err := parseTime(message.CreatedAt)
		if err != nil {
			return 0, err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO agent_messages (chat_id, role, body, created_at) VALUES ($1, $2, $3, $4)`, chatID, message.Role, message.Text, createdAt); err != nil {
			return 0, err
		}
	}
	if _, err = tx.Exec(ctx, `UPDATE agent_chats SET updated_at = $2 WHERE chat_id = $1`, chatID, time.Now().UTC()); err != nil {
		return 0, err
	}
	var count int
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM agent_messages WHERE chat_id = $1`, chatID).Scan(&count); err != nil {
		return 0, err
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return count, nil
}

func (h *History) UpdateSummary(ctx context.Context, userID, chatID, summary string) error {
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err = ensureChat(ctx, tx, userID, chatID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE agent_chats SET summary = $2, updated_at = $3 WHERE chat_id = $1`, chatID, summary, time.Now().UTC()); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (h *History) Find(ctx context.Context, userID, chatID string) (entity.Chat, error) {
	var chat entity.Chat
	err := h.pool.QueryRow(ctx, `SELECT chat_id, user_id, summary, created_at, updated_at FROM agent_chats WHERE chat_id = $1`, chatID).
		Scan(&chat.ID, &chat.UserID, &chat.Summary, &chat.CreatedAt, &chat.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Chat{}, session.ErrNotFound
	}
	if err != nil {
		return entity.Chat{}, err
	}
	if chat.UserID != userID {
		return entity.Chat{}, session.ErrForbidden
	}
	rows, err := h.pool.Query(ctx, `SELECT role, body, created_at FROM agent_messages WHERE chat_id = $1 ORDER BY id`, chatID)
	if err != nil {
		return entity.Chat{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var message entity.Message
		var createdAt time.Time
		if err = rows.Scan(&message.Role, &message.Text, &createdAt); err != nil {
			return entity.Chat{}, err
		}
		message.CreatedAt = createdAt.UTC().Format(time.RFC3339Nano)
		chat.Messages = append(chat.Messages, message)
	}
	return chat, rows.Err()
}

func (h *History) PreviousSummaries(ctx context.Context, userID, chatID string, limit int) ([]entity.PreviousChat, error) {
	if limit <= 0 {
		return nil, nil
	}
	rows, err := h.pool.Query(ctx, `
		SELECT chat_id, summary, updated_at
		FROM agent_chats
		WHERE user_id = $1 AND chat_id <> $2 AND summary <> ''
		ORDER BY updated_at DESC, chat_id
		LIMIT $3`, userID, chatID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var found []entity.PreviousChat
	for rows.Next() {
		var item entity.PreviousChat
		if err = rows.Scan(&item.ChatID, &item.Summary, &item.UpdatedAt); err != nil {
			return nil, err
		}
		found = append(found, item)
	}
	return found, rows.Err()
}

func ensureChat(ctx context.Context, tx pgx.Tx, userID, chatID string) error {
	var owner string
	err := tx.QueryRow(ctx, `SELECT user_id FROM agent_chats WHERE chat_id = $1`, chatID).Scan(&owner)
	if errors.Is(err, pgx.ErrNoRows) {
		now := time.Now().UTC()
		_, err = tx.Exec(ctx, `INSERT INTO agent_chats (chat_id, user_id, summary, created_at, updated_at) VALUES ($1, $2, '', $3, $3)`, chatID, userID, now)
		return err
	}
	if err != nil {
		return err
	}
	if owner != userID {
		return session.ErrForbidden
	}
	return nil
}

func parseTime(value string) (time.Time, error) {
	if value == "" {
		return time.Now().UTC(), nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339, value)
	}
	if err != nil {
		return time.Time{}, session.ErrInvalid
	}
	return parsed.UTC(), nil
}
