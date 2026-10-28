package devtool

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type API struct {
	URL    string
	Client *http.Client
}

func NewAPI(url string) *API {
	return &API{URL: strings.TrimSuffix(url, "/"), Client: &http.Client{Timeout: 10 * time.Second}}
}
func (a *API) Call(ctx context.Context, method, path, token string, body any, want int) (map[string]any, error) {
	var encoded []byte
	var err error
	if body != nil {
		encoded, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, a.URL+path, bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if response.StatusCode != want {
		return nil, fmt.Errorf("%s %s: HTTP %d, expected %d", method, path, response.StatusCode, want)
	}
	if len(raw) == 0 {
		return nil, nil
	}
	var result map[string]any
	if err = json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("%s %s: invalid JSON response", method, path)
	}
	return result, nil
}
func ExampleTrip(now time.Time) map[string]any {
	return map[string]any{
		"schema_version": "1.0", "locale": "ru", "currency": "RUB",
		"origin": map[string]any{"city": "Москва", "country_code": "RU"}, "destination": map[string]any{"city": "Казань", "country_code": "RU"},
		"dates":     map[string]any{"departure_date": now.AddDate(0, 0, 7).Format(time.DateOnly), "return_date": now.AddDate(0, 0, 9).Format(time.DateOnly), "nights": 2},
		"travelers": map[string]any{"adults": 2, "children_ages": []int{8}}, "transport": map[string]any{"modes": []string{"train", "plane"}, "round_trip": true, "direct_only": true},
		"hotel": map[string]any{"needed": true, "rooms": 1, "star_rating_min": 3, "meal": "breakfast", "free_cancellation": true, "price_per_night_max": 8000},
	}
}
func ExampleResolved() map[string]any {
	return map[string]any{"origin": map[string]any{"iata": "MOW", "yandex_code": "c213", "latitude": 55.7558, "longitude": 37.6173}, "destination": map[string]any{"iata": "KZN", "yandex_code": "c43", "latitude": 55.7987, "longitude": 49.1059}}
}
func Smoke(ctx context.Context, url string, out io.Writer) error {
	api := NewAPI(url)
	login, err := api.Call(ctx, "POST", "/api/auth/register", "", map[string]string{"email": "demo-" + strings.ToLower(rand.Text()) + "@example.com", "password": rand.Text()}, 201)
	if err != nil {
		return err
	}
	token, _ := login["access_token"].(string)
	if token == "" {
		return fmt.Errorf("registration did not return a token")
	}
	session, err := api.Call(ctx, "POST", "/api/sessions", token, map[string]any{}, 200)
	if err != nil {
		return err
	}
	id, _ := session["session_id"].(string)
	if id == "" {
		return fmt.Errorf("session id missing")
	}
	path := "/api/sessions/" + id
	trip := ExampleTrip(time.Now())
	for _, step := range []struct {
		method, path string
		body         any
	}{{"POST", "/api/providers/map", map[string]any{"config": trip, "resolved": ExampleResolved()}}, {"PATCH", path + "/config", map[string]any{"config": map[string]any{"version": "1.0", "request": trip}}}, {"POST", path + "/confirm", map[string]any{}}} {
		if _, err = api.Call(ctx, step.method, step.path, token, step.body, 200); err != nil {
			return err
		}
	}
	result, err := api.Call(ctx, "POST", path+"/search", token, map[string]any{}, 200)
	if err != nil {
		return err
	}
	parts, ok := result["parts"].(map[string]any)
	if !ok {
		return fmt.Errorf("parts missing")
	}
	hotels, _ := parts["hotel_parts"].([]any)
	transport, _ := parts["transport_parts"].([]any)
	if len(hotels) == 0 || len(transport) == 0 {
		return fmt.Errorf("demo parts missing; smoke expects demo mode")
	}
	if _, err = api.Call(ctx, "POST", "/api/auth/logout", token, nil, 204); err != nil {
		return err
	}
	fmt.Fprintf(out, "PASS: register, mapping, session, confirmation, %d transport / %d hotel parts, logout\n", len(transport), len(hotels))
	return nil
}
