package service

import (
	"context"
	"errors"
	"regexp"

	"github.com/majcek210/monsee/internal/domain"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

type MonitoringService struct {
	services domain.ServiceRepository
}

func NewMonitoringService(services domain.ServiceRepository) *MonitoringService {
	return &MonitoringService{services: services}
}

func (s *MonitoringService) Create(ctx context.Context, p domain.CreateServiceParams) (*domain.Service, error) {
	if p.Name == "" {
		return nil, domain.ValidationErr("name", "name is required")
	}
	return s.services.Create(ctx, p)
}

func (s *MonitoringService) GetByID(ctx context.Context, id string) (*domain.Service, error) {
	return s.services.GetByID(ctx, id)
}

func (s *MonitoringService) List(ctx context.Context) ([]*domain.Service, error) {
	return s.services.List(ctx)
}

func (s *MonitoringService) Update(ctx context.Context, id string, p domain.UpdateServiceParams) (*domain.Service, error) {
	if p.Name == "" {
		return nil, domain.ValidationErr("name", "name is required")
	}
	if p.Slug != nil && *p.Slug != "" && !slugPattern.MatchString(*p.Slug) {
		return nil, domain.ValidationErr("slug", "slug must be lowercase letters, numbers, and hyphens only")
	}
	if p.UptimeRangeDays != nil && (*p.UptimeRangeDays < 1 || *p.UptimeRangeDays > 365) {
		return nil, domain.ValidationErr("uptime_range_days", "uptime_range_days must be between 1 and 365")
	}
	if p.StatusOverride != nil && *p.StatusOverride != "" {
		valid := map[string]bool{"operational": true, "degraded": true, "outage": true, "maintenance": true}
		if !valid[*p.StatusOverride] {
			return nil, domain.ValidationErr("status_override", "status_override must be operational, degraded, outage, or maintenance")
		}
	}
	existing, err := s.services.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	dedicatedEnabled := existing.DedicatedPageEnabled
	if p.DedicatedPageEnabled != nil {
		dedicatedEnabled = *p.DedicatedPageEnabled
	}
	slug := existing.Slug
	if p.Slug != nil {
		if *p.Slug == "" {
			slug = nil
		} else {
			slug = p.Slug
		}
	}
	if dedicatedEnabled && (slug == nil || *slug == "") {
		return nil, domain.ValidationErr("slug", "slug is required when dedicated page is enabled")
	}
	return s.services.Update(ctx, id, p)
}

func (s *MonitoringService) GetBySlug(ctx context.Context, slug string) (*domain.Service, error) {
	return s.services.GetBySlug(ctx, slug)
}

func (s *MonitoringService) GetByCustomDomain(ctx context.Context, domain string) (*domain.Service, error) {
	return s.services.GetByCustomDomain(ctx, domain)
}

func (s *MonitoringService) Archive(ctx context.Context, id string) error {
	_, err := s.services.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.services.Archive(ctx, id)
}

func (s *MonitoringService) UpdateStatus(ctx context.Context, id, status string) error {
	valid := map[string]bool{"operational": true, "degraded": true, "down": true}
	if !valid[status] {
		return domain.ValidationErr("status", "status must be operational, degraded, or down")
	}
	return s.services.UpdateStatus(ctx, id, status)
}

// IsNotFound checks whether an error is a domain.ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, domain.ErrNotFound)
}
