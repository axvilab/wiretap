package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStatusNoSetup(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store)

	req := httptest.NewRequest("GET", "/api/auth/status", nil)
	w := httptest.NewRecorder()
	am.statusHandler(w, req)

	var result map[string]bool
	json.NewDecoder(w.Body).Decode(&result)

	if result["setup_done"] {
		t.Error("expected setup_done=false")
	}
	if result["authenticated"] {
		t.Error("expected authenticated=false")
	}
}

func TestSetupCreatesPassword(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store)

	body := `{"password":"test123"}`
	req := httptest.NewRequest("POST", "/api/auth/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	am.setupHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Check cookie is set
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "wiretap_session" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Error("expected session cookie after setup")
	}

	// Check setup is done
	if !am.isSetupDone() {
		t.Error("expected setup to be done")
	}
}

func TestSetupRejectsShortPassword(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store)

	body := `{"password":"abc"}`
	req := httptest.NewRequest("POST", "/api/auth/setup", strings.NewReader(body))
	w := httptest.NewRecorder()

	am.setupHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSetupCannotRunTwice(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store)

	// First setup
	body := `{"password":"test123"}`
	req := httptest.NewRequest("POST", "/api/auth/setup", strings.NewReader(body))
	w := httptest.NewRecorder()
	am.setupHandler(w, req)

	// Second setup
	req2 := httptest.NewRequest("POST", "/api/auth/setup", strings.NewReader(body))
	w2 := httptest.NewRecorder()
	am.setupHandler(w2, req2)

	if w2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 on second setup, got %d", w2.Code)
	}
}

func TestLoginSuccess(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store)

	// Setup first
	setupReq := httptest.NewRequest("POST", "/api/auth/setup", strings.NewReader(`{"password":"test123"}`))
	am.setupHandler(httptest.NewRecorder(), setupReq)

	// Login
	body := `{"password":"test123"}`
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	w := httptest.NewRecorder()
	am.loginHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
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
	am := newAuthManager(store)

	setupReq := httptest.NewRequest("POST", "/api/auth/setup", strings.NewReader(`{"password":"test123"}`))
	am.setupHandler(httptest.NewRecorder(), setupReq)

	body := `{"password":"wrong"}`
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	w := httptest.NewRecorder()
	am.loginHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestMiddlewareAllowsWithoutSetup(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store)

	handler := am.middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 without setup, got %d", w.Code)
	}
}

func TestMiddlewareBlocksAfterSetup(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store)

	setupReq := httptest.NewRequest("POST", "/api/auth/setup", strings.NewReader(`{"password":"test123"}`))
	am.setupHandler(httptest.NewRecorder(), setupReq)

	handler := am.middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	// No cookie
	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without cookie, got %d", w.Code)
	}
}

func TestMiddlewareAllowsWithCookie(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store)

	// Setup
	setupReq := httptest.NewRequest("POST", "/api/auth/setup", strings.NewReader(`{"password":"test123"}`))
	setupW := httptest.NewRecorder()
	am.setupHandler(setupW, setupReq)

	// Get session cookie
	var sessionCookie *http.Cookie
	for _, c := range setupW.Result().Cookies() {
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
	am := newAuthManager(store)

	setupReq := httptest.NewRequest("POST", "/api/auth/setup", strings.NewReader(`{"password":"test123"}`))
	am.setupHandler(httptest.NewRecorder(), setupReq)

	handler := am.middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	for _, path := range []string{"/api/auth/login", "/api/auth/setup", "/api/auth/status", "/api/health"} {
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
	am := newAuthManager(store)

	// Setup + get cookie
	setupReq := httptest.NewRequest("POST", "/api/auth/setup", strings.NewReader(`{"password":"test123"}`))
	setupW := httptest.NewRecorder()
	am.setupHandler(setupW, setupReq)

	var sessionCookie *http.Cookie
	for _, c := range setupW.Result().Cookies() {
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

func TestChangePassword(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store)

	// Setup
	setupReq := httptest.NewRequest("POST", "/api/auth/setup", strings.NewReader(`{"password":"test123"}`))
	am.setupHandler(httptest.NewRecorder(), setupReq)

	// Change password
	body := `{"current_password":"test123","new_password":"newpass123"}`
	req := httptest.NewRequest("POST", "/api/auth/change-password", strings.NewReader(body))
	w := httptest.NewRecorder()
	am.changePasswordHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Login with new password
	loginReq := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"password":"newpass123"}`))
	loginW := httptest.NewRecorder()
	am.loginHandler(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Errorf("expected login with new password to succeed, got %d", loginW.Code)
	}

	// Old password should fail
	loginReq2 := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"password":"test123"}`))
	loginW2 := httptest.NewRecorder()
	am.loginHandler(loginW2, loginReq2)

	if loginW2.Code != http.StatusUnauthorized {
		t.Errorf("expected old password to fail, got %d", loginW2.Code)
	}
}

func TestChangePasswordWrongCurrent(t *testing.T) {
	store := testStore(t)
	am := newAuthManager(store)

	setupReq := httptest.NewRequest("POST", "/api/auth/setup", strings.NewReader(`{"password":"test123"}`))
	am.setupHandler(httptest.NewRecorder(), setupReq)

	body := `{"current_password":"wrong","new_password":"newpass123"}`
	req := httptest.NewRequest("POST", "/api/auth/change-password", strings.NewReader(body))
	w := httptest.NewRecorder()
	am.changePasswordHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
