package service

import (
	"context"

	"github.com/majcek210/monsee/internal/domain"
)

type MonitorService struct {
	monitors domain.MonitorRepository
	services domain.ServiceRepository
}

func NewMonitorService(monitors domain.MonitorRepository, services domain.ServiceRepository) *MonitorService {
	return &MonitorService{monitors: monitors, services: services}
}

func (s *MonitorService) Create(ctx context.Context, p domain.CreateMonitorParams) (*domain.Monitor, error) {
	if p.Name == "" {
		return nil, domain.ValidationErr("name", "name is required")
	}
	if p.Type != "http" && p.Type != "tcp" {
		return nil, domain.ValidationErr("type", "type must be http or tcp")
	}

	// Validate service exists
	if _, err := s.services.GetByID(ctx, p.ServiceID); err != nil {
		return nil, err
	}

	// Apply defaults
	if p.IntervalSeconds == 0 {
		p.IntervalSeconds = 60
	}
	if p.TimeoutMs == 0 {
		p.TimeoutMs = 5000
	}
	if p.RetryCount == 0 {
		p.RetryCount = 2
	}

	return s.monitors.Create(ctx, p)
}

func (s *MonitorService) GetByID(ctx context.Context, id string) (*domain.Monitor, error) {
	return s.monitors.GetByID(ctx, id)
}

func (s *MonitorService) ListByService(ctx context.Context, serviceID string) ([]*domain.Monitor, error) {
	if _, err := s.services.GetByID(ctx, serviceID); err != nil {
		return nil, err
	}
	return s.monitors.ListByService(ctx, serviceID)
}

func (s *MonitorService) Update(ctx context.Context, id string, p domain.UpdateMonitorParams) (*domain.Monitor, error) {
	if p.Name == "" {
		return nil, domain.ValidationErr("name", "name is required")
	}
	if _, err := s.monitors.GetByID(ctx, id); err != nil {
		return nil, err
	}
	return s.monitors.Update(ctx, id, p)
}

func (s *MonitorService) Archive(ctx context.Context, id string) error {
	if _, err := s.monitors.GetByID(ctx, id); err != nil {
		return err
	}
	return s.monitors.Archive(ctx, id)
}
