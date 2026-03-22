package analyzer

import (
	"encoding/json"
	"fmt"

	"github.com/axvi/wiretap/internal/storage"
)

func detectDuplicate(events []storage.Event) []Warning {
	type queryKey struct {
		SQL      string
		Bindings string
	}

	counts := make(map[queryKey]int)

	for _, ev := range events {
		if ev.Type != "query" {
			continue
		}
		var data struct {
			SQL      string `json:"sql"`
			Bindings []any  `json:"bindings"`
		}
		json.Unmarshal(ev.Data, &data)

		bindJSON, _ := json.Marshal(data.Bindings)
		key := queryKey{SQL: data.SQL, Bindings: string(bindJSON)}
		counts[key]++
	}

	var warnings []Warning
	for key, count := range counts {
		if count >= 2 {
			warnings = append(warnings, Warning{
				Type:       "duplicate_query",
				Message:    fmt.Sprintf("Duplicate query executed %d times", count),
				Suggestion: "Consider caching the result to avoid repeated identical queries",
				Count:      count,
				SQL:        key.SQL,
			})
		}
	}
	return warnings
}
