package analyzer

import (
	"encoding/json"
	"testing"

	"github.com/axvi/wiretap/internal/storage"
)

func makeTrace(durationMs float64, events []storage.Event) *storage.Trace {
	evJSON, _ := json.Marshal(events)
	return &storage.Trace{
		DurationMs: durationMs,
		EventsJSON: string(evJSON),
	}
}

func queryEvent(sql string, timeMs float64) storage.Event {
	data, _ := json.Marshal(map[string]any{"sql": sql, "time_ms": timeMs})
	return storage.Event{Type: "query", Data: data}
}

func TestDetectSlowQuery(t *testing.T) {
	thresholds := Thresholds{SlowQueryMs: 100, SlowRequestMs: 1000}
	trace := makeTrace(50, []storage.Event{
		queryEvent("SELECT 1", 150),
		queryEvent("SELECT 2", 50),
	})

	warnings := Analyze(trace, thresholds)

	slowQueries := 0
	for _, w := range warnings {
		if w.Type == "slow_query" {
			slowQueries++
		}
	}
	if slowQueries != 1 {
		t.Errorf("expected 1 slow query warning, got %d", slowQueries)
	}
}

func TestDetectSlowRequest(t *testing.T) {
	thresholds := Thresholds{SlowQueryMs: 100, SlowRequestMs: 500}
	trace := makeTrace(600, []storage.Event{})

	warnings := Analyze(trace, thresholds)

	found := false
	for _, w := range warnings {
		if w.Type == "slow_request" {
			found = true
		}
	}
	if !found {
		t.Error("expected slow_request warning")
	}
}

func TestNoWarningsUnderThreshold(t *testing.T) {
	thresholds := Thresholds{SlowQueryMs: 100, SlowRequestMs: 1000}
	trace := makeTrace(50, []storage.Event{
		queryEvent("SELECT 1", 10),
	})

	warnings := Analyze(trace, thresholds)
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %d", len(warnings))
	}
}

func TestCustomThresholds(t *testing.T) {
	// Low thresholds
	thresholds := Thresholds{SlowQueryMs: 5, SlowRequestMs: 10}
	trace := makeTrace(20, []storage.Event{
		queryEvent("SELECT 1", 8),
	})

	warnings := Analyze(trace, thresholds)
	if len(warnings) != 2 { // slow query + slow request
		t.Errorf("expected 2 warnings, got %d", len(warnings))
	}
}

func TestDetectNPlusOne(t *testing.T) {
	events := []storage.Event{}
	for i := 0; i < 15; i++ {
		events = append(events, queryEvent("SELECT * FROM users WHERE id = ?", 1))
	}
	trace := makeTrace(50, events)
	thresholds := Thresholds{SlowQueryMs: 100, SlowRequestMs: 1000}

	warnings := Analyze(trace, thresholds)

	found := false
	for _, w := range warnings {
		if w.Type == "n_plus_one" {
			found = true
		}
	}
	if !found {
		t.Error("expected n_plus_one warning")
	}
}

func TestDetectDuplicateQuery(t *testing.T) {
	events := []storage.Event{
		queryEvent("SELECT * FROM settings", 1),
		queryEvent("SELECT * FROM settings", 1),
		queryEvent("SELECT * FROM settings", 1),
	}
	trace := makeTrace(50, events)
	thresholds := Thresholds{SlowQueryMs: 100, SlowRequestMs: 1000}

	warnings := Analyze(trace, thresholds)

	found := false
	for _, w := range warnings {
		if w.Type == "duplicate_query" {
			found = true
		}
	}
	if !found {
		t.Error("expected duplicate_query warning")
	}
}
