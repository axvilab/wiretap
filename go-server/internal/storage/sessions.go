package storage

import "time"

func (s *Store) CreateSession(token string, ttl time.Duration) error {
	expiresAt := float64(time.Now().Add(ttl).Unix())
	_, err := s.db.Exec("INSERT INTO sessions (token, expires_at) VALUES (?, ?)", token, expiresAt)
	return err
}

func (s *Store) ValidateSession(token string) bool {
	var expiresAt float64
	err := s.db.QueryRow("SELECT expires_at FROM sessions WHERE token = ?", token).Scan(&expiresAt)
	if err != nil {
		return false
	}
	if float64(time.Now().Unix()) > expiresAt {
		s.DeleteSession(token)
		return false
	}
	return true
}

func (s *Store) DeleteSession(token string) {
	s.db.Exec("DELETE FROM sessions WHERE token = ?", token)
}

func (s *Store) CleanExpiredSessions() {
	now := float64(time.Now().Unix())
	s.db.Exec("DELETE FROM sessions WHERE expires_at < ?", now)
}
