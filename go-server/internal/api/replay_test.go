package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/axvi/wiretap/internal/config"
)

func TestConfigHandlerDisabled(t *testing.T) {
	rh := newReplayHandler(&config.Config{})

	req := httptest.NewRequest("GET", "/api/replay/config", nil)
	w := httptest.NewRecorder()
	rh.configHandler(w, req)

	if w.Body.String() != `{"enabled":false}` {
		t.Errorf("expected disabled, got %s", w.Body.String())
	}
}

func TestConfigHandlerEnabled(t *testing.T) {
	rh := newReplayHandler(&config.Config{
		AppURL:       "http://localhost:8000",
		ReplaySecret: "wt_test",
	})

	req := httptest.NewRequest("GET", "/api/replay/config", nil)
	w := httptest.NewRecorder()
	rh.configHandler(w, req)

	if w.Body.String() != `{"enabled":true}` {
		t.Errorf("expected enabled, got %s", w.Body.String())
	}
}

func TestConfigHandlerPartialConfig(t *testing.T) {
	rh := newReplayHandler(&config.Config{AppURL: "http://localhost:8000"})

	req := httptest.NewRequest("GET", "/api/replay/config", nil)
	w := httptest.NewRecorder()
	rh.configHandler(w, req)

	if w.Body.String() != `{"enabled":false}` {
		t.Errorf("expected disabled with partial config, got %s", w.Body.String())
	}
}

func TestReplayRequestNoConfig(t *testing.T) {
	rh := newReplayHandler(&config.Config{})

	req := httptest.NewRequest("POST", "/api/replay/request", nil)
	w := httptest.NewRecorder()
	rh.replayRequest(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestProxyNoConfig(t *testing.T) {
	rh := newReplayHandler(&config.Config{})

	handler := rh.proxy("retry-job")
	req := httptest.NewRequest("POST", "/api/replay/job", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}