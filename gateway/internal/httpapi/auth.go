package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/godnota/gateway/internal/auth"
	"google.golang.org/grpc/codes"
)

type limiter struct {
	mu      sync.Mutex
	entries map[string]attempt
}
type attempt struct {
	count int
	until time.Time
}

func (l *limiter) allow(r *http.Request) bool {
	key, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		key = r.RemoteAddr
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	for k, v := range l.entries {
		if !now.Before(v.until) {
			delete(l.entries, k)
		}
	}
	a := l.entries[key]
	if a.count == 0 {
		if len(l.entries) >= 1024 {
			return false
		}
		a.until = now.Add(time.Minute)
	}
	a.count++
	l.entries[key] = a
	return a.count <= 10
}
func bearer(r *http.Request) string {
	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
func authenticate(store *auth.Store, r *http.Request) (auth.User, error) {
	if store == nil {
		return auth.User{}, &Error{Code: int(codes.Unauthenticated), Message: "authentication required"}
	}
	user, err := store.Authenticate(bearer(r))
	if err != nil {
		return auth.User{}, &Error{Code: int(codes.Unauthenticated), Message: "invalid or expired token"}
	}
	return user, nil
}
func registerAuth(mux *http.ServeMux, store *auth.Store) {
	limit := &limiter{entries: map[string]attempt{}}
	for _, action := range []string{"register", "login"} {
		mux.HandleFunc("POST /api/auth/"+action, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			if store == nil {
				writeError(w, &Error{Code: int(codes.Unavailable), Message: "authentication unavailable"})
				return
			}
			if !limit.allow(r) {
				writeError(w, &Error{Code: int(codes.ResourceExhausted), Message: "too many authentication attempts"})
				return
			}
			raw, err := requestDTO(w, r, route{})
			if err != nil {
				writeError(w, err)
				return
			}
			var input auth.Credentials
			dec := json.NewDecoder(strings.NewReader(string(raw)))
			dec.DisallowUnknownFields()
			if dec.Decode(&input) != nil {
				writeError(w, &Error{Code: invalidArgument, Message: "email and password required"})
				return
			}
			var result auth.Login
			if action == "register" {
				result, err = store.Register(input)
			} else {
				result, err = store.Login(input)
			}
			if err != nil {
				code, message := internalCode, "account storage unavailable"
				switch {
				case errors.Is(err, auth.ErrInput):
					code, message = invalidArgument, err.Error()
				case errors.Is(err, auth.ErrExists):
					code, message = int(codes.AlreadyExists), err.Error()
				case errors.Is(err, auth.ErrCredentials):
					code, message = int(codes.Unauthenticated), err.Error()
				}
				writeError(w, &Error{Code: code, Message: message})
				return
			}
			if action == "register" {
				w.WriteHeader(http.StatusCreated)
			}
			_ = json.NewEncoder(w).Encode(result)
		})
	}
	mux.HandleFunc("GET /api/auth/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		u, e := authenticate(store, r)
		if e != nil {
			writeError(w, e)
			return
		}
		_ = json.NewEncoder(w).Encode(u)
	})
	mux.HandleFunc("POST /api/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if _, e := authenticate(store, r); e != nil {
			writeError(w, e)
			return
		}
		if err := store.Logout(bearer(r)); err != nil {
			writeError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func authorize(ctx context.Context, backend Backend, user auth.User, spec route, raw json.RawMessage) (json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	get := func(key string) string { var value string; _ = json.Unmarshal(fields[key], &value); return value }
	denied := func() (json.RawMessage, error) {
		return nil, &Error{Code: int(codes.PermissionDenied), Message: "access denied"}
	}
	if value := get("user_id"); value != "" && value != user.ID {
		return denied()
	}
	if spec.operation == "/v1/session/create" || spec.operation == "/v1/order/create" || strings.HasPrefix(spec.operation, "/v1/travel/") {
		fields["user_id"], _ = json.Marshal(user.ID)
	}
	check := func(operation, idKey, id, resource string) error {
		body, _ := json.Marshal(map[string]string{idKey: id})
		response, err := backend.Call(ctx, operation, body)
		if err != nil {
			return err
		}
		var owner struct {
			UserID string `json:"user_id"`
		}
		if json.Unmarshal(response, &owner) != nil || owner.UserID != user.ID {
			return &Error{Code: notFound, Message: resource + " not found"}
		}
		return nil
	}
	if sid := get("session_id"); sid != "" {
		if err := check("/v1/session/get", "session_id", sid, "session"); err != nil {
			return nil, err
		}
		if spec.operation == "/v1/order/create" {
			if _, exists := fields["route_snapshot"]; exists {
				return nil, &Error{Code: invalidArgument, Message: "route snapshot is supplied by the server"}
			}
			id := get("route_id")
			if id == "" {
				return nil, &Error{Code: invalidArgument, Message: "route_id required"}
			}
			body, _ := json.Marshal(map[string]string{"session_id": sid, "user_id": user.ID})
			response, err := backend.Call(ctx, "/v1/travel/routes", body)
			if err != nil {
				return nil, err
			}
			var routes struct {
				Candidates []json.RawMessage `json:"candidates"`
			}
			if json.Unmarshal(response, &routes) != nil {
				return nil, &Error{Code: internalCode, Message: "invalid route response"}
			}
			for _, raw := range routes.Candidates {
				var candidate struct {
					ID string `json:"id"`
				}
				if json.Unmarshal(raw, &candidate) == nil && candidate.ID == id {
					fields["route_snapshot"] = raw
					break
				}
			}
			if _, ok := fields["route_snapshot"]; !ok {
				return nil, &Error{Code: notFound, Message: "route not found in current search"}
			}
		}
	} else if strings.HasPrefix(spec.operation, "/v1/travel/") || spec.operation == "/v1/order/create" {
		return nil, &Error{Code: invalidArgument, Message: "session_id required"}
	}
	if oid := get("order_id"); oid != "" {
		if err := check("/v1/order/get", "order_id", oid, "order"); err != nil {
			return nil, err
		}
	}
	return json.Marshal(fields)
}
