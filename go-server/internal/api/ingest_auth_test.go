package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIngestNoKeyConfigured(t *testing.T) {
	store := testStore(t)
	middleware := ingestKeyMiddleware(store)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("POST", "/ingest", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 without key configured, got %d", w.Code)
	}
}

func TestIngestWithKeyMissing(t *testing.T) {
	store := testStore(t)
	store.SetSetting("ingest_api_key", "secret-key-123")

	middleware := ingestKeyMiddleware(store)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("POST", "/ingest", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without key, got %d", w.Code)
	}
}

func TestIngestWithWrongKey(t *testing.T) {
	store := testStore(t)
	store.SetSetting("ingest_api_key", "secret-key-123")

	middleware := ingestKeyMiddleware(store)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("POST", "/ingest", nil)
	req.Header.Set("X-Wiretap-Key", "wrong-key")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with wrong key, got %d", w.Code)
	}
}

func TestIngestWithCorrectKey(t *testing.T) {
	store := testStore(t)
	store.SetSetting("ingest_api_key", "secret-key-123")

	middleware := ingestKeyMiddleware(store)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("POST", "/ingest", nil)
	req.Header.Set("X-Wiretap-Key", "secret-key-123")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with correct key, got %d", w.Code)
	}
}

func TestIngestHealthAlwaysPublic(t *testing.T) {
	store := testStore(t)
	store.SetSetting("ingest_api_key", "secret-key-123")

	middleware := ingestKeyMiddleware(store)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for /health without key, got %d", w.Code)
	}
}
