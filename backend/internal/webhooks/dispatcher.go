package webhooks

import (
	"context"
	"log"
	"time"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/pkg/encrypt"
)

// Dispatcher fires outgoing webhook deliveries for platform events.
type Dispatcher struct {
	webhooks domain.WebhookRepository
	encKey   []byte
}

func NewDispatcher(webhooks domain.WebhookRepository, encKey []byte) *Dispatcher {
	return &Dispatcher{webhooks: webhooks, encKey: encKey}
}

// Dispatch sends the event to all matching enabled webhooks.
func (d *Dispatcher) Dispatch(ctx context.Context, event domain.AlertEvent, data map[string]any) {
	hooks, err := d.webhooks.ListByEvent(ctx, event.Event)
	if err != nil {
		log.Printf("webhooks: list by event %s: %v", event.Event, err)
		return
	}

	payload := EventPayload{
		Event:     event.Event,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      data,
	}

	for _, hook := range hooks {
		go func() {
			// Detach from ctx's cancellation: the caller (the asynq task
			// handler) cancels ctx as soon as it returns, but delivery runs
			// in the background after that. Keep ctx's values, drop its
			// deadline, and apply our own.
			deliverCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
			defer cancel()
			if err := d.deliver(deliverCtx, hook, event.Event, payload); err != nil {
				log.Printf("webhooks: deliver to %s: %v", hook.ID, err)
			}
		}()
	}
}

func (d *Dispatcher) deliver(ctx context.Context, hook *domain.Webhook, event string, payload EventPayload) error {
	// Decrypt URL
	url, err := encrypt.Decrypt(d.encKey, hook.URL)
	if err != nil {
		return err
	}

	// Decrypt secret if set
	secret := ""
	if hook.Secret != nil {
		secret, err = encrypt.Decrypt(d.encKey, *hook.Secret)
		if err != nil {
			return err
		}
	}

	result := Deliver(ctx, url, secret, payload, 3)

	// Log delivery
	var statusCode *int32
	if result.StatusCode > 0 {
		sc := int32(result.StatusCode)
		statusCode = &sc
	}
	var errStr *string
	if result.Err != "" {
		errStr = &result.Err
	}
	dur := result.DurationMs

	_, logErr := d.webhooks.InsertLog(ctx, domain.InsertWebhookLogParams{
		WebhookID:  hook.ID,
		Event:      event,
		StatusCode: statusCode,
		Error:      errStr,
		DurationMs: &dur,
	})
	if logErr != nil {
		log.Printf("webhooks: insert log: %v", logErr)
	}
	return nil
}
