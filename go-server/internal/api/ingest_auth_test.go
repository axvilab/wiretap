package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/axvi/wiretap/internal/config"
)

func TestIngestNoKeyConfigured(t *testing.T) {
	cfg := &config.Config{IngestKey: ""}
	middleware := ingestKeyMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("POST", "/ingest", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 without key, got %d", w.Code)
	}
}

func TestIngestWithKeyMissing(t *testing.T) {
	cfg := &config.Config{IngestKey: "secret-key-123"}
	middleware := ingestKeyMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("POST", "/ingest", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestIngestWithCorrectKey(t *testing.T) {
	cfg := &config.Config{IngestKey: "secret-key-123"}
	middleware := ingestKeyMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("POST", "/ingest", nil)
	req.Header.Set("X-Wiretap-Key", "secret-key-123")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestIngestHealthAlwaysPublic(t *testing.T) {
	cfg := &config.Config{IngestKey: "secret-key-123"}
	middleware := ingestKeyMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for /health, got %d", w.Code)
	}
}