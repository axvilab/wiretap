package api

import (
	"io/fs"
	"net/http"

	"github.com/axvi/wiretap/internal/correlator"
	"github.com/axvi/wiretap/internal/hub"
	"github.com/axvi/wiretap/internal/receiver"
	"github.com/axvi/wiretap/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewIngestRouter(recv *receiver.HTTPReceiver, store *storage.Store) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(ingestKeyMiddleware(store))
	r.Post("/ingest", recv.Handler())
	r.Get("/health", healthHandler)
	return r
}

func NewAPIRouter(store *storage.Store, corr *correlator.Correlator, wsHub *hub.Hub, frontendFS ...fs.FS) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	// Auth
	am := newAuthManager(store)
	r.Use(am.middleware)

	r.Post("/api/auth/setup", am.setupHandler)
	r.Post("/api/auth/login", am.loginHandler)
	r.Get("/api/auth/status", am.statusHandler)
	r.Post("/api/auth/logout", am.logoutHandler)
	r.Post("/api/auth/change-password", am.changePasswordHandler)

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

	// Settings API
	r.Get("/api/settings", h.getSettings)
	r.Put("/api/settings", h.updateSettings)
	r.Get("/api/storage/stats", h.storageStats)
	r.Post("/api/storage/clear", h.clearData)

	// Replay proxy
	rh := newReplayHandler(store)
	r.Get("/api/replay/config", rh.configHandler)
	r.Post("/api/replay/request", rh.replayRequest)
	r.Post("/api/replay/job", rh.proxy("retry-job"))
	r.Post("/api/replay/dispatch-job", rh.proxy("dispatch-job"))
	r.Post("/api/replay/broadcast", rh.proxy("replay-broadcast"))
	r.Post("/api/replay/command", rh.proxy("replay-command"))

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
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
