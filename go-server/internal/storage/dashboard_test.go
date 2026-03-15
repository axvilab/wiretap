package storage

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// insertTestTrace is a helper that inserts a trace directly into the DB for testing.
func insertTestTrace(t *testing.T, s *Store, rid, traceType, method, uri string, status int, durationMs, tsStart float64, eventsJSON string) {
	t.Helper()
	if eventsJSON == "" {
		eventsJSON = "[]"
	}
	_, err := s.db.Exec(`
		INSERT INTO traces (rid, type, method, uri, status, duration_ms, ts_start, ts_end, meta_json, events_json, warnings_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, '{}', ?, '[]')`,
		rid, traceType, method, uri, status, durationMs, tsStart, tsStart+durationMs/1000, eventsJSON,
	)
	if err != nil {
		t.Fatalf("insertTestTrace %s: %v", rid, err)
	}
}

func TestDashboardSlowQueriesSorting(t *testing.T) {
	s := tempStore(t)
	now := float64(time.Now().Unix())

	// Create events JSON with query events of varying durations
	makeEventsJSON := func(queries []struct{ sql string; timeMs float64 }) string {
		var events []Event
		for _, q := range queries {
			data, _ := json.Marshal(map[string]any{
				"sql":     q.sql,
				"time_ms": q.timeMs,
			})
			events = append(events, Event{
				Type: "query",
				Ts:   now,
				Data: data,
			})
		}
		b, _ := json.Marshal(events)
		return string(b)
	}

	// Trace 1: two queries, slow (150ms) and fast (5ms)
	events1 := makeEventsJSON([]struct{ sql string; timeMs float64 }{
		{"SELECT * FROM users", 150.0},
		{"SELECT * FROM posts", 5.0},
	})
	insertTestTrace(t, s, "rid-1", "request", "GET", "/users", 200, 200, now-10, events1)

	// Trace 2: one query, medium (75ms)
	events2 := makeEventsJSON([]struct{ sql string; timeMs float64 }{
		{"SELECT * FROM orders WHERE id = 1", 75.0},
	})
	insertTestTrace(t, s, "rid-2", "request", "GET", "/orders", 200, 100, now-5, events2)

	// Trace 3: one query, slowest (300ms)
	events3 := makeEventsJSON([]struct{ sql string; timeMs float64 }{
		{"SELECT * FROM analytics GROUP BY date", 300.0},
	})
	insertTestTrace(t, s, "rid-3", "request", "GET", "/analytics", 200, 350, now-2, events3)

	since := now - 3600 // last hour
	results, err := s.DashboardSlowQueries(since, 10)
	if err != nil {
		t.Fatal(err)
	}

	// Should have 4 query events total
	if len(results) != 4 {
		t.Fatalf("expected 4 slow queries, got %d", len(results))
	}

	// Verify sorted by time_ms descending
	if results[0].TimeMs != 300.0 {
		t.Errorf("expected first query at 300ms, got %.2f", results[0].TimeMs)
	}
	if results[1].TimeMs != 150.0 {
		t.Errorf("expected second query at 150ms, got %.2f", results[1].TimeMs)
	}
	if results[2].TimeMs != 75.0 {
		t.Errorf("expected third query at 75ms, got %.2f", results[2].TimeMs)
	}
	if results[3].TimeMs != 5.0 {
		t.Errorf("expected fourth query at 5ms, got %.2f", results[3].TimeMs)
	}
}

func TestDashboardSlowQueriesLimit(t *testing.T) {
	s := tempStore(t)
	now := float64(time.Now().Unix())

	// Insert a trace with 5 query events
	var events []Event
	for i := 0; i < 5; i++ {
		data, _ := json.Marshal(map[string]any{
			"sql":     fmt.Sprintf("SELECT %d FROM t", i),
			"time_ms": float64((i + 1) * 10),
		})
		events = append(events, Event{
			Type: "query",
			Ts:   now,
			Data: data,
		})
	}
	evJSON, _ := json.Marshal(events)
	insertTestTrace(t, s, "rid-many", "request", "GET", "/many", 200, 100, now-10, string(evJSON))

	since := now - 3600
	results, err := s.DashboardSlowQueries(since, 3)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results with limit=3, got %d", len(results))
	}

	// Top 3 should be the slowest: 50ms, 40ms, 30ms
	if results[0].TimeMs != 50.0 {
		t.Errorf("expected first at 50ms, got %.2f", results[0].TimeMs)
	}
	if results[1].TimeMs != 40.0 {
		t.Errorf("expected second at 40ms, got %.2f", results[1].TimeMs)
	}
	if results[2].TimeMs != 30.0 {
		t.Errorf("expected third at 30ms, got %.2f", results[2].TimeMs)
	}
}

