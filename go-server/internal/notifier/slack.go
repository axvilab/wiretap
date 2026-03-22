package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Slack struct {
	WebhookURL string
}

type slackConfig struct {
	WebhookURL string `json:"webhook_url"`
}

func NewSlack(configJSON string) (*Slack, error) {
	var cfg slackConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("invalid slack config: %w", err)
	}
	if cfg.WebhookURL == "" {
		return nil, fmt.Errorf("slack requires webhook_url")
	}
	return &Slack{WebhookURL: cfg.WebhookURL}, nil
}

func (s *Slack) Type() string { return "slack" }

func (s *Slack) Send(ctx context.Context, msg Message) error {
	icon := ":information_source:"
	switch msg.Level {
	case "warning":
		icon = ":warning:"
	case "critical":
		icon = ":red_circle:"
	}

	text := fmt.Sprintf("%s *%s*\n%s", icon, msg.Title, msg.Body)
	if msg.URL != "" {
		text += fmt.Sprintf("\n<%s|View in Wiretap>", msg.URL)
	}

	body, _ := json.Marshal(map[string]string{"text": text})

	req, err := http.NewRequestWithContext(ctx, "POST", s.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("slack webhook error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned %d", resp.StatusCode)
	}
	return nil
}
