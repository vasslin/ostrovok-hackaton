package entity

import commonv1 "github.com/godnota/pkg/generated/common/v1"

// Session is the dialogue context owned by session-service.
type Session struct {
	ID        string
	UserID    string
	Status    string
	Summary   string
	CreatedAt string
	UpdatedAt string
	Config    *commonv1.TripConfig
	Messages  []Message
}

// Message is one turn of the planning dialogue.
type Message struct {
	Role      string
	Text      string
	CreatedAt string
}
