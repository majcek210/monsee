package domain

import (
	"context"
	"time"
)

type NotificationChannel struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	Config     string     `json:"-"`
	Enabled    bool       `json:"enabled"`
	CreatedAt  time.Time  `json:"created_at"`
	ArchivedAt *time.Time `json:"archived_at"`
}

type CreateNotificationChannelParams struct {
	Name   string
	Type   string
	Config string // already encrypted
}

// UpdateNotificationChannelParams is a partial update: nil fields leave the
// existing column value unchanged (COALESCE in the SQL layer).
type UpdateNotificationChannelParams struct {
	Name    *string
	Config  *string // already encrypted
	Enabled *bool
}

// AlertEvent is the payload passed to dispatchers when a monitor state changes.
type AlertEvent struct {
	Event       string // "monitor.down" | "monitor.recovered"
	MonitorID   string
	MonitorName string
	ServiceID   string
}

type NotificationChannelRepository interface {
	Create(ctx context.Context, p CreateNotificationChannelParams) (*NotificationChannel, error)
	GetByID(ctx context.Context, id string) (*NotificationChannel, error)
	List(ctx context.Context) ([]*NotificationChannel, error)
	ListEnabled(ctx context.Context) ([]*NotificationChannel, error)
	Update(ctx context.Context, id string, p UpdateNotificationChannelParams) (*NotificationChannel, error)
	Archive(ctx context.Context, id string) error
}
