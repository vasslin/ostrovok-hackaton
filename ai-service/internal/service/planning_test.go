package service

import (
	"ai-service/internal/adapter/agent"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func completeConfig() PlanningConfig {
	var c PlanningConfig
	_ = json.Unmarshal([]byte(`{"version":"1.0","trip":{"origin":{"city":"Москва"},"destinations":[{"city":"Сочи"}],"date_range":{"start_date":"2026-11-10","end_date":"2026-11-17"},"travelers":{"adults":2},"budget":{"amount":80000,"currency":"RUB","scope":"total"},"transport":{"max_transfers":0},"preferences":{"avoid_overnight_transfers":false}}}`), &c)
	return c
}
func planningService(t *testing.T, completion any, inspect func(map[string]any)) (*Service, *int) {
	t.Helper()
	calls := new(int)
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*calls++
		if inspect != nil {
			var body struct {
				Messages []struct {
					Content string `json:"content"`
				} `json:"messages"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			var in map[string]any
			if len(body.Messages) < 2 {
				t.Error("missing model messages")
			} else {
				_ = json.Unmarshal([]byte(body.Messages[1].Content), &in)
				inspect(in)
			}
		}
		data, _ := json.Marshal(completion)
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": string(data)}, "finish_reason": "stop"}}})
	}))
	t.Cleanup(provider.Close)
	return New(agent.New(provider.URL, "test", "model")), calls
}
func TestPlanningReadinessAndPreservation(t *testing.T) {
	base := completeConfig()
	patch := PlanningConfig{Trip: &PlanningTrip{Budget: &PlanningBudget{Amount: ptr(60000.0)}}}
	svc, calls := planningService(t, planningCompletion{Status: "questions", Config: &patch, Evidence: []PlanningEvidence{{Field: "trip.budget.amount", Quote: "60000"}}, Questions: []PlanningQuestion{{Field: "trip.origin.city", Text: "Откуда?"}}}, func(in map[string]any) {
		if in["current_date"] != "2026-10-28" || in["text"] != "Уменьши бюджет до 60000" {
			t.Error("latest message or date missing")
		}
		ctx, _ := in["context"].(map[string]any)
		if ctx["summary"] != "Избегать ночных пересадок" {
			t.Error("summary not forwarded")
		}
	})
	out, err := svc.Analyze(context.Background(), PlanningInput{Text: "Уменьши бюджет до 60000", Config: base, Context: PlanningContext{Summary: "Избегать ночных пересадок"}, CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != "generate" || len(out.Questions) != 0 || len(out.MissingFields) != 0 {
		t.Fatalf("wrong readiness: %+v", out)
	}
	if *out.Config.Trip.Budget.Amount != 60000 || *out.Config.Trip.Budget.Currency != "RUB" || out.Config.Trip.Origin.City != "Москва" || *out.Config.Trip.Transport.MaxTransfers != 0 || *out.Config.Trip.Preferences.AvoidOvernightTransfers {
		t.Fatalf("lost old fields: %+v", out.Config.Trip)
	}
	if *calls != 1 {
		t.Fatal("more than one model call")
	}
}
func ptr[T any](value T) *T { return &value }
func TestPlanningCannotForceIncompleteConfig(t *testing.T) {
	for _, force := range []bool{false, true} {
		partial := PlanningConfig{Trip: &PlanningTrip{Origin: &PlanningPlace{City: "Москва"}}}
		svc, _ := planningService(t, planningCompletion{Status: "generate", Config: &partial}, nil)
		out, err := svc.Analyze(context.Background(), PlanningInput{Text: "Давай искать", ForceGenerate: force, CurrentDate: "2026-10-28"})
		if err != nil {
			t.Fatal(err)
		}
		if out.Status != "questions" || len(out.Questions) != 2 || len(out.MissingFields) < 4 || (out.Config.Trip != nil && out.Config.Trip.DateRange != nil) {
			t.Fatalf("unsafe readiness: %+v", out)
		}
	}
}
func TestPlanningRejectsInvalidDatesAndChildAges(t *testing.T) {
	c := completeConfig()
	c.Trip.DateRange.EndDate = ptr("2026-11-01")
	c.Trip.Travelers.Children = ptr(1)
	svc, _ := planningService(t, planningCompletion{Status: "generate", Config: &c}, nil)
	out, err := svc.Analyze(context.Background(), PlanningInput{Text: "Проверь", Config: c, CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != "questions" || !containsString(out.MissingFields, "trip.date_range.end_date") || !containsString(out.MissingFields, "trip.travelers.children_ages") {
		t.Fatalf("invalid config marked ready: %+v", out)
	}
}
func sampleIdeas() []TravelIdea {
	a, b := completeConfig(), completeConfig()
	b.Trip.Destinations[0].City = "Казань"
	return []TravelIdea{{ID: "sea", Title: "Море", Description: "Идея, требуется проверка", Config: &a}, {ID: "city", Title: "Город", Description: "Идея, требуется проверка", Config: &b}}
}
func TestAdviceAndSelection(t *testing.T) {
	empty := PlanningConfig{Trip: &PlanningTrip{Origin: &PlanningPlace{City: "Москва"}}}
	ideas := sampleIdeas()
	svc, _ := planningService(t, planningCompletion{Status: "advice", Config: &empty, Advice: ideas}, nil)
	out, err := svc.Analyze(context.Background(), PlanningInput{Text: "Предложи поездку", CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != "advice" || hasDestination(*out.Config) || len(out.Advice) != 0 {
		t.Fatalf("wrong advice: %+v", out)
	}
	patch := PlanningConfig{Trip: &PlanningTrip{Budget: &PlanningBudget{Amount: ptr(50000.0)}}}
	selected, _ := planningService(t, planningCompletion{Status: "generate", Config: &patch, Evidence: []PlanningEvidence{{Field: "trip.budget.amount", Quote: "50000"}}}, nil)
	result, err := selected.Analyze(context.Background(), PlanningInput{Text: "Первый вариант, бюджет 50000", Config: empty, SelectedIdea: &ideas[0], CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "generate" || result.Config.Trip.Destinations[0].City != "Сочи" || *result.Config.Trip.Budget.Amount != 50000 {
		t.Fatalf("selection failed: %+v", result)
	}
}
func TestAnswerBatchAndExplicitClear(t *testing.T) {
	base := completeConfig()
	svc, _ := planningService(t, planningCompletion{Status: "generate", Config: &PlanningConfig{}, ClearedFields: []string{"trip.budget"}, Evidence: []PlanningEvidence{{Field: "trip.budget", Quote: "Ограничение бюджета убираем"}}}, func(in map[string]any) {
		if len(in["answers"].([]any)) != 2 {
			t.Error("answer order not forwarded")
		}
	})
	out, err := svc.Analyze(context.Background(), PlanningInput{Config: base, Questions: []PlanningQuestion{{ID: "budget", Field: "trip.budget.amount"}}, Answers: []PlanningAnswer{{QuestionID: "budget", Text: "60000"}, {QuestionID: "budget", Text: "Ограничение бюджета убираем"}}, CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Config.Trip.Budget != nil || out.Status != "generate" {
		t.Fatal("explicit clear failed")
	}
}
func TestInvalidAnswerDoesNotConsumeModelCall(t *testing.T) {
	svc, calls := planningService(t, planningCompletion{}, nil)
	_, err := svc.Analyze(context.Background(), PlanningInput{Answers: []PlanningAnswer{{QuestionID: "missing", Text: "answer"}}})
	var e *Error
	if !errors.As(err, &e) || e.Status != 400 || *calls != 0 {
		t.Fatalf("wrong validation: %v, calls=%d", err, *calls)
	}
}
func TestRecommendationsUseHistoryAndKeepConstraints(t *testing.T) {
	ideas := recommendationsCompletion{Advice: []ideaProposal{{ID: "sea", City: "Сочи", Theme: "relax"}, {ID: "city", City: "Казань", Theme: "culture"}}}
	base := completeConfig()
	base.Trip.Budget.Amount = ptr(50000.0)
	svc, _ := planningService(t, ideas, func(in map[string]any) {
		if len(in["last_sessions"].([]any)) != 1 {
			t.Error("history missing")
		}
	})
	out, err := svc.Recommendations(context.Background(), RecommendationsInput{Text: "Придумай поездку", Config: base, LastSessions: []PreviousSession{{Context: PlanningContext{Summary: "Любит музеи"}}}, CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	for _, idea := range out.Advice {
		if *idea.Config.Trip.Budget.Amount != 50000 {
			t.Fatal("recommendation ignored known budget")
		}
	}
	ideas.Advice[1].City = "Сочи"
	bad, _ := planningService(t, ideas, nil)
	_, err = bad.Recommendations(context.Background(), RecommendationsInput{CurrentDate: "2026-10-28"})
	var e *Error
	if !errors.As(err, &e) || e.Status != 502 {
		t.Fatal("duplicate ideas accepted")
	}
}

func TestUngroundedDefaultsNeverEnterConfig(t *testing.T) {
	patch := completeConfig()
	patch.Trip.Travelers.Adults = ptr(1)
	patch.Trip.Accommodation = &PlanningAccommodation{MinStars: ptr(3)}
	completion := planningCompletion{Status: "generate", Config: &patch, Evidence: []PlanningEvidence{
		{Field: "trip.origin.city", Quote: "Из Москвы"},
		{Field: "trip.destinations", Quote: "в Сочи"},
		{Field: "trip.travelers.adults", Quote: "Из Москвы в Сочи"},
		{Field: "trip.accommodation.min_stars", Quote: "Из Москвы в Сочи"},
		{Field: "trip.date_range.start_date", Quote: "в ноябре 2026"},
	}}
	svc, _ := planningService(t, completion, nil)
	out, err := svc.Analyze(context.Background(), PlanningInput{Text: "Из Москвы в Сочи в ноябре 2026, даты и состав уточняем", CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != "questions" || out.Config.Trip.Travelers != nil || out.Config.Trip.Accommodation != nil || out.Config.Trip.DateRange != nil {
		t.Fatalf("invented defaults survived: %+v", out.Config.Trip)
	}
	if out.Config.Trip.Origin.City != "Москва" || out.Config.Trip.Destinations[0].City != "Сочи" {
		t.Fatal("supported city facts lost")
	}
}
func TestColloquialOriginAndStartDate(t *testing.T) {
	patch := PlanningConfig{Trip: &PlanningTrip{Origin: &PlanningPlace{City: "Санкт-Петербург"}, DateRange: &PlanningDates{StartDate: ptr("2026-11-01")}}}
	completion := planningCompletion{Status: "questions", Config: &patch, Evidence: []PlanningEvidence{{Field: "trip.origin.city", Quote: "питер"}, {Field: "trip.date_range.start_date", Quote: "1 ноября"}}}
	svc, _ := planningService(t, completion, nil)
	out, err := svc.Analyze(context.Background(), PlanningInput{Text: "город отправления - питер\nдата начала - 1 ноября", CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Config == nil || out.Config.Trip == nil || out.Config.Trip.Origin == nil || out.Config.Trip.Origin.City != "Санкт-Петербург" || out.Config.Trip.DateRange == nil || out.Config.Trip.DateRange.StartDate == nil || *out.Config.Trip.DateRange.StartDate != "2026-11-01" {
		t.Fatalf("colloquial answer was dropped: %+v", out.Config)
	}
}
func TestParaphrasedOriginUpdatesConfig(t *testing.T) {
	patch := PlanningConfig{Trip: &PlanningTrip{Origin: &PlanningPlace{City: "Санкт-Петербург"}, DateRange: &PlanningDates{StartDate: ptr("2026-11-04")}}}
	completion := planningCompletion{Status: "questions", Config: &patch, Evidence: []PlanningEvidence{{Field: "trip.origin.city", Quote: "северной столицы"}, {Field: "trip.date_range.start_date", Quote: "через неделю"}}}
	svc, _ := planningService(t, completion, nil)
	out, err := svc.Analyze(context.Background(), PlanningInput{Text: "Выезжаю из северной столицы через неделю", CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Config == nil || out.Config.Trip == nil || out.Config.Trip.Origin == nil || out.Config.Trip.Origin.City != "Санкт-Петербург" || out.Config.Trip.DateRange == nil || out.Config.Trip.DateRange.StartDate == nil || *out.Config.Trip.DateRange.StartDate != "2026-11-04" {
		t.Fatalf("free wording was dropped: %+v", out.Config)
	}
}
func TestTypoMonthBecomesNovember(t *testing.T) {
	paris := PlanningConfig{Version: "1.0", Trip: &PlanningTrip{Destinations: []PlanningPlace{{City: "Париж"}}}}
	wrong := PlanningConfig{Trip: &PlanningTrip{Origin: &PlanningPlace{City: "Санкт-Петербург"}, DateRange: &PlanningDates{StartDate: ptr("2026-10-29")}, Travelers: &PlanningTravelers{Adults: ptr(2)}}}
	completion := planningCompletion{Status: "questions", Summary: "Вот описание вашей поездки: вы едете из санкт-петербург в париж с выездом 2026-10-29, 2 взрослых.", Config: &wrong, Evidence: []PlanningEvidence{{Field: "trip.origin.city", Quote: "питер"}, {Field: "trip.date_range.start_date", Quote: "1 нябра"}, {Field: "trip.travelers.adults", Quote: "1 нябра"}}}
	svc, _ := planningService(t, completion, nil)
	out, err := svc.AnalyzeCanonical(context.Background(), PlanningInput{Text: "питер, 1 нябра", Config: paris, CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	request := out.Config.Request
	if request.Origin.City != "Санкт-Петербург" || request.Dates.DepartureDate != "2026-11-01" || request.Travelers.Adults != 0 || strings.Contains(out.Summary, "2026-10-29") || strings.Contains(out.Summary, "2 взрослых") {
		t.Fatalf("typo was not corrected: %+v summary=%s", request, out.Summary)
	}
}
func TestBareNicknameFillsMissingOrigin(t *testing.T) {
	paris := PlanningConfig{Version: "1.0", Trip: &PlanningTrip{Destinations: []PlanningPlace{{City: "Париж"}}, DateRange: &PlanningDates{StartDate: ptr("2026-11-03")}}}
	echo := paris
	summary := "Вот описание вашей поездки: вы едете в париж с выездом 2026-11-03. Вам ещё нужно назвать: город отправления, дата возвращения, число взрослых."
	svc, _ := planningService(t, planningCompletion{Status: "questions", Config: &echo, Summary: summary}, nil)
	out, err := svc.AnalyzeCanonical(context.Background(), PlanningInput{Text: "питер", Config: paris, Context: PlanningContext{Summary: summary}, CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	request := out.Config.Request
	if request.Origin.City != "Санкт-Петербург" || request.Destination == nil || request.Destination.City != "Париж" || request.Dates.DepartureDate != "2026-11-03" || request.Dates.ReturnDate != "" || request.Travelers.Adults != 0 {
		t.Fatalf("nickname was not stored: %+v", request)
	}
	if !strings.Contains(out.Summary, "Санкт-Петербург") || !strings.Contains(out.Summary, "2026-11-03") || !strings.Contains(out.Summary, "дата возвращения") || !strings.Contains(out.Summary, "число взрослых") {
		t.Fatal(out.Summary)
	}
}
func TestNamedOriginReplacesPreviousCity(t *testing.T) {
	two := 2
	current := PlanningConfig{Version: "1.0", Trip: &PlanningTrip{Origin: &PlanningPlace{City: "Санкт-Петербург"}, Destinations: []PlanningPlace{{City: "Париж"}}, DateRange: &PlanningDates{StartDate: ptr("2026-11-05"), EndDate: ptr("2026-11-08")}, Travelers: &PlanningTravelers{Adults: &two}}}
	echo := current
	summary := "Вот описание вашей поездки: вы едете из Санкт-Петербург в париж с 2026-11-05 по 2026-11-08, 2 взрослых."
	svc, _ := planningService(t, planningCompletion{Status: "generate", Config: &echo, Summary: summary, Evidence: []PlanningEvidence{{Field: "trip.origin.city", Quote: "Санкт-Петербург"}, {Field: "trip.destinations", Quote: "париж"}, {Field: "trip.date_range.start_date", Quote: "2026-11-05"}, {Field: "trip.date_range.end_date", Quote: "2026-11-08"}, {Field: "trip.travelers.adults", Quote: "2 взрослых"}}}, nil)
	out, err := svc.AnalyzeCanonical(context.Background(), PlanningInput{Text: "измени город отправления на москву", Config: current, Context: PlanningContext{Summary: summary}, CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	request := out.Config.Request
	if request.Origin.City != "Москва" || request.Destination == nil || request.Destination.City != "Париж" || request.Dates.DepartureDate != "2026-11-05" || request.Dates.ReturnDate != "2026-11-08" || request.Travelers.Adults != 2 {
		t.Fatalf("origin was not replaced: %+v", request)
	}
	if !strings.Contains(out.Summary, "Москва") {
		t.Fatal(out.Summary)
	}
}
func TestSingleDateFillsReturnAndKeepsDeparture(t *testing.T) {
	current := PlanningConfig{Version: "1.0", Trip: &PlanningTrip{Origin: &PlanningPlace{City: "Санкт-Петербург"}, Destinations: []PlanningPlace{{City: "Париж"}}, DateRange: &PlanningDates{StartDate: ptr("2026-11-03")}}}
	both := "2026-11-10"
	echo := PlanningConfig{Version: "1.0", Trip: &PlanningTrip{Origin: &PlanningPlace{City: "Санкт-Петербург"}, Destinations: []PlanningPlace{{City: "Париж"}}, DateRange: &PlanningDates{StartDate: &both, EndDate: &both}, Travelers: &PlanningTravelers{Adults: ptr(2)}}}
	summary := "Вот описание вашей поездки: вы едете из Санкт-Петербург в париж с выездом 2026-11-03. Вам ещё нужно назвать: дата возвращения, число взрослых."
	svc, _ := planningService(t, planningCompletion{Status: "questions", Config: &echo, Summary: summary, Evidence: []PlanningEvidence{{Field: "trip.date_range.start_date", Quote: "10 нябра"}, {Field: "trip.date_range.end_date", Quote: "10 нябра"}, {Field: "trip.travelers.adults", Quote: "10 нябра, 2"}}}, nil)
	out, err := svc.AnalyzeCanonical(context.Background(), PlanningInput{Text: "10 нябра, 2", Config: current, Context: PlanningContext{Summary: summary}, CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	request := out.Config.Request
	if request.Dates.DepartureDate != "2026-11-03" || request.Dates.ReturnDate != "2026-11-10" || request.Travelers.Adults != 2 || request.Origin.City != "Санкт-Петербург" {
		t.Fatalf("return date replaced departure: %+v", request)
	}
}
func TestNamedDepartureReplacesPreviousDate(t *testing.T) {
	two := 2
	current := PlanningConfig{Version: "1.0", Trip: &PlanningTrip{Origin: &PlanningPlace{City: "Санкт-Петербург"}, Destinations: []PlanningPlace{{City: "Париж"}}, DateRange: &PlanningDates{StartDate: ptr("2026-11-03"), EndDate: ptr("2026-11-08")}, Travelers: &PlanningTravelers{Adults: &two}}}
	next := "2026-11-10"
	echo := PlanningConfig{Version: "1.0", Trip: &PlanningTrip{Origin: &PlanningPlace{City: "Санкт-Петербург"}, Destinations: []PlanningPlace{{City: "Париж"}}, DateRange: &PlanningDates{StartDate: &next, EndDate: ptr("2026-11-08")}, Travelers: &PlanningTravelers{Adults: &two}}}
	summary := "Вот описание вашей поездки: вы едете из Санкт-Петербург в париж с 2026-11-03 по 2026-11-08, 2 взрослых."
	svc, _ := planningService(t, planningCompletion{Status: "generate", Config: &echo, Summary: summary, Evidence: []PlanningEvidence{{Field: "trip.date_range.start_date", Quote: "10 ноября"}, {Field: "trip.date_range.end_date", Quote: "2026-11-08"}, {Field: "trip.travelers.adults", Quote: "2 взрослых"}}}, nil)
	out, err := svc.AnalyzeCanonical(context.Background(), PlanningInput{Text: "измени дату выезда на 10 ноября", Config: current, Context: PlanningContext{Summary: summary}, CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	request := out.Config.Request
	if request.Dates.DepartureDate != "2026-11-10" || request.Dates.ReturnDate != "2026-11-08" || request.Travelers.Adults != 2 {
		t.Fatalf("named departure was not replaced: %+v", request)
	}
}
func TestInventedQuoteIsNotEvidence(t *testing.T) {
	patch := PlanningConfig{Trip: &PlanningTrip{Travelers: &PlanningTravelers{Adults: ptr(2)}}}
	svc, _ := planningService(t, planningCompletion{Status: "generate", Config: &patch, Evidence: []PlanningEvidence{{Field: "trip.travelers.adults", Quote: "двое взрослых"}}}, nil)
	out, err := svc.Analyze(context.Background(), PlanningInput{Text: "Хочу в Сочи", CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Config.Trip != nil && out.Config.Trip.Travelers != nil {
		t.Fatal("invented evidence accepted")
	}
}

func TestLiteralAnswerPriorityAndNoDefaults(t *testing.T) {
	base := completeConfig()
	out, err := applyLiteralPlanning(base, PlanningInput{Text: "Бюджет всей поездки 60000 рублей", Questions: []PlanningQuestion{{ID: "date", Field: "trip.date_range.start_date"}}, Answers: []PlanningAnswer{{QuestionID: "date", Text: "2026-11-12"}}})
	if err != nil {
		t.Fatal(err)
	}
	if *out.Trip.Budget.Amount != 60000 || *out.Trip.DateRange.StartDate != "2026-11-12" || out.Trip.Origin.City != "Москва" {
		t.Fatal("literal answers were lost")
	}
	empty, _ := applyLiteralPlanning(PlanningConfig{}, PlanningInput{Text: "Хочу из Москвы в Сочи. Даты и состав пока уточняем."})
	if empty.Trip.Travelers != nil || empty.Trip.DateRange != nil || empty.Trip.Destinations[0].City != "Сочи" {
		t.Fatal("literal parser invented unknown facts")
	}
}

func TestChildWithoutAgeCannotGenerate(t *testing.T) {
	base := completeConfig()
	svc, _ := planningService(t, planningCompletion{Status: "generate", Config: &PlanningConfig{}}, nil)
	out, err := svc.Analyze(context.Background(), PlanningInput{Text: "С нами поедет ребенок, возраст пока уточним", Config: base, CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != "questions" || !containsString(out.MissingFields, "trip.travelers.children_ages") {
		t.Fatal("child without age did not block generation")
	}
}
