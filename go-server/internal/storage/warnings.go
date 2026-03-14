package storage

import "encoding/json"

func (s *Store) UpdateWarnings(rid string, warnings any) error {
	wJSON, err := json.Marshal(warnings)
	if err != nil {
		return err
	}

	// Detect flags from warnings
	var hasNPlusOne, hasSlowQuery, hasDuplicate bool
	if ws, ok := warnings.([]struct {
		Type string `json:"type"`
	}); ok {
		for _, w := range ws {
			switch w.Type {
			case "n_plus_one":
				hasNPlusOne = true
			case "slow_query", "slow_request":
				hasSlowQuery = true
			case "duplicate_query":
				hasDuplicate = true
			}
		}
	} else {
		// Try raw JSON approach
		var raw []map[string]any
		json.Unmarshal(wJSON, &raw)
		for _, w := range raw {
			switch w["type"] {
			case "n_plus_one":
				hasNPlusOne = true
			case "slow_query", "slow_request":
				hasSlowQuery = true
			case "duplicate_query":
				hasDuplicate = true
			}
		}
	}

	_, err = s.db.Exec(`
		UPDATE traces SET
			warnings_json = ?,
			has_n_plus_one = ?,
			has_slow_query = ?,
			has_duplicate_query = ?
		WHERE rid = ?`,
		string(wJSON), hasNPlusOne, hasSlowQuery, hasDuplicate, rid,
	)
	return err
}
