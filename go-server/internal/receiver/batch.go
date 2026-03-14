package receiver

import (
	"log"
	"time"

	"github.com/axvi/wiretap/internal/analyzer"
	"github.com/axvi/wiretap/internal/correlator"
	"github.com/axvi/wiretap/internal/hub"
	"github.com/axvi/wiretap/internal/storage"
)

const (
	batchSize    = 100
	flushInterval = 500 * time.Millisecond
)

type BatchWriter struct {
	store      *storage.Store
	correlator *correlator.Correlator
	hub        *hub.Hub
	ch         <-chan *storage.TracePayload
	buf        []*storage.TracePayload
	ticker     *time.Ticker
}

func NewBatchWriter(store *storage.Store, corr *correlator.Correlator, h *hub.Hub, ch <-chan *storage.TracePayload) *BatchWriter {
	return &BatchWriter{
		store:      store,
		correlator: corr,
		hub:        h,
		ch:         ch,
		buf:        make([]*storage.TracePayload, 0, batchSize),
		ticker:     time.NewTicker(flushInterval),
	}
}

func (bw *BatchWriter) Run() {
	for {
		select {
		case p, ok := <-bw.ch:
			if !ok {
				bw.flush()
				return
			}
			bw.buf = append(bw.buf, p)
			if len(bw.buf) >= batchSize {
				bw.flush()
			}
		case <-bw.ticker.C:
			if len(bw.buf) > 0 {
				bw.flush()
			}
		}
	}
}

func (bw *BatchWriter) flush() {
	if len(bw.buf) == 0 {
		return
	}
	if err := bw.store.SaveTraceBatch(bw.buf); err != nil {
		log.Printf("[batch] error saving %d traces: %v", len(bw.buf), err)
		return
	}
	log.Printf("[batch] flushed %d traces", len(bw.buf))

	// Post-save: extract emails, correlate jobs, run analyzer
	for _, p := range bw.buf {
		bw.store.ExtractEmails(p)
		bw.correlator.ProcessTrace(p)

		trace, err := bw.store.GetTrace(p.RID)
		if err != nil {
			continue
		}
		warnings := analyzer.Analyze(trace)
		if len(warnings) > 0 {
			bw.store.UpdateWarnings(p.RID, warnings)
			trace, _ = bw.store.GetTrace(p.RID) // reload with warnings
		}

		// Broadcast to WebSocket clients
		bw.hub.Broadcast("trace.new", trace)
	}

	bw.buf = bw.buf[:0]
}
