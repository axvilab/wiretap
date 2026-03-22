package storage

func (s *Store) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS traces (
			rid              VARCHAR PRIMARY KEY,
			parent_rid       VARCHAR DEFAULT '',
			type             VARCHAR NOT NULL,
			method           VARCHAR DEFAULT '',
			uri              VARCHAR DEFAULT '',
			status           INTEGER DEFAULT 0,
			duration_ms      DOUBLE DEFAULT 0,
			ts_start         DOUBLE NOT NULL,
			ts_end           DOUBLE DEFAULT 0,
			query_count      INTEGER DEFAULT 0,
			event_count      INTEGER DEFAULT 0,
			job_count        INTEGER DEFAULT 0,
			has_exception    BOOLEAN DEFAULT FALSE,
			has_n_plus_one   BOOLEAN DEFAULT FALSE,
			has_slow_query   BOOLEAN DEFAULT FALSE,
			has_duplicate_query BOOLEAN DEFAULT FALSE,
			meta_json        VARCHAR DEFAULT '{}',
			events_json      VARCHAR DEFAULT '[]',
			warnings_json    VARCHAR DEFAULT '[]'
		)`,
		`CREATE TABLE IF NOT EXISTS job_links (
			job_uuid      VARCHAR PRIMARY KEY,
			job_class     VARCHAR NOT NULL,
			queue         VARCHAR DEFAULT 'default',
			parent_rid    VARCHAR NOT NULL,
			child_rid     VARCHAR DEFAULT '',
			status        VARCHAR DEFAULT 'pending',
			dispatched_at DOUBLE NOT NULL,
			started_at    DOUBLE DEFAULT 0,
			finished_at   DOUBLE DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key   VARCHAR PRIMARY KEY,
			value VARCHAR NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			token      VARCHAR PRIMARY KEY,
			expires_at DOUBLE NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS emails (
			id              VARCHAR PRIMARY KEY,
			from_addr       VARCHAR NOT NULL,
			to_addrs        VARCHAR DEFAULT '[]',
			cc_addrs        VARCHAR DEFAULT '[]',
			subject         VARCHAR DEFAULT '',
			html            VARCHAR DEFAULT '',
			text_body       VARCHAR DEFAULT '',
			attachments_json VARCHAR DEFAULT '[]',
			trace_rid       VARCHAR DEFAULT '',
			created_at      DOUBLE NOT NULL
		)`,
	// Monitoring & notifications tables
	`CREATE SEQUENCE IF NOT EXISTS monitor_sites_seq START 1`,
	`CREATE TABLE IF NOT EXISTS monitor_sites (
		id           INTEGER PRIMARY KEY DEFAULT nextval('monitor_sites_seq'),
		url          VARCHAR NOT NULL,
		name         VARCHAR NOT NULL DEFAULT '',
		check_ssl    BOOLEAN NOT NULL DEFAULT true,
		health_url   VARCHAR NOT NULL DEFAULT '',
		auth_type    VARCHAR NOT NULL DEFAULT 'none',
		auth_json    VARCHAR NOT NULL DEFAULT '{}',
		interval_s   INTEGER NOT NULL DEFAULT 30,
		is_primary   BOOLEAN NOT NULL DEFAULT false,
		created_at   DOUBLE NOT NULL DEFAULT 0
	)`,
	`CREATE SEQUENCE IF NOT EXISTS monitor_checks_seq START 1`,
	`CREATE TABLE IF NOT EXISTS monitor_checks (
		id          INTEGER PRIMARY KEY DEFAULT nextval('monitor_checks_seq'),
		site_id     INTEGER NOT NULL,
		status      INTEGER NOT NULL DEFAULT 0,
		response_ms INTEGER NOT NULL DEFAULT 0,
		error       VARCHAR NOT NULL DEFAULT '',
		checked_at  DOUBLE NOT NULL DEFAULT 0
	)`,
	`CREATE TABLE IF NOT EXISTS monitor_ssl (
		site_id         INTEGER PRIMARY KEY,
		issuer          VARCHAR NOT NULL DEFAULT '',
		subject         VARCHAR NOT NULL DEFAULT '',
		expires_at      DOUBLE NOT NULL DEFAULT 0,
		days_remaining  INTEGER NOT NULL DEFAULT 0,
		is_valid        BOOLEAN NOT NULL DEFAULT true,
		error           VARCHAR NOT NULL DEFAULT '',
		checked_at      DOUBLE NOT NULL DEFAULT 0
	)`,
	`CREATE TABLE IF NOT EXISTS monitor_health (
		site_id     INTEGER PRIMARY KEY,
		status      VARCHAR NOT NULL DEFAULT 'unknown',
		data_json   VARCHAR NOT NULL DEFAULT '{}',
		checked_at  DOUBLE NOT NULL DEFAULT 0
	)`,
	`CREATE TABLE IF NOT EXISTS monitor_domains (
		site_id         INTEGER PRIMARY KEY,
		domain          VARCHAR NOT NULL DEFAULT '',
		registrar       VARCHAR NOT NULL DEFAULT '',
		expires_at      DOUBLE NOT NULL DEFAULT 0,
		days_remaining  INTEGER NOT NULL DEFAULT 0,
		error           VARCHAR NOT NULL DEFAULT '',
		checked_at      DOUBLE NOT NULL DEFAULT 0
	)`,
	`CREATE SEQUENCE IF NOT EXISTS schedule_heartbeats_seq START 1`,
	`CREATE TABLE IF NOT EXISTS schedule_heartbeats (
		id           INTEGER PRIMARY KEY DEFAULT nextval('schedule_heartbeats_seq'),
		command      VARCHAR NOT NULL,
		mode         VARCHAR NOT NULL DEFAULT 'every',
		max_gap_m    INTEGER NOT NULL DEFAULT 60,
		at_time      VARCHAR NOT NULL DEFAULT '',
		grace_m      INTEGER NOT NULL DEFAULT 15,
		last_seen_at DOUBLE DEFAULT 0,
		alerted      BOOLEAN NOT NULL DEFAULT false
	)`,
	`CREATE SEQUENCE IF NOT EXISTS notification_channels_seq START 1`,
	`CREATE TABLE IF NOT EXISTS notification_channels (
		id      INTEGER PRIMARY KEY DEFAULT nextval('notification_channels_seq'),
		type    VARCHAR NOT NULL,
		name    VARCHAR NOT NULL DEFAULT '',
		config  VARCHAR NOT NULL DEFAULT '{}',
		enabled BOOLEAN NOT NULL DEFAULT true
	)`,
	`CREATE SEQUENCE IF NOT EXISTS alert_rules_seq START 1`,
	`CREATE TABLE IF NOT EXISTS alert_rules (
		id          INTEGER PRIMARY KEY DEFAULT nextval('alert_rules_seq'),
		type        VARCHAR NOT NULL,
		channel_id  INTEGER NOT NULL,
		config      VARCHAR NOT NULL DEFAULT '{}',
		enabled     BOOLEAN NOT NULL DEFAULT true,
		cooldown_m  INTEGER NOT NULL DEFAULT 5
	)`,
	`CREATE SEQUENCE IF NOT EXISTS alert_log_seq START 1`,
	`CREATE TABLE IF NOT EXISTS alert_log (
		id       INTEGER PRIMARY KEY DEFAULT nextval('alert_log_seq'),
		rule_id  INTEGER NOT NULL,
		message  VARCHAR NOT NULL DEFAULT '',
		sent_at  DOUBLE NOT NULL DEFAULT 0
	)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}
