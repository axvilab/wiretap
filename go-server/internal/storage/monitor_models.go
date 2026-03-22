package storage

// MonitorSite is a URL to monitor for uptime, SSL, and health.
type MonitorSite struct {
	ID          int     `json:"id"`
	URL         string  `json:"url"`
	Name        string  `json:"name"`
	CheckSSL    bool    `json:"check_ssl"`
	HealthURL   string  `json:"health_url"`
	AuthType    string  `json:"auth_type"`  // none, bearer, basic, header, query
	AuthJSON    string  `json:"auth_json"`  // JSON config per auth type
	IntervalS   int     `json:"interval_s"`
	IsPrimary   bool    `json:"is_primary"`
	CreatedAt   float64 `json:"created_at"`
}

// SiteAuth is the parsed auth config from AuthJSON.
type SiteAuth struct {
	Token    string `json:"token,omitempty"`    // bearer
	Username string `json:"username,omitempty"` // basic
	Password string `json:"password,omitempty"` // basic
	Header   string `json:"header,omitempty"`   // header
	Value    string `json:"value,omitempty"`    // header / query
	Param    string `json:"param,omitempty"`    // query
}

// MonitorCheck is a single uptime check result.
type MonitorCheck struct {
	ID         int     `json:"id"`
	SiteID     int     `json:"site_id"`
	Status     int     `json:"status"`
	ResponseMs int    `json:"response_ms"`
	Error      string  `json:"error"`
	CheckedAt  float64 `json:"checked_at"`
}

// MonitorSSL holds SSL certificate info for a site.
type MonitorSSL struct {
	SiteID        int     `json:"site_id"`
	Issuer        string  `json:"issuer"`
	Subject       string  `json:"subject"`
	ExpiresAt     float64 `json:"expires_at"`
	DaysRemaining int     `json:"days_remaining"`
	IsValid       bool    `json:"is_valid"`
	Error         string  `json:"error"`
	CheckedAt     float64 `json:"checked_at"`
}

// MonitorHealth holds health endpoint data for a site.
type MonitorHealth struct {
	SiteID    int     `json:"site_id"`
	Status    string  `json:"status"`
	DataJSON  string  `json:"data_json"`
	CheckedAt float64 `json:"checked_at"`
}

// MonitorDomain holds WHOIS domain expiry info.
type MonitorDomain struct {
	SiteID        int     `json:"site_id"`
	Domain        string  `json:"domain"`
	Registrar     string  `json:"registrar"`
	ExpiresAt     float64 `json:"expires_at"`
	DaysRemaining int     `json:"days_remaining"`
	Error         string  `json:"error"`
	CheckedAt     float64 `json:"checked_at"`
}

// ScheduleHeartbeat tracks expected scheduled tasks.
type ScheduleHeartbeat struct {
	ID         int     `json:"id"`
	Command    string  `json:"command"`
	Mode       string  `json:"mode"`        // "every" or "at"
	MaxGapM    int     `json:"max_gap_m"`   // for "every": interval + grace
	AtTime     string  `json:"at_time"`     // for "at": "HH:MM"
	GraceM     int     `json:"grace_m"`     // grace period minutes
	LastSeenAt float64 `json:"last_seen_at"`
	Alerted    bool    `json:"alerted"`
}

// NotificationChannel is a configured notification destination.
type NotificationChannel struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Config  string `json:"config"`
	Enabled bool   `json:"enabled"`
}

// AlertRule defines when to send a notification.
type AlertRule struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	ChannelID int    `json:"channel_id"`
	Config    string `json:"config"`
	Enabled   bool   `json:"enabled"`
	CooldownM int   `json:"cooldown_m"`
}

// AlertLog is a record of a sent alert.
type AlertLog struct {
	ID      int     `json:"id"`
	RuleID  int     `json:"rule_id"`
	Message string  `json:"message"`
	SentAt  float64 `json:"sent_at"`
}

// MonitorStatus is an aggregated status view of a site.
type MonitorStatus struct {
	Site   MonitorSite    `json:"site"`
	SSL    *MonitorSSL    `json:"ssl,omitempty"`
	Health *MonitorHealth `json:"health,omitempty"`
	Domain *MonitorDomain `json:"domain,omitempty"`
	LastCheck *MonitorCheck `json:"last_check,omitempty"`
	UptimePct float64      `json:"uptime_pct"`
}
