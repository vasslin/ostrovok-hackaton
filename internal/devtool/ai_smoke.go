package devtool

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func AISmoke(ctx context.Context, base string, out io.Writer) error {
	api := NewAPI(base)
	api.Client = &http.Client{Timeout: 390 * time.Second}
	account := map[string]string{"email": strings.ToLower(rand.Text()) + "@example.com", "password": rand.Text()}
	registered, err := api.Call(ctx, "POST", "/api/auth/register", "", account, 201)
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}
	token, _ := registered["access_token"].(string)
	if token == "" {
		return fmt.Errorf("missing access token")
	}
	created, err := api.Call(ctx, "POST", "/api/sessions", token, map[string]any{}, 200)
	if err != nil {
		return err
	}
	id, _ := created["session_id"].(string)
	if id == "" {
		return fmt.Errorf("missing session ID")
	}
	path := "/api/sessions/" + id
	start := time.Now()
	initial, err := api.Call(ctx, "POST", path+"/messages", token, map[string]any{"text": "Хочу из Москвы в Сочи. Даты и состав пока уточняем."}, 200)
	if err != nil {
		return fmt.Errorf("real AI clarification: %w", err)
	}
	questions, _ := initial["questions"].([]any)
	if len(questions) == 0 || len(questions) > 2 || initial["ready_for_search"] == true {
		return fmt.Errorf("AI did not ask for missing trip fields")
	}
	fmt.Fprintf(out, "PASS: real AI clarification (%0.1fs)\n", time.Since(start).Seconds())
	from := time.Now().UTC().AddDate(0, 0, 10).Format(time.DateOnly)
	to := time.Now().UTC().AddDate(0, 0, 17).Format(time.DateOnly)
	text := fmt.Sprintf("Из Москвы в Сочи с %s по %s, двое взрослых, бюджет 80000 рублей на всю поездку.", from, to)
	answers := []map[string]any{}
	for _, question := range questions {
		q, ok := question.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid question")
		}
		answers = append(answers, map[string]any{"question_id": q["id"], "text": text})
	}
	completed, err := api.Call(ctx, "POST", path+"/answers", token, map[string]any{"answers": answers}, 200)
	if err != nil {
		return fmt.Errorf("real AI answers: %w", err)
	}
	if completed["ready_for_search"] != true {
		return fmt.Errorf("AI did not complete trip: missing fields %v", completed["missing_fields"])
	}
	stored, err := api.Call(ctx, "GET", path+"/config", token, nil, 200)
	if err != nil {
		return err
	}
	config, _ := stored["config"].(map[string]any)
	request, _ := config["request"].(map[string]any)
	travelers, _ := request["travelers"].(map[string]any)
	dates, _ := request["dates"].(map[string]any)
	if travelers["adults"] != float64(2) || dates["departure_date"] != from || dates["return_date"] != to {
		return fmt.Errorf("saved config differs from user facts")
	}
	fmt.Fprintln(out, "PASS: real AI answers and session config persistence")
	if _, err = api.Call(ctx, "POST", path+"/confirm", token, map[string]any{}, 200); err != nil {
		return fmt.Errorf("confirm: %w", err)
	}
	found, err := api.Call(ctx, "POST", path+"/search", token, map[string]any{}, 200)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}
	parts, _ := found["parts"].(map[string]any)
	hotels, _ := parts["hotel_parts"].([]any)
	transports, _ := parts["transport_parts"].([]any)
	if len(hotels)+len(transports) == 0 {
		raw, _ := json.Marshal(found)
		return fmt.Errorf("no provider parts: %s", raw)
	}
	fmt.Fprintf(out, "PASS: confirmed search, %d transport / %d hotel parts; provider mode follows configuration\n", len(transports), len(hotels))
	return nil
}
