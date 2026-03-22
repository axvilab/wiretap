package api

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/axvi/wiretap/internal/config"
	"github.com/axvi/wiretap/internal/correlator"
	"github.com/axvi/wiretap/internal/hub"
	"github.com/axvi/wiretap/internal/notifier"
	"github.com/axvi/wiretap/internal/receiver"
	"github.com/axvi/wiretap/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewIngestRouter(recv *receiver.HTTPReceiver, cfg *config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(ingestKeyMiddleware(cfg))
	r.Post("/ingest", recv.Handler())
	r.Get("/health", healthHandler)
	return r
}

func NewAPIRouter(store *storage.Store, corr *correlator.Correlator, wsHub *hub.Hub, cfg *config.Config, notifyMgr *notifier.Manager, siteReloader SiteReloader, frontendFS ...fs.FS) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)
	r.Use(maxBodySize(1024 * 1024)) // 1MB max for API requests

	// Auth
	am := newAuthManager(store, cfg)
	r.Use(am.middleware)

	r.Post("/api/auth/login", am.loginHandler)
	r.Get("/api/auth/status", am.statusHandler)
	r.Post("/api/auth/logout", am.logoutHandler)

	h := &handlers{store: store, correlator: corr}

	r.Get("/api/health", healthHandler)
	r.Get("/api/dashboard", h.dashboard)
	r.Get("/api/requests", h.listRequests)
	r.Get("/api/trace/{rid}", h.getTrace)
	r.Get("/api/mail", h.listMail)
	r.Get("/api/mail/by-trace/{rid}", h.getMailByTrace)
	r.Get("/api/mail/{id}", h.getMail)
	r.Get("/api/mail/{id}/attachment", h.downloadAttachment)
	r.Get("/api/ws", wsHub.HandleWS)

	// Rate limiter for write operations: 30 requests per minute per IP
	rl := newRateLimiter(30, time.Minute)

	// Settings API
	r.Get("/api/settings", h.getSettings)
	r.Put("/api/settings", writeRateLimit(rl, h.updateSettings))
	r.Get("/api/storage/stats", h.storageStats)
	r.Post("/api/storage/clear", writeRateLimit(rl, h.clearData))

	// Replay proxy
	rh := newReplayHandler(cfg)
	r.Get("/api/replay/config", rh.configHandler)
	r.Post("/api/replay/request", writeRateLimit(rl, rh.replayRequest))
	r.Post("/api/replay/job", writeRateLimit(rl, rh.proxy("retry-job")))
	r.Post("/api/replay/dispatch-job", writeRateLimit(rl, rh.proxy("dispatch-job")))
	r.Post("/api/replay/broadcast", writeRateLimit(rl, rh.proxy("replay-broadcast")))
	r.Post("/api/replay/command", writeRateLimit(rl, rh.proxy("replay-command")))

	// Monitor API
	mh := &monitorHandlers{store: store, reloader: siteReloader}
	r.Get("/api/monitor/sites", mh.listSites)
	r.Post("/api/monitor/sites", writeRateLimit(rl, mh.createSite))
	r.Put("/api/monitor/sites/{id}", writeRateLimit(rl, mh.updateSite))
	r.Delete("/api/monitor/sites/{id}", writeRateLimit(rl, mh.deleteSite))
	r.Get("/api/monitor/status", mh.status)
	r.Get("/api/monitor/checks/{siteID}", mh.checks)
	r.Get("/api/monitor/ssl/{siteID}", mh.ssl)
	r.Get("/api/monitor/health/{siteID}", mh.health)
	r.Get("/api/monitor/domain/{siteID}", mh.domain)
	r.Get("/api/monitor/schedule-commands", mh.knownScheduleCommands)
	r.Get("/api/monitor/schedules", mh.listSchedules)
	r.Post("/api/monitor/schedules", writeRateLimit(rl, mh.createSchedule))
	r.Put("/api/monitor/schedules/{id}", writeRateLimit(rl, mh.updateSchedule))
	r.Delete("/api/monitor/schedules/{id}", writeRateLimit(rl, mh.deleteSchedule))

	// Notification API
	nh := &notificationHandlers{store: store, notify: notifyMgr}
	r.Get("/api/notifications/channels", nh.listChannels)
	r.Post("/api/notifications/channels", writeRateLimit(rl, nh.createChannel))
	r.Put("/api/notifications/channels/{id}", writeRateLimit(rl, nh.updateChannel))
	r.Delete("/api/notifications/channels/{id}", writeRateLimit(rl, nh.deleteChannel))
	r.Post("/api/notifications/channels/{id}/test", writeRateLimit(rl, nh.testChannel))
	r.Get("/api/notifications/rules", nh.listRules)
	r.Post("/api/notifications/rules", writeRateLimit(rl, nh.createRule))
	r.Put("/api/notifications/rules/{id}", writeRateLimit(rl, nh.updateRule))
	r.Delete("/api/notifications/rules/{id}", writeRateLimit(rl, nh.deleteRule))
	r.Get("/api/notifications/logs", nh.listLogs)

	// Serve embedded frontend (prod) — fallback to index.html for SPA
	if len(frontendFS) > 0 && frontendFS[0] != nil {
		dist := frontendFS[0]
		fileServer := http.FileServer(http.FS(dist))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path[1:]
			if path == "" {
				path = "index.html"
			}
			if f, err := dist.Open(path); err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
			} else {
				// SPA fallback
				r.URL.Path = "/"
				fileServer.ServeHTTP(w, r)
			}
		})
	}

	return r
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"healthy"}`))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// Allow same-origin and Vite dev server
		if origin != "" {
			// In production, frontend is served from same origin (go:embed).
			// In dev, Vite proxies API calls so Origin matches the web server.
			// Only allow the actual origin making the request.
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}