package service

import (
	"ai-service/internal/canonical"
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

func legacyFromRequest(r canonical.TripRequest) PlanningConfig {
	t := &PlanningTrip{Origin: &PlanningPlace{City: r.Origin.City}, DateRange: &PlanningDates{}, Travelers: &PlanningTravelers{}}
	if r.Origin.CountryCode != "" {
		t.Origin.CountryCode = &r.Origin.CountryCode
	}
	if r.Destination != nil && r.Destination.City != "" {
		t.Destinations = []PlanningPlace{{City: r.Destination.City}}
	}
	if r.Dates.DepartureDate != "" {
		t.DateRange.StartDate = &r.Dates.DepartureDate
	}
	if r.Dates.ReturnDate != "" {
		t.DateRange.EndDate = &r.Dates.ReturnDate
	}
	if r.Travelers.Adults != 0 {
		t.Travelers.Adults = &r.Travelers.Adults
	}
	for _, a := range r.Travelers.ChildrenAges {
		t.Travelers.ChildrenAges = append(t.Travelers.ChildrenAges, int(a))
	}
	if r.Budget != nil {
		b := r.Budget
		amount := float64(b.Total)
		t.Budget = &PlanningBudget{Amount: &amount, Currency: &r.Currency, Scope: &b.Scope, IsHardLimit: &b.IsHard}
	}
	if r.Transport != nil {
		t.Transport = &PlanningTransport{Modes: r.Transport.Modes, MaxTransfers: r.Transport.MaxTransfers}
	}
	if r.Hotel != nil {
		t.Accommodation = &PlanningAccommodation{Required: r.Hotel.Needed, MinStars: r.Hotel.StarRatingMin, PreferredTypes: r.Hotel.Kinds}
	}
	return PlanningConfig{Version: "1.0", Trip: t}
}

func changedPatch(before, after map[string]any) map[string]any {
	result := map[string]any{}
	for key, value := range after {
		if reflect.DeepEqual(value, before[key]) {
			continue
		}
		if child, ok := value.(map[string]any); ok {
			old, _ := before[key].(map[string]any)
			result[key] = changedPatch(old, child)
		} else {
			result[key] = value
		}
	}
	return result
}

func canonicalIssues(r canonical.TripRequest, day time.Time) []PlanningQuestion {
	issues := []PlanningQuestion{}
	seen := map[string]bool{}
	add := func(path, text string) {
		if !seen[path] {
			seen[path] = true
			issues = append(issues, PlanningQuestion{ID: strings.ReplaceAll(strings.TrimPrefix(path, "/"), "/", "_"), Field: path, Text: text, Required: true, InputType: "text"})
		}
	}
	if r.Origin.City == "" {
		add("/origin/city", "Из какого города отправляемся?")
	}
	if r.Destination == nil || r.Destination.City == "" {
		add("/destination/city", "В какой город хотите поехать?")
	}
	if r.Dates.DepartureDate == "" {
		add("/dates/departure_date", "Когда хотите отправиться?")
	}
	if r.Dates.ReturnDate == "" && (r.Transport == nil || r.Transport.RoundTrip == nil || *r.Transport.RoundTrip || r.Hotel == nil || r.Hotel.Needed == nil || *r.Hotel.Needed) {
		add("/dates/return_date", "Укажите дату возвращения или число ночей.")
	}
	if r.Travelers.Adults == 0 {
		add("/travelers/adults", "Сколько взрослых поедет?")
	}
	var validation validator.ValidationErrors
	if err := r.ValidateAt(day); err != nil {
		if errors.As(err, &validation) {
			for _, v := range validation {
				parts := strings.Split(v.Namespace(), ".")
				path := "/" + strings.Join(parts[1:], "/")
				if path == "/dates/departure_month" && r.Dates.DepartureDate == "" {
					continue
				}
				add(path, "Уточните это значение своими словами.")
			}
		}
		for _, line := range strings.Split(err.Error(), "\n") {
			if strings.HasPrefix(line, "/") {
				path, _, _ := strings.Cut(line, ":")
				add(path, "Уточните поле "+path+": дата должна быть в будущем, в пределах года; длительность — до 30 ночей и согласована с датами.")
			}
		}
	}
	return issues
}

func (s *Service) AnalyzeCanonical(ctx context.Context, in PlanningInput) (*PlanningOutput, error) {
	base, err := requestFromConfig(in.Config)
	if err != nil {
		return nil, err
	}
	if base.SchemaVersion == "" {
		base.SchemaVersion = "1.0"
	}
	if base.Locale == "" {
		base.Locale = "ru"
	}
	if base.Currency == "" {
		base.Currency = "RUB"
	}
	original := cloneRequest(base)
	if in.SelectedIdea != nil {
		city := strings.TrimSpace(in.SelectedIdea.City)
		if in.SelectedIdea.Config != nil {
			idea, e := requestFromConfig(*in.SelectedIdea.Config)
			if e != nil {
				return nil, e
			}
			if idea.Destination != nil {
				city = idea.Destination.City
			}
		}
		if in.SelectedIdea.ID == "" || city == "" {
			return nil, invalid("selected_idea requires id and city")
		}
		base.Destination = &canonical.Destination{City: city}
	}
	base, err = canonicalLiteral(base, in)
	if err != nil {
		return nil, err
	}
	legacy := legacyFromRequest(base)
	modelInput := in
	modelInput.Config = legacy
	modelInput.canonicalRequest = &base
	modelInput.SelectedIdea = nil
	modelInput.Questions = append([]PlanningQuestion(nil), in.Questions...)
	for i := range modelInput.Questions {
		modelInput.Questions[i].Field = legacyPath(modelInput.Questions[i].Field)
	}
	if strings.TrimSpace(modelInput.Text) == "" && in.SelectedIdea != nil {
		modelInput.Text = "Выбрано направление: " + base.Destination.City
	}
	out, err := s.Analyze(ctx, modelInput)
	if err != nil {
		return nil, err
	}
	before, _ := requestFromConfig(legacy)
	after, err := requestFromConfig(*out.Config)
	if err != nil {
		return nil, upstream(err)
	}
	patch := changedPatch(requestMap(before), requestMap(after))
	delete(patch, "completeness")
	base, err = mergeRequest(base, patch, deletedPaths(requestMap(before), requestMap(after), ""))
	if err != nil {
		return nil, upstream(err)
	}
	base, err = canonicalLiteral(base, in)
	if err != nil {
		return nil, err
	}
	if base.Dates.DepartureDate != "" && base.Dates.ReturnDate == "" && base.Dates.Nights > 0 && base.Dates.Nights <= 30 {
		if date, e := time.Parse(time.DateOnly, base.Dates.DepartureDate); e == nil {
			base.Dates.ReturnDate = date.AddDate(0, 0, base.Dates.Nights).Format(time.DateOnly)
		}
	}
	day, _ := planningDate(in.CurrentDate)
	issues := canonicalIssues(base, day)
	if mentionsChildren(in) && len(base.Travelers.ChildrenAges) == 0 {
		issues = append(issues, PlanningQuestion{ID: "travelers_children_ages", Field: "/travelers/children_ages", Text: "Какой возраст у каждого ребёнка?", Required: true, InputType: "text"})
	}
	out.Status = "questions"
	out.Questions = []PlanningQuestion{}
	out.MissingFields = []string{}
	for _, q := range issues {
		out.MissingFields = append(out.MissingFields, q.Field)
		if len(out.Questions) < 2 {
			out.Questions = append(out.Questions, q)
		}
	}
	base.Completeness = &canonical.Completeness{MissingRequired: out.MissingFields}
	resolved := in.Config.Resolved
	original.Completeness = nil
	check := cloneRequest(base)
	check.Completeness = nil
	if !reflect.DeepEqual(original, check) {
		resolved = nil
	}
	out.Config = &PlanningConfig{Version: "1.0", Request: &base, Resolved: resolved}
	if text := strings.TrimSpace(out.Summary); text == "" || notesMemory(text) {
		out.Summary = describeTrip(base)
	}
	for i := range out.Assumptions {
		out.Assumptions[i].Field = canonicalPath(out.Assumptions[i].Field)
	}
	if len(issues) == 0 {
		out.Status = "generate"
	} else if !in.ForceGenerate && (base.Destination == nil || base.Destination.City == "") && hasAny(norm(in.Text), "предложи", "посоветуй", "придумай", "не знаю куда", "не определил") {
		out.Status = "advice"
		out.Questions = []PlanningQuestion{}
	}
	return out, nil
}

func (s *Service) RecommendationsCanonical(ctx context.Context, in RecommendationsInput) (*RecommendationsOutput, error) {
	base, err := requestFromConfig(in.Config)
	if err != nil {
		return nil, err
	}
	legacy := in
	legacy.Config = legacyFromRequest(base)
	legacy.canonicalRequest = &base
	legacy.LastSessions = append([]PreviousSession(nil), in.LastSessions...)
	for i := range legacy.LastSessions {
		r, e := requestFromConfig(legacy.LastSessions[i].Config)
		if e != nil {
			return nil, e
		}
		legacy.LastSessions[i].Config = legacyFromRequest(r)
	}
	out, err := s.Recommendations(ctx, legacy)
	if err != nil {
		return nil, err
	}
	for i := range out.Advice {
		idea := &out.Advice[i]
		city := idea.Config.Trip.Destinations[0].City
		r := cloneRequest(base)
		r.Destination = &canonical.Destination{City: city}
		r.Completeness = nil
		idea.Config = &PlanningConfig{Version: "1.0", Request: &r}
		idea.City = city
		idea.Reason = "Идея направления для обсуждения; цены, наличие и соответствие пожеланиям требуют проверки."
	}
	return out, nil
}

func deletedPaths(before, after map[string]any, prefix string) []string {
	paths := []string{}
	for key, value := range before {
		path := prefix + "/" + key
		other, ok := after[key]
		if !ok {
			paths = append(paths, path)
			continue
		}
		if child, ok := value.(map[string]any); ok {
			next, _ := other.(map[string]any)
			paths = append(paths, deletedPaths(child, next, path)...)
		}
	}
	return paths
}
