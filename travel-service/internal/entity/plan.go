package entity

// Plan is the orchestration state of one planning step.
type Plan struct {
	SessionID string
	Text      string
	IdeaID    string
	Comment   string
}

// Answer is a traveller reply bound to a question id.
type Answer struct {
	QuestionID string
	Text       string
}
