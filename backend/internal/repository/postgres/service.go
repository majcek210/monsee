package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majcek210/monsee/internal/domain"
	sqlcdb "github.com/majcek210/monsee/db/sqlc"
)

type ServiceRepo struct {
	q *sqlcdb.Queries
}

func NewServiceRepo(pool *pgxpool.Pool) *ServiceRepo {
	return &ServiceRepo{q: sqlcdb.New(pool)}
}

func (r *ServiceRepo) Create(ctx context.Context, p domain.CreateServiceParams) (*domain.Service, error) {
	row, err := r.q.CreateService(ctx, sqlcdb.CreateServiceParams{
		Name:        p.Name,
		Description: p.Description,
	})
	if err != nil {
		return nil, err
	}
	return serviceToDomain(row), nil
}

func (r *ServiceRepo) GetByID(ctx context.Context, id string) (*domain.Service, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("service not found")
	}
	row, err := r.q.GetServiceByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("service not found")
		}
		return nil, err
	}
	return serviceToDomain(row), nil
}

func (r *ServiceRepo) List(ctx context.Context) ([]*domain.Service, error) {
	rows, err := r.q.ListServices(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Service, len(rows))
	for i, row := range rows {
		out[i] = serviceToDomain(row)
	}
	return out, nil
}

func (r *ServiceRepo) Update(ctx context.Context, id string, p domain.UpdateServiceParams) (*domain.Service, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("service not found")
	}
	row, err := r.q.UpdateService(ctx, sqlcdb.UpdateServiceParams{
		ID:          uid,
		Name:        &p.Name,
		Description: p.Description,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("service not found")
		}
		return nil, err
	}
	return serviceToDomain(row), nil
}

func (r *ServiceRepo) UpdateStatus(ctx context.Context, id, status string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return domain.NotFound("service not found")
	}
	_, err = r.q.UpdateService(ctx, sqlcdb.UpdateServiceParams{
		ID:     uid,
		Status: &status,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.NotFound("service not found")
	}
	return err
}

func (r *ServiceRepo) Archive(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return domain.NotFound("service not found")
	}
	return r.q.ArchiveService(ctx, uid)
}

func serviceToDomain(s sqlcdb.Service) *domain.Service {
	return &domain.Service{
		ID:          uuidStr(s.ID),
		Name:        s.Name,
		Description: s.Description,
		Status:      s.Status,
		CreatedAt:   tsToTime(s.CreatedAt),
		ArchivedAt:  tsToTimePtr(s.ArchivedAt),
	}
}