func TestDashboardSlowQueriesTraceMetadata(t *testing.T) {
	s := tempStore(t)
	now := float64(time.Now().Unix())

	data, _ := json.Marshal(map[string]any{
		"sql":     "SELECT * FROM products",
		"time_ms": 42.5,
	})
	events := []Event{{Type: "query", Ts: now, Data: data}}
	evJSON, _ := json.Marshal(events)

	insertTestTrace(t, s, "rid-meta", "request", "POST", "/api/products", 200, 50, now-1, string(evJSON))

	since := now - 3600
	results, err := s.DashboardSlowQueries(since, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	q := results[0]
	if q.SQL != "SELECT * FROM products" {
		t.Errorf("expected SQL 'SELECT * FROM products', got %q", q.SQL)
	}
	if q.TimeMs != 42.5 {
		t.Errorf("expected 42.5ms, got %.2f", q.TimeMs)
	}
	if q.TraceRID != "rid-meta" {
		t.Errorf("expected trace_rid 'rid-meta', got %q", q.TraceRID)
	}
	if q.TraceType != "request" {
		t.Errorf("expected trace_type 'request', got %q", q.TraceType)
	}
	if q.TraceURI != "/api/products" {
		t.Errorf("expected trace_uri '/api/products', got %q", q.TraceURI)
	}
}

func TestDashboardSlowQueriesEmpty(t *testing.T) {
	s := tempStore(t)
	now := float64(time.Now().Unix())

	results, err := s.DashboardSlowQueries(now-3600, 10)
	if err != nil {
		t.Fatal(err)
	}

	if results == nil {
		t.Error("expected non-nil empty slice")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestDashboardSlowQueriesIgnoresNonQueryEvents(t *testing.T) {
	s := tempStore(t)
	now := float64(time.Now().Unix())

	// Mix of query and non-query events
	queryData, _ := json.Marshal(map[string]any{"sql": "SELECT 1", "time_ms": 10.0})
	cacheData, _ := json.Marshal(map[string]any{"key": "users.1", "hit": true})
	logData, _ := json.Marshal(map[string]any{"message": "test log"})

	events := []Event{
		{Type: "query", Ts: now, Data: queryData},
		{Type: "cache", Ts: now, Data: cacheData},
		{Type: "log", Ts: now, Data: logData},
	}
	evJSON, _ := json.Marshal(events)
	insertTestTrace(t, s, "rid-mixed", "request", "GET", "/mixed", 200, 50, now-1, string(evJSON))

	results, err := s.DashboardSlowQueries(now-3600, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 query result (ignoring non-query events), got %d", len(results))
	}
	if results[0].SQL != "SELECT 1" {
		t.Errorf("expected 'SELECT 1', got %q", results[0].SQL)
	}
}

func TestDashboardSlowQueriesRespectsTimeFilter(t *testing.T) {
	s := tempStore(t)
	now := float64(time.Now().Unix())

	// Old trace (2 hours ago)
	oldData, _ := json.Marshal(map[string]any{"sql": "SELECT old", "time_ms": 500.0})
	oldEvents, _ := json.Marshal([]Event{{Type: "query", Ts: now - 7200, Data: oldData}})
	insertTestTrace(t, s, "rid-old", "request", "GET", "/old", 200, 50, now-7200, string(oldEvents))

	// Recent trace (5 minutes ago)
	newData, _ := json.Marshal(map[string]any{"sql": "SELECT new", "time_ms": 10.0})
	newEvents, _ := json.Marshal([]Event{{Type: "query", Ts: now - 300, Data: newData}})
	insertTestTrace(t, s, "rid-new", "request", "GET", "/new", 200, 20, now-300, string(newEvents))

	// Query with since = 1 hour ago; should only get the recent trace
	since := now - 3600
	results, err := s.DashboardSlowQueries(since, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result (filtered by time), got %d", len(results))
	}
	if results[0].SQL != "SELECT new" {
		t.Errorf("expected 'SELECT new', got %q", results[0].SQL)
	}
}

func TestDashboardSlowQueriesFromJobTraces(t *testing.T) {
	s := tempStore(t)
	now := float64(time.Now().Unix())

	data, _ := json.Marshal(map[string]any{"sql": "UPDATE jobs SET status = 'done'", "time_ms": 88.0})
	events, _ := json.Marshal([]Event{{Type: "query", Ts: now, Data: data}})
	insertTestTrace(t, s, "rid-job", "job", "", "", 0, 100, now-1, string(events))

	results, err := s.DashboardSlowQueries(now-3600, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result from job trace, got %d", len(results))
	}
	if results[0].TraceType != "job" {
		t.Errorf("expected trace_type 'job', got %q", results[0].TraceType)
	}
	if results[0].SQL != "UPDATE jobs SET status = 'done'" {
		t.Errorf("unexpected SQL: %q", results[0].SQL)
	}
}
