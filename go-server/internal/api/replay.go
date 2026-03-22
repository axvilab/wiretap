package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/axvi/wiretap/internal/config"
)

type replayHandler struct {
	appURL       string
	replaySecret string
	client       *http.Client
}

func newReplayHandler(cfg *config.Config) *replayHandler {
	return &replayHandler{
		appURL:       cfg.AppURL,
		replaySecret: cfg.ReplaySecret,
		client:       &http.Client{Timeout: 30 * time.Second},
	}
}

// replayRequest sends the HTTP request directly from Go
func (rh *replayHandler) replayRequest(w http.ResponseWriter, r *http.Request) {
	if rh.appURL == "" {
		http.Error(w, `{"error":"WIRETAP_APP_URL not configured"}`, http.StatusServiceUnavailable)
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
	targetURL := strings.TrimRight(rh.appURL, "/") + body.URI

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

// proxy forwards request to Laravel's ReplayController
func (rh *replayHandler) proxy(laravelPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rh.replaySecret == "" || rh.appURL == "" {
			http.Error(w, `{"error":"WIRETAP_APP_URL and WIRETAP_REPLAY_SECRET must be configured"}`, http.StatusServiceUnavailable)
			return
		}

		targetURL := strings.TrimRight(rh.appURL, "/") + "/_wiretap/" + laravelPath

		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, targetURL, r.Body)
		if err != nil {
			writeReplayError(w, "failed to create request")
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Wiretap-Secret", rh.replaySecret)

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
	enabled := rh.appURL != "" && rh.replaySecret != ""
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"enabled":%t}`, enabled)
}

func writeReplayError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}