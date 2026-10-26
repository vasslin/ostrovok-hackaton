package service

type Config struct {
	Origin        *string  `json:"origin"`
	Destination   *string  `json:"destination"`
	DepartureDate *string  `json:"departure_date"`
	ReturnDate    *string  `json:"return_date"`
	Adults        *int     `json:"adults"`
	ChildrenAges  []int    `json:"children_ages"`
	Budget        *float64 `json:"budget"`
	Currency      *string  `json:"currency"`
	Preferences   []string `json:"preferences"`
}

type ExtractInput struct {
	Message string `json:"message"`
	Config  Config `json:"config"`
}

type ExtractOutput struct {
	Config        *Config  `json:"config"`
	MissingFields []string `json:"missing_fields"`
	Questions     []string `json:"questions"`
}

type Hotel struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Details map[string]any `json:"details"`
}

type RankInput struct {
	Config Config  `json:"config"`
	Hotels []Hotel `json:"hotels"`
}

type Ranked struct {
	ID     string   `json:"id"`
	Score  *float64 `json:"score"`
	Reason string   `json:"reason"`
	Risks  []string `json:"risks"`
}

type RankOutput struct {
	Results []Ranked `json:"results"`
}
