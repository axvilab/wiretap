package analyzer

import (
	"encoding/json"
	"strconv"

	"github.com/axvi/wiretap/internal/storage"
)

type Warning struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
	Count      int    `json:"count,omitempty"`
	SQL        string `json:"sql,omitempty"`
}

type Thresholds struct {
	SlowQueryMs   float64
	SlowRequestMs float64
}

func LoadThresholds(store *storage.Store) Thresholds {
	t := Thresholds{
		SlowQueryMs:   100,
		SlowRequestMs: 1000,
	}
	if v, _ := store.GetSetting("slow_query_ms"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil && n > 0 {
			t.SlowQueryMs = n
		}
	}
	if v, _ := store.GetSetting("slow_request_ms"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil && n > 0 {
			t.SlowRequestMs = n
		}
	}
	return t
}

// Analyze runs all sub-analyzers on a trace and returns warnings.
func Analyze(trace *storage.Trace, thresholds Thresholds) []Warning {
	var events []storage.Event
	json.Unmarshal([]byte(trace.EventsJSON), &events)

	var warnings []Warning
	warnings = append(warnings, detectNPlusOne(events)...)
	warnings = append(warnings, detectDuplicate(events)...)
	warnings = append(warnings, detectSlow(events, trace.DurationMs, thresholds)...)

	return warnings
}