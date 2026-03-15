package analyzer

import (
	"encoding/json"
	"fmt"

	"github.com/axvi/wiretap/internal/storage"
)

func detectSlow(events []storage.Event, requestDurationMs float64, thresholds Thresholds) []Warning {
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

		if data.TimeMs >= thresholds.SlowQueryMs {
			warnings = append(warnings, Warning{
				Type:    "slow_query",
				Message: fmt.Sprintf("Slow query: %.1fms", data.TimeMs),
				SQL:     data.SQL,
			})
		}
	}

	if requestDurationMs >= thresholds.SlowRequestMs {
		warnings = append(warnings, Warning{
			Type:    "slow_request",
			Message: fmt.Sprintf("Slow request: %.1fms", requestDurationMs),
		})
	}

	return warnings
}