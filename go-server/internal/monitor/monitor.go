package monitor

import (
	"context"
	"log"

	"github.com/axvi/wiretap/internal/hub"
	"github.com/axvi/wiretap/internal/notifier"
	"github.com/axvi/wiretap/internal/storage"
)

// Monitor coordinates all monitoring goroutines.
type Monitor struct {
	store   *storage.Store
	hub     *hub.Hub
	notify  *notifier.Manager
	uptime  *UptimeMonitor
	ssl     *SSLMonitor
	health  *HealthMonitor
	domain  *DomainMonitor
	sched   *ScheduleMonitor
}

// New creates a new Monitor.
func New(store *storage.Store, h *hub.Hub, notify *notifier.Manager) *Monitor {
	return &Monitor{
		store:  store,
		hub:    h,
		notify: notify,
		uptime: &UptimeMonitor{store: store, hub: h, notify: notify},
		ssl:    &SSLMonitor{store: store, notify: notify},
		health: &HealthMonitor{store: store, hub: h},
		domain: &DomainMonitor{store: store, notify: notify},
		sched:  &ScheduleMonitor{store: store, notify: notify},
	}
}

// Run starts all monitoring goroutines. Blocks until ctx is cancelled.
func (m *Monitor) Run(ctx context.Context) {
	log.Println("[monitor] starting monitoring goroutines")

	go m.uptime.Run(ctx)
	go m.ssl.Run(ctx)
	go m.health.Run(ctx)
	go m.domain.Run(ctx)
	go m.sched.Run(ctx)

	<-ctx.Done()
	log.Println("[monitor] shutting down")
}

// ReloadSites signals monitors to reload their site list.
func (m *Monitor) ReloadSites() {
	m.uptime.reload.Store(true)
	m.ssl.reload.Store(true)
	m.health.reload.Store(true)
	m.domain.reload.Store(true)
}
