package domain

import (
	"context"
	"time"
)

type Webhook struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	URL        string     `json:"-"`
	Secret     *string    `json:"-"`
	Events     []string   `json:"events"`
	Enabled    bool       `json:"enabled"`
	CreatedAt  time.Time  `json:"created_at"`
	ArchivedAt *time.Time `json:"archived_at"`
}

type WebhookLog struct {
	ID          string    `json:"id"`
	WebhookID   string    `json:"webhook_id"`
	Event       string    `json:"event"`
	StatusCode  *int32    `json:"status_code"`
	Error       *string   `json:"error"`
	DurationMs  *int32    `json:"duration_ms"`
	DeliveredAt time.Time `json:"delivered_at"`
}

type CreateWebhookParams struct {
	Name   string
	URL    string // already encrypted
	Secret *string
	Events []string
}

// UpdateWebhookParams is a partial update: nil fields leave the existing
// column value unchanged (COALESCE in the SQL layer).
type UpdateWebhookParams struct {
	Name    *string
	URL     *string // already encrypted
	Secret  *string // already encrypted
	Events  []string
	Enabled *bool
}

type InsertWebhookLogParams struct {
	WebhookID  string
	Event      string
	StatusCode *int32
	Error      *string
	DurationMs *int32
}

type WebhookRepository interface {
	Create(ctx context.Context, p CreateWebhookParams) (*Webhook, error)
	GetByID(ctx context.Context, id string) (*Webhook, error)
	List(ctx context.Context) ([]*Webhook, error)
	ListByEvent(ctx context.Context, event string) ([]*Webhook, error)
	Update(ctx context.Context, id string, p UpdateWebhookParams) (*Webhook, error)
	Archive(ctx context.Context, id string) error
	InsertLog(ctx context.Context, p InsertWebhookLogParams) (*WebhookLog, error)
	ListLogs(ctx context.Context, webhookID string, limit int32) ([]*WebhookLog, error)
}
