package api

import (
	"net/http"

	"github.com/axvi/wiretap/internal/config"
)

func ingestKeyMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Health endpoint is always public
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			// If no key configured, allow all
			if cfg.IngestKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			provided := r.Header.Get("X-Wiretap-Key")
			if provided != cfg.IngestKey {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"invalid ingest key"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}