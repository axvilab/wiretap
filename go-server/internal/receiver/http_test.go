package receiver

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"testing"
)

func TestReadBodyPlain(t *testing.T) {
	body := []byte(`{"rid":"r_123","type":"request"}`)
	req, _ := http.NewRequest("POST", "/ingest", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	result, err := readBody(req)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != string(body) {
		t.Errorf("expected %s, got %s", body, result)
	}
}

func TestReadBodyGzip(t *testing.T) {
	original := []byte(`{"rid":"r_456","type":"job","events":[]}`)

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write(original)
	gw.Close()

	req, _ := http.NewRequest("POST", "/ingest", &buf)
	req.Header.Set("Content-Encoding", "gzip")

	result, err := readBody(req)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != string(original) {
		t.Errorf("expected %s, got %s", original, result)
	}
}

func TestReadBodyGzipCompression(t *testing.T) {
	// Test that gzip actually reduces size
	original := []byte(`{"rid":"r_789","type":"request","events":[` +
		`{"type":"query","data":{"sql":"SELECT * FROM users WHERE id = ?","time_ms":1.5}},` +
		`{"type":"query","data":{"sql":"SELECT * FROM users WHERE id = ?","time_ms":2.1}},` +
		`{"type":"query","data":{"sql":"SELECT * FROM users WHERE id = ?","time_ms":0.8}}` +
		`]}`)

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write(original)
	gw.Close()

	compressed := buf.Bytes()
	if len(compressed) >= len(original) {
		t.Logf("warning: compressed (%d) >= original (%d)", len(compressed), len(original))
	}

	req, _ := http.NewRequest("POST", "/ingest", bytes.NewReader(compressed))
	req.Header.Set("Content-Encoding", "gzip")

	result, err := readBody(req)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != string(original) {
		t.Error("decompressed data doesn't match original")
	}
}

func TestReadBodyEmptyGzip(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Close()

	req, _ := http.NewRequest("POST", "/ingest", &buf)
	req.Header.Set("Content-Encoding", "gzip")

	result, err := readBody(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty, got %d bytes", len(result))
	}
}

func TestReadBodyNoBody(t *testing.T) {
	req, _ := http.NewRequest("POST", "/ingest", io.NopCloser(bytes.NewReader(nil)))

	result, err := readBody(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty, got %d bytes", len(result))
	}
}
