package api

import (
	"net/http"

	"github.com/axvi/wiretap/internal/storage"
)

func ingestKeyMiddleware(store *storage.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Health endpoint is always public
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			key, _ := store.GetSetting("ingest_api_key")
			// If no key configured, allow all (backwards compatible)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			provided := r.Header.Get("X-Wiretap-Key")
			if provided != key {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"invalid ingest key"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
