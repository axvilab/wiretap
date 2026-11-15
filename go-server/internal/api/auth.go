package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/axvi/wiretap/internal/config"
	"github.com/axvi/wiretap/internal/storage"
)

const sessionTTL = 24 * time.Hour

type authManager struct {
	store    *storage.Store
	password string
}

func newAuthManager(store *storage.Store, cfg *config.Config) *authManager {
	return &authManager{
		store:    store,
		password: cfg.Password,
	}
}

func (am *authManager) isAuthEnabled() bool {
	return am.password != ""
}

func (am *authManager) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// The UI itself (HTML, JS, CSS) is public so the browser can show the login form;
		// all data goes through /api, which stays protected.
		if !strings.HasPrefix(path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		// Always allow auth endpoints and health
		if path == "/api/auth/login" || path == "/api/auth/status" || path == "/api/health" {
			next.ServeHTTP(w, r)
			return
		}

		// If no password configured, allow everything
		if !am.isAuthEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		// Check session cookie
		cookie, err := r.Cookie("wiretap_session")
		if err != nil || !am.store.ValidateSession(cookie.Value) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (am *authManager) createSession() string {
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)
	am.store.CreateSession(token, sessionTTL)
	return token
}

func (am *authManager) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "wiretap_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
}

func (am *authManager) statusHandler(w http.ResponseWriter, r *http.Request) {
	authenticated := false
	if cookie, err := r.Cookie("wiretap_session"); err == nil {
		authenticated = am.store.ValidateSession(cookie.Value)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"auth_enabled":  am.isAuthEnabled(),
		"authenticated": authenticated,
	})
}

func (am *authManager) loginHandler(w http.ResponseWriter, r *http.Request) {
	if !am.isAuthEnabled() {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true}`))
		return
	}

	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Password == "" {
		http.Error(w, `{"error":"password is required"}`, http.StatusBadRequest)
		return
	}

	if body.Password != am.password {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"wrong password"}`))
		return
	}

	token := am.createSession()
	am.setSessionCookie(w, token)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

func (am *authManager) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("wiretap_session"); err == nil {
		am.store.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "wiretap_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}
