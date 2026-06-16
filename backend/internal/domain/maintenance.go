package domain

import (
	"context"
	"time"
)

type MaintenanceWindow struct {
	ID          string     `json:"id"`
	ServiceID   string     `json:"service_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	StartsAt    time.Time  `json:"starts_at"`
	EndsAt      time.Time  `json:"ends_at"`
	CreatedAt   time.Time  `json:"created_at"`
	ArchivedAt  *time.Time `json:"archived_at"`
}

type CreateMaintenanceWindowParams struct {
	ServiceID   string
	Title       string
	Description *string
	StartsAt    time.Time
	EndsAt      time.Time
}

type UpdateMaintenanceWindowParams struct {
	Title       *string
	Description *string
	StartsAt    *time.Time
	EndsAt      *time.Time
}

type MaintenanceWindowRepository interface {
	Create(ctx context.Context, p CreateMaintenanceWindowParams) (*MaintenanceWindow, error)
	GetByID(ctx context.Context, id string) (*MaintenanceWindow, error)
	ListByService(ctx context.Context, serviceID string) ([]*MaintenanceWindow, error)
	ListActive(ctx context.Context) ([]*MaintenanceWindow, error)
	ListActiveForService(ctx context.Context, serviceID string) ([]*MaintenanceWindow, error)
	IsActiveForService(ctx context.Context, serviceID string) (bool, error)
	Update(ctx context.Context, id string, p UpdateMaintenanceWindowParams) (*MaintenanceWindow, error)
	Archive(ctx context.Context, id string) error
}
