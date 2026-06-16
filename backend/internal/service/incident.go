package service

import (
	"context"
	"time"

	"github.com/majcek210/monsee/internal/domain"
)

type IncidentService struct {
	incidents domain.IncidentRepository
	services  domain.ServiceRepository
}

func NewIncidentService(incidents domain.IncidentRepository, services domain.ServiceRepository) *IncidentService {
	return &IncidentService{incidents: incidents, services: services}
}

func (s *IncidentService) Create(ctx context.Context, p domain.CreateIncidentParams) (*domain.Incident, error) {
	if p.Title == "" {
		return nil, domain.ValidationErr("title", "title is required")
	}
	if _, err := s.services.GetByID(ctx, p.ServiceID); err != nil {
		return nil, err
	}
	return s.incidents.Create(ctx, p)
}

func (s *IncidentService) GetByID(ctx context.Context, id string) (*domain.Incident, error) {
	return s.incidents.GetByID(ctx, id)
}

func (s *IncidentService) List(ctx context.Context, serviceID string) ([]*domain.Incident, error) {
	if serviceID != "" {
		if _, err := s.services.GetByID(ctx, serviceID); err != nil {
			return nil, err
		}
		return s.incidents.List(ctx, serviceID)
	}
	return s.incidents.ListAll(ctx)
}

func (s *IncidentService) Resolve(ctx context.Context, id string) (*domain.Incident, error) {
	inc, err := s.incidents.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inc.Status == "resolved" {
		return inc, nil
	}
	return s.incidents.Resolve(ctx, id, time.Now())
}

func (s *IncidentService) Update(ctx context.Context, id string, p domain.UpdateIncidentParams) (*domain.Incident, error) {
	if p.Title == "" {
		return nil, domain.ValidationErr("title", "title is required")
	}
	if _, err := s.incidents.GetByID(ctx, id); err != nil {
		return nil, err
	}
	return s.incidents.Update(ctx, id, p)
}

var validUpdateStatuses = map[string]bool{
	"investigating": true,
	"identified":    true,
	"monitoring":    true,
	"resolved":      true,
}

func (s *IncidentService) PostUpdate(ctx context.Context, incidentID, status, message string) (*domain.IncidentUpdate, error) {
	if !validUpdateStatuses[status] {
		return nil, domain.ValidationErr("status", "status must be investigating, identified, monitoring, or resolved")
	}
	if message == "" {
		return nil, domain.ValidationErr("message", "message is required")
	}
	if _, err := s.incidents.GetByID(ctx, incidentID); err != nil {
		return nil, err
	}
	update, err := s.incidents.CreateUpdate(ctx, domain.CreateIncidentUpdateParams{
		IncidentID: incidentID,
		Status:     status,
		Message:    message,
	})
	if err != nil {
		return nil, err
	}
	if status == "resolved" {
		if _, err := s.incidents.Resolve(ctx, incidentID, time.Now()); err != nil {
			// Non-fatal if already resolved
			_ = err
		}
	}
	return update, nil
}

func (s *IncidentService) ListUpdates(ctx context.Context, incidentID string) ([]*domain.IncidentUpdate, error) {
	return s.incidents.ListUpdates(ctx, incidentID)
}
