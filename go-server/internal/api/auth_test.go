package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/axvi/wiretap/internal/config"
)

func testConfig(password string) *config.Config {
	return &config.Config{Password: password}
}

func TestStatusNoAuth(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store, testConfig(""))

	req := httptest.NewRequest("GET", "/api/auth/status", nil)
	w := httptest.NewRecorder()
	am.statusHandler(w, req)

	var result map[string]bool
	json.NewDecoder(w.Body).Decode(&result)

	if result["auth_enabled"] {
		t.Error("expected auth_enabled=false when no password")
	}
}

func TestStatusWithAuth(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store, testConfig("secret"))

	req := httptest.NewRequest("GET", "/api/auth/status", nil)
	w := httptest.NewRecorder()
	am.statusHandler(w, req)

	var result map[string]bool
	json.NewDecoder(w.Body).Decode(&result)

	if !result["auth_enabled"] {
		t.Error("expected auth_enabled=true")
	}
	if result["authenticated"] {
		t.Error("expected authenticated=false without cookie")
	}
}

func TestLoginSuccess(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store, testConfig("mypassword"))

	body := `{"password":"mypassword"}`
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	w := httptest.NewRecorder()
	am.loginHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	found := false
	for _, c := range w.Result().Cookies() {
		if c.Name == "wiretap_session" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Error("expected session cookie")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store, testConfig("mypassword"))

	body := `{"password":"wrong"}`
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	w := httptest.NewRecorder()
	am.loginHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLoginNoPasswordConfigured(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store, testConfig(""))

	body := `{"password":"anything"}`
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	w := httptest.NewRecorder()
	am.loginHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 when no password configured, got %d", w.Code)
	}
}

func TestMiddlewareAllowsWithoutPassword(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store, testConfig(""))

	handler := am.middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 without password, got %d", w.Code)
	}
}

func TestMiddlewareBlocksWithPassword(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store, testConfig("secret"))

	handler := am.middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without cookie, got %d", w.Code)
	}
}

func TestMiddlewareAllowsWithCookie(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store, testConfig("secret"))

	// Login first
	loginReq := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"password":"secret"}`))
	loginW := httptest.NewRecorder()
	am.loginHandler(loginW, loginReq)

	var sessionCookie *http.Cookie
	for _, c := range loginW.Result().Cookies() {
		if c.Name == "wiretap_session" {
			sessionCookie = c
		}
	}

	handler := am.middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	req.AddCookie(sessionCookie)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with cookie, got %d", w.Code)
	}
}

func TestMiddlewareAllowsAuthEndpoints(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store, testConfig("secret"))

	handler := am.middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	for _, path := range []string{"/api/auth/login", "/api/auth/status", "/api/health"} {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200 for %s, got %d", path, w.Code)
		}
	}
}

func TestLogoutInvalidatesSession(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store, testConfig("secret"))

	// Login
	loginReq := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"password":"secret"}`))
	loginW := httptest.NewRecorder()
	am.loginHandler(loginW, loginReq)

	var sessionCookie *http.Cookie
	for _, c := range loginW.Result().Cookies() {
		if c.Name == "wiretap_session" {
			sessionCookie = c
		}
	}

	// Logout
	logoutReq := httptest.NewRequest("POST", "/api/auth/logout", nil)
	logoutReq.AddCookie(sessionCookie)
	am.logoutHandler(httptest.NewRecorder(), logoutReq)

	// Try using old cookie
	handler := am.middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	req.AddCookie(sessionCookie)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 after logout, got %d", w.Code)
	}
}