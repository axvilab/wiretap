package storage

import "time"

// --- Notification Channels ---

func (s *Store) ListNotificationChannels() ([]NotificationChannel, error) {
	rows, err := s.db.Query(`SELECT id, type, name, config, enabled FROM notification_channels ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []NotificationChannel
	for rows.Next() {
		var ch NotificationChannel
		if err := rows.Scan(&ch.ID, &ch.Type, &ch.Name, &ch.Config, &ch.Enabled); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, nil
}

func (s *Store) GetNotificationChannel(id int) (*NotificationChannel, error) {
	var ch NotificationChannel
	err := s.db.QueryRow(
		`SELECT id, type, name, config, enabled FROM notification_channels WHERE id=?`, id,
	).Scan(&ch.ID, &ch.Type, &ch.Name, &ch.Config, &ch.Enabled)
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func (s *Store) CreateNotificationChannel(ch NotificationChannel) (int, error) {
	var id int
	err := s.db.QueryRow(
		`INSERT INTO notification_channels (type, name, config, enabled) VALUES (?, ?, ?, ?) RETURNING id`,
		ch.Type, ch.Name, ch.Config, ch.Enabled,
	).Scan(&id)
	return id, err
}

func (s *Store) UpdateNotificationChannel(ch NotificationChannel) error {
	_, err := s.db.Exec(
		`UPDATE notification_channels SET type=?, name=?, config=?, enabled=? WHERE id=?`,
		ch.Type, ch.Name, ch.Config, ch.Enabled, ch.ID,
	)
	return err
}

func (s *Store) DeleteNotificationChannel(id int) error {
	_, err := s.db.Exec(`DELETE FROM notification_channels WHERE id=?`, id)
	if err != nil {
		return err
	}
	// Cleanup rules referencing this channel
	s.db.Exec(`DELETE FROM alert_rules WHERE channel_id=?`, id)
	return nil
}

// --- Alert Rules ---

func (s *Store) ListAlertRules() ([]AlertRule, error) {
	rows, err := s.db.Query(`SELECT id, type, channel_id, config, enabled, cooldown_m FROM alert_rules ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []AlertRule
	for rows.Next() {
		var r AlertRule
		if err := rows.Scan(&r.ID, &r.Type, &r.ChannelID, &r.Config, &r.Enabled, &r.CooldownM); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, nil
}

func (s *Store) CreateAlertRule(r AlertRule) (int, error) {
	var id int
	err := s.db.QueryRow(
		`INSERT INTO alert_rules (type, channel_id, config, enabled, cooldown_m) VALUES (?, ?, ?, ?, ?) RETURNING id`,
		r.Type, r.ChannelID, r.Config, r.Enabled, r.CooldownM,
	).Scan(&id)
	return id, err
}

func (s *Store) UpdateAlertRule(r AlertRule) error {
	_, err := s.db.Exec(
		`UPDATE alert_rules SET type=?, channel_id=?, config=?, enabled=?, cooldown_m=? WHERE id=?`,
		r.Type, r.ChannelID, r.Config, r.Enabled, r.CooldownM, r.ID,
	)
	return err
}

func (s *Store) DeleteAlertRule(id int) error {
	_, err := s.db.Exec(`DELETE FROM alert_rules WHERE id=?`, id)
	return err
}

// --- Alert Log ---

func (s *Store) SaveAlertLog(ruleID int, message string) error {
	now := float64(time.Now().Unix())
	_, err := s.db.Exec(
		`INSERT INTO alert_log (rule_id, message, sent_at) VALUES (?, ?, ?)`,
		ruleID, message, now,
	)
	return err
}

func (s *Store) GetLastAlertTime(ruleID int) (float64, error) {
	var sentAt float64
	err := s.db.QueryRow(
		`SELECT sent_at FROM alert_log WHERE rule_id=? ORDER BY sent_at DESC LIMIT 1`, ruleID,
	).Scan(&sentAt)
	if err != nil {
		return 0, nil // no previous alert
	}
	return sentAt, nil
}

func (s *Store) ListAlertLogs(limit int) ([]AlertLog, error) {
	rows, err := s.db.Query(`SELECT id, rule_id, message, sent_at FROM alert_log ORDER BY sent_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AlertLog
	for rows.Next() {
		var l AlertLog
		if err := rows.Scan(&l.ID, &l.RuleID, &l.Message, &l.SentAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
