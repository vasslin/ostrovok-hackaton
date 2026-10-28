package service

import (
	"ai-service/internal/canonical"
	"bytes"
	"encoding/json"
	"fmt"
)

func (c *PlanningConfig) UnmarshalJSON(data []byte) error {
	type wire PlanningConfig
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	wrapped := fields["request"] != nil || fields["trip"] != nil || fields["version"] != nil || fields["resolved"] != nil
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if wrapped {
		var value wire
		if err := decoder.Decode(&value); err != nil {
			return err
		}
		*c = PlanningConfig(value)
		if c.Request != nil && c.Trip != nil {
			return fmt.Errorf("use config.request or legacy config.trip, not both")
		}
		return nil
	}
	var request canonical.TripRequest
	if err := decoder.Decode(&request); err != nil {
		return err
	}
	*c = PlanningConfig{Version: "1.0", Request: &request}
	return nil
}

type AIAdvice struct {
	ID     string `json:"id"`
	City   string `json:"city"`
	Reason string `json:"reason"`
}
type AIQuestion struct {
	ID       string `json:"id"`
	Field    string `json:"field"`
	Text     string `json:"text"`
	Required bool   `json:"required"`
}
type AIResponse struct {
	Status    string                 `json:"status"`
	Config    *canonical.TripRequest `json:"config"`
	Advice    []AIAdvice             `json:"advice"`
	Questions []AIQuestion           `json:"questions"`
}

func StrictAIResponse(out *PlanningOutput) AIResponse {
	result := AIResponse{Status: out.Status, Config: out.Config.Request, Advice: []AIAdvice{}, Questions: []AIQuestion{}}
	for _, idea := range out.Advice {
		city := idea.City
		if idea.Config != nil && idea.Config.Request != nil && idea.Config.Request.Destination != nil {
			city = idea.Config.Request.Destination.City
		}
		result.Advice = append(result.Advice, AIAdvice{ID: idea.ID, City: city, Reason: idea.Reason})
	}
	for _, q := range out.Questions {
		result.Questions = append(result.Questions, AIQuestion{ID: q.ID, Field: q.Field, Text: q.Text, Required: q.Required})
	}
	return result
}
