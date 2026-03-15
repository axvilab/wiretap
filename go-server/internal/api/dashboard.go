package api

import (
	"net/http"
	"time"
)

func (h *handlers) dashboard(w http.ResponseWriter, r *http.Request) {
	minutes := 60
	if p := r.URL.Query().Get("period"); p != "" {
		switch p {
		case "1m":
			minutes = 1
		case "5m":
			minutes = 5
		case "15m":
			minutes = 15
		case "30m":
			minutes = 30
		case "1h":
			minutes = 60
		case "24h":
			minutes = 60 * 24
		case "7d":
			minutes = 60 * 24 * 7
		default:
			minutes = 60
		}
	}

	hours := minutes / 60
	if hours < 1 {
		hours = 1
	}
	since := float64(time.Now().Add(-time.Duration(minutes) * time.Minute).Unix())

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
		"period":           r.URL.Query().Get("period"),
		"stats":            stats,
		"timeline":         timeline,
		"status_breakdown": statusBreakdown,
		"slow_endpoints":   slowEndpoints,
		"slow_queries":     slowQueries,
	})
}
