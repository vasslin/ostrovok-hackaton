package service

import (
	"ai-service/internal/adapter/agent"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type Service struct{ client *agent.Client }

func New(client *agent.Client) *Service { return &Service{client: client} }
func (s *Service) Configured() bool     { return s.client.Configured() }
func (s *Service) Model() string        { return s.client.Model }

type Error struct {
	Status     int
	Message    string
	RetryAfter string
}

func (e *Error) Error() string     { return e.Message }
func invalid(message string) error { return &Error{Status: 400, Message: message} }

func upstream(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return &Error{Status: 504, Message: "model request timed out"}
	}

	if errors.Is(err, context.Canceled) {
		return &Error{Status: 499, Message: "request canceled"}
	}

	var provider *agent.Error
	if errors.As(err, &provider) && provider.Status == 429 {
		return &Error{Status: 429, Message: "provider quota exceeded; retry later", RetryAfter: provider.RetryAfter}
	}

	return &Error{Status: 502, Message: "model unavailable or returned invalid data"}
}

func validate(c Config) error {
	for _, field := range []struct {
		name  string
		value *string
	}{{"origin", c.Origin}, {"destination", c.Destination}} {
		if field.value != nil && (strings.TrimSpace(*field.value) == "" || utf8.RuneCountInString(*field.value) > 200) {
			return fmt.Errorf("%s must contain 1..200 characters", field.name)
		}
	}

	if c.Currency != nil && (len(*c.Currency) != 3 || strings.Trim(*c.Currency, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "") {
		return fmt.Errorf("currency must be a three-letter uppercase code")
	}

	if len(c.ChildrenAges) > 20 || len(c.Preferences) > 30 {
		return fmt.Errorf("too many children or preferences")
	}

	for _, preference := range c.Preferences {
		if strings.TrimSpace(preference) == "" || utf8.RuneCountInString(preference) > 300 {
			return fmt.Errorf("preferences must contain 1..300 characters")
		}
	}

	var start, end time.Time
	var err error
	if c.DepartureDate != nil {
		start, err = time.Parse("2006-01-02", *c.DepartureDate)
		if err != nil {
			return fmt.Errorf("invalid departure_date")
		}
	}

	if c.ReturnDate != nil {
		end, err = time.Parse("2006-01-02", *c.ReturnDate)
		if err != nil {
			return fmt.Errorf("invalid return_date")
		}
	}

	if !start.IsZero() && !end.IsZero() && end.Before(start) {
		return fmt.Errorf("return_date precedes departure_date")
	}

	if c.Adults != nil && (*c.Adults < 1 || *c.Adults > 20) {
		return fmt.Errorf("adults must be 1..20")
	}

	if c.Budget != nil && *c.Budget <= 0 {
		return fmt.Errorf("budget must be positive")
	}

	for _, age := range c.ChildrenAges {
		if age < 0 || age > 17 {
			return fmt.Errorf("children_ages must be 0..17")
		}
	}

	return nil
}

func (s *Service) Extract(ctx context.Context, in ExtractInput) (*ExtractOutput, error) {
	if !s.Configured() {
		return nil, &Error{Status: 503, Message: "LLM_API_KEY is not configured"}
	}

	if strings.TrimSpace(in.Message) == "" || utf8.RuneCountInString(in.Message) > 6000 {
		return nil, invalid("message must contain 1..6000 characters")
	}

	if err := validate(in.Config); err != nil {
		return nil, invalid(err.Error())
	}

	var out ExtractOutput
	input := map[string]any{"message": in.Message, "config": in.Config, "current_date": time.Now().UTC().Format("2006-01-02")}
	if err := s.client.Complete(ctx, extractPrompt, input, &out); err != nil {
		return nil, upstream(err)
	}

	if out.Config == nil || len(out.Questions) > 2 {
		return nil, upstream(fmt.Errorf("invalid extraction"))
	}

	if err := validate(*out.Config); err != nil {
		return nil, upstream(err)
	}

	out.MissingFields = []string{}
	fields := []struct {
		name  string
		value *string
	}{{"origin", out.Config.Origin}, {"destination", out.Config.Destination}, {"departure_date", out.Config.DepartureDate}, {"return_date", out.Config.ReturnDate}}
	for _, f := range fields {
		if f.value == nil || strings.TrimSpace(*f.value) == "" {
			out.MissingFields = append(out.MissingFields, f.name)
		}
	}

	if out.Config.Adults == nil {
		out.MissingFields = append(out.MissingFields, "adults")
	}

	if out.Questions == nil {
		out.Questions = []string{}
	}

	if out.Config.ChildrenAges == nil {
		out.Config.ChildrenAges = []int{}
	}

	if out.Config.Preferences == nil {
		out.Config.Preferences = []string{}
	}

	return &out, nil
}

func (s *Service) RankHotels(ctx context.Context, in RankInput) (*RankOutput, error) {
	if !s.Configured() {
		return nil, &Error{Status: 503, Message: "LLM_API_KEY is not configured"}
	}

	if err := validate(in.Config); err != nil {
		return nil, invalid(err.Error())
	}

	if len(in.Hotels) < 1 || len(in.Hotels) > 10 {
		return nil, invalid("provide 1..10 hotels")
	}

	ids := map[string]bool{}
	for _, c := range in.Hotels {
		if strings.TrimSpace(c.ID) == "" || utf8.RuneCountInString(c.ID) > 200 || ids[c.ID] {
			return nil, invalid("hotel IDs must be nonempty and unique")
		}
		if strings.TrimSpace(c.Name) == "" || utf8.RuneCountInString(c.Name) > 300 {
			return nil, invalid("hotel name must contain 1..300 characters")
		}
		ids[c.ID] = true
	}

	var out RankOutput
	if err := s.client.Complete(ctx, rankPrompt, in, &out); err != nil {
		return nil, upstream(err)
	}

	if len(out.Results) != len(ids) {
		return nil, upstream(fmt.Errorf("missing ranked hotels"))
	}

	last := 101.0
	for _, item := range out.Results {
		if !ids[item.ID] || item.Score == nil || *item.Score < 0 || *item.Score > 100 || *item.Score > last || strings.TrimSpace(item.Reason) == "" {
			return nil, upstream(fmt.Errorf("invalid ranking"))
		}
		delete(ids, item.ID)
		last = *item.Score
	}

	for i := range out.Results {
		if out.Results[i].Risks == nil {
			out.Results[i].Risks = []string{}
		}
	}

	return &out, nil
}
