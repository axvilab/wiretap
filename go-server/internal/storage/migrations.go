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
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}
