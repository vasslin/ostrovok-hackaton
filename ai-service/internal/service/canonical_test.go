package service

import (
	"ai-service/internal/canonical"
	"context"
	"encoding/json"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"strings"
	"testing"
)

func canonicalConfig(t *testing.T) PlanningConfig {
	t.Helper()
	request, err := requestFromConfig(completeConfig())
	if err != nil {
		t.Fatal(err)
	}
	request.Travelers.Pets = []canonical.Pet{{Kind: "dog", WeightKg: 4}}
	request.Travelers.Residency = "ru"
	request.Hotel = &canonical.Hotel{Rooms: 2, Meal: "breakfast", Amenities: []string{"wifi"}}
	return PlanningConfig{Version: "1.0", Request: &request, Resolved: map[string]any{"private_resolver_code": "never-to-model"}}
}
func TestCanonicalPreservesFieldsAndSeparatesResolved(t *testing.T) {
	config := canonicalConfig(t)
	svc, _ := planningService(t, planningCompletion{Status: "generate", Config: &PlanningConfig{}}, func(in map[string]any) {
		data, _ := json.Marshal(in)
		if strings.Contains(string(data), "never-to-model") || strings.Contains(string(data), "resolved") {
			t.Error("backend resolver leaked to model")
		}
	})
	out, err := svc.AnalyzeCanonical(context.Background(), PlanningInput{Text: "Проверь конфиг", Config: config, CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	r := out.Config.Request
	if out.Status != "generate" || r.Travelers.Adults != 2 || len(r.Travelers.Pets) != 1 || r.Hotel.Rooms != 2 || r.Hotel.Meal != "breakfast" || out.Config.Resolved["private_resolver_code"] != "never-to-model" {
		t.Fatalf("lost fields: %+v", out)
	}
	config.Request.Origin.City = "Москва"
	changed, err := svc.AnalyzeCanonical(context.Background(), PlanningInput{Text: "Бюджет всей поездки 60000 рублей", Config: config, CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	if changed.Config.Request.Budget.Total != 60000 || changed.Config.Resolved != nil {
		t.Fatal("stale resolver or old budget survived")
	}
}
func TestCanonicalIncompleteAndStrictResponse(t *testing.T) {
	svc, _ := planningService(t, planningCompletion{Status: "generate", Config: &PlanningConfig{}}, nil)
	out, err := svc.AnalyzeCanonical(context.Background(), PlanningInput{Text: "Хочу из Москвы в Сочи", ForceGenerate: true, CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != "questions" || out.Config.Request.Travelers.Adults != 0 || len(out.Questions) > 2 {
		t.Fatal("invented mandatory fields")
	}
	data, _ := json.Marshal(StrictAIResponse(out))
	var m map[string]json.RawMessage
	_ = json.Unmarshal(data, &m)
	if len(m) != 4 || m["summary"] != nil || m["status"] == nil || m["config"] == nil || m["advice"] == nil || m["questions"] == nil {
		t.Fatal("strict response does not match main contract")
	}
}
func TestCanonicalAcceptsRawAndRejectsAmbiguousOrUnknown(t *testing.T) {
	for _, data := range []string{`{"request":{},"trip":{}}`, `{"schema_version":"1.0","resolver_code":"made-up"}`, `{"request":{"hotel":{"invented":true}}}`} {
		var c PlanningConfig
		if json.Unmarshal([]byte(data), &c) == nil {
			t.Fatalf("accepted %s", data)
		}
	}
	var raw PlanningConfig
	if err := json.Unmarshal([]byte(`{"schema_version":"1.0","origin":{"city":"Москва"},"dates":{"departure_month":"2026-11"},"travelers":{"adults":2,"pets":[{"kind":"dog"}]}}`), &raw); err != nil || raw.Request == nil || raw.Trip != nil {
		t.Fatal("raw TripRequest rejected", err)
	}
}
func TestCanonicalSelectedIdeaAndLiterals(t *testing.T) {
	config := canonicalConfig(t)
	config.Request.Destination = nil
	svc, _ := planningService(t, planningCompletion{Status: "generate", Config: &PlanningConfig{}}, nil)
	out, err := svc.AnalyzeCanonical(context.Background(), PlanningInput{Config: config, SelectedIdea: &TravelIdea{ID: "idea-1", City: "Казань"}, Text: "С нами собака 5 кг, нужен отель с бесплатной отменой, 2 номера", CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	r := out.Config.Request
	if r.Destination.City != "Казань" || !r.Hotel.FreeCancellation || r.Hotel.Rooms != 2 || r.Travelers.Pets[0].WeightKg != 5 {
		t.Fatalf("literal/idea lost: %+v", r)
	}
}
func TestCanonicalExplicitBudgetRemoval(t *testing.T) {
	config := canonicalConfig(t)
	svc, _ := planningService(t, planningCompletion{Status: "generate", Config: &PlanningConfig{}, ClearedFields: []string{"trip.budget"}, Evidence: []PlanningEvidence{{Field: "trip.budget", Quote: "Ограничение бюджета убираем"}}}, nil)
	out, err := svc.AnalyzeCanonical(context.Background(), PlanningInput{Config: config, Text: "Ограничение бюджета убираем", CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Config.Request.Budget != nil {
		t.Fatal("explicit removal ignored")
	}
}

func TestGenerateMatchesMainSchemas(t *testing.T) {
	config := canonicalConfig(t)
	svc, _ := planningService(t, planningCompletion{Status: "generate", Config: &PlanningConfig{}}, nil)
	out, err := svc.AnalyzeCanonical(context.Background(), PlanningInput{Config: config, Text: "Проверь", CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != "generate" {
		t.Fatal("config is incomplete")
	}
	for _, test := range []struct {
		file  string
		value any
	}{{"trip_request.schema.json", out.Config.Request}, {"ai_response.schema.json", StrictAIResponse(out)}} {
		compiler := jsonschema.NewCompiler()
		compiler.AssertFormat()
		schema, err := compiler.Compile("../../schemas/" + test.file)
		if err != nil {
			t.Fatal(err)
		}
		data, _ := json.Marshal(test.value)
		var value any
		_ = json.Unmarshal(data, &value)
		if err := schema.Validate(value); err != nil {
			t.Fatalf("%s: %v", test.file, err)
		}
	}
}
