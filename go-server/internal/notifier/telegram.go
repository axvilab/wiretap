package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Telegram struct {
	Token  string
	ChatID string
}

type telegramConfig struct {
	Token  string `json:"token"`
	ChatID string `json:"chat_id"`
}

func NewTelegram(configJSON string) (*Telegram, error) {
	var cfg telegramConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("invalid telegram config: %w", err)
	}
	if cfg.Token == "" || cfg.ChatID == "" {
		return nil, fmt.Errorf("telegram requires token and chat_id")
	}
	return &Telegram{Token: cfg.Token, ChatID: cfg.ChatID}, nil
}

func (t *Telegram) Type() string { return "telegram" }

func (t *Telegram) Send(ctx context.Context, msg Message) error {
	text := formatText(msg)

	body, _ := json.Marshal(map[string]string{
		"chat_id":    t.ChatID,
		"text":       text,
		"parse_mode": "HTML",
	})

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.Token)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram api error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api returned %d", resp.StatusCode)
	}
	return nil
}

func formatText(msg Message) string {
	icon := "ℹ️"
	switch msg.Level {
	case "warning":
		icon = "⚠️"
	case "critical":
		icon = "🔴"
	}

	text := fmt.Sprintf("%s <b>%s</b>\n%s", icon, msg.Title, msg.Body)
	if msg.URL != "" {
		text += fmt.Sprintf("\n\n<a href=\"%s\">View in Wiretap</a>", msg.URL)
	}
	return text
}
