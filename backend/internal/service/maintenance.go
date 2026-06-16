package service

import (
	"context"
	"time"

	"github.com/majcek210/monsee/internal/domain"
)

type MaintenanceService struct {
	repo     domain.MaintenanceWindowRepository
	services domain.ServiceRepository
}

func NewMaintenanceService(repo domain.MaintenanceWindowRepository, services domain.ServiceRepository) *MaintenanceService {
	return &MaintenanceService{repo: repo, services: services}
}

func (s *MaintenanceService) Create(ctx context.Context, p domain.CreateMaintenanceWindowParams) (*domain.MaintenanceWindow, error) {
	if p.Title == "" {
		return nil, domain.ValidationErr("title", "title is required")
	}
	if !p.EndsAt.After(p.StartsAt) {
		return nil, domain.ValidationErr("ends_at", "ends_at must be after starts_at")
	}
	if _, err := s.services.GetByID(ctx, p.ServiceID); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, p)
}

func (s *MaintenanceService) GetByID(ctx context.Context, id string) (*domain.MaintenanceWindow, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MaintenanceService) ListByService(ctx context.Context, serviceID string) ([]*domain.MaintenanceWindow, error) {
	return s.repo.ListByService(ctx, serviceID)
}

func (s *MaintenanceService) ListActive(ctx context.Context) ([]*domain.MaintenanceWindow, error) {
	return s.repo.ListActive(ctx)
}

func (s *MaintenanceService) ListActiveForService(ctx context.Context, serviceID string) ([]*domain.MaintenanceWindow, error) {
	return s.repo.ListActiveForService(ctx, serviceID)
}

func (s *MaintenanceService) Update(ctx context.Context, id string, p domain.UpdateMaintenanceWindowParams) (*domain.MaintenanceWindow, error) {
	mw, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	startsAt := mw.StartsAt
	endsAt := mw.EndsAt
	if p.StartsAt != nil {
		startsAt = *p.StartsAt
	}
	if p.EndsAt != nil {
		endsAt = *p.EndsAt
	}
	if !endsAt.After(startsAt) {
		return nil, domain.ValidationErr("ends_at", "ends_at must be after starts_at")
	}
	return s.repo.Update(ctx, id, p)
}

func (s *MaintenanceService) Archive(ctx context.Context, id string) error {
	return s.repo.Archive(ctx, id)
}

func (s *MaintenanceService) IsActiveForService(ctx context.Context, serviceID string) (bool, error) {
	return s.repo.IsActiveForService(ctx, serviceID)
}

func (s *MaintenanceService) ListAll(ctx context.Context) ([]*domain.MaintenanceWindow, error) {
	return s.repo.ListActive(ctx)
}

// Unused field avoidance
var _ = time.Now
