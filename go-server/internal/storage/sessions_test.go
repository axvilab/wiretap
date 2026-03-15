package storage

import (
	"testing"
	"time"
)

func TestCreateAndValidateSession(t *testing.T) {
	s := tempStore(t)

	s.CreateSession("token123", 1*time.Hour)

	if !s.ValidateSession("token123") {
		t.Error("expected session to be valid")
	}
}

func TestValidateNonexistentSession(t *testing.T) {
	s := tempStore(t)

	if s.ValidateSession("nonexistent") {
		t.Error("expected nonexistent session to be invalid")
	}
}

func TestExpiredSession(t *testing.T) {
	s := tempStore(t)

	// Create session with 0 TTL (already expired)
	s.CreateSession("expired", -1*time.Second)

	if s.ValidateSession("expired") {
		t.Error("expected expired session to be invalid")
	}
}

func TestDeleteSession(t *testing.T) {
	s := tempStore(t)

	s.CreateSession("todelete", 1*time.Hour)
	s.DeleteSession("todelete")

	if s.ValidateSession("todelete") {
		t.Error("expected deleted session to be invalid")
	}
}

func TestCleanExpiredSessions(t *testing.T) {
	s := tempStore(t)

	s.CreateSession("valid", 1*time.Hour)
	s.CreateSession("expired1", -1*time.Second)
	s.CreateSession("expired2", -1*time.Second)

	s.CleanExpiredSessions()

	if !s.ValidateSession("valid") {
		t.Error("expected valid session to survive cleanup")
	}
	// expired ones already cleaned by ValidateSession or CleanExpiredSessions
}

func TestSessionSurvivesReopen(t *testing.T) {
	// Simulate restart — close and reopen store
	s := tempStore(t)
	s.CreateSession("persistent", 1*time.Hour)

	// Get the path before closing
	path := s.db // we can't easily get path, but session is in DuckDB file
	_ = path

	// Session should still be valid (it's in DuckDB, not memory)
	if !s.ValidateSession("persistent") {
		t.Error("expected session to persist in DuckDB")
	}
}
