package storage

import "fmt"

func (s *Store) UpdateJobLink(jobUUID, childRID string) error {
	_, err := s.db.Exec(
		`UPDATE job_links SET child_rid = ?, status = 'processing' WHERE job_uuid = ?`,
		childRID, jobUUID,
	)
	if err != nil {
		return fmt.Errorf("update job_link: %w", err)
	}
	return nil
}

func (s *Store) CompleteJobLink(jobUUID, status string, finishedAt float64) error {
	_, err := s.db.Exec(
		`UPDATE job_links SET status = ?, finished_at = ? WHERE job_uuid = ?`,
		status, finishedAt, jobUUID,
	)
	return err
}

func (s *Store) GetJobParentRID(jobUUID string) (string, error) {
	var parentRID string
	err := s.db.QueryRow(`SELECT parent_rid FROM job_links WHERE job_uuid = ?`, jobUUID).Scan(&parentRID)
	if err != nil {
		return "", err
	}
	return parentRID, nil
}

func (s *Store) SetTraceParent(rid, parentRID string) error {
	_, err := s.db.Exec(`UPDATE traces SET parent_rid = ? WHERE rid = ?`, parentRID, rid)
	return err
}

func (s *Store) GetChildTraces(parentRID string) ([]Trace, error) {
	rows, err := s.db.Query(
		`SELECT rid, parent_rid, type, method, uri, status, duration_ms, ts_start, ts_end,
			query_count, event_count, job_count, has_exception, has_n_plus_one, has_slow_query,
			has_duplicate_query, meta_json, events_json, warnings_json
		FROM traces WHERE parent_rid = ? ORDER BY ts_start ASC`, parentRID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var traces []Trace
	for rows.Next() {
		var t Trace
		err := rows.Scan(
			&t.RID, &t.ParentRID, &t.Type, &t.Method, &t.URI, &t.Status, &t.DurationMs,
			&t.TsStart, &t.TsEnd, &t.QueryCount, &t.EventCount, &t.JobCount,
			&t.HasException, &t.HasNPlusOne, &t.HasSlowQuery, &t.HasDuplicateQuery,
			&t.MetaJSON, &t.EventsJSON, &t.WarningsJSON,
		)
		if err != nil {
			return nil, err
		}
		traces = append(traces, t)
	}
	return traces, nil
}
