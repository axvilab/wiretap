package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/axvi/wiretap/internal/storage"
)

func testStore(t *testing.T) *storage.Store {
	t.Helper()
	f, err := os.CreateTemp("", "wiretap-api-test-*.duckdb")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()
	os.Remove(path) // DuckDB needs to create the file itself

	t.Cleanup(func() {
		os.Remove(path)
		os.Remove(path + ".wal")
	})

	s, err := storage.New(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestConfigHandlerDisabled(t *testing.T) {
	store := testStore(t)
	rh := newReplayHandler(store)

	req := httptest.NewRequest("GET", "/api/replay/config", nil)
	w := httptest.NewRecorder()

	rh.configHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != `{"enabled":false}` {
		t.Errorf("expected disabled, got %s", w.Body.String())
	}
}

func TestConfigHandlerEnabled(t *testing.T) {
	store := testStore(t)
	store.SetSetting("app_url", "http://localhost:8000")
	store.SetSetting("replay_secret", "wt_test123")

	rh := newReplayHandler(store)

	req := httptest.NewRequest("GET", "/api/replay/config", nil)
	w := httptest.NewRecorder()

	rh.configHandler(w, req)

	if w.Body.String() != `{"enabled":true}` {
		t.Errorf("expected enabled, got %s", w.Body.String())
	}
}

func TestConfigHandlerPartialConfig(t *testing.T) {
	store := testStore(t)
	// Only app_url set, no secret
	store.SetSetting("app_url", "http://localhost:8000")

	rh := newReplayHandler(store)

	req := httptest.NewRequest("GET", "/api/replay/config", nil)
	w := httptest.NewRecorder()

	rh.configHandler(w, req)

	if w.Body.String() != `{"enabled":false}` {
		t.Errorf("expected disabled with partial config, got %s", w.Body.String())
	}
}

func TestReplayRequestNoConfig(t *testing.T) {
	store := testStore(t)
	rh := newReplayHandler(store)

	req := httptest.NewRequest("POST", "/api/replay/request", nil)
	w := httptest.NewRecorder()

	rh.replayRequest(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestProxyNoConfig(t *testing.T) {
	store := testStore(t)
	rh := newReplayHandler(store)

	handler := rh.proxy("retry-job")
	req := httptest.NewRequest("POST", "/api/replay/job", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}
