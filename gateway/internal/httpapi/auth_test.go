package httpapi

import (
	"context"
	"encoding/json"
	"github.com/godnota/gateway/internal/auth"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

type backendFunc func(context.Context, string, json.RawMessage) (json.RawMessage, error)

func (f backendFunc) Call(ctx context.Context, op string, raw json.RawMessage) (json.RawMessage, error) {
	return f(ctx, op, raw)
}
func TestAuthorization(t *testing.T) {
	s, _ := auth.Open(filepath.Join(t.TempDir(), "auth.json"))
	login, e := s.Register(auth.Credentials{Email: "a@example.com", Password: "long-demo-password"})
	if e != nil {
		t.Fatal(e)
	}
	calls := 0
	b := backendFunc(func(_ context.Context, op string, raw json.RawMessage) (json.RawMessage, error) {
		calls++
		if op == "/v1/session/get" {
			return json.RawMessage(`{"user_id":"someone-else"}`), nil
		}
		if op == "/v1/session/create" {
			var body map[string]string
			_ = json.Unmarshal(raw, &body)
			if body["user_id"] != login.User.ID {
				t.Error("identity not injected")
			}
		}
		return json.RawMessage(`{}`), nil
	})
	h := HandlerWithAuth(b, s)
	for _, test := range []struct {
		method, path, body, token string
		code                      int
	}{{"POST", "/api/sessions", "{}", "", 401}, {"POST", "/api/sessions", "{}", login.AccessToken, 200}, {"POST", "/api/sessions", `{"user_id":"forged"}`, login.AccessToken, 403}, {"GET", "/api/sessions/foreign", "", login.AccessToken, 404}, {"POST", "/api/sessions/foreign/search", "{}", login.AccessToken, 404}, {"GET", "/api/users/foreign", "", login.AccessToken, 403}, {"POST", "/api/sessions", `{"userId":"forged"}`, login.AccessToken, 400}} {
		r := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
		if test.token != "" {
			r.Header.Set("Authorization", "Bearer "+test.token)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != test.code {
			t.Errorf("%s %s got %d: %s", test.method, test.path, w.Code, w.Body.String())
		}
	}
	if calls != 3 {
		t.Fatalf("unauthorized backend call, count=%d", calls)
	}
}
func TestTravelRequestCarriesUserID(t *testing.T) {
	user := auth.User{ID: "mine"}
	backend := backendFunc(func(_ context.Context, op string, _ json.RawMessage) (json.RawMessage, error) {
		if op == "/v1/session/get" {
			return json.RawMessage(`{"user_id":"mine"}`), nil
		}
		return json.RawMessage(`{}`), nil
	})
	raw, err := authorize(context.Background(), backend, user, route{operation: "/v1/travel/search"}, json.RawMessage(`{"session_id":"s"}`))
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]string
	if err = json.Unmarshal(raw, &body); err != nil || body["user_id"] != "mine" || body["session_id"] != "s" {
		t.Fatalf("%s %v", raw, err)
	}
}
func TestCORS(t *testing.T) {
	h := Handler(backendFunc(func(context.Context, string, json.RawMessage) (json.RawMessage, error) { return nil, nil }))
	for _, origin := range []string{"https://untrusted.example", "http://localhost:5173"} {
		r := httptest.NewRequest("OPTIONS", "/api/sessions", nil)
		r.Header.Set("Origin", origin)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if origin == "https://untrusted.example" && w.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatal("untrusted origin allowed")
		}
		if origin == "http://localhost:5173" && !strings.Contains(w.Header().Get("Access-Control-Allow-Headers"), "Authorization") {
			t.Fatal("bearer header not allowed")
		}
	}
}

func TestForeignOrdersDeniedBeforeMutation(t *testing.T) {
	user := auth.User{ID: "mine"}
	mutations := 0
	backend := backendFunc(func(_ context.Context, op string, _ json.RawMessage) (json.RawMessage, error) {
		if op == "/v1/order/get" || op == "/v1/session/get" {
			return json.RawMessage(`{"user_id":"someone-else"}`), nil
		}
		mutations++
		return json.RawMessage(`{}`), nil
	})
	for _, op := range []string{"/v1/order/get", "/v1/order/pay", "/v1/order/cancel"} {
		if _, err := authorize(context.Background(), backend, user, route{operation: op}, json.RawMessage(`{"order_id":"foreign"}`)); err == nil {
			t.Fatalf("foreign order allowed: %s", op)
		}
	}
	if _, err := authorize(context.Background(), backend, user, route{operation: "/v1/order/create"}, json.RawMessage(`{"session_id":"foreign","route_id":"r"}`)); err == nil {
		t.Fatal("order on foreign session allowed")
	}
	if mutations != 0 {
		t.Fatal("unauthorized mutation reached backend")
	}
}
