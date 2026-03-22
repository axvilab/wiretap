package api

import (
	"encoding/json"
	"net/http"

	"github.com/axvi/wiretap/internal/notifier"
	"github.com/axvi/wiretap/internal/storage"
)

type notificationHandlers struct {
	store  *storage.Store
	notify *notifier.Manager
}

// --- Channels ---

func (h *notificationHandlers) listChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := h.store.ListNotificationChannels()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if channels == nil {
		channels = []storage.NotificationChannel{}
	}
	writeJSON(w, channels)
}

func (h *notificationHandlers) createChannel(w http.ResponseWriter, r *http.Request) {
	var ch storage.NotificationChannel
	if err := json.NewDecoder(r.Body).Decode(&ch); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	validTypes := map[string]bool{"telegram": true, "slack": true, "teams": true}
	if !validTypes[ch.Type] {
		http.Error(w, "type must be telegram, slack, or teams", http.StatusBadRequest)
		return
	}
	if ch.Config == "" {
		http.Error(w, "config is required", http.StatusBadRequest)
		return
	}
	// Validate config is valid JSON
	var configTest map[string]any
	if err := json.Unmarshal([]byte(ch.Config), &configTest); err != nil {
		http.Error(w, "config must be valid JSON", http.StatusBadRequest)
		return
	}
	// Validate required config fields
	switch ch.Type {
	case "telegram":
		if configTest["token"] == nil || configTest["chat_id"] == nil {
			http.Error(w, "telegram config requires token and chat_id", http.StatusBadRequest)
			return
		}
	case "slack", "teams":
		if configTest["webhook_url"] == nil {
			http.Error(w, ch.Type+" config requires webhook_url", http.StatusBadRequest)
			return
		}
	}
	ch.Name = maxLen(ch.Name, 255)
	ch.Config = maxLen(ch.Config, 4096)
	if ch.Name == "" {
		ch.Name = ch.Type
	}
	id, err := h.store.CreateNotificationChannel(ch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ch.ID = id
	h.notify.Reload()
	writeJSON(w, ch)
}

func (h *notificationHandlers) updateChannel(w http.ResponseWriter, r *http.Request) {
	id, ok := urlParamID(w, r, "id"); if !ok { return }; _ = id
	var ch storage.NotificationChannel
	if err := json.NewDecoder(r.Body).Decode(&ch); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	ch.ID = id
	if err := h.store.UpdateNotificationChannel(ch); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.notify.Reload()
	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *notificationHandlers) deleteChannel(w http.ResponseWriter, r *http.Request) {
	id, ok := urlParamID(w, r, "id"); if !ok { return }; _ = id
	if err := h.store.DeleteNotificationChannel(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.notify.Reload()
	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *notificationHandlers) testChannel(w http.ResponseWriter, r *http.Request) {
	id, ok := urlParamID(w, r, "id"); if !ok { return }; _ = id
	if err := h.notify.SendTest(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// --- Rules ---

func (h *notificationHandlers) listRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.store.ListAlertRules()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rules == nil {
		rules = []storage.AlertRule{}
	}
	writeJSON(w, rules)
}

func (h *notificationHandlers) createRule(w http.ResponseWriter, r *http.Request) {
	var rule storage.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	validRuleTypes := map[string]bool{"exception": true, "site_down": true, "ssl_expiry": true, "domain_expiry": true, "schedule_missed": true}
	if !validRuleTypes[rule.Type] {
		http.Error(w, "type must be exception, site_down, ssl_expiry, domain_expiry, or schedule_missed", http.StatusBadRequest)
		return
	}
	if rule.ChannelID == 0 {
		http.Error(w, "channel_id is required", http.StatusBadRequest)
		return
	}
	// Verify channel exists
	if _, err := h.store.GetNotificationChannel(rule.ChannelID); err != nil {
		http.Error(w, "channel not found", http.StatusBadRequest)
		return
	}
	if rule.CooldownM < 1 {
		rule.CooldownM = 5
	}
	id, err := h.store.CreateAlertRule(rule)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rule.ID = id
	writeJSON(w, rule)
}

func (h *notificationHandlers) updateRule(w http.ResponseWriter, r *http.Request) {
	id, ok := urlParamID(w, r, "id"); if !ok { return }; _ = id
	var rule storage.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	rule.ID = id
	if err := h.store.UpdateAlertRule(rule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *notificationHandlers) deleteRule(w http.ResponseWriter, r *http.Request) {
	id, ok := urlParamID(w, r, "id"); if !ok { return }; _ = id
	if err := h.store.DeleteAlertRule(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// --- Alert Log ---

func (h *notificationHandlers) listLogs(w http.ResponseWriter, r *http.Request) {
	limit := intParam(r, "limit", 50)
	logs, err := h.store.ListAlertLogs(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if logs == nil {
		logs = []storage.AlertLog{}
	}
	writeJSON(w, logs)
}
