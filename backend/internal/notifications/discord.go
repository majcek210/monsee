package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/majcek210/monsee/pkg/netguard"
)

type discordPayload struct {
	Embeds []discordEmbed `json:"embeds"`
}

type discordEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       int    `json:"color"`
	Timestamp   string `json:"timestamp"`
}

// SendDiscord posts an embed to a Discord webhook URL.
func SendDiscord(ctx context.Context, webhookURL, title, description string, isDown bool) error {
	// SSRF guard: re-checked at send time (not just at create/update) so DNS
	// rebinding after the channel was configured can't reach internal hosts.
	if err := netguard.CheckPublicURL(webhookURL); err != nil {
		return fmt.Errorf("blocked target: %w", err)
	}

	color := 0x57F287 // green — recovered
	if isDown {
		color = 0xED4245 // red — down
	}

	payload := discordPayload{
		Embeds: []discordEmbed{{
			Title:       title,
			Description: description,
			Color:       color,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook returned %d", resp.StatusCode)
	}
	return nil
}
