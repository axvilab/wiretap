package storage

import (
	"os"
	"testing"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	f, err := os.CreateTemp("", "wiretap-test-*.duckdb")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()
	os.Remove(path) // DuckDB needs to create the file itself

	t.Cleanup(func() {
		os.Remove(path)
		os.Remove(path + ".wal")
	})

	s, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestGetSettingEmpty(t *testing.T) {
	s := tempStore(t)
	val, err := s.GetSetting("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if val != "" {
		t.Errorf("expected empty, got %q", val)
	}
}

func TestSetAndGetSetting(t *testing.T) {
	s := tempStore(t)
	if err := s.SetSetting("app_url", "http://localhost:8000"); err != nil {
		t.Fatal(err)
	}
	val, err := s.GetSetting("app_url")
	if err != nil {
		t.Fatal(err)
	}
	if val != "http://localhost:8000" {
		t.Errorf("expected http://localhost:8000, got %q", val)
	}
}

func TestSetSettingOverwrite(t *testing.T) {
	s := tempStore(t)
	s.SetSetting("key", "old")
	s.SetSetting("key", "new")

	val, _ := s.GetSetting("key")
	if val != "new" {
		t.Errorf("expected new, got %q", val)
	}
}

func TestGetAllSettings(t *testing.T) {
	s := tempStore(t)
	s.SetSetting("a", "1")
	s.SetSetting("b", "2")
	s.SetSetting("c", "3")

	all, err := s.GetAllSettings()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 settings, got %d", len(all))
	}
	if all["a"] != "1" || all["b"] != "2" || all["c"] != "3" {
		t.Errorf("unexpected values: %v", all)
	}
}

func TestGetAllSettingsEmpty(t *testing.T) {
	s := tempStore(t)
	all, err := s.GetAllSettings()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 0 {
		t.Errorf("expected empty map, got %d items", len(all))
	}
}
