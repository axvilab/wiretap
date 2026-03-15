package api

import (
	"net/http"
	"strconv"
	"time"
)

func (h *handlers) dashboard(w http.ResponseWriter, r *http.Request) {
	hours := 1
	if p := r.URL.Query().Get("period"); p != "" {
		switch p {
		case "24h":
			hours = 24
		case "7d":
			hours = 24 * 7
		default:
			hours = 1
		}
	}

	since := float64(time.Now().Add(-time.Duration(hours) * time.Hour).Unix())

	stats, err := h.store.DashboardStats(since)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeline, err := h.store.DashboardTimeline(since, hours)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	statusBreakdown, err := h.store.DashboardStatusBreakdown(since)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slowEndpoints, err := h.store.DashboardSlowEndpoints(since, 10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slowQueries, err := h.store.DashboardSlowQueries(since, 10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"period":           strconv.Itoa(hours) + "h",
		"stats":            stats,
		"timeline":         timeline,
		"status_breakdown": statusBreakdown,
		"slow_endpoints":   slowEndpoints,
		"slow_queries":     slowQueries,
	})
}
