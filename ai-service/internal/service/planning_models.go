package service

import "ai-service/internal/canonical"

type PlanningConfig struct {
	Version  string                 `json:"version"`
	Trip     *PlanningTrip          `json:"trip,omitempty"`
	Request  *canonical.TripRequest `json:"request,omitempty"`
	Resolved map[string]any         `json:"resolved,omitempty"`
}
type PlanningTrip struct {
	Origin        *PlanningPlace         `json:"origin"`
	Destinations  []PlanningPlace        `json:"destinations"`
	DateRange     *PlanningDates         `json:"date_range"`
	Travelers     *PlanningTravelers     `json:"travelers"`
	Budget        *PlanningBudget        `json:"budget"`
	Transport     *PlanningTransport     `json:"transport"`
	Accommodation *PlanningAccommodation `json:"accommodation"`
	Activities    *PlanningActivities    `json:"activities"`
	Preferences   *PlanningPreferences   `json:"preferences"`
}
type PlanningPlace struct {
	City        string  `json:"city"`
	CountryCode *string `json:"country_code"`
}
type PlanningDates struct {
	StartDate     *string `json:"start_date"`
	EndDate       *string `json:"end_date"`
	FlexibleDates *bool   `json:"flexible_dates"`
}
type PlanningTravelers struct {
	Adults       *int  `json:"adults"`
	Children     *int  `json:"children"`
	ChildrenAges []int `json:"children_ages"`
}
type PlanningBudget struct {
	Amount      *float64 `json:"amount"`
	Currency    *string  `json:"currency"`
	Scope       *string  `json:"scope"`
	IsHardLimit *bool    `json:"is_hard_limit"`
}
type PlanningTransport struct {
	Modes        []string `json:"modes"`
	MaxTransfers *int     `json:"max_transfers"`
	MaxWaitHours *int     `json:"max_wait_hours"`
}
type PlanningAccommodation struct {
	Required         *bool    `json:"required"`
	MinStars         *int     `json:"min_stars"`
	MaxPricePerNight *float64 `json:"max_price_per_night"`
	PreferredTypes   []string `json:"preferred_types"`
}
type PlanningActivities struct {
	Interests []string `json:"interests"`
	Pace      *string  `json:"pace"`
}
type PlanningPreferences struct {
	Language                *string `json:"language"`
	AvoidOvernightTransfers *bool   `json:"avoid_overnight_transfers"`
}
type PlanningQuestion struct {
	ID        string `json:"id"`
	Field     string `json:"field"`
	Text      string `json:"text"`
	Required  bool   `json:"required"`
	InputType string `json:"input_type"`
}
type PlanningAnswer struct {
	QuestionID string `json:"question_id"`
	Text       string `json:"text"`
}
type PlanningMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type PlanningContext struct {
	Summary  string            `json:"summary"`
	Messages []PlanningMessage `json:"messages"`
}
type PlanningAssumption struct {
	Field  string `json:"field"`
	Value  string `json:"value"`
	Reason string `json:"reason"`
}
type TravelIdea struct {
	City        string               `json:"city,omitempty"`
	Reason      string               `json:"reason,omitempty"`
	ID          string               `json:"id"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Config      *PlanningConfig      `json:"config"`
	Reasons     []string             `json:"reasons"`
	Assumptions []PlanningAssumption `json:"assumptions"`
	Source      string               `json:"source"`
}
type PlanningInput struct {
	canonicalRequest *canonical.TripRequest
	SessionID        string             `json:"session_id"`
	Text             string             `json:"text"`
	Config           PlanningConfig     `json:"config"`
	Context          PlanningContext    `json:"context"`
	Questions        []PlanningQuestion `json:"questions"`
	Answers          []PlanningAnswer   `json:"answers"`
	SelectedIdea     *TravelIdea        `json:"selected_idea"`
	CurrentDate      string             `json:"current_date"`
	ForceGenerate    bool               `json:"force_generate"`
}
type PlanningOutput struct {
	Status        string               `json:"status"`
	Config        *PlanningConfig      `json:"config"`
	Advice        []TravelIdea         `json:"advice"`
	Questions     []PlanningQuestion   `json:"questions"`
	MissingFields []string             `json:"missing_fields"`
	Assumptions   []PlanningAssumption `json:"assumptions"`
	Summary       string               `json:"summary"`
	Warnings      []string             `json:"warnings"`
}
type planningCompletion struct {
	Evidence      []PlanningEvidence   `json:"evidence"`
	Status        string               `json:"status"`
	Config        *PlanningConfig      `json:"config"`
	Advice        []TravelIdea         `json:"advice"`
	Questions     []PlanningQuestion   `json:"questions"`
	Assumptions   []PlanningAssumption `json:"assumptions"`
	ClearedFields []string             `json:"cleared_fields"`
	Summary       string               `json:"summary"`
	Warnings      []string             `json:"warnings"`
}
type PreviousSession struct {
	Config  PlanningConfig  `json:"config"`
	Context PlanningContext `json:"context"`
}
type RecommendationsInput struct {
	canonicalRequest *canonical.TripRequest
	Text             string            `json:"text"`
	Config           PlanningConfig    `json:"config"`
	Context          PlanningContext   `json:"context"`
	LastSessions     []PreviousSession `json:"last_sessions"`
	CurrentDate      string            `json:"current_date"`
}
type RecommendationsOutput struct {
	Advice []TravelIdea `json:"advice"`
}

type ideaProposal struct {
	ID    string `json:"id"`
	City  string `json:"city"`
	Theme string `json:"theme"`
}
type recommendationsCompletion struct {
	Advice []ideaProposal `json:"advice"`
}
