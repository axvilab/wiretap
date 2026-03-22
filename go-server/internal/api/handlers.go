package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/axvi/wiretap/internal/correlator"
	"github.com/axvi/wiretap/internal/storage"
	"github.com/go-chi/chi/v5"
)

type handlers struct {
	store      *storage.Store
	correlator *correlator.Correlator
}

func (h *handlers) listRequests(w http.ResponseWriter, r *http.Request) {
	params := storage.ListParams{
		Page:    intParam(r, "page", 1),
		PerPage: intParam(r, "per_page", 50),
		Type:    r.URL.Query().Get("type"),
		Status:  intParam(r, "status", 0),
		Method:  r.URL.Query().Get("method"),
		URI:     r.URL.Query().Get("uri"),
	}

	result, err := h.store.ListTraces(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, result)
}

func (h *handlers) getTrace(w http.ResponseWriter, r *http.Request) {
	rid := chi.URLParam(r, "rid")
	if rid == "" {
		http.Error(w, "rid is required", http.StatusBadRequest)
		return
	}

	tree, err := h.correlator.BuildTree(rid)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "trace not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, tree)
}

func (h *handlers) listMail(w http.ResponseWriter, r *http.Request) {
	page := intParam(r, "page", 1)
	perPage := intParam(r, "per_page", 50)

	emails, total, err := h.store.ListEmails(page, perPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"emails":   emails,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

func (h *handlers) getMailByTrace(w http.ResponseWriter, r *http.Request) {
	rid := chi.URLParam(r, "rid")
	email, err := h.store.GetEmailByTrace(rid)
	if err == sql.ErrNoRows {
		http.Error(w, "email not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, email)
}

func (h *handlers) getMail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	email, err := h.store.GetEmail(id)
	if err == sql.ErrNoRows {
		http.Error(w, "email not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, email)
}

func (h *handlers) downloadAttachment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idx := intParam(r, "index", -1)
	if idx < 0 {
		http.Error(w, "index is required", http.StatusBadRequest)
		return
	}

	att, err := h.store.GetEmailAttachment(id, idx)
	if err != nil {
		http.Error(w, "attachment not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", att.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, att.Name))
	w.Header().Set("Content-Length", strconv.Itoa(len(att.Content)))
	w.Write(att.Content)
}

func (h *handlers) getSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.store.GetAllSettings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, settings)
}

func (h *handlers) updateSettings(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	allowed := map[string]bool{
		"app_url":                   true,
		"replay_secret":             true,
		"slow_query_ms":             true,
		"slow_request_ms":           true,
		"retention_days":            true,
		"monitor_enabled":           true,
		"monitor_default_interval":  true,
		"ssl_warning_days":          true,
	}

	for k, v := range body {
		if !allowed[k] {
			continue
		}
		if err := h.store.SetSetting(k, v); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *handlers) storageStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.StorageStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, stats)
}

func (h *handlers) clearData(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Days int `json:"days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Days < 0 {
		// Clear all
		body.Days = 0
	}

	if err := h.store.ClearData(body.Days); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func intParam(r *http.Request, key string, fallback int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}
