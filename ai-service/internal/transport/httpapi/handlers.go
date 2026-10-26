package httpapi

import (
	"ai-service/internal/service"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

func reply(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func fail(w http.ResponseWriter, status int, message string) {
	reply(w, status, map[string]string{"error": message})
}

func decode(w http.ResponseWriter, r *http.Request, target any) bool {
	if contentType := r.Header.Get("Content-Type"); contentType != "" {
		media, _, err := mime.ParseMediaType(contentType)
		if err != nil || media != "application/json" {
			fail(w, 415, "Content-Type must be application/json")
			return false
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		var limit *http.MaxBytesError
		if errors.As(err, &limit) {
			fail(w, 413, "body exceeds 32 KiB")
		} else {
			fail(w, 400, "invalid JSON")
		}

		return false
	}

	if err := d.Decode(new(any)); err != io.EOF {
		var limit *http.MaxBytesError
		if errors.As(err, &limit) {
			fail(w, 413, "body exceeds 32 KiB")
			return false
		}

		fail(w, 400, "expected one JSON object")
		return false
	}

	return true
}
func New(svc *service.Service, serviceKey string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { reply(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		if !svc.Configured() {
			fail(w, 503, "LLM_API_KEY is not configured")
			return
		}
		reply(w, 200, map[string]string{"status": "configured", "model": svc.Model()})
	})

	webDir := os.Getenv("WEB_DIR")
	if webDir == "" {
		webDir = "web"
	}

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(webDir, "index.html"))
	})
	mux.HandleFunc("GET /docs/{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, filepath.Join(webDir, "openapi.json"))
	})
	mux.Handle("GET /docs/assets/", http.StripPrefix("/docs/assets/", http.FileServer(http.Dir(filepath.Join(webDir, "swagger")))))
	protect := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if serviceKey != "" && subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+serviceKey)) != 1 {
				fail(w, 401, "unauthorized")
				return
			}
			if !svc.Configured() {
				fail(w, 503, "LLM_API_KEY is not configured")
				return
			}
			next(w, r)
		}
	}
	mux.HandleFunc("POST /v1/config", protect(func(w http.ResponseWriter, r *http.Request) {
		var in service.ExtractInput
		if !decode(w, r, &in) {
			return
		}
		out, err := svc.Extract(r.Context(), in)
		if err != nil {
			serviceError(w, err)
			return
		}
		reply(w, 200, out)
	}))
	mux.HandleFunc("POST /v1/rank", protect(func(w http.ResponseWriter, r *http.Request) {
		var in service.RankInput
		if !decode(w, r, &in) {
			return
		}
		out, err := svc.RankHotels(r.Context(), in)
		if err != nil {
			serviceError(w, err)
			return
		}
		reply(w, 200, out)
	}))
	return mux
}

func serviceError(w http.ResponseWriter, err error) {
	var e *service.Error
	if errors.As(err, &e) {
		if e.RetryAfter != "" {
			w.Header().Set("Retry-After", e.RetryAfter)
		}
		fail(w, e.Status, e.Message)
		return
	}
	fail(w, 500, "internal error")
}
