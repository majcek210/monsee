package service

import (
	"context"
	"errors"

	"github.com/majcek210/monsee/internal/domain"
)

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
	_, err := s.services.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.services.Update(ctx, id, p)
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
