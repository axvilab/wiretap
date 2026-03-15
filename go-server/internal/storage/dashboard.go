package storage

import (
	"fmt"
	"math"
	"time"
)

type DashboardStatsResult struct {
	TotalRequests  int     `json:"total_requests"`
	TotalJobs      int     `json:"total_jobs"`
	TotalEmails    int     `json:"total_emails"`
	TotalExceptions int    `json:"total_exceptions"`
	AvgDuration    float64 `json:"avg_duration"`
	MedianDuration float64 `json:"median_duration"`
	P95Duration    float64 `json:"p95_duration"`
	RequestsPerMin float64 `json:"requests_per_min"`
}

type TimelinePoint struct {
	Timestamp  string  `json:"timestamp"`
	Count      int     `json:"count"`
	AvgMs      float64 `json:"avg_ms"`
	P95Ms      float64 `json:"p95_ms"`
	ErrorCount int     `json:"error_count"`
}

type StatusGroup struct {
	Group string `json:"group"`
	Count int    `json:"count"`
}

type SlowEndpoint struct {
	Method      string  `json:"method"`
	URI         string  `json:"uri"`
	Count       int     `json:"count"`
	AvgMs       float64 `json:"avg_ms"`
	MaxMs       float64 `json:"max_ms"`
	P95Ms       float64 `json:"p95_ms"`
}

func (s *Store) DashboardStats(since float64) (*DashboardStatsResult, error) {
	var stats DashboardStatsResult

	s.db.QueryRow(`SELECT COUNT(*) FROM traces WHERE type = 'request' AND ts_start >= ?`, since).Scan(&stats.TotalRequests)
	s.db.QueryRow(`SELECT COUNT(*) FROM traces WHERE type = 'job' AND ts_start >= ?`, since).Scan(&stats.TotalJobs)
	s.db.QueryRow(`SELECT COUNT(*) FROM emails WHERE created_at >= ?`, since).Scan(&stats.TotalEmails)
	s.db.QueryRow(`SELECT COUNT(*) FROM traces WHERE has_exception = true AND ts_start >= ?`, since).Scan(&stats.TotalExceptions)

	s.db.QueryRow(`SELECT COALESCE(AVG(duration_ms), 0) FROM traces WHERE type = 'request' AND ts_start >= ?`, since).Scan(&stats.AvgDuration)
	stats.AvgDuration = math.Round(stats.AvgDuration*100) / 100

	s.db.QueryRow(`SELECT COALESCE(quantile_cont(duration_ms, 0.5), 0) FROM traces WHERE type = 'request' AND ts_start >= ?`, since).Scan(&stats.MedianDuration)
	stats.MedianDuration = math.Round(stats.MedianDuration*100) / 100

	s.db.QueryRow(`SELECT COALESCE(quantile_cont(duration_ms, 0.95), 0) FROM traces WHERE type = 'request' AND ts_start >= ?`, since).Scan(&stats.P95Duration)
	stats.P95Duration = math.Round(stats.P95Duration*100) / 100

	// Requests per minute
	var minutes float64
	s.db.QueryRow(`SELECT (EXTRACT(EPOCH FROM NOW()) - ?) / 60`, since).Scan(&minutes)
	if minutes > 0 && stats.TotalRequests > 0 {
		stats.RequestsPerMin = math.Round(float64(stats.TotalRequests)/minutes*100) / 100
	}

	return &stats, nil
}

func (s *Store) DashboardTimeline(since float64, hours int) ([]TimelinePoint, error) {
	// Determine bucket size based on period
	var bucketMinutes int
	switch {
	case hours <= 1:
		bucketMinutes = 1  // 1 minute buckets for 1h
	case hours <= 24:
		bucketMinutes = 15 // 15 minute buckets for 24h
	default:
		bucketMinutes = 60 // 1 hour buckets for 7d
	}

	bucketSeconds := bucketMinutes * 60
	rows, err := s.db.Query(fmt.Sprintf(`
		SELECT
			CAST(FLOOR(ts_start / %d) * %d AS BIGINT) as bucket,
			COUNT(*) as cnt,
			COALESCE(AVG(duration_ms), 0) as avg_ms,
			COALESCE(quantile_cont(duration_ms, 0.95), 0) as p95_ms,
			SUM(CASE WHEN status >= 500 THEN 1 ELSE 0 END) as errors
		FROM traces
		WHERE type = 'request' AND ts_start >= ?
		GROUP BY bucket
		ORDER BY bucket ASC
	`, bucketSeconds, bucketSeconds), since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []TimelinePoint
	for rows.Next() {
		var bucket int64
		var p TimelinePoint
		rows.Scan(&bucket, &p.Count, &p.AvgMs, &p.P95Ms, &p.ErrorCount)
		p.Timestamp = time.Unix(bucket, 0).Format("15:04")
		p.AvgMs = math.Round(p.AvgMs*100) / 100
		p.P95Ms = math.Round(p.P95Ms*100) / 100
		points = append(points, p)
	}

	if points == nil {
		points = []TimelinePoint{}
	}
	return points, nil
}

func (s *Store) DashboardStatusBreakdown(since float64) ([]StatusGroup, error) {
	rows, err := s.db.Query(`
		SELECT
			CASE
				WHEN status >= 500 THEN '5xx'
				WHEN status >= 400 THEN '4xx'
				WHEN status >= 300 THEN '3xx'
				WHEN status >= 200 THEN '2xx'
				ELSE 'other'
			END as grp,
			COUNT(*) as cnt
		FROM traces
		WHERE type = 'request' AND ts_start >= ?
		GROUP BY grp
		ORDER BY grp
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []StatusGroup
	for rows.Next() {
		var g StatusGroup
		rows.Scan(&g.Group, &g.Count)
		groups = append(groups, g)
	}
	if groups == nil {
		groups = []StatusGroup{}
	}
	return groups, nil
}

func (s *Store) DashboardSlowEndpoints(since float64, limit int) ([]SlowEndpoint, error) {
	rows, err := s.db.Query(`
		SELECT
			method,
			uri,
			COUNT(*) as cnt,
			ROUND(AVG(duration_ms), 2) as avg_ms,
			ROUND(MAX(duration_ms), 2) as max_ms,
			ROUND(COALESCE(quantile_cont(duration_ms, 0.95), 0), 2) as p95_ms
		FROM traces
		WHERE type = 'request' AND ts_start >= ? AND uri != ''
		GROUP BY method, uri
		ORDER BY avg_ms DESC
		LIMIT ?
	`, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endpoints []SlowEndpoint
	for rows.Next() {
		var e SlowEndpoint
		rows.Scan(&e.Method, &e.URI, &e.Count, &e.AvgMs, &e.MaxMs, &e.P95Ms)
		endpoints = append(endpoints, e)
	}
	if endpoints == nil {
		endpoints = []SlowEndpoint{}
	}
	return endpoints, nil
}
