package correlator

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/axvi/wiretap/internal/storage"
)

type Correlator struct {
	store *storage.Store
}

func New(store *storage.Store) *Correlator {
	return &Correlator{store: store}
}

// ProcessTrace links job traces to their parent request traces.
// Called after a trace is saved. If this trace is a job execution,
// find the job_link by job_uuid and set parent_rid + child_rid.
func (c *Correlator) ProcessTrace(p *storage.TracePayload) {
	if p.Type != "job" {
		return
	}

	// Find job_uuid from meta
	var meta struct {
		JobUUID string `json:"job_uuid"`
	}
	if err := json.Unmarshal(p.Meta, &meta); err != nil || meta.JobUUID == "" {
		return
	}

	// Update job_link: set child_rid and status
	err := c.store.UpdateJobLink(meta.JobUUID, p.RID)
	if err != nil {
		log.Printf("[correlator] failed to link job %s → %s: %v", meta.JobUUID, p.RID, err)
		return
	}

	// Set parent_rid on this trace
	parentRID, err := c.store.GetJobParentRID(meta.JobUUID)
	if err != nil {
		log.Printf("[correlator] failed to get parent for job %s: %v", meta.JobUUID, err)
		return
	}
	if parentRID != "" {
		if err := c.store.SetTraceParent(p.RID, parentRID); err != nil {
			log.Printf("[correlator] failed to set parent %s on trace %s: %v", parentRID, p.RID, err)
		}
	}

	log.Printf("[correlator] linked job %s: %s → %s", meta.JobUUID, parentRID, p.RID)
}

// TraceNode is a recursive trace tree for the API response.
type TraceNode struct {
	Trace    *storage.Trace `json:"trace"`
	Children []*TraceNode   `json:"children"`
}

// BuildTree returns a recursive tree starting from the given rid.
func (c *Correlator) BuildTree(rid string) (*TraceNode, error) {
	trace, err := c.store.GetTrace(rid)
	if err != nil {
		return nil, fmt.Errorf("get trace %s: %w", rid, err)
	}

	node := &TraceNode{
		Trace:    trace,
		Children: []*TraceNode{},
	}

	children, err := c.store.GetChildTraces(rid)
	if err != nil {
		return node, nil // return node without children on error
	}

	for _, child := range children {
		childNode, err := c.BuildTree(child.RID)
		if err != nil {
			continue
		}
		node.Children = append(node.Children, childNode)
	}

	return node, nil
}
