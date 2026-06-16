package domain

import (
	"context"
	"time"
)

type Service struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	Description         *string    `json:"description"`
	Status              string     `json:"status"`
	PublicVisible       bool       `json:"public_visible"`
	ShowUptime          bool       `json:"show_uptime"`
	DedicatedPageEnabled bool      `json:"dedicated_page_enabled"`
	Slug                *string    `json:"slug"`
	CustomDomain        *string    `json:"custom_domain"`
	UptimeRangeDays     int32      `json:"uptime_range_days"`
	StatusOverride      *string    `json:"status_override"`
	CreatedAt           time.Time  `json:"created_at"`
	ArchivedAt          *time.Time `json:"archived_at"`
}

func (s *Service) EffectiveStatus() string {
	if s.StatusOverride != nil && *s.StatusOverride != "" {
		return *s.StatusOverride
	}
	return s.Status
}

type CreateServiceParams struct {
	Name        string
	Description *string
}

type UpdateServiceParams struct {
	Name                 string
	Description          *string
	Status               *string
	PublicVisible        *bool
	ShowUptime           *bool
	DedicatedPageEnabled *bool
	Slug                 *string
	CustomDomain         *string
	UptimeRangeDays      *int32
	StatusOverride       *string
}

type ServiceRepository interface {
	Create(ctx context.Context, p CreateServiceParams) (*Service, error)
	GetByID(ctx context.Context, id string) (*Service, error)
	GetBySlug(ctx context.Context, slug string) (*Service, error)
	GetByCustomDomain(ctx context.Context, domain string) (*Service, error)
	List(ctx context.Context) ([]*Service, error)
	Update(ctx context.Context, id string, p UpdateServiceParams) (*Service, error)
	UpdateStatus(ctx context.Context, id, status string) error
	Archive(ctx context.Context, id string) error
}
