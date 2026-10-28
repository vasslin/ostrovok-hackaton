package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

type Minor int64

func (m Minor) MarshalJSON() ([]byte, error) { return json.Marshal(strconv.FormatInt(int64(m), 10)) }
func (m *Minor) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return fmt.Errorf("amount_minor cannot be null")
	}
	var text string
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
	} else {
		text = string(data)
	}
	value, err := strconv.ParseInt(text, 10, 64)
	if err != nil || value < 0 {
		return fmt.Errorf("amount_minor must be a nonnegative int64")
	}
	*m = Minor(value)
	return nil
}

type RouteCandidate struct {
	ID                  string   `json:"id"`
	TransportPartIDs    []string `json:"transport_part_ids"`
	HotelPartIDs        []string `json:"hotel_part_ids"`
	ConnectionsVerified bool     `json:"connections_verified"`
}
type RankRoutesInput struct {
	Config      PlanningConfig      `json:"config"`
	Context     PlanningContext     `json:"context"`
	Parts       SearchPartsResponse `json:"parts"`
	Candidates  []RouteCandidate    `json:"candidates"`
	Limit       int                 `json:"limit"`
	CurrentDate string              `json:"current_date"`
}
type RankedRoute struct {
	ID                   string   `json:"id"`
	Score                float64  `json:"score"`
	Reason               string   `json:"reason"`
	Risks                []string `json:"risks"`
	TotalPrice           *Money   `json:"total_price,omitempty"`
	PriceComplete        bool     `json:"price_complete"`
	RequiresVerification bool     `json:"requires_verification"`
}
type RejectedRoute struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}
type RankRoutesOutput struct {
	Results  []RankedRoute       `json:"results"`
	Rejected []RejectedRoute     `json:"rejected"`
	Parts    SearchPartsResponse `json:"parts"`
}
type routeScore struct {
	ID    string   `json:"id"`
	Score *float64 `json:"score"`
}
type routeScores struct {
	Results []routeScore `json:"results"`
}
