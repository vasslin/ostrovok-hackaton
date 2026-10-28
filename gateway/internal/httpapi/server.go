package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/godnota/gateway/docs"
	"github.com/godnota/gateway/internal/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Backend interface {
	Call(context.Context, string, json.RawMessage) (json.RawMessage, error)
}

const (
	invalidArgument     = 3
	notFound            = 5
	unimplemented       = 12
	internalCode        = 13
	maxRequestBytes     = 1 << 20
	backendTimeout      = 8 * time.Second
	clientClosedRequest = 499
	missingResource     = "resource not found"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string { return e.Message }

type route struct {
	pattern    string
	operation  string
	pathFields []string
}

var routes = []route{
	{pattern: "POST /api/providers/map", operation: "/v1/map/config"},
	{pattern: "POST /api/sessions", operation: "/v1/session/create"},
	{pattern: "GET /api/sessions/{session_id}", operation: "/v1/session/get", pathFields: []string{"session_id"}},
	{pattern: "POST /api/sessions/{session_id}/messages", operation: "/v1/travel/plan", pathFields: []string{"session_id"}},
	{pattern: "POST /api/sessions/{session_id}/answers", operation: "/v1/travel/answers", pathFields: []string{"session_id"}},
	{pattern: "POST /api/sessions/{session_id}/ideas", operation: "/v1/travel/idea", pathFields: []string{"session_id"}},
	{pattern: "GET /api/sessions/{session_id}/config", operation: "/v1/travel/config/get", pathFields: []string{"session_id"}},
	{pattern: "PATCH /api/sessions/{session_id}/config", operation: "/v1/travel/config/update", pathFields: []string{"session_id"}},
	{pattern: "POST /api/sessions/{session_id}/confirm", operation: "/v1/travel/config/confirm", pathFields: []string{"session_id"}},
	{pattern: "POST /api/sessions/{session_id}/search", operation: "/v1/travel/search", pathFields: []string{"session_id"}},
	{pattern: "GET /api/sessions/{session_id}/routes", operation: "/v1/travel/routes", pathFields: []string{"session_id"}},
	{pattern: "POST /api/orders", operation: "/v1/order/create"},
	{pattern: "GET /api/orders/{order_id}", operation: "/v1/order/get", pathFields: []string{"order_id"}},
	{pattern: "POST /api/orders/{order_id}/pay", operation: "/v1/order/pay", pathFields: []string{"order_id"}},
	{pattern: "POST /api/orders/{order_id}/cancel", operation: "/v1/order/cancel", pathFields: []string{"order_id"}},
	{pattern: "GET /api/users/{user_id}", operation: "/v1/user/get", pathFields: []string{"user_id"}},
	{pattern: "PATCH /api/users/{user_id}", operation: "/v1/user/update", pathFields: []string{"user_id"}},
}

func Handler(backend Backend) http.Handler {
	return HandlerWithAuth(backend, nil)
}

func HandlerWithAuth(backend Backend, store *auth.Store) http.Handler {
	mux := http.NewServeMux()
	registerAuth(mux, store)
	docs.Register(mux)
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs", http.StatusFound)
	})
	for _, spec := range routes {
		mux.HandleFunc(spec.pattern, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			user, err := authenticate(store, r)
			if err != nil {
				writeError(w, err)
				return
			}
			raw, err := requestDTO(w, r, spec)
			if err != nil {
				writeError(w, err)
				return
			}
			slog.Info("gateway request", "method", r.Method, "operation", spec.operation)
			timeout := backendTimeout
			if spec.operation == "/v1/travel/search" {
				timeout = 30 * time.Second
			}
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			raw, err = authorize(ctx, backend, user, spec, raw)
			if err != nil {
				writeError(w, err)
				return
			}
			response, err := backend.Call(ctx, spec.operation, raw)
			if err != nil {
				writeError(w, err)
				return
			}
			if len(response) == 0 {
				response = json.RawMessage(`{}`)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(response)
		})
	}
	return corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if path.Clean(r.URL.EscapedPath()) != r.URL.EscapedPath() {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			writeError(w, &Error{Code: notFound, Message: missingResource})
			return
		}
		fallback, pattern := mux.Handler(r)
		if pattern != "" {
			mux.ServeHTTP(w, r)
			return
		}
		capture := &routingError{header: make(http.Header)}
		fallback.ServeHTTP(capture, r)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if allow := capture.header.Get("Allow"); allow != "" {
			w.Header().Set("Allow", allow)
		}
		code, message := notFound, missingResource
		if capture.status == http.StatusMethodNotAllowed {
			code, message = unimplemented, "method not allowed"
		}
		w.WriteHeader(capture.status)
		_ = json.NewEncoder(w).Encode(&Error{Code: code, Message: message})
	}))
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

