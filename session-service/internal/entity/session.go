package entity

// Session is the dialogue context owned by session-service.
type Session struct {
	ID        string
	UserID    string
	Status    string
	Summary   string
	CreatedAt string
	UpdatedAt string
}

// Message is one turn of the planning dialogue.
type Message struct {
	Role      string
	Text      string
	CreatedAt string
}
