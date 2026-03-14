package storage

import (
	"fmt"
	"strings"
)

type ListParams struct {
	Page    int
	PerPage int
	Type    string
	Status  int
	Method  string
}

type ListResult struct {
	Traces []Trace `json:"traces"`
	Total  int     `json:"total"`
	Page   int     `json:"page"`
	PerPage int    `json:"per_page"`
}

func (s *Store) ListTraces(p ListParams) (*ListResult, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 100 {
		p.PerPage = 50
	}

	where, args := buildWhere(p)

	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM traces"+where, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count traces: %w", err)
	}

	offset := (p.Page - 1) * p.PerPage
	rows, err := s.db.Query(
		"SELECT rid, parent_rid, type, method, uri, status, duration_ms, ts_start, ts_end, "+
			"query_count, event_count, job_count, has_exception, has_n_plus_one, has_slow_query, "+
			"has_duplicate_query, meta_json, events_json, warnings_json "+
			"FROM traces"+where+" ORDER BY ts_start DESC LIMIT ? OFFSET ?",
		append(args, p.PerPage, offset)...,
	)
	if err != nil {
		return nil, fmt.Errorf("list traces: %w", err)
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
			return nil, fmt.Errorf("scan trace: %w", err)
		}
		traces = append(traces, t)
	}
	if traces == nil {
		traces = []Trace{}
	}

	return &ListResult{
		Traces:  traces,
		Total:   total,
		Page:    p.Page,
		PerPage: p.PerPage,
	}, nil
}

func (s *Store) GetTrace(rid string) (*Trace, error) {
	var t Trace
	err := s.db.QueryRow(
		"SELECT rid, parent_rid, type, method, uri, status, duration_ms, ts_start, ts_end, "+
			"query_count, event_count, job_count, has_exception, has_n_plus_one, has_slow_query, "+
			"has_duplicate_query, meta_json, events_json, warnings_json "+
			"FROM traces WHERE rid = ?", rid,
	).Scan(
		&t.RID, &t.ParentRID, &t.Type, &t.Method, &t.URI, &t.Status, &t.DurationMs,
		&t.TsStart, &t.TsEnd, &t.QueryCount, &t.EventCount, &t.JobCount,
		&t.HasException, &t.HasNPlusOne, &t.HasSlowQuery, &t.HasDuplicateQuery,
		&t.MetaJSON, &t.EventsJSON, &t.WarningsJSON,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func buildWhere(p ListParams) (string, []any) {
	var conds []string
	var args []any

	if p.Type != "" {
		conds = append(conds, "type = ?")
		args = append(args, p.Type)
	}
	if p.Status > 0 {
		conds = append(conds, "status = ?")
		args = append(args, p.Status)
	}
	if p.Method != "" {
		conds = append(conds, "method = ?")
		args = append(args, strings.ToUpper(p.Method))
	}

	if len(conds) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conds, " AND "), args
}
