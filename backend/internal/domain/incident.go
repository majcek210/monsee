package domain

import (
	"context"
	"time"
)

type Incident struct {
	ID         string     `json:"id"`
	ServiceID  string     `json:"service_id"`
	MonitorID  *string    `json:"monitor_id"`
	Title      string     `json:"title"`
	Severity   string     `json:"severity"`
	Status     string     `json:"status"`
	ResolvedAt *time.Time `json:"resolved_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type CreateIncidentParams struct {
	ServiceID string
	MonitorID *string
	Title     string
	Severity  string
}

type UpdateIncidentParams struct {
	Title    string
	Severity string
	Status   string
}

type IncidentRepository interface {
	Create(ctx context.Context, p CreateIncidentParams) (*Incident, error)
	GetByID(ctx context.Context, id string) (*Incident, error)
	GetOpenForMonitor(ctx context.Context, monitorID string) (*Incident, error)
	List(ctx context.Context, serviceID string) ([]*Incident, error)
	ListAll(ctx context.Context) ([]*Incident, error)
	Resolve(ctx context.Context, id string, resolvedAt time.Time) (*Incident, error)
	Update(ctx context.Context, id string, p UpdateIncidentParams) (*Incident, error)
}
