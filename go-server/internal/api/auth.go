package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/axvi/wiretap/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type authManager struct {
	store    *storage.Store
	sessions map[string]time.Time
	mu       sync.RWMutex
}

func newAuthManager(store *storage.Store) *authManager {
	return &authManager{
		store:    store,
		sessions: make(map[string]time.Time),
	}
}

func (am *authManager) isSetupDone() bool {
	hash, _ := am.store.GetSetting("auth_password_hash")
	return hash != ""
}

func (am *authManager) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Always allow auth endpoints, health, and ingest
		if path == "/api/auth/login" || path == "/api/auth/setup" || path == "/api/auth/status" || path == "/api/health" {
			next.ServeHTTP(w, r)
			return
		}

		// If setup not done, allow everything (first-time setup)
		if !am.isSetupDone() {
			next.ServeHTTP(w, r)
			return
		}

		// Check session cookie
		cookie, err := r.Cookie("wiretap_session")
		if err != nil || !am.validSession(cookie.Value) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (am *authManager) validSession(token string) bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	expiry, ok := am.sessions[token]
	if !ok {
		return false
	}
	if time.Now().After(expiry) {
		delete(am.sessions, token)
		return false
	}
	return true
}

func (am *authManager) createSession() string {
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)

	am.mu.Lock()
	am.sessions[token] = time.Now().Add(24 * time.Hour)
	am.mu.Unlock()

	return token
}

func (am *authManager) statusHandler(w http.ResponseWriter, r *http.Request) {
	setupDone := am.isSetupDone()

	authenticated := false
	if cookie, err := r.Cookie("wiretap_session"); err == nil {
		authenticated = am.validSession(cookie.Value)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"setup_done":    setupDone,
		"authenticated": authenticated,
	})
}

func (am *authManager) setupHandler(w http.ResponseWriter, r *http.Request) {
	if am.isSetupDone() {
		http.Error(w, `{"error":"setup already completed"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Password == "" {
		http.Error(w, `{"error":"password is required"}`, http.StatusBadRequest)
		return
	}

	if len(body.Password) < 6 {
		http.Error(w, `{"error":"password must be at least 6 characters"}`, http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	am.store.SetSetting("auth_password_hash", string(hash))

	// Auto-login after setup
	token := am.createSession()
	http.SetCookie(w, &http.Cookie{
		Name:     "wiretap_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

func (am *authManager) loginHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Password == "" {
		http.Error(w, `{"error":"password is required"}`, http.StatusBadRequest)
		return
	}

	hash, _ := am.store.GetSetting("auth_password_hash")
	if hash == "" {
		http.Error(w, `{"error":"setup not completed"}`, http.StatusBadRequest)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.Password)); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"wrong password"}`))
		return
	}

	token := am.createSession()
	http.SetCookie(w, &http.Cookie{
		Name:     "wiretap_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

func (am *authManager) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("wiretap_session"); err == nil {
		am.mu.Lock()
		delete(am.sessions, cookie.Value)
		am.mu.Unlock()
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
