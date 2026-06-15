package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/pkg/encrypt"
)

// Dispatcher fires notifications to all enabled channels for a given alert event.
type Dispatcher struct {
	channels    domain.NotificationChannelRepository
	encKey      []byte
	smtpConfig  SMTPConfig
	frontendURL string // public URL of the admin dashboard, used to link back from alerts; "" disables links
}

func NewDispatcher(
	channels domain.NotificationChannelRepository,
	encKey []byte,
	smtp SMTPConfig,
	frontendURL string,
) *Dispatcher {
	return &Dispatcher{channels: channels, encKey: encKey, smtpConfig: smtp, frontendURL: frontendURL}
}

// Dispatch sends an alert to all enabled channels.
// The data map is unused by notifications (it uses the AlertEvent fields directly).
func (d *Dispatcher) Dispatch(ctx context.Context, event domain.AlertEvent, _ map[string]any) {
	channels, err := d.channels.ListEnabled(ctx)
	if err != nil {
		log.Printf("notifications: list channels: %v", err)
		return
	}

	title, body := formatMessage(event)

	link := ""
	if d.frontendURL != "" {
		link = fmt.Sprintf("%s/admin/services/%s", d.frontendURL, event.ServiceID)
	}

	for _, ch := range channels {
		go func() {
			// Detach from ctx's cancellation: the caller (the asynq task
			// handler) cancels ctx as soon as it returns, but delivery runs
			// in the background after that. Keep ctx's values, drop its
			// deadline, and apply our own.
			sendCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
			defer cancel()
			if err := d.send(sendCtx, ch, title, body, link, event.Event == "monitor.down"); err != nil {
				log.Printf("notifications: send %s via %s: %v", event.Event, ch.Type, err)
			}
		}()
	}
}

func (d *Dispatcher) send(ctx context.Context, ch *domain.NotificationChannel, title, body, link string, isDown bool) error {
	// Decrypt config
	raw, err := encrypt.Decrypt(d.encKey, ch.Config)
	if err != nil {
		return fmt.Errorf("decrypt config: %w", err)
	}

	var cfg map[string]string
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	switch ch.Type {
	case "discord":
		url := cfg["webhook_url"]
		if url == "" {
			return fmt.Errorf("discord channel missing webhook_url")
		}
		return SendDiscord(ctx, url, title, body, link, isDown)

	case "email":
		to := cfg["to"]
		if to == "" {
			return fmt.Errorf("email channel missing to")
		}
		if link != "" {
			body += "\n\nView details: " + link
		}
		return SendEmail(ctx, d.smtpConfig, to, title, body)

	default:
		return fmt.Errorf("unknown channel type: %s", ch.Type)
	}
}

func formatMessage(e domain.AlertEvent) (title, body string) {
	switch e.Event {
	case "monitor.down":
		title = fmt.Sprintf("🔴 %s is DOWN", e.MonitorName)
		body = fmt.Sprintf("Monitor \"%s\" (service ID: %s) has failed its health check.", e.MonitorName, e.ServiceID)
	case "monitor.recovered":
		title = fmt.Sprintf("✅ %s has RECOVERED", e.MonitorName)
		body = fmt.Sprintf("Monitor \"%s\" (service ID: %s) is back up.", e.MonitorName, e.ServiceID)
	default:
		title = fmt.Sprintf("Alert: %s", e.Event)
		body = fmt.Sprintf("Monitor: %s", e.MonitorName)
	}
	return
}
