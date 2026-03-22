package monitor

import (
	"context"
	"fmt"
	"log"
	"time"

	"strconv"
	"strings"

	"github.com/axvi/wiretap/internal/notifier"
	"github.com/axvi/wiretap/internal/storage"
)

// ScheduleMonitor checks if expected scheduled tasks have run.
type ScheduleMonitor struct {
	store  *storage.Store
	notify *notifier.Manager
}

func (m *ScheduleMonitor) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkAll()
		}
	}
}

func (m *ScheduleMonitor) checkAll() {
	heartbeats, err := m.store.ListScheduleHeartbeats()
	if err != nil {
		log.Printf("[monitor:schedule] failed to list heartbeats: %v", err)
		return
	}

	now := time.Now()
	for _, hb := range heartbeats {
		if hb.Alerted {
			continue
		}
		if hb.LastSeenAt == 0 {
			// Never seen — don't alert yet, wait for first run
			continue
		}

		shouldAlert := false
		var body string

		if hb.Mode == "at" && hb.AtTime != "" {
			// "at" mode: expect task to run daily at HH:MM + grace
			shouldAlert, body = m.checkAtMode(now, hb)
		} else {
			// "every" mode: expect task within max_gap_m
			lastSeen := time.Unix(int64(hb.LastSeenAt), 0)
			gap := now.Sub(lastSeen)
			maxGap := time.Duration(hb.MaxGapM) * time.Minute
			if gap > maxGap {
				shouldAlert = true
				body = fmt.Sprintf("Task '%s' has not run for %s (expected every %dm)", hb.Command, gap.Round(time.Minute), hb.MaxGapM)
			}
		}

		if shouldAlert {
			m.notify.Notify("schedule_missed", notifier.Message{
				Title: fmt.Sprintf("Scheduled Task Missed: %s", hb.Command),
				Body:  body,
				Level: "critical",
			})
			m.store.SetHeartbeatAlerted(hb.ID)
		}
	}
}

// checkAtMode checks if a daily task ran today at/after the expected time.
func (m *ScheduleMonitor) checkAtMode(now time.Time, hb storage.ScheduleHeartbeat) (bool, string) {
	// Parse "HH:MM"
	parts := strings.SplitN(hb.AtTime, ":", 2)
	if len(parts) != 2 {
		return false, ""
	}
	hour, _ := strconv.Atoi(parts[0])
	min, _ := strconv.Atoi(parts[1])

	// Expected run time today
	expected := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, now.Location())
	deadline := expected.Add(time.Duration(hb.GraceM) * time.Minute)

	// Only alert if we're past the deadline
	if now.Before(deadline) {
		return false, ""
	}

	// Check if task ran today after expected time
	lastSeen := time.Unix(int64(hb.LastSeenAt), 0)
	if lastSeen.After(expected) {
		return false, "" // ran today, all good
	}

	return true, fmt.Sprintf("Task '%s' expected at %s, not seen today (grace: %dm)", hb.Command, hb.AtTime, hb.GraceM)
}
