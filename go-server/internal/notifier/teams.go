package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Teams struct {
	WebhookURL string
}

type teamsConfig struct {
	WebhookURL string `json:"webhook_url"`
}

func NewTeams(configJSON string) (*Teams, error) {
	var cfg teamsConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("invalid teams config: %w", err)
	}
	if cfg.WebhookURL == "" {
		return nil, fmt.Errorf("teams requires webhook_url")
	}
	return &Teams{WebhookURL: cfg.WebhookURL}, nil
}

func (t *Teams) Type() string { return "teams" }

func (t *Teams) Send(ctx context.Context, msg Message) error {
	color := "0078D7" // blue
	switch msg.Level {
	case "warning":
		color = "FFC107"
	case "critical":
		color = "DC3545"
	}

	// Adaptive Card format for Teams webhooks
	card := map[string]any{
		"type":    "message",
		"summary": msg.Title,
		"attachments": []map[string]any{
			{
				"contentType": "application/vnd.microsoft.card.adaptive",
				"content": map[string]any{
					"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
					"type":    "AdaptiveCard",
					"version": "1.4",
					"body": []map[string]any{
						{
							"type":   "TextBlock",
							"text":   msg.Title,
							"weight": "Bolder",
							"size":   "Medium",
							"color":  "Default",
						},
						{
							"type": "TextBlock",
							"text": msg.Body,
							"wrap": true,
						},
					},
					"actions": func() []map[string]any {
						if msg.URL != "" {
							return []map[string]any{
								{"type": "Action.OpenUrl", "title": "View in Wiretap", "url": msg.URL},
							}
						}
						return nil
					}(),
					"msteams": map[string]any{
						"width": "Full",
					},
				},
			},
		},
	}
	_ = color // color used in legacy format, Adaptive Cards use styles instead

	body, _ := json.Marshal(card)

	req, err := http.NewRequestWithContext(ctx, "POST", t.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("teams webhook error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("teams webhook returned %d", resp.StatusCode)
	}
	return nil
}
