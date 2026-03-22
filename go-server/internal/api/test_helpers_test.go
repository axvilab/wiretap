package api

import (
	"os"
	"testing"

	"github.com/axvi/wiretap/internal/storage"
)

func testStore(t *testing.T) *storage.Store {
	t.Helper()
	f, err := os.CreateTemp("", "wiretap-api-test-*.duckdb")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()
	os.Remove(path)

	t.Cleanup(func() {
		os.Remove(path)
		os.Remove(path + ".wal")
	})

	s, err := storage.New(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}