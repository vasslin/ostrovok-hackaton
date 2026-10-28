package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/godnota/gateway/internal/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func request(t *testing.T, h http.Handler, method, path, token, body string, code int) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != code {
		t.Fatalf("%s %s: HTTP %d, want %d: %s", method, path, w.Code, code, w.Body.String())
	}
	return w
}
func TestHTTPAuthenticationIntegration(t *testing.T) {
	s, err := auth.Open(filepath.Join(t.TempDir(), "auth.json"))
	if err != nil {
		t.Fatal(err)
	}
	b := backendFunc(func(context.Context, string, json.RawMessage) (json.RawMessage, error) { return nil, nil })
	h := HandlerWithAuth(b, s)
	request(t, h, "POST", "/api/auth/register", "", `{"email":"not-email","password":"short"}`, 400)
	request(t, h, "POST", "/api/auth/register", "", `{"extra":true}`, 400)
	request(t, h, "POST", "/api/auth/register", "", `null`, 400)
	credentials := `{"email":"a@example.com","password":"long-demo-password"}`
	response := request(t, h, "POST", "/api/auth/register", "", credentials, 201)
	var registered auth.Login
	if json.Unmarshal(response.Body.Bytes(), &registered) != nil || registered.AccessToken == "" {
		t.Fatal("missing token")
	}
	request(t, h, "POST", "/api/auth/register", "", credentials, 409)
	request(t, h, "POST", "/api/auth/login", "", `{"email":"a@example.com","password":"wrong-password"}`, 401)
	loggedIn := request(t, h, "POST", "/api/auth/login", "", credentials, 200)
	var login auth.Login
	_ = json.Unmarshal(loggedIn.Body.Bytes(), &login)
	request(t, h, "GET", "/api/auth/me", registered.AccessToken, "", 401)
	request(t, h, "GET", "/api/auth/me", login.AccessToken, "", 200)
	request(t, h, "POST", "/api/auth/logout", login.AccessToken, "", 204)
	request(t, h, "POST", "/api/auth/logout", login.AccessToken, "", 401)
	request(t, h, "GET", "/api/auth/me", login.AccessToken, "", 401)
	for i := 0; i < 6; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{}`)))
		if i == 5 && w.Code != 429 {
			t.Fatalf("auth limiter missing: %d", w.Code)
		}
	}
	noStore := Handler(b)
	request(t, noStore, "POST", "/api/auth/login", "", credentials, 503)
}
func TestRequestValidationAndRouting(t *testing.T) {
	s, _ := auth.Open(filepath.Join(t.TempDir(), "auth.json"))
	login, err := s.Register(auth.Credentials{Email: "owner@example.com", Password: "long-demo-password"})
	if err != nil {
		t.Fatal(err)
	}
	backend := backendFunc(func(_ context.Context, op string, _ json.RawMessage) (json.RawMessage, error) {
		if op == "/v1/session/get" {
			raw, _ := json.Marshal(map[string]string{"user_id": login.User.ID})
			return raw, nil
		}
		return nil, nil
	})
	h := HandlerWithAuth(backend, s)
	token := login.AccessToken
	request(t, h, "GET", "/health", "", "", 200)
	request(t, h, "GET", "/", "", "", 302)
	request(t, h, "POST", "/api/sessions", token, `[]`, 400)
	request(t, h, "POST", "/api/sessions", token, `{broken}`, 400)
	request(t, h, "POST", "/api/sessions/s/messages", token, `{"session_id":"forged"}`, 400)
	request(t, h, "GET", "/api/sessions/s", token, `{}`, 400)
	request(t, h, "POST", "/api/sessions", token, strings.Repeat("x", (1<<20)+1), 400)
	request(t, h, "POST", "/api/sessions", token, `{}`, 200)
	request(t, h, "GET", "/api/sessions/s", token, "", 200)
	request(t, h, "GET", "/unknown", "", "", 404)
	request(t, h, "PUT", "/api/sessions", token, `{}`, 405)
	request(t, h, "GET", "/api//sessions", token, "", 404)
}
func TestErrorStatusesAndAuthorizationFailures(t *testing.T) {
	for _, test := range []struct {
		err  error
		code int
	}{{errors.New("secret"), 500}, {&Error{Code: 999, Message: "unknown"}, 500}, {status.Error(codes.NotFound, "missing"), 404}, {status.Error(codes.PermissionDenied, "denied"), 403}, {context.Canceled, 499}, {context.DeadlineExceeded, 504}, {status.Error(codes.Unauthenticated, "login"), 401}, {status.Error(codes.Unavailable, "down"), 503}} {
		w := httptest.NewRecorder()
		writeError(w, test.err)
		if w.Code != test.code {
			t.Fatalf("%v => %d", test.err, w.Code)
		}
		if strings.Contains(w.Body.String(), "secret") {
			t.Fatal("private error leaked")
		}
	}
	user := auth.User{ID: "owner"}
	for _, response := range []string{`{broken}`, `{"user_id":"other"}`} {
		b := backendFunc(func(context.Context, string, json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(response), nil
		})
		if _, err := authorize(context.Background(), b, user, route{operation: "/v1/travel/search"}, json.RawMessage(`{"session_id":"s"}`)); err == nil {
			t.Fatal("invalid ownership response accepted")
		}
	}
	failed := backendFunc(func(context.Context, string, json.RawMessage) (json.RawMessage, error) {
		return nil, context.DeadlineExceeded
	})
	if _, err := authorize(context.Background(), failed, user, route{operation: "/v1/travel/search"}, json.RawMessage(`{"session_id":"s"}`)); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("ownership failure lost")
	}
	if _, err := authorize(context.Background(), failed, user, route{operation: "/v1/order/create"}, json.RawMessage(`{}`)); err == nil {
		t.Fatal("order without session accepted")
	}
}
