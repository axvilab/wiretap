package notifier

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/axvi/wiretap/internal/storage"
)

// Message is the payload sent to notification channels.
type Message struct {
	Title string
	Body  string
	Level string // info, warning, critical
	URL   string // link to Wiretap page
}

// Channel sends a notification message.
type Channel interface {
	Send(ctx context.Context, msg Message) error
	Type() string
}

// Manager loads channels from DB and dispatches alerts based on rules.
type Manager struct {
	store    *storage.Store
	mu       sync.RWMutex
	channels map[int]Channel
}

// NewManager creates a notification manager and loads channels from DB.
func NewManager(store *storage.Store) *Manager {
	m := &Manager{store: store, channels: make(map[int]Channel)}
	m.Reload()
	return m
}

// Reload loads notification channels from database.
func (m *Manager) Reload() {
	channels, err := m.store.ListNotificationChannels()
	if err != nil {
		log.Printf("[notify] failed to load channels: %v", err)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels = make(map[int]Channel)
	for _, ch := range channels {
		if !ch.Enabled {
			continue
		}
		c, err := buildChannel(ch)
		if err != nil {
			log.Printf("[notify] skip channel %d (%s): %v", ch.ID, ch.Type, err)
			continue
		}
		m.channels[ch.ID] = c
	}
}

// Notify checks alert rules and sends notifications for the given rule type.
func (m *Manager) Notify(ruleType string, msg Message) {
	rules, err := m.store.ListAlertRules()
	if err != nil {
		log.Printf("[notify] failed to load rules: %v", err)
		return
	}

	for _, rule := range rules {
		if !rule.Enabled || rule.Type != ruleType {
			continue
		}

		// Check cooldown
		lastSent, _ := m.store.GetLastAlertTime(rule.ID)
		if lastSent > 0 {
			cooldown := time.Duration(rule.CooldownM) * time.Minute
			if time.Since(time.Unix(int64(lastSent), 0)) < cooldown {
				continue
			}
		}

		m.mu.RLock()
		ch, ok := m.channels[rule.ChannelID]
		m.mu.RUnlock()
		if !ok {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := ch.Send(ctx, msg); err != nil {
			log.Printf("[notify] failed to send to channel %d: %v", rule.ChannelID, err)
			cancel()
			continue
		}
		cancel()

		// Log the alert
		logMsg := fmt.Sprintf("[%s] %s: %s", ruleType, msg.Title, msg.Body)
		m.store.SaveAlertLog(rule.ID, logMsg)
	}
}

// SendTest sends a test message to a specific channel by ID.
func (m *Manager) SendTest(channelID int) error {
	ch, err := m.store.GetNotificationChannel(channelID)
	if err != nil {
		return fmt.Errorf("channel not found: %w", err)
	}
	c, err := buildChannel(*ch)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return c.Send(ctx, Message{
		Title: "Wiretap Test",
		Body:  "This is a test notification from Wiretap.",
		Level: "info",
	})
}

func buildChannel(ch storage.NotificationChannel) (Channel, error) {
	switch ch.Type {
	case "telegram":
		return NewTelegram(ch.Config)
	case "slack":
		return NewSlack(ch.Config)
	case "teams":
		return NewTeams(ch.Config)
	default:
		return nil, fmt.Errorf("unknown channel type: %s", ch.Type)
	}
}
