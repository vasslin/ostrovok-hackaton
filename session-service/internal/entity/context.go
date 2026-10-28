package entity

import "time"

// Chat is one agent dialogue owned by a user.
// History and summary live here. The trip config is stored separately.
type Chat struct {
	ID        string
	UserID    string
	Summary   string
	Messages  []Message
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PreviousChat is a short reminder of an earlier dialogue of the same user.
type PreviousChat struct {
	ChatID    string
	Summary   string
	UpdatedAt time.Time
}
