package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/axvi/wiretap/internal/storage"
	"github.com/go-chi/chi/v5"
)

// SiteReloader is called when sites are added/updated/deleted.
type SiteReloader interface {
	ReloadSites()
}

type monitorHandlers struct {
	store    *storage.Store
	reloader SiteReloader
}

// --- Sites ---

func (h *monitorHandlers) listSites(w http.ResponseWriter, r *http.Request) {
	sites, err := h.store.ListMonitorSites()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if sites == nil {
		sites = []storage.MonitorSite{}
	}
	writeJSON(w, sites)
}

func (h *monitorHandlers) createSite(w http.ResponseWriter, r *http.Request) {
	var site storage.MonitorSite
	if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if site.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}
	u, err := url.Parse(site.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		http.Error(w, "url must be a valid http/https URL", http.StatusBadRequest)
		return
	}
	site.URL = strings.TrimRight(site.URL, "/")
	// Auto-disable SSL check for non-HTTPS URLs
	if !strings.HasPrefix(site.URL, "https://") {
		site.CheckSSL = false
	}
	if site.IntervalS < 5 {
		site.IntervalS = 30
	}
	id, err := h.store.CreateMonitorSite(site)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	site.ID = id
	if h.reloader != nil {
		h.reloader.ReloadSites()
	}
	writeJSON(w, site)
}

func (h *monitorHandlers) updateSite(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var site storage.MonitorSite
	if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	site.ID = id
	if err := h.store.UpdateMonitorSite(site); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if h.reloader != nil {
		h.reloader.ReloadSites()
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *monitorHandlers) deleteSite(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.store.DeleteMonitorSite(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if h.reloader != nil {
		h.reloader.ReloadSites()
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// --- Status ---

func (h *monitorHandlers) status(w http.ResponseWriter, r *http.Request) {
	sites, err := h.store.ListMonitorSites()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	since24h := float64(time.Now().Add(-24 * time.Hour).Unix())
	var statuses []storage.MonitorStatus
	for _, site := range sites {
		s := storage.MonitorStatus{Site: site}
		s.SSL, _ = h.store.GetMonitorSSL(site.ID)
		s.Health, _ = h.store.GetMonitorHealth(site.ID)
		s.Domain, _ = h.store.GetMonitorDomain(site.ID)
		s.LastCheck, _ = h.store.GetLastMonitorCheck(site.ID)
		s.UptimePct, _ = h.store.GetUptimePercent(site.ID, since24h)
		statuses = append(statuses, s)
	}
	if statuses == nil {
		statuses = []storage.MonitorStatus{}
	}
	writeJSON(w, statuses)
}

// --- Checks history ---

func (h *monitorHandlers) checks(w http.ResponseWriter, r *http.Request) {
	siteID, _ := strconv.Atoi(chi.URLParam(r, "siteID"))
	limit := intParam(r, "limit", 100)
	checks, err := h.store.ListMonitorChecks(siteID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if checks == nil {
		checks = []storage.MonitorCheck{}
	}
	writeJSON(w, checks)
}

// --- SSL ---

func (h *monitorHandlers) ssl(w http.ResponseWriter, r *http.Request) {
	siteID, _ := strconv.Atoi(chi.URLParam(r, "siteID"))
	ssl, err := h.store.GetMonitorSSL(siteID)
	if err != nil {
		writeJSON(w, nil)
		return
	}
	writeJSON(w, ssl)
}

// --- Health ---

func (h *monitorHandlers) health(w http.ResponseWriter, r *http.Request) {
	siteID, _ := strconv.Atoi(chi.URLParam(r, "siteID"))
	health, err := h.store.GetMonitorHealth(siteID)
	if err != nil {
		writeJSON(w, nil)
		return
	}
	writeJSON(w, health)
}

// --- Domain ---

func (h *monitorHandlers) domain(w http.ResponseWriter, r *http.Request) {
	siteID, _ := strconv.Atoi(chi.URLParam(r, "siteID"))
	domain, err := h.store.GetMonitorDomain(siteID)
	if err != nil {
		writeJSON(w, nil)
		return
	}
	writeJSON(w, domain)
}

// --- Schedule Heartbeats ---

func (h *monitorHandlers) listSchedules(w http.ResponseWriter, r *http.Request) {
	beats, err := h.store.ListScheduleHeartbeats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if beats == nil {
		beats = []storage.ScheduleHeartbeat{}
	}
	writeJSON(w, beats)
}

func (h *monitorHandlers) createSchedule(w http.ResponseWriter, r *http.Request) {
	var b storage.ScheduleHeartbeat
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	b.Command = strings.TrimSpace(b.Command)
	if b.Command == "" {
		http.Error(w, "command is required", http.StatusBadRequest)
		return
	}
	if b.MaxGapM < 1 {
		b.MaxGapM = 60
	}
	id, err := h.store.CreateScheduleHeartbeat(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b.ID = id
	writeJSON(w, b)
}

func (h *monitorHandlers) updateSchedule(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var b storage.ScheduleHeartbeat
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	b.ID = id
	if err := h.store.UpdateScheduleHeartbeat(b); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *monitorHandlers) knownScheduleCommands(w http.ResponseWriter, r *http.Request) {
	commands, err := h.store.ListKnownScheduleCommands()
	if err != nil {
		commands = []string{}
	}
	if commands == nil {
		commands = []string{}
	}
	writeJSON(w, commands)
}

func (h *monitorHandlers) deleteSchedule(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.store.DeleteScheduleHeartbeat(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}
