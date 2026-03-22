package monitor

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"math"
	"net"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/axvi/wiretap/internal/notifier"
	"github.com/axvi/wiretap/internal/storage"
)

// SSLMonitor checks SSL certificate expiry.
type SSLMonitor struct {
	store  *storage.Store
	notify *notifier.Manager
	reload atomic.Bool
}

func (m *SSLMonitor) Run(ctx context.Context) {
	// Check immediately on start, then every hour (or on reload signal)
	m.checkAll()

	ticker := time.NewTicker(1 * time.Hour)
	reloadTicker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	defer reloadTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkAll()
		case <-reloadTicker.C:
			if m.reload.CompareAndSwap(true, false) {
				m.checkAll()
			}
		}
	}
}

func (m *SSLMonitor) checkAll() {
	sites, err := m.store.ListMonitorSites()
	if err != nil {
		log.Printf("[monitor:ssl] failed to list sites: %v", err)
		return
	}

	for _, site := range sites {
		if !site.CheckSSL {
			continue
		}
		m.checkSite(site)
	}
}

func (m *SSLMonitor) checkSite(site storage.MonitorSite) {
	u, err := url.Parse(site.URL)
	if err != nil || u.Scheme != "https" {
		return
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "443"
	}

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp",
		net.JoinHostPort(host, port),
		&tls.Config{InsecureSkipVerify: false},
	)
	if err != nil {
		ssl := storage.MonitorSSL{
			SiteID:  site.ID,
			Error:   err.Error(),
			IsValid: false,
		}
		m.store.SaveMonitorSSL(ssl)
		return
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		m.store.SaveMonitorSSL(storage.MonitorSSL{
			SiteID: site.ID, Error: "no certificates", IsValid: false,
		})
		return
	}

	leaf := certs[0]
	daysRemaining := int(math.Floor(time.Until(leaf.NotAfter).Hours() / 24))

	ssl := storage.MonitorSSL{
		SiteID:        site.ID,
		Issuer:        leaf.Issuer.CommonName,
		Subject:       leaf.Subject.CommonName,
		ExpiresAt:     float64(leaf.NotAfter.Unix()),
		DaysRemaining: daysRemaining,
		IsValid:       daysRemaining > 0,
	}
	m.store.SaveMonitorSSL(ssl)

	// Check warning threshold
	warningDays := 14
	if v, _ := m.store.GetSetting("ssl_warning_days"); v != "" {
		fmt.Sscanf(v, "%d", &warningDays)
	}

	if daysRemaining <= warningDays && daysRemaining > 0 {
		m.notify.Notify("ssl_expiry", notifier.Message{
			Title: fmt.Sprintf("SSL Expiring: %s", site.Name),
			Body:  fmt.Sprintf("SSL certificate for %s expires in %d days", host, daysRemaining),
			Level: "warning",
		})
	} else if daysRemaining <= 0 {
		m.notify.Notify("ssl_expiry", notifier.Message{
			Title: fmt.Sprintf("SSL Expired: %s", site.Name),
			Body:  fmt.Sprintf("SSL certificate for %s has expired!", host),
			Level: "critical",
		})
	}
}
