package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majcek210/monsee/internal/domain"
	sqlcdb "github.com/majcek210/monsee/db/sqlc"
)

type IncidentRepo struct {
	q *sqlcdb.Queries
}

func NewIncidentRepo(pool *pgxpool.Pool) *IncidentRepo {
	return &IncidentRepo{q: sqlcdb.New(pool)}
}

func (r *IncidentRepo) Create(ctx context.Context, p domain.CreateIncidentParams) (*domain.Incident, error) {
	sid, err := parseUUID(p.ServiceID)
	if err != nil {
		return nil, domain.ValidationErr("service_id", "invalid service_id")
	}
	row, err := r.q.CreateIncident(ctx, sqlcdb.CreateIncidentParams{
		ServiceID: sid,
		MonitorID: parseUUIDPtr(p.MonitorID),
		Title:     p.Title,
		Severity:  p.Severity,
	})
	if err != nil {
		return nil, err
	}
	return incidentToDomain(row), nil
}

func (r *IncidentRepo) GetByID(ctx context.Context, id string) (*domain.Incident, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("incident not found")
	}
	row, err := r.q.GetIncidentByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("incident not found")
		}
		return nil, err
	}
	return incidentToDomain(row), nil
}

func (r *IncidentRepo) GetOpenForMonitor(ctx context.Context, monitorID string) (*domain.Incident, error) {
	mid, err := parseUUID(monitorID)
	if err != nil {
		return nil, domain.NotFound("incident not found")
	}
	row, err := r.q.GetOpenIncidentForMonitor(ctx, mid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("incident not found")
		}
		return nil, err
	}
	return incidentToDomain(row), nil
}

func (r *IncidentRepo) List(ctx context.Context, serviceID string) ([]*domain.Incident, error) {
	sid, err := parseUUID(serviceID)
	if err != nil {
		return nil, domain.ValidationErr("service_id", "invalid service_id")
	}
	rows, err := r.q.ListIncidentsByService(ctx, sid)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Incident, len(rows))
	for i, row := range rows {
		out[i] = incidentToDomain(row)
	}
	return out, nil
}

func (r *IncidentRepo) ListAll(ctx context.Context) ([]*domain.Incident, error) {
	rows, err := r.q.ListAllIncidents(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Incident, len(rows))
	for i, row := range rows {
		out[i] = incidentToDomain(row)
	}
	return out, nil
}

func (r *IncidentRepo) Resolve(ctx context.Context, id string, resolvedAt time.Time) (*domain.Incident, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("incident not found")
	}
	row, err := r.q.ResolveIncident(ctx, sqlcdb.ResolveIncidentParams{
		ID:         uid,
		ResolvedAt: timeToPgtz(resolvedAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("incident not found")
		}
		return nil, err
	}
	return incidentToDomain(row), nil
}

func (r *IncidentRepo) Update(ctx context.Context, id string, p domain.UpdateIncidentParams) (*domain.Incident, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("incident not found")
	}
	row, err := r.q.UpdateIncident(ctx, sqlcdb.UpdateIncidentParams{
		ID:       uid,
		Title:    p.Title,
		Severity: p.Severity,
		Status:   p.Status,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("incident not found")
		}
		return nil, err
	}
	return incidentToDomain(row), nil
}

func (r *IncidentRepo) CreateUpdate(ctx context.Context, p domain.CreateIncidentUpdateParams) (*domain.IncidentUpdate, error) {
	iid, err := parseUUID(p.IncidentID)
	if err != nil {
		return nil, domain.ValidationErr("incident_id", "invalid incident_id")
	}
	row, err := r.q.CreateIncidentUpdate(ctx, sqlcdb.CreateIncidentUpdateParams{
		IncidentID: iid,
		Status:     p.Status,
		Message:    p.Message,
	})
	if err != nil {
		return nil, err
	}
	return incidentUpdateToDomain(row), nil
}

func (r *IncidentRepo) ListUpdates(ctx context.Context, incidentID string) ([]*domain.IncidentUpdate, error) {
	iid, err := parseUUID(incidentID)
	if err != nil {
		return nil, domain.ValidationErr("incident_id", "invalid incident_id")
	}
	rows, err := r.q.ListIncidentUpdates(ctx, iid)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.IncidentUpdate, len(rows))
	for i, row := range rows {
		out[i] = incidentUpdateToDomain(row)
	}
	return out, nil
}

func incidentToDomain(i sqlcdb.Incident) *domain.Incident {
	return &domain.Incident{
		ID:         uuidStr(i.ID),
		ServiceID:  uuidStr(i.ServiceID),
		MonitorID:  uuidStrPtr(i.MonitorID),
		Title:      i.Title,
		Severity:   i.Severity,
		Status:     i.Status,
		ResolvedAt: tsToTimePtr(i.ResolvedAt),
		CreatedAt:  tsToTime(i.CreatedAt),
		UpdatedAt:  tsToTime(i.UpdatedAt),
	}
}

func incidentUpdateToDomain(u sqlcdb.IncidentUpdate) *domain.IncidentUpdate {
	return &domain.IncidentUpdate{
		ID:         uuidStr(u.ID),
		IncidentID: uuidStr(u.IncidentID),
		Status:     u.Status,
		Message:    u.Message,
		CreatedAt:  tsToTime(u.CreatedAt),
	}
}
