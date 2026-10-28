package service

import (
	"ai-service/internal/canonical"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

func planningDate(value string) (time.Time, error) {
	if value == "" {
		value = time.Now().UTC().Format("2006-01-02")
	}
	day, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, invalid("current_date must be YYYY-MM-DD")
	}
	return day, nil
}
func checkContext(ctx PlanningContext) error {
	if len(ctx.Messages) > 20 || utf8.RuneCountInString(ctx.Summary) > 3000 {
		return invalid("context supports 20 messages and summary up to 3000 characters")
	}
	total := 0
	for _, m := range ctx.Messages {
		if m.Role != "user" && m.Role != "assistant" {
			return invalid("context role must be user or assistant")
		}
		total += utf8.RuneCountInString(m.Content)
	}
	if total > 12000 {
		return invalid("context messages exceed 12000 characters")
	}
	return nil
}
func (s *Service) Analyze(ctx context.Context, in PlanningInput) (*PlanningOutput, error) {
	if !s.Configured() {
		return nil, &Error{Status: 503, Message: "LLM_API_KEY is not configured"}
	}
	if utf8.RuneCountInString(in.Text) > 6000 || (strings.TrimSpace(in.Text) == "" && len(in.Answers) == 0 && in.SelectedIdea == nil && !in.ForceGenerate) {
		return nil, invalid("provide text, answers, selected_idea or force_generate")
	}
	if err := checkContext(in.Context); err != nil {
		return nil, err
	}
	if len(in.Answers) > 20 || len(in.Questions) > 20 {
		return nil, invalid("at most 20 answers and questions")
	}
	ids := map[string]bool{}
	for _, q := range in.Questions {
		if q.ID == "" || ids[q.ID] {
			return nil, invalid("question IDs must be nonempty and unique")
		}
		ids[q.ID] = true
	}
	for _, a := range in.Answers {
		if !ids[a.QuestionID] || strings.TrimSpace(a.Text) == "" || utf8.RuneCountInString(a.Text) > 2000 {
			return nil, invalid("answers require a matching question_id and 1..2000 characters of text")
		}
	}
	day, err := planningDate(in.CurrentDate)
	if err != nil {
		return nil, err
	}
	in.CurrentDate = day.Format("2006-01-02")
	if in.Config.Version != "" && in.Config.Version != "1.0" {
		return nil, invalid("config.version must be 1.0")
	}
	base := in.Config
	if in.SelectedIdea != nil {
		if in.SelectedIdea.ID == "" || in.SelectedIdea.Config == nil || !hasDestination(*in.SelectedIdea.Config) {
			return nil, invalid("selected_idea requires id and config with destination")
		}
		if in.SelectedIdea.Config.Version != "" && in.SelectedIdea.Config.Version != "1.0" {
			return nil, invalid("selected_idea.config.version must be 1.0")
		}
		base, err = mergePlanning(base, *in.SelectedIdea.Config, nil)
		if err != nil {
			return nil, err
		}
	}
	base, err = applyLiteralPlanning(base, in)
	if err != nil {
		return nil, err
	}
	in.Config = base
	var completion planningCompletion
	var modelInput any = in
	if in.canonicalRequest != nil {
		modelInput = struct {
			PlanningInput
			Request *canonical.TripRequest `json:"canonical_request"`
		}{in, in.canonicalRequest}
	}
	if err := s.client.Complete(ctx, planningPrompt, modelInput, &completion); err != nil {
		return nil, upstream(err)
	}
	if completion.Config == nil || (completion.Status != "advice" && completion.Status != "questions" && completion.Status != "generate") {
		return nil, upstream(fmt.Errorf("invalid planning completion"))
	}
	patch, cleared, assumptions := groundedPlanningPatch(in, completion)
	config, err := mergePlanning(base, patch, cleared)
	if err != nil {
		return nil, upstream(err)
	}
	config, err = applyLiteralPlanning(config, in)
	if err != nil {
		return nil, err
	}
	issues := planningIssues(config, day)
	if mentionsChildren(in) && (config.Trip == nil || config.Trip.Travelers == nil || len(config.Trip.Travelers.ChildrenAges) == 0) {
		exists := false
		for _, issue := range issues {
			exists = exists || issue.Field == "trip.travelers.children_ages"
		}
		if !exists {
			issues = append(issues, PlanningQuestion{ID: "trip_travelers_children_ages", Field: "trip.travelers.children_ages", Text: "Сколько детей поедет и какой возраст у каждого?", Required: true, InputType: "text"})
		}
	}

	summary := strings.TrimSpace(completion.Summary)
	if summary == "" || notesMemory(summary) {
		summary = planningSummary(in)
	}
	out := &PlanningOutput{Config: &config, Status: "questions", Advice: []TravelIdea{}, Questions: []PlanningQuestion{}, MissingFields: []string{}, Assumptions: assumptions, Summary: summary, Warnings: []string{}}
	if out.Assumptions == nil {
		out.Assumptions = []PlanningAssumption{}
	}
	if out.Warnings == nil {
		out.Warnings = []string{}
	}
	if utf8.RuneCountInString(out.Summary) > 3000 {
		return nil, upstream(fmt.Errorf("summary too long"))
	}
	for _, issue := range issues {
		out.MissingFields = append(out.MissingFields, issue.Field)
	}
	if len(issues) == 0 {
		out.Status = "generate"
		return out, nil
	}
	if (completion.Status == "advice" || hasAny(norm(in.Text), "предложи", "посоветуй", "придумай", "не знаю куда", "не определился", "не определилась")) && !hasDestination(config) && !in.ForceGenerate {
		out.Status = "advice"
		return out, nil
	}
	out.Questions = planningQuestions(issues, completion.Questions)
	if in.ForceGenerate {
		out.Warnings = append(out.Warnings, "Для поиска не хватает обязательных или корректных полей; случайные значения не подставлены.")
	}
	return out, nil
}
func (s *Service) Recommendations(ctx context.Context, in RecommendationsInput) (*RecommendationsOutput, error) {
	if !s.Configured() {
		return nil, &Error{Status: 503, Message: "LLM_API_KEY is not configured"}
	}
	if in.Config.Version != "" && in.Config.Version != "1.0" {
		return nil, invalid("config.version must be 1.0")
	}
	if utf8.RuneCountInString(in.Text) > 6000 || len(in.LastSessions) > 5 {
		return nil, invalid("text exceeds 6000 characters or more than 5 last_sessions")
	}
	if err := checkContext(in.Context); err != nil {
		return nil, err
	}
	for _, session := range in.LastSessions {
		if err := checkContext(session.Context); err != nil {
			return nil, err
		}
	}
	day, err := planningDate(in.CurrentDate)
	if err != nil {
		return nil, err
	}
	in.CurrentDate = day.Format("2006-01-02")
	var completion recommendationsCompletion
	var modelInput any = in
	if in.canonicalRequest != nil {
		modelInput = struct {
			RecommendationsInput
			Request *canonical.TripRequest `json:"canonical_request"`
		}{in, in.canonicalRequest}
	}
	if err := s.client.Complete(ctx, recommendationsPrompt, modelInput, &completion); err != nil {
		return nil, upstream(err)
	}
	out := RecommendationsOutput{Advice: []TravelIdea{}}
	labels := map[string]string{"culture": "Культурная поездка", "nature": "Прогулки и природа", "relax": "Спокойный отдых", "active": "Активный отдых"}
	for _, proposal := range completion.Advice {
		label, ok := labels[proposal.Theme]
		if !ok || strings.TrimSpace(proposal.City) == "" || utf8.RuneCountInString(proposal.City) > 100 {
			return nil, upstream(fmt.Errorf("invalid idea proposal"))
		}
		city := strings.TrimSpace(proposal.City)
		ideaConfig := PlanningConfig{Version: "1.0", Trip: &PlanningTrip{Destinations: []PlanningPlace{{City: city}}}}
		out.Advice = append(out.Advice, TravelIdea{ID: proposal.ID, Title: label + ": " + city, Description: "Идея направления: " + city + ". Предлагаемый сценарий — " + strings.ToLower(label) + ". Это не найденный тур: цены, наличие, погоду, условия въезда и соответствие пожеланиям необходимо проверить через источники.", Config: &ideaConfig, Reasons: []string{"Альтернативное направление для обсуждения, а не подтверждённое предложение.", "Известные даты, состав гостей и бюджет сохранены как пожелания пользователя, не как проверенные условия поездки."}, Assumptions: []PlanningAssumption{{Field: "trip.destinations", Value: city, Reason: "Направление предложено AI; условия и соответствие ограничениям ещё не проверены."}}, Source: "ai_generated_idea"})
	}
	out.Advice, err = prepareIdeas(out.Advice, in.Config, day)
	if err != nil {
		return nil, upstream(err)
	}

	return &out, nil
}
func hasDestination(c PlanningConfig) bool {
	if c.Trip == nil {
		return false
	}
	for _, p := range c.Trip.Destinations {
		if strings.TrimSpace(p.City) != "" {
			return true
		}
	}
	return false
}
func mergePlanning(base, patch PlanningConfig, cleared []string) (PlanningConfig, error) {
	toMap := func(c PlanningConfig) map[string]any {
		b, _ := json.Marshal(c)
		m := map[string]any{}
		_ = json.Unmarshal(b, &m)
		return m
	}
	var merge func(map[string]any, map[string]any)
	merge = func(dst, src map[string]any) {
		for k, v := range src {
			if v == nil || v == "" {
				continue
			}
			if values, ok := v.([]any); ok && len(values) == 0 {
				continue
			}
			if child, ok := v.(map[string]any); ok {
				old, _ := dst[k].(map[string]any)
				if old == nil {
					old = map[string]any{}
				}
				merge(old, child)
				dst[k] = old
			} else {
				dst[k] = v
			}
		}
	}
	result := toMap(base)
	merge(result, toMap(patch))
	allowed := map[string]bool{}
	var paths func(reflect.Type, string)
	paths = func(t reflect.Type, prefix string) {
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			path := prefix + name
			allowed[path] = true
			paths(f.Type, path+".")
		}
	}
	paths(reflect.TypeOf(PlanningConfig{}), "")
	for _, path := range cleared {
		if !allowed[path] || path == "version" || !strings.HasPrefix(path, "trip.") {
			return PlanningConfig{}, fmt.Errorf("unsupported cleared field")
		}
		parts := strings.Split(path, ".")
		current := result
		for _, part := range parts[:len(parts)-1] {
			next, _ := current[part].(map[string]any)
			if next == nil {
				current = nil
				break
			}
			current = next
		}
		if current != nil {
			delete(current, parts[len(parts)-1])
		}
	}
	result["version"] = "1.0"
	data, _ := json.Marshal(result)
	var out PlanningConfig
	if err := json.Unmarshal(data, &out); err != nil {
		return out, err
	}
	if out.Trip != nil && out.Trip.Travelers != nil {
		guests := out.Trip.Travelers
		if patch.Trip != nil && patch.Trip.Travelers != nil && patch.Trip.Travelers.Children != nil && *patch.Trip.Travelers.Children == 0 && len(patch.Trip.Travelers.ChildrenAges) == 0 {
			guests.ChildrenAges = []int{}
		}
		if guests.Children == nil && len(guests.ChildrenAges) > 0 {
			count := len(guests.ChildrenAges)
			guests.Children = &count
		}
	}
	return out, nil
}
func prepareIdeas(ideas []TravelIdea, base PlanningConfig, day time.Time) ([]TravelIdea, error) {
	if len(ideas) < 2 || len(ideas) > 3 {
		return nil, fmt.Errorf("provide 2..3 ideas")
	}
	ids, destinations := map[string]bool{}, map[string]bool{}
	for i := range ideas {
		idea := &ideas[i]
		if idea.ID == "" || ids[idea.ID] || strings.TrimSpace(idea.Title) == "" || strings.TrimSpace(idea.Description) == "" || idea.Config == nil || !hasDestination(*idea.Config) {
			return nil, fmt.Errorf("invalid idea")
		}
		ids[idea.ID] = true
		proposal := PlanningConfig{Trip: &PlanningTrip{Destinations: idea.Config.Trip.Destinations}}
		merged, err := mergePlanning(base, proposal, nil)
		if err != nil {
			return nil, err
		}
		idea.Config = &merged

		key := strings.ToLower(strings.TrimSpace(idea.Config.Trip.Destinations[0].City))
		if destinations[key] {
			return nil, fmt.Errorf("ideas must have different destinations")
		}
		destinations[key] = true
		for _, issue := range planningIssues(*idea.Config, day) {
			if !issue.Required {
				return nil, fmt.Errorf("invalid optional idea field")
			}
		}
		idea.Source = "ai_generated_idea"
		if idea.Reasons == nil {
			idea.Reasons = []string{}
		}
		if idea.Assumptions == nil {
			idea.Assumptions = []PlanningAssumption{}
		}
	}
	return ideas, nil
}
