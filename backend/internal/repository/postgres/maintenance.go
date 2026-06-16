package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majcek210/monsee/internal/domain"
	sqlcdb "github.com/majcek210/monsee/db/sqlc"
)

type MaintenanceRepo struct {
	q *sqlcdb.Queries
}

func NewMaintenanceRepo(pool *pgxpool.Pool) *MaintenanceRepo {
	return &MaintenanceRepo{q: sqlcdb.New(pool)}
}

func (r *MaintenanceRepo) Create(ctx context.Context, p domain.CreateMaintenanceWindowParams) (*domain.MaintenanceWindow, error) {
	sid, err := parseUUID(p.ServiceID)
	if err != nil {
		return nil, domain.ValidationErr("service_id", "invalid service_id")
	}
	row, err := r.q.CreateMaintenanceWindow(ctx, sqlcdb.CreateMaintenanceWindowParams{
		ServiceID:   sid,
		Title:       p.Title,
		Description: p.Description,
		StartsAt:    timeToPgtz(p.StartsAt),
		EndsAt:      timeToPgtz(p.EndsAt),
	})
	if err != nil {
		return nil, err
	}
	return maintenanceToDomain(row), nil
}

func (r *MaintenanceRepo) GetByID(ctx context.Context, id string) (*domain.MaintenanceWindow, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("maintenance window not found")
	}
	row, err := r.q.GetMaintenanceWindowByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("maintenance window not found")
		}
		return nil, err
	}
	return maintenanceToDomain(row), nil
}

func (r *MaintenanceRepo) ListByService(ctx context.Context, serviceID string) ([]*domain.MaintenanceWindow, error) {
	sid, err := parseUUID(serviceID)
	if err != nil {
		return nil, domain.ValidationErr("service_id", "invalid service_id")
	}
	rows, err := r.q.ListMaintenanceWindowsByService(ctx, sid)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.MaintenanceWindow, len(rows))
	for i, row := range rows {
		out[i] = maintenanceToDomain(row)
	}
	return out, nil
}

func (r *MaintenanceRepo) ListActive(ctx context.Context) ([]*domain.MaintenanceWindow, error) {
	rows, err := r.q.ListActiveMaintenanceWindows(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.MaintenanceWindow, len(rows))
	for i, row := range rows {
		out[i] = maintenanceToDomain(row)
	}
	return out, nil
}

func (r *MaintenanceRepo) ListActiveForService(ctx context.Context, serviceID string) ([]*domain.MaintenanceWindow, error) {
	sid, err := parseUUID(serviceID)
	if err != nil {
		return nil, domain.ValidationErr("service_id", "invalid service_id")
	}
	rows, err := r.q.ListActiveForService(ctx, sid)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.MaintenanceWindow, len(rows))
	for i, row := range rows {
		out[i] = maintenanceToDomain(row)
	}
	return out, nil
}

func (r *MaintenanceRepo) IsActiveForService(ctx context.Context, serviceID string) (bool, error) {
	sid, err := parseUUID(serviceID)
	if err != nil {
		return false, nil
	}
	active, err := r.q.IsMaintenanceActiveForService(ctx, sid)
	if err != nil {
		return false, err
	}
	return active, nil
}

func (r *MaintenanceRepo) Update(ctx context.Context, id string, p domain.UpdateMaintenanceWindowParams) (*domain.MaintenanceWindow, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("maintenance window not found")
	}

	var startsAt, endsAt pgtype.Timestamptz
	if p.StartsAt != nil {
		startsAt = timeToPgtz(*p.StartsAt)
	}
	if p.EndsAt != nil {
		endsAt = timeToPgtz(*p.EndsAt)
	}

	row, err := r.q.UpdateMaintenanceWindow(ctx, sqlcdb.UpdateMaintenanceWindowParams{
		ID:          uid,
		Title:       p.Title,
		Description: p.Description,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("maintenance window not found")
		}
		return nil, err
	}
	return maintenanceToDomain(row), nil
}

func (r *MaintenanceRepo) Archive(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return domain.NotFound("maintenance window not found")
	}
	return r.q.ArchiveMaintenanceWindow(ctx, uid)
}

func maintenanceToDomain(m sqlcdb.MaintenanceWindow) *domain.MaintenanceWindow {
	return &domain.MaintenanceWindow{
		ID:          uuidStr(m.ID),
		ServiceID:   uuidStr(m.ServiceID),
		Title:       m.Title,
		Description: m.Description,
		StartsAt:    tsToTime(m.StartsAt),
		EndsAt:      tsToTime(m.EndsAt),
		CreatedAt:   tsToTime(m.CreatedAt),
		ArchivedAt:  tsToTimePtr(m.ArchivedAt),
	}
}
