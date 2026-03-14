package storage

import "encoding/json"

// TracePayload is what the PHP SDK sends to /ingest.
type TracePayload struct {
	Version  int             `json:"v"`
	PID      int             `json:"pid"`
	RID      string          `json:"rid"`
	TsStart  float64         `json:"ts_start"`
	TsEnd    float64         `json:"ts_end"`
	Type     string          `json:"type"`
	Meta     json.RawMessage `json:"meta"`
	Events   []Event         `json:"events"`
}

// Event is an individual event within a trace (query, cache, job, etc.).
type Event struct {
	Type string          `json:"type"`
	Ts   float64         `json:"ts"`
	Data json.RawMessage `json:"data"`
}

// Trace is what we store in DuckDB.
type Trace struct {
	RID              string  `json:"rid"`
	ParentRID        string  `json:"parent_rid"`
	Type             string  `json:"type"`
	Method           string  `json:"method"`
	URI              string  `json:"uri"`
	Status           int     `json:"status"`
	DurationMs       float64 `json:"duration_ms"`
	TsStart          float64 `json:"ts_start"`
	TsEnd            float64 `json:"ts_end"`
	QueryCount       int     `json:"query_count"`
	EventCount       int     `json:"event_count"`
	JobCount         int     `json:"job_count"`
	HasException     bool    `json:"has_exception"`
	HasNPlusOne      bool    `json:"has_n_plus_one"`
	HasSlowQuery     bool    `json:"has_slow_query"`
	HasDuplicateQuery bool   `json:"has_duplicate_query"`
	MetaJSON         string  `json:"meta_json"`
	EventsJSON       string  `json:"events_json"`
	WarningsJSON     string  `json:"warnings_json"`
}

// Email is an intercepted email from the SMTP server.
type Email struct {
	ID          string            `json:"id"`
	From        string            `json:"from"`
	To          []string          `json:"to"`
	Cc          []string          `json:"cc"`
	Subject     string            `json:"subject"`
	HTML        string            `json:"html"`
	Text        string            `json:"text"`
	Attachments []EmailAttachment `json:"attachments"`
	TraceRID    string            `json:"trace_rid"`
	CreatedAt   float64           `json:"created_at"`
}

// EmailAttachment metadata (content not stored in DB).
type EmailAttachment struct {
	Index         int    `json:"index"`
	Name          string `json:"name"`
	ContentType   string `json:"content_type"`
	Size          int    `json:"size"`
	ContentBase64 string `json:"content_base64,omitempty"`
	Content       []byte `json:"-"`
}

// JobLink maps a dispatched job to its execution trace.
type JobLink struct {
	JobUUID    string  `json:"job_uuid"`
	JobClass   string  `json:"job_class"`
	Queue      string  `json:"queue"`
	ParentRID  string  `json:"parent_rid"`
	ChildRID   string  `json:"child_rid"`
	Status     string  `json:"status"`
	DispatchedAt float64 `json:"dispatched_at"`
	StartedAt  float64 `json:"started_at"`
	FinishedAt float64 `json:"finished_at"`
}
