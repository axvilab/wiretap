package main

import (
	"context"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/axvi/wiretap/internal/api"
	"github.com/axvi/wiretap/internal/config"
	"github.com/axvi/wiretap/internal/correlator"
	"github.com/axvi/wiretap/internal/hub"
	"github.com/axvi/wiretap/internal/monitor"
	"github.com/axvi/wiretap/internal/notifier"
	"github.com/axvi/wiretap/internal/receiver"
	wiretapsmtp "github.com/axvi/wiretap/internal/smtp"
	"github.com/axvi/wiretap/internal/storage"

	wiretap "github.com/axvi/wiretap"
)

func main() {
	cfg := config.Load()

	store, err := storage.New(cfg.StoragePath)
	if err != nil {
		log.Fatalf("failed to init storage: %v", err)
	}
	defer store.Close()
	log.Printf("DuckDB opened at %s", cfg.StoragePath)

	// Retention
	store.StartRetention(cfg.RetentionDays)

	// Correlator + WebSocket hub
	corr := correlator.New(store)
	wsHub := hub.New()

	// Notification manager
	notifyMgr := notifier.NewManager(store)

	// Ingest channel + batch writer
	ingestCh := make(chan *storage.TracePayload, 1000)
	bw := receiver.NewBatchWriter(store, corr, wsHub, notifyMgr, ingestCh)
	go bw.Run()

	// Site monitor (created here so it can be passed to API router)
	monitorCtx, monitorCancel := context.WithCancel(context.Background())
	mon := monitor.New(store, wsHub, notifyMgr)

	// HTTP receiver
	recv := receiver.NewHTTP(ingestCh)

	// Servers
	ingestServer := &http.Server{
		Addr:    ":" + cfg.IngestPort,
		Handler: api.NewIngestRouter(recv, cfg),
	}
	webServer := &http.Server{
		Addr:    ":" + cfg.WebPort,
		Handler: api.NewAPIRouter(store, corr, wsHub, cfg, notifyMgr, mon, mustSubFS(wiretap.FrontendDist, "frontend/dist")),
	}

	go func() {
		log.Printf("Ingest server listening on :%s", cfg.IngestPort)
		if err := ingestServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ingest server: %v", err)
		}
	}()

	go func() {
		log.Printf("Web/API server listening on :%s", cfg.WebPort)
		if err := webServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("web server: %v", err)
		}
	}()

	go func() {
		if err := wiretapsmtp.StartSMTPServer(":" + cfg.SMTPPort); err != nil {
			log.Fatalf("smtp server: %v", err)
		}
	}()

	// Redis receiver (optional)
	var redisCancel context.CancelFunc
	if cfg.RedisAddr != "" {
		var redisCtx context.Context
		redisCtx, redisCancel = context.WithCancel(context.Background())
		redisRecv := receiver.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, cfg.RedisQueue, ingestCh)
		go redisRecv.Run(redisCtx)
		log.Printf("Redis receiver connected to %s (queue: %s)", cfg.RedisAddr, cfg.RedisQueue)
	}

	// Start monitor goroutines
	go mon.Run(monitorCtx)

	log.Printf("Wiretap ready — UI :%s | Ingest :%s | SMTP :%s", cfg.WebPort, cfg.IngestPort, cfg.SMTPPort)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	monitorCancel()
	if redisCancel != nil {
		redisCancel()
	}
	close(ingestCh)
	ingestServer.Shutdown(ctx)
	webServer.Shutdown(ctx)

	log.Println("bye")
}

func mustSubFS(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		return nil
	}
	return sub
}
