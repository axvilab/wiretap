package receiver

import (
	"encoding/json"
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

		var payload storage.TracePayload
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
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
