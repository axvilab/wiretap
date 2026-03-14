package analyzer

import (
	"encoding/json"

	"github.com/axvi/wiretap/internal/storage"
)

type Warning struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
	Count      int    `json:"count,omitempty"`
	SQL        string `json:"sql,omitempty"`
}

// Analyze runs all sub-analyzers on a trace and returns warnings.
func Analyze(trace *storage.Trace) []Warning {
	var events []storage.Event
	json.Unmarshal([]byte(trace.EventsJSON), &events)

	var warnings []Warning
	warnings = append(warnings, detectNPlusOne(events)...)
	warnings = append(warnings, detectDuplicate(events)...)
	warnings = append(warnings, detectSlow(events, trace.DurationMs)...)

	return warnings
}
