package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majcek210/monsee/internal/domain"
	sqlcdb "github.com/majcek210/monsee/db/sqlc"
)

type NotificationRepo struct {
	q *sqlcdb.Queries
}

func NewNotificationRepo(pool *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{q: sqlcdb.New(pool)}
}

func (r *NotificationRepo) Create(ctx context.Context, p domain.CreateNotificationChannelParams) (*domain.NotificationChannel, error) {
	row, err := r.q.CreateNotificationChannel(ctx, sqlcdb.CreateNotificationChannelParams{
		Name:   p.Name,
		Type:   p.Type,
		Config: p.Config,
	})
	if err != nil {
		return nil, err
	}
	return notifToDomain(row), nil
}

func (r *NotificationRepo) GetByID(ctx context.Context, id string) (*domain.NotificationChannel, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("notification channel not found")
	}
	row, err := r.q.GetNotificationChannelByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("notification channel not found")
		}
		return nil, err
	}
	return notifToDomain(row), nil
}

func (r *NotificationRepo) List(ctx context.Context) ([]*domain.NotificationChannel, error) {
	rows, err := r.q.ListNotificationChannels(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.NotificationChannel, len(rows))
	for i, row := range rows {
		out[i] = notifToDomain(row)
	}
	return out, nil
}

func (r *NotificationRepo) ListEnabled(ctx context.Context) ([]*domain.NotificationChannel, error) {
	rows, err := r.q.ListEnabledNotificationChannels(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.NotificationChannel, len(rows))
	for i, row := range rows {
		out[i] = notifToDomain(row)
	}
	return out, nil
}

func (r *NotificationRepo) Update(ctx context.Context, id string, p domain.UpdateNotificationChannelParams) (*domain.NotificationChannel, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("notification channel not found")
	}
	row, err := r.q.UpdateNotificationChannel(ctx, sqlcdb.UpdateNotificationChannelParams{
		ID:      uid,
		Name:    p.Name,
		Config:  p.Config,
		Enabled: p.Enabled,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("notification channel not found")
		}
		return nil, err
	}
	return notifToDomain(row), nil
}

func (r *NotificationRepo) Archive(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return domain.NotFound("notification channel not found")
	}
	return r.q.ArchiveNotificationChannel(ctx, uid)
}

func notifToDomain(n sqlcdb.NotificationChannel) *domain.NotificationChannel {
	return &domain.NotificationChannel{
		ID:         uuidStr(n.ID),
		Name:       n.Name,
		Type:       n.Type,
		Config:     n.Config,
		Enabled:    n.Enabled,
		CreatedAt:  tsToTime(n.CreatedAt),
		ArchivedAt: tsToTimePtr(n.ArchivedAt),
	}
}
