package storage

import "time"

// --- Monitor Sites ---

func (s *Store) ListMonitorSites() ([]MonitorSite, error) {
	rows, err := s.db.Query(`SELECT id, url, name, check_ssl, health_url, auth_type, auth_json, interval_s, is_primary, created_at FROM monitor_sites ORDER BY is_primary DESC, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []MonitorSite
	for rows.Next() {
		var site MonitorSite
		if err := rows.Scan(&site.ID, &site.URL, &site.Name, &site.CheckSSL, &site.HealthURL, &site.AuthType, &site.AuthJSON, &site.IntervalS, &site.IsPrimary, &site.CreatedAt); err != nil {
			return nil, err
		}
		sites = append(sites, site)
	}
	return sites, nil
}

func (s *Store) CreateMonitorSite(site MonitorSite) (int, error) {
	now := float64(time.Now().Unix())
	var id int
	err := s.db.QueryRow(
		`INSERT INTO monitor_sites (url, name, check_ssl, health_url, auth_type, auth_json, interval_s, is_primary, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
		site.URL, site.Name, site.CheckSSL, site.HealthURL, site.AuthType, site.AuthJSON, site.IntervalS, site.IsPrimary, now,
	).Scan(&id)
	return id, err
}

func (s *Store) UpdateMonitorSite(site MonitorSite) error {
	_, err := s.db.Exec(
		`UPDATE monitor_sites SET url=?, name=?, check_ssl=?, health_url=?, auth_type=?, auth_json=?, interval_s=?, is_primary=? WHERE id=?`,
		site.URL, site.Name, site.CheckSSL, site.HealthURL, site.AuthType, site.AuthJSON, site.IntervalS, site.IsPrimary, site.ID,
	)
	return err
}

func (s *Store) DeleteMonitorSite(id int) error {
	_, err := s.db.Exec(`DELETE FROM monitor_sites WHERE id=?`, id)
	if err != nil {
		return err
	}
	// Cleanup related data
	s.db.Exec(`DELETE FROM monitor_checks WHERE site_id=?`, id)
	s.db.Exec(`DELETE FROM monitor_ssl WHERE site_id=?`, id)
	s.db.Exec(`DELETE FROM monitor_health WHERE site_id=?`, id)
	s.db.Exec(`DELETE FROM monitor_domains WHERE site_id=?`, id)
	return nil
}

// --- Monitor Checks ---

func (s *Store) SaveMonitorCheck(check MonitorCheck) error {
	now := float64(time.Now().Unix())
	_, err := s.db.Exec(
		`INSERT INTO monitor_checks (site_id, status, response_ms, error, checked_at) VALUES (?, ?, ?, ?, ?)`,
		check.SiteID, check.Status, check.ResponseMs, check.Error, now,
	)
	return err
}

func (s *Store) ListMonitorChecks(siteID int, limit int) ([]MonitorCheck, error) {
	rows, err := s.db.Query(
		`SELECT id, site_id, status, response_ms, error, checked_at FROM monitor_checks WHERE site_id=? ORDER BY checked_at DESC LIMIT ?`,
		siteID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []MonitorCheck
	for rows.Next() {
		var c MonitorCheck
		if err := rows.Scan(&c.ID, &c.SiteID, &c.Status, &c.ResponseMs, &c.Error, &c.CheckedAt); err != nil {
			return nil, err
		}
		checks = append(checks, c)
	}
	return checks, nil
}

func (s *Store) GetLastMonitorCheck(siteID int) (*MonitorCheck, error) {
	var c MonitorCheck
	err := s.db.QueryRow(
		`SELECT id, site_id, status, response_ms, error, checked_at FROM monitor_checks WHERE site_id=? ORDER BY checked_at DESC LIMIT 1`,
		siteID,
	).Scan(&c.ID, &c.SiteID, &c.Status, &c.ResponseMs, &c.Error, &c.CheckedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) GetUptimePercent(siteID int, sinceUnix float64) (float64, error) {
	var total, up int
	err := s.db.QueryRow(
		`SELECT COUNT(*), COUNT(*) FILTER (WHERE status >= 200 AND status < 400) FROM monitor_checks WHERE site_id=? AND checked_at >= ?`,
		siteID, sinceUnix,
	).Scan(&total, &up)
	if err != nil || total == 0 {
		return 100, err
	}
	return float64(up) / float64(total) * 100, nil
}

// --- SSL ---

func (s *Store) SaveMonitorSSL(ssl MonitorSSL) error {
	now := float64(time.Now().Unix())
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO monitor_ssl (site_id, issuer, subject, expires_at, days_remaining, is_valid, error, checked_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		ssl.SiteID, ssl.Issuer, ssl.Subject, ssl.ExpiresAt, ssl.DaysRemaining, ssl.IsValid, ssl.Error, now,
	)
	return err
}

func (s *Store) GetMonitorSSL(siteID int) (*MonitorSSL, error) {
	var ssl MonitorSSL
	err := s.db.QueryRow(
		`SELECT site_id, issuer, subject, expires_at, days_remaining, is_valid, error, checked_at FROM monitor_ssl WHERE site_id=?`,
		siteID,
	).Scan(&ssl.SiteID, &ssl.Issuer, &ssl.Subject, &ssl.ExpiresAt, &ssl.DaysRemaining, &ssl.IsValid, &ssl.Error, &ssl.CheckedAt)
	if err != nil {
		return nil, err
	}
	return &ssl, nil
}

// --- Health ---

func (s *Store) SaveMonitorHealth(h MonitorHealth) error {
	now := float64(time.Now().Unix())
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO monitor_health (site_id, status, data_json, checked_at) VALUES (?, ?, ?, ?)`,
		h.SiteID, h.Status, h.DataJSON, now,
	)
	return err
}

func (s *Store) GetMonitorHealth(siteID int) (*MonitorHealth, error) {
	var h MonitorHealth
	err := s.db.QueryRow(
		`SELECT site_id, status, data_json, checked_at FROM monitor_health WHERE site_id=?`,
		siteID,
	).Scan(&h.SiteID, &h.Status, &h.DataJSON, &h.CheckedAt)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// --- Domain ---

func (s *Store) SaveMonitorDomain(d MonitorDomain) error {
	now := float64(time.Now().Unix())
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO monitor_domains (site_id, domain, registrar, expires_at, days_remaining, error, checked_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		d.SiteID, d.Domain, d.Registrar, d.ExpiresAt, d.DaysRemaining, d.Error, now,
	)
	return err
}

func (s *Store) GetMonitorDomain(siteID int) (*MonitorDomain, error) {
	var d MonitorDomain
	err := s.db.QueryRow(
		`SELECT site_id, domain, registrar, expires_at, days_remaining, error, checked_at FROM monitor_domains WHERE site_id=?`,
		siteID,
	).Scan(&d.SiteID, &d.Domain, &d.Registrar, &d.ExpiresAt, &d.DaysRemaining, &d.Error, &d.CheckedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// --- Retention for monitor data ---

func (s *Store) CleanMonitorChecks(retainDays int) error {
	cutoff := float64(time.Now().AddDate(0, 0, -retainDays).Unix())
	_, err := s.db.Exec(`DELETE FROM monitor_checks WHERE checked_at < ?`, cutoff)
	return err
}
