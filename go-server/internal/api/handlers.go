package api

import (
	"database/sql"
	"encoding/json"
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