type routingError struct {
	header http.Header
	status int
}

func (r *routingError) Header() http.Header { return r.header }

func (r *routingError) WriteHeader(code int) { r.status = code }

func (r *routingError) Write(data []byte) (int, error) { return len(data), nil }

func requestDTO(w http.ResponseWriter, r *http.Request, spec route) (json.RawMessage, error) {
	invalid := func(message string) (json.RawMessage, error) {
		return nil, &Error{Code: invalidArgument, Message: message}
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRequestBytes))
	if err != nil {
		return invalid("invalid body (maximum 1 MiB)")
	}
	fields := map[string]json.RawMessage{}
	if len(strings.TrimSpace(string(body))) > 0 {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodDelete {
			return invalid("this method does not accept a body")
		}
		if json.Unmarshal(body, &fields) != nil || fields == nil {
			return invalid("body must be a JSON object")
		}
	}
	for _, key := range spec.pathFields {
		if _, exists := fields[key]; exists {
			return invalid("path identifiers must not be repeated in the body")
		}
		fields[key], _ = json.Marshal(r.PathValue(key))
	}
	if spec.operation == "/v1/order/create" {
		key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if key == "" || len(key) > 128 {
			return invalid("Idempotency-Key header is required (maximum 128 characters)")
		}
		if _, exists := fields["idempotency_key"]; exists {
			return invalid("use the Idempotency-Key header")
		}
		fields["idempotency_key"], _ = json.Marshal(key)
	}
	for _, alias := range []string{"userId", "sessionId", "orderId"} {
		if _, exists := fields[alias]; exists {
			return invalid("use snake_case identifiers")
		}
	}
	return json.Marshal(fields)
}

func writeError(w http.ResponseWriter, err error) {
	public := &Error{Code: internalCode, Message: "internal error"}
	if exposed, ok := errors.AsType[*Error](err); ok {
		public = exposed
	}
	if st, ok := status.FromError(err); ok && st.Code() != codes.OK {
		public = &Error{Code: int(st.Code()), Message: st.Message()}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		public = &Error{Code: int(codes.DeadlineExceeded), Message: "deadline exceeded"}
	}
	if errors.Is(err, context.Canceled) {
		public = &Error{Code: int(codes.Canceled), Message: "request canceled"}
	}

	httpStatus := map[int]int{
		int(codes.Canceled):           clientClosedRequest,
		int(codes.InvalidArgument):    http.StatusBadRequest,
		int(codes.DeadlineExceeded):   http.StatusGatewayTimeout,
		int(codes.NotFound):           http.StatusNotFound,
		int(codes.AlreadyExists):      http.StatusConflict,
		int(codes.PermissionDenied):   http.StatusForbidden,
		int(codes.ResourceExhausted):  http.StatusTooManyRequests,
		int(codes.FailedPrecondition): http.StatusConflict,
		int(codes.Aborted):            http.StatusConflict,
		int(codes.OutOfRange):         http.StatusBadRequest,
		int(codes.Unimplemented):      http.StatusNotImplemented,
		int(codes.Internal):           http.StatusInternalServerError,
		int(codes.Unavailable):        http.StatusServiceUnavailable,
		int(codes.DataLoss):           http.StatusInternalServerError,
		int(codes.Unauthenticated):    http.StatusUnauthorized,
	}[public.Code]
	if httpStatus == 0 {
		httpStatus = http.StatusInternalServerError
	}
	if httpStatus == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", "Bearer")
	}
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(public)
}
