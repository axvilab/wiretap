package monitor

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/axvi/wiretap/internal/hub"
	"github.com/axvi/wiretap/internal/notifier"
	"github.com/axvi/wiretap/internal/storage"
)

// UptimeMonitor checks site availability.
type UptimeMonitor struct {
	store  *storage.Store
	hub    *hub.Hub
	notify *notifier.Manager
	reload atomic.Bool
}

func (m *UptimeMonitor) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var (
		mu       sync.Mutex
		cancels  = make(map[int]context.CancelFunc)
		lastDown = make(map[int]bool)
	)

	startSite := func(site storage.MonitorSite) {
		siteCtx, cancel := context.WithCancel(ctx)
		mu.Lock()
		if old, ok := cancels[site.ID]; ok {
			old()
		}
		cancels[site.ID] = cancel
		mu.Unlock()

		go m.checkLoop(siteCtx, site, lastDown)
	}

	// Initial load
	if sites, err := m.store.ListMonitorSites(); err == nil {
		for _, site := range sites {
			startSite(site)
		}
	}

	for {
		select {
		case <-ctx.Done():
			mu.Lock()
			for _, cancel := range cancels {
				cancel()
			}
			mu.Unlock()
			return
		case <-ticker.C:
			if !m.reload.CompareAndSwap(true, false) {
				continue
			}
			// Reload sites
			sites, err := m.store.ListMonitorSites()
			if err != nil {
				continue
			}
			mu.Lock()
			// Cancel removed sites
			active := make(map[int]bool)
			for _, s := range sites {
				active[s.ID] = true
			}
			for id, cancel := range cancels {
				if !active[id] {
					cancel()
					delete(cancels, id)
				}
			}
			mu.Unlock()
			// Start new/updated sites
			for _, site := range sites {
				startSite(site)
			}
		}
	}
}

func (m *UptimeMonitor) checkLoop(ctx context.Context, site storage.MonitorSite, lastDown map[int]bool) {
	interval := time.Duration(site.IntervalS) * time.Second
	if interval < 5*time.Second {
		interval = 5 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Immediate first check
	m.doCheck(ctx, site, lastDown)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.doCheck(ctx, site, lastDown)
		}
	}
}

func (m *UptimeMonitor) doCheck(ctx context.Context, site storage.MonitorSite, lastDown map[int]bool) {
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	start := time.Now()
	req, err := http.NewRequestWithContext(reqCtx, "GET", site.URL, nil)
	if err != nil {
		m.saveResult(site, 0, 0, err.Error(), lastDown)
		return
	}
	req.Header.Set("User-Agent", "Wiretap-Monitor/1.0")
	applyAuth(req, site)

	resp, err := http.DefaultClient.Do(req)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		m.saveResult(site, 0, int(elapsed), err.Error(), lastDown)
		return
	}
	resp.Body.Close()
	m.saveResult(site, resp.StatusCode, int(elapsed), "", lastDown)
}

func (m *UptimeMonitor) saveResult(site storage.MonitorSite, status int, responseMs int, errMsg string, lastDown map[int]bool) {
	check := storage.MonitorCheck{
		SiteID:     site.ID,
		Status:     status,
		ResponseMs: responseMs,
		Error:      errMsg,
	}
	if err := m.store.SaveMonitorCheck(check); err != nil {
		log.Printf("[monitor:uptime] failed to save check for site %d: %v", site.ID, err)
	}

	// Broadcast to UI
	m.hub.Broadcast("monitor.check", check)

	isDown := status == 0 || status >= 500
	wasDown := lastDown[site.ID]

	if isDown && !wasDown {
		// Site just went down
		msg := notifier.Message{
			Title: fmt.Sprintf("Site Down: %s", site.Name),
			Body:  fmt.Sprintf("%s is not responding (status: %d, error: %s)", site.URL, status, errMsg),
			Level: "critical",
		}
		m.notify.Notify("site_down", msg)
	}

	lastDown[site.ID] = isDown
}
