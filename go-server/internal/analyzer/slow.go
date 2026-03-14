package analyzer

import (
	"encoding/json"
	"fmt"

	"github.com/axvi/wiretap/internal/storage"
)

const (
	slowQueryThresholdMs   = 100
	slowRequestThresholdMs = 1000
)

func detectSlow(events []storage.Event, requestDurationMs float64) []Warning {
	var warnings []Warning

	for _, ev := range events {
		if ev.Type != "query" {
			continue
		}
		var data struct {
			SQL    string  `json:"sql"`
			TimeMs float64 `json:"time_ms"`
		}
		json.Unmarshal(ev.Data, &data)

		if data.TimeMs >= slowQueryThresholdMs {
			warnings = append(warnings, Warning{
				Type:    "slow_query",
				Message: fmt.Sprintf("Slow query: %.1fms", data.TimeMs),
				SQL:     data.SQL,
			})
		}
	}

	if requestDurationMs >= slowRequestThresholdMs {
		warnings = append(warnings, Warning{
			Type:    "slow_request",
			Message: fmt.Sprintf("Slow request: %.1fms", requestDurationMs),
		})
	}

	return warnings
}
