package httpapi

import (
	"net/http"
	"os"
	"strings"
)

func corsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false
		origins := os.Getenv("CORS_ALLOWED_ORIGINS")
		if origins == "" {
			origins = "http://localhost:8080,http://127.0.0.1:8080,http://localhost:5173"
		}
		for _, value := range strings.Split(origins, ",") {
			if origin != "" && origin == strings.TrimSpace(value) {
				allowed = true
			}
		}
		w.Header().Add("Vary", "Origin")
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization, Idempotency-Key")
		w.Header().Set("Access-Control-Expose-Headers", "Location")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
