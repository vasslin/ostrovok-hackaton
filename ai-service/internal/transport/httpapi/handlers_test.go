package httpapi

import (
	"ai-service/internal/adapter/agent"
	"ai-service/internal/service"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPI(t *testing.T) {
	tests := []struct {
		name, path, body, output string
		upstream, want           int
	}{
		{"extract", "/v1/config", `{"message":"В Сочи"}`, `{"config":{"destination":"Сочи"},"questions":["Откуда едете?"],"missing_fields":[]}`, 200, 200},
		{"bad dates from model", "/v1/config", `{"message":"поездка"}`, `{"config":{"departure_date":"2026-11-10","return_date":"2026-11-01"}}`, 200, 502},
		{"rank", "/v1/rank", `{"hotels":[{"name":"Hotel A","id":"a","details":{"price":10}}]}`, `{"results":[{"id":"a","score":90,"reason":"Подходит","risks":[]}]}`, 200, 200},
		{"invented ID", "/v1/rank", `{"hotels":[{"name":"Hotel A","id":"a"}]}`, `{"results":[{"id":"invented","score":90,"reason":"ok"}]}`, 200, 502},
		{"duplicate IDs", "/v1/rank", `{"hotels":[{"name":"Hotel A","id":"a"},{"name":"Hotel A","id":"a"}]}`, `{}`, 200, 400},
		{"invalid JSON", "/v1/config", `{"message":`, `{}`, 200, 400},
		{"quota", "/v1/config", `{"message":"Привет"}`, `{}`, 429, 429},
		{"bad JSON from model", "/v1/config", `{"message":"Привет"}`, `oops`, 200, 502},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != "/chat/completions" || r.Header.Get("Authorization") != "Bearer test-key" {
					t.Error("invalid provider request")
				}
				w.Header().Set("Retry-After", "2")
				w.WriteHeader(tt.upstream)
				_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": tt.output}, "finish_reason": "stop"}}})
			}))
			defer provider.Close()
			api := New(service.New(agent.New(provider.URL, "test-key", "test-model")), "secret")
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(tt.body))
			req.Header.Set("Authorization", "Bearer secret")
			rec := httptest.NewRecorder()
			api.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("status %d want %d: %s", rec.Code, tt.want, rec.Body.String())
			}
			if tt.want == 400 && calls != 0 {
				t.Error("invalid request consumed quota")
			}
			if calls > 1 {
				t.Error("unexpected retry")
			}
			if tt.want == 429 && rec.Header().Get("Retry-After") != "2" {
				t.Error("missing retry hint")
			}
			if tt.name == "extract" {
				var out service.ExtractOutput
				_ = json.Unmarshal(rec.Body.Bytes(), &out)
				if len(out.MissingFields) != 4 {
					t.Errorf("missing fields not computed: %s", rec.Body.String())
				}
			}
		})
	}
}
func TestConfigurationAndAuth(t *testing.T) {
	api := New(service.New(agent.New("http://unused", "", "test")), "secret")
	for _, tt := range []struct {
		method, path, auth string
		status             int
	}{{"GET", "/health", "", 200}, {"GET", "/ready", "", 503}, {"POST", "/v1/config", "", 401}, {"POST", "/v1/config", "Bearer secret", 503}} {
		r := httptest.NewRequest(tt.method, tt.path, nil)
		r.Header.Set("Authorization", tt.auth)
		w := httptest.NewRecorder()
		api.ServeHTTP(w, r)
		if w.Code != tt.status {
			t.Errorf("%s: got %d want %d", tt.path, w.Code, tt.status)
		}
	}
}

func TestDocumentation(t *testing.T) {
	t.Setenv("WEB_DIR", "../../../web")
	api := New(service.New(agent.New("http://unused", "", "test")), "secret")
	for _, test := range []struct {
		path, contains string
		status         int
	}{
		{"/", "", 307},
		{"/docs", "SwaggerUIBundle", 200},
		{"/openapi.json", `"openapi": "3.0.3"`, 200},
		{"/docs/assets/swagger-ui-bundle.js", "SwaggerUIBundle", 200},
		{"/docs/assets/swagger-ui.css", ".swagger-ui", 200},
	} {
		w := httptest.NewRecorder()
		api.ServeHTTP(w, httptest.NewRequest("GET", test.path, nil))
		if w.Code != test.status || !strings.Contains(w.Body.String(), test.contains) {
			t.Errorf("%s: status=%d", test.path, w.Code)
		}
	}
	w := httptest.NewRecorder()
	api.ServeHTTP(w, httptest.NewRequest("GET", "/openapi.json", nil))
	var spec struct {
		Paths map[string]map[string]struct {
			OperationID string `json:"operationId"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/health", "/ready", "/v1/config", "/v1/rank"} {
		if len(spec.Paths[path]) != 1 {
			t.Errorf("missing documented operation %s", path)
		}
	}
}

func TestInvalidRequestsDoNotCallProvider(t *testing.T) {
	api := New(service.New(agent.New("http://unused", "test-key", "test")), "")
	for _, test := range []struct {
		body, contentType string
		status            int
	}{
		{`{"message":"test"}`, "text/plain", 415},
		{`{"message":"` + strings.Repeat("x", 33000) + `"}`, "application/json", 413},
		{`{"message":"test","config":{"currency":"rub"}}`, "application/json", 400},
		{`{"message":"test","config":{"origin":" "}}`, "application/json", 400},
		{`{"message":"test","extra":true}`, "application/json", 400},
		{`{"message":"test"} {}`, "application/json", 400},
		{`null`, "application/json", 400},
	} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/v1/config", strings.NewReader(test.body))
		r.Header.Set("Content-Type", test.contentType)
		api.ServeHTTP(w, r)
		if w.Code != test.status {
			t.Errorf("got %d want %d: %s", w.Code, test.status, w.Body.String())
		}
	}
}
