package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestLocalCompletion(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("local request must not send cloud credentials")
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		format, _ := payload["response_format"].(map[string]any)
		if format["type"] != "json_schema" {
			t.Error("local output must use structured schema")
		}
		if payload["reasoning_effort"] != "none" || payload["model"] != "qwen3:4b" {
			t.Error("wrong local parameters")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": `{"ok":true}`}, "finish_reason": "stop"}}})
	}))
	defer provider.Close()
	client := NewLocal(provider.URL, "qwen3:4b")
	var result struct {
		OK bool `json:"ok"`
	}
	if !client.Configured() {
		t.Fatal("local mode should not require key")
	}
	if err := client.Complete(context.Background(), "Return JSON", map[string]string{"message": "test"}, &result); err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatal("invalid decoded result")
	}
}
func TestLocalQueueHonorsDeadline(t *testing.T) {
	var calls atomic.Int32
	started, release := make(chan struct{}), make(chan struct{})
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		close(started)
		<-release
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{}"},"finish_reason":"stop"}]}`))
	}))
	defer provider.Close()
	defer close(release)
	client := NewLocal(provider.URL, "qwen3:4b")
	done := make(chan error, 1)
	go func() { done <- client.Complete(context.Background(), "JSON", nil, new(map[string]any)) }()
	<-started
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := client.Complete(ctx, "JSON", nil, new(map[string]any)); err != context.DeadlineExceeded {
		t.Fatalf("got %v", err)
	}
	if calls.Load() != 1 {
		t.Fatal("queued call reached model")
	}
	go func() { <-done }()
}

func TestOfflineCompletion(t *testing.T) {
	client := NewOffline()
	if !client.Configured() {
		t.Fatal("offline mode must be configured without a key")
	}
	var plan struct {
		Status  string `json:"status"`
		Summary string `json:"summary"`
	}
	if err := client.Complete(context.Background(), "Ты извлекаешь", map[string]string{"message": "Москва Сочи"}, &plan); err != nil {
		t.Fatal(err)
	}
	if plan.Status != "questions" || plan.Summary == "" {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}
