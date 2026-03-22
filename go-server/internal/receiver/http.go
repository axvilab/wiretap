package receiver

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/axvi/wiretap/internal/storage"
)

type HTTPReceiver struct {
	ingestCh chan<- *storage.TracePayload
}

func NewHTTP(ch chan<- *storage.TracePayload) *HTTPReceiver {
	return &HTTPReceiver{ingestCh: ch}
}

func (r *HTTPReceiver) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := readBody(req)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		if len(body) >= maxIngestBody {
			log.Printf("[ingest] payload too large (%d bytes), rejected", len(body))
			http.Error(w, "payload too large", http.StatusRequestEntityTooLarge)
			return
		}

		var payload storage.TracePayload
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if payload.RID == "" {
			http.Error(w, "rid is required", http.StatusBadRequest)
			return
		}

		log.Printf("[ingest] %s %s %s (%d events)", payload.Type, payload.RID, metaURI(&payload), len(payload.Events))

		r.ingestCh <- &payload

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"status":"accepted"}`))
	}
}

const maxIngestBody = 50 * 1024 * 1024 // 50MB

// readBody handles both plain and gzip-compressed request bodies
func readBody(req *http.Request) ([]byte, error) {
	if req.Header.Get("Content-Encoding") == "gzip" {
		gr, err := gzip.NewReader(req.Body)
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		return io.ReadAll(io.LimitReader(gr, maxIngestBody))
	}
	return io.ReadAll(io.LimitReader(req.Body, maxIngestBody))
}

func metaURI(p *storage.TracePayload) string {
	var meta struct {
		Method string `json:"method"`
		URI    string `json:"uri"`
	}
	json.Unmarshal(p.Meta, &meta)
	if meta.Method != "" {
		return meta.Method + " " + meta.URI
	}
	return ""
}