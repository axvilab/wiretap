package analyzer

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/axvi/wiretap/internal/storage"
)

var valuePattern = regexp.MustCompile(`'[^']*'|\b\d+\b`)

func normalizeSQL(sql string) string {
	s := strings.TrimSpace(sql)
	s = valuePattern.ReplaceAllString(s, "?")
	return s
}

func detectNPlusOne(events []storage.Event) []Warning {
	selectCounts := make(map[string]int)

	for _, ev := range events {
		if ev.Type != "query" {
			continue
		}
		var data struct {
			SQL string `json:"sql"`
		}
		json.Unmarshal(ev.Data, &data)

		normalized := normalizeSQL(data.SQL)
		if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(normalized)), "SELECT") {
			selectCounts[normalized]++
		}
	}

	var warnings []Warning
	for sql, count := range selectCounts {
		if count >= 3 {
			warnings = append(warnings, Warning{
				Type:       "n_plus_one",
				Message:    fmt.Sprintf("Same SELECT executed %d times", count),
				Suggestion: "Consider eager loading with ::with()",
				Count:      count,
				SQL:        sql,
			})
		}
	}
	return warnings
}
