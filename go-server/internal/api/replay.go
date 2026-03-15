package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/axvi/wiretap/internal/storage"
)

type replayHandler struct {
	store  *storage.Store
	client *http.Client
}

func newReplayHandler(store *storage.Store) *replayHandler {
	return &replayHandler{
		store:  store,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (rh *replayHandler) getConfig() (appURL, secret string) {
	appURL, _ = rh.store.GetSetting("app_url")
	secret, _ = rh.store.GetSetting("replay_secret")
	return
}

// replayRequest sends the HTTP request directly from Go — no PHP self-call deadlock
func (rh *replayHandler) replayRequest(w http.ResponseWriter, r *http.Request) {
	appURL, _ := rh.getConfig()
	if appURL == "" {
		http.Error(w, `{"error":"replay not configured — set App URL in Settings"}`, http.StatusServiceUnavailable)
		return
	}

	var body struct {
		Method  string            `json:"method"`
		URI     string            `json:"uri"`
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	method := strings.ToUpper(body.Method)
	if method == "" {
		method = "GET"
	}
	targetURL := strings.TrimRight(appURL, "/") + body.URI

	var reqBody io.Reader
	if body.Body != "" {
		reqBody = strings.NewReader(body.Body)
	}

	req, err := http.NewRequestWithContext(r.Context(), method, targetURL, reqBody)
	if err != nil {
		writeReplayError(w, "failed to create request: "+err.Error())
		return
	}

	for k, v := range body.Headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("X-Wiretap-Replay", "1")
	if reqBody != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := rh.client.Do(req)
	if err != nil {
		writeReplayError(w, "failed to reach app: "+err.Error())
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": resp.StatusCode >= 200 && resp.StatusCode < 400,
		"status":  resp.StatusCode,
	})
}

// proxy forwards request to Laravel's ReplayController (for job/broadcast/command)
func (rh *replayHandler) proxy(laravelPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appURL, secret := rh.getConfig()
		if secret == "" || appURL == "" {
			http.Error(w, `{"error":"replay not configured — set App URL and Replay Secret in Settings"}`, http.StatusServiceUnavailable)
			return
		}

		targetURL := strings.TrimRight(appURL, "/") + "/_wiretap/" + laravelPath

		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, targetURL, r.Body)
		if err != nil {
			writeReplayError(w, "failed to create request")
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Wiretap-Secret", secret)

		resp, err := rh.client.Do(req)
		if err != nil {
			writeReplayError(w, "failed to reach app: "+err.Error())
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}

func (rh *replayHandler) configHandler(w http.ResponseWriter, r *http.Request) {
	appURL, secret := rh.getConfig()
	enabled := appURL != "" && secret != ""
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"enabled":%t}`, enabled)
}

func writeReplayError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}