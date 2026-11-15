package monitor

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/axvi/wiretap/internal/hub"
	"github.com/axvi/wiretap/internal/storage"
)

// HealthMonitor polls /health endpoints for sites that have check_health enabled.
type HealthMonitor struct {
	store  *storage.Store
	hub    *hub.Hub
	reload atomic.Bool
}

func (m *HealthMonitor) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Immediate first check
	m.checkAll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkAll(ctx)
		}
	}
}

func (m *HealthMonitor) checkAll(ctx context.Context) {
	sites, err := m.store.ListMonitorSites()
	if err != nil {
		log.Printf("[monitor:health] failed to list sites: %v", err)
		return
	}

	for _, site := range sites {
		if site.HealthURL == "" {
			continue
		}
		m.checkSite(ctx, site)
	}
}

func (m *HealthMonitor) checkSite(ctx context.Context, site storage.MonitorSite) {
	healthURL := site.HealthURL
	// If health URL is a path, prepend the site URL
	if len(healthURL) > 0 && healthURL[0] == '/' {
		healthURL = strings.TrimRight(site.URL, "/") + healthURL
	}

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", healthURL, nil)
	if err != nil {
		m.save(site.ID, "unreachable", "{}")
		return
	}
	req.Header.Set("User-Agent", "Wiretap-Monitor/1.0")
	applyAuth(req, site)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.save(site.ID, "unreachable", "{}")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024)) // 64KB max

	status := "healthy"
	if resp.StatusCode >= 400 {
		status = "unhealthy"
	}

	m.save(site.ID, status, string(body))
}

func (m *HealthMonitor) save(siteID int, status string, dataJSON string) {
	h := storage.MonitorHealth{
		SiteID:   siteID,
		Status:   status,
		DataJSON: dataJSON,
	}
	if err := m.store.SaveMonitorHealth(h); err != nil {
		log.Printf("[monitor:health] failed to save for site %d: %v", siteID, err)
	}
	m.hub.Broadcast("monitor.health", h)
}
