package domain

import (
	"context"
	"time"
)

type Settings struct {
	ID                   int       `json:"id"`
	SiteTitle            string    `json:"site_title"`
	LogoURL              string    `json:"logo_url"`
	PublicStatusEnabled  bool      `json:"public_status_enabled"`
	CustomDomainsEnabled bool      `json:"custom_domains_enabled"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type UpdateSettingsParams struct {
	SiteTitle            *string
	LogoURL              *string
	PublicStatusEnabled  *bool
	CustomDomainsEnabled *bool
}

type SettingsRepository interface {
	Get(ctx context.Context) (*Settings, error)
	Update(ctx context.Context, p UpdateSettingsParams) (*Settings, error)
}
