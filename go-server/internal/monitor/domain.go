package monitor

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/axvi/wiretap/internal/notifier"
	"github.com/axvi/wiretap/internal/storage"
)

// DomainMonitor checks domain WHOIS expiry.
type DomainMonitor struct {
	store  *storage.Store
	notify *notifier.Manager
	reload atomic.Bool
}

func (m *DomainMonitor) Run(ctx context.Context) {
	// Check on start, then daily
	m.checkAll()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkAll()
		}
	}
}

func (m *DomainMonitor) checkAll() {
	sites, err := m.store.ListMonitorSites()
	if err != nil {
		log.Printf("[monitor:domain] failed to list sites: %v", err)
		return
	}

	// Deduplicate domains
	seen := make(map[string]bool)
	for _, site := range sites {
		domain := extractDomain(site.URL)
		if domain == "" || seen[domain] {
			continue
		}
		seen[domain] = true
		m.checkDomain(site, domain)
	}
}

func (m *DomainMonitor) checkDomain(site storage.MonitorSite, domain string) {
	expiresAt, registrar, err := whoisLookup(domain)
	if err != nil {
		m.store.SaveMonitorDomain(storage.MonitorDomain{
			SiteID: site.ID, Domain: domain, Error: err.Error(),
		})
		return
	}

	daysRemaining := int(math.Floor(time.Until(expiresAt).Hours() / 24))

	d := storage.MonitorDomain{
		SiteID:        site.ID,
		Domain:        domain,
		Registrar:     registrar,
		ExpiresAt:     float64(expiresAt.Unix()),
		DaysRemaining: daysRemaining,
	}
	m.store.SaveMonitorDomain(d)

	// Alert if domain expiring within 30 days
	if daysRemaining <= 30 && daysRemaining > 0 {
		m.notify.Notify("domain_expiry", notifier.Message{
			Title: fmt.Sprintf("Domain Expiring: %s", domain),
			Body:  fmt.Sprintf("Domain %s expires in %d days (registrar: %s)", domain, daysRemaining, registrar),
			Level: "warning",
		})
	}
}

// extractDomain extracts the registrable domain from a URL.
func extractDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := u.Hostname()
	// Simple: take last two parts (works for .com, .org, etc.)
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "." + parts[len(parts)-1]
	}
	return host
}

// whoisLookup performs a raw WHOIS lookup via TCP port 43.
func whoisLookup(domain string) (expiresAt time.Time, registrar string, err error) {
	conn, err := net.DialTimeout("tcp", "whois.verisign-grs.com:43", 10*time.Second)
	if err != nil {
		// Try generic whois server
		conn, err = net.DialTimeout("tcp", "whois.iana.org:43", 10*time.Second)
		if err != nil {
			return time.Time{}, "", fmt.Errorf("whois connection failed: %w", err)
		}
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(10 * time.Second))
	fmt.Fprintf(conn, "%s\r\n", domain)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lower := strings.ToLower(line)

		if strings.HasPrefix(lower, "registry expiry date:") || strings.HasPrefix(lower, "registrar registration expiration date:") || strings.HasPrefix(lower, "expiration date:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				dateStr := strings.TrimSpace(parts[1])
				for _, layout := range []string{
					time.RFC3339,
					"2006-01-02T15:04:05Z",
					"2006-01-02",
				} {
					if t, err := time.Parse(layout, dateStr); err == nil {
						expiresAt = t
						break
					}
				}
			}
		}
		if strings.HasPrefix(lower, "registrar:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				registrar = strings.TrimSpace(parts[1])
			}
		}
	}

	if expiresAt.IsZero() {
		return time.Time{}, registrar, fmt.Errorf("could not parse expiry date from whois for %s", domain)
	}
	return expiresAt, registrar, nil
}
