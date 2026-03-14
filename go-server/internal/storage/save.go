package storage

import (
	"encoding/json"
	"fmt"
	"log"
)

func (s *Store) SaveTrace(p *TracePayload) error {
	t := traceFromPayload(p)
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO traces (
			rid, parent_rid, type, method, uri, status, duration_ms,
			ts_start, ts_end, query_count, event_count, job_count,
			has_exception, meta_json, events_json, warnings_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.RID, t.ParentRID, t.Type, t.Method, t.URI, t.Status, t.DurationMs,
		t.TsStart, t.TsEnd, t.QueryCount, t.EventCount, t.JobCount,
		t.HasException, t.MetaJSON, t.EventsJSON, t.WarningsJSON,
	)
	if err != nil {
		return fmt.Errorf("insert trace: %w", err)
	}

	// Extract job.queued events → insert job_links
	for _, ev := range p.Events {
		if ev.Type == "job.queued" {
			if err := s.saveJobLink(p.RID, ev, p.TsStart); err != nil {
				return fmt.Errorf("insert job_link: %w", err)
			}
		}
	}
	return nil
}

func (s *Store) SaveTraceBatch(payloads []*TracePayload) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	traceStmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO traces (
			rid, parent_rid, type, method, uri, status, duration_ms,
			ts_start, ts_end, query_count, event_count, job_count,
			has_exception, meta_json, events_json, warnings_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare trace stmt: %w", err)
	}
	defer traceStmt.Close()

	jobStmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO job_links (
			job_uuid, job_class, queue, parent_rid, status, dispatched_at
		) VALUES (?, ?, ?, ?, 'pending', ?)`)
	if err != nil {
		return fmt.Errorf("prepare job stmt: %w", err)
	}
	defer jobStmt.Close()

	for _, p := range payloads {
		t := traceFromPayload(p)
		_, err := traceStmt.Exec(
			t.RID, t.ParentRID, t.Type, t.Method, t.URI, t.Status, t.DurationMs,
			t.TsStart, t.TsEnd, t.QueryCount, t.EventCount, t.JobCount,
			t.HasException, t.MetaJSON, t.EventsJSON, t.WarningsJSON,
		)
		if err != nil {
			return fmt.Errorf("exec trace %s: %w", t.RID, err)
		}

		for _, ev := range p.Events {
			if ev.Type != "job.queued" {
				continue
			}
			var data struct {
				JobUUID  string `json:"job_uuid"`
				JobClass string `json:"job_class"`
				Queue    string `json:"queue"`
			}
			if err := json.Unmarshal(ev.Data, &data); err != nil {
				continue
			}
			if data.JobUUID == "" {
				continue
			}
			queue := data.Queue
			if queue == "" {
				queue = "default"
			}
			_, err := jobStmt.Exec(data.JobUUID, data.JobClass, queue, p.RID, p.TsStart)
			if err != nil {
				return fmt.Errorf("exec job_link: %w", err)
			}
		}
	}

	return tx.Commit()
}

// ExtractEmails saves any mail.sent events as Email records.
// Must be called after the batch transaction is committed.
func (s *Store) ExtractEmails(p *TracePayload) {
	for _, ev := range p.Events {
		if ev.Type == "mail.sent" {
			if err := s.saveEmailFromEvent(p.RID, ev, p.TsStart); err != nil {
				log.Printf("[storage] failed to save email from trace %s: %v", p.RID, err)
			}
		}
	}
}

func (s *Store) saveEmailFromEvent(traceRID string, ev Event, ts float64) error {
	var data struct {
		From        string `json:"from"`
		To          []string `json:"to"`
		Cc          []string `json:"cc"`
		Subject     string `json:"subject"`
		HTML        string `json:"html"`
		Text        string `json:"text"`
		Attachments []EmailAttachment `json:"attachments"`
	}
	if err := json.Unmarshal(ev.Data, &data); err != nil {
		return fmt.Errorf("unmarshal mail.sent: %w", err)
	}

	email := &Email{
		From:        data.From,
		To:          data.To,
		Cc:          data.Cc,
		Subject:     data.Subject,
		HTML:        data.HTML,
		Text:        data.Text,
		Attachments: data.Attachments,
		TraceRID:    traceRID,
		CreatedAt:   ts,
	}
	return s.SaveEmail(email)
}

func (s *Store) saveJobLink(parentRID string, ev Event, ts float64) error {
	var data struct {
		JobUUID  string `json:"job_uuid"`
		JobClass string `json:"job_class"`
		Queue    string `json:"queue"`
	}
	if err := json.Unmarshal(ev.Data, &data); err != nil {
		return nil // skip malformed
	}
	if data.JobUUID == "" {
		return nil
	}
	queue := data.Queue
	if queue == "" {
		queue = "default"
	}
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO job_links (job_uuid, job_class, queue, parent_rid, status, dispatched_at)
		VALUES (?, ?, ?, ?, 'pending', ?)`,
		data.JobUUID, data.JobClass, queue, parentRID, ts,
	)
	return err
}

func traceFromPayload(p *TracePayload) Trace {
	var meta struct {
		Method     string  `json:"method"`
		URI        string  `json:"uri"`
		Status     int     `json:"status"`
		DurationMs float64 `json:"duration_ms"`
	}
	json.Unmarshal(p.Meta, &meta)

	var queryCount, jobCount int
	var hasException bool
	for _, ev := range p.Events {
		switch ev.Type {
		case "query":
			queryCount++
		case "job.queued":
			jobCount++
		case "exception":
			hasException = true
		}
	}

	evJSON, _ := json.Marshal(p.Events)
	metaJSON := string(p.Meta)
	if metaJSON == "" {
		metaJSON = "{}"
	}

	durationMs := meta.DurationMs
	if durationMs == 0 && p.TsEnd > p.TsStart {
		durationMs = (p.TsEnd - p.TsStart) * 1000
	}

	return Trace{
		RID:          p.RID,
		Type:         p.Type,
		Method:       meta.Method,
		URI:          meta.URI,
		Status:       meta.Status,
		DurationMs:   durationMs,
		TsStart:      p.TsStart,
		TsEnd:        p.TsEnd,
		QueryCount:   queryCount,
		EventCount:   len(p.Events),
		JobCount:     jobCount,
		HasException: hasException,
		MetaJSON:     metaJSON,
		EventsJSON:   string(evJSON),
		WarningsJSON: "[]",
	}
}
