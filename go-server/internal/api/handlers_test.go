package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetSettingsEmpty(t *testing.T) {
	store := testStore(t)
	h := &handlers{store: store}

	req := httptest.NewRequest("GET", "/api/settings", nil)
	w := httptest.NewRecorder()

	h.getSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 0 {
		t.Errorf("expected empty settings, got %v", result)
	}
}

func TestUpdateAndGetSettings(t *testing.T) {
	store := testStore(t)
	h := &handlers{store: store}

	// PUT settings
	body := `{"app_url":"http://localhost:8000","replay_secret":"wt_test"}`
	req := httptest.NewRequest("PUT", "/api/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.updateSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// GET settings
	req2 := httptest.NewRequest("GET", "/api/settings", nil)
	w2 := httptest.NewRecorder()

	h.getSettings(w2, req2)

	var result map[string]string
	json.NewDecoder(w2.Body).Decode(&result)

	if result["app_url"] != "http://localhost:8000" {
		t.Errorf("expected app_url=http://localhost:8000, got %s", result["app_url"])
	}
	if result["replay_secret"] != "wt_test" {
		t.Errorf("expected replay_secret=wt_test, got %s", result["replay_secret"])
	}
}

func TestUpdateSettingsIgnoresUnknownKeys(t *testing.T) {
	store := testStore(t)
	h := &handlers{store: store}

	body := `{"app_url":"http://test","unknown_key":"should_be_ignored"}`
	req := httptest.NewRequest("PUT", "/api/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.updateSettings(w, req)

	// Check that unknown_key was not saved
	val, _ := store.GetSetting("unknown_key")
	if val != "" {
		t.Errorf("unknown_key should not be saved, got %q", val)
	}

	// Check that app_url was saved
	val, _ = store.GetSetting("app_url")
	if val != "http://test" {
		t.Errorf("expected http://test, got %q", val)
	}
}

func TestUpdateSettingsThresholds(t *testing.T) {
	store := testStore(t)
	h := &handlers{store: store}

	body := `{"slow_query_ms":"50","slow_request_ms":"500"}`
	req := httptest.NewRequest("PUT", "/api/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.updateSettings(w, req)

	q, _ := store.GetSetting("slow_query_ms")
	r, _ := store.GetSetting("slow_request_ms")
	if q != "50" || r != "500" {
		t.Errorf("expected 50/500, got %s/%s", q, r)
	}
}

func TestUpdateSettingsRetentionDays(t *testing.T) {
	store := testStore(t)
	h := &handlers{store: store}

	body := `{"retention_days":"30"}`
	req := httptest.NewRequest("PUT", "/api/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.updateSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	val, _ := store.GetSetting("retention_days")
	if val != "30" {
		t.Errorf("expected retention_days=30, got %q", val)
	}
}

func TestUpdateSettingsAllAllowedKeys(t *testing.T) {
	store := testStore(t)
	h := &handlers{store: store}

	body := `{
		"app_url":"http://example.com",
		"replay_secret":"secret",
		"slow_query_ms":"100",
		"slow_request_ms":"1000",
		"retention_days":"7"
	}`
	req := httptest.NewRequest("PUT", "/api/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.updateSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	checks := map[string]string{
		"app_url":         "http://example.com",
		"replay_secret":   "secret",
		"slow_query_ms":   "100",
		"slow_request_ms": "1000",
		"retention_days":  "7",
	}
	for key, expected := range checks {
		val, _ := store.GetSetting(key)
		if val != expected {
			t.Errorf("expected %s=%s, got %q", key, expected, val)
		}
	}
}

func TestUpdateSettingsInvalidJSON(t *testing.T) {
	store := testStore(t)
	h := &handlers{store: store}

	req := httptest.NewRequest("PUT", "/api/settings", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	h.updateSettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
