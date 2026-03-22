package storage

// ListKnownScheduleCommands returns unique schedule commands seen in traces.
func (s *Store) ListKnownScheduleCommands() ([]string, error) {
	rows, err := s.db.Query(`SELECT DISTINCT json_extract_string(meta_json, '$.command') as cmd FROM traces WHERE type = 'schedule' AND cmd != '' ORDER BY cmd`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []string
	for rows.Next() {
		var cmd string
		if err := rows.Scan(&cmd); err != nil {
			continue
		}
		commands = append(commands, cmd)
	}
	return commands, nil
}

// --- Schedule Heartbeats ---

func (s *Store) ListScheduleHeartbeats() ([]ScheduleHeartbeat, error) {
	rows, err := s.db.Query(`SELECT id, command, mode, max_gap_m, at_time, grace_m, last_seen_at, alerted FROM schedule_heartbeats ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var beats []ScheduleHeartbeat
	for rows.Next() {
		var b ScheduleHeartbeat
		if err := rows.Scan(&b.ID, &b.Command, &b.Mode, &b.MaxGapM, &b.AtTime, &b.GraceM, &b.LastSeenAt, &b.Alerted); err != nil {
			return nil, err
		}
		beats = append(beats, b)
	}
	return beats, nil
}

func (s *Store) CreateScheduleHeartbeat(b ScheduleHeartbeat) (int, error) {
	var id int
	err := s.db.QueryRow(
		`INSERT INTO schedule_heartbeats (command, mode, max_gap_m, at_time, grace_m) VALUES (?, ?, ?, ?, ?) RETURNING id`,
		b.Command, b.Mode, b.MaxGapM, b.AtTime, b.GraceM,
	).Scan(&id)
	return id, err
}

func (s *Store) UpdateScheduleHeartbeat(b ScheduleHeartbeat) error {
	_, err := s.db.Exec(
		`UPDATE schedule_heartbeats SET command=?, mode=?, max_gap_m=?, at_time=?, grace_m=? WHERE id=?`,
		b.Command, b.Mode, b.MaxGapM, b.AtTime, b.GraceM, b.ID,
	)
	return err
}

func (s *Store) DeleteScheduleHeartbeat(id int) error {
	_, err := s.db.Exec(`DELETE FROM schedule_heartbeats WHERE id=?`, id)
	return err
}

func (s *Store) TouchScheduleHeartbeat(command string, ts float64) error {
	_, err := s.db.Exec(
		`UPDATE schedule_heartbeats SET last_seen_at=?, alerted=false WHERE command=?`,
		ts, command,
	)
	return err
}

func (s *Store) SetHeartbeatAlerted(id int) error {
	_, err := s.db.Exec(`UPDATE schedule_heartbeats SET alerted=true WHERE id=?`, id)
	return err
}
