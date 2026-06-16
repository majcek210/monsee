package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majcek210/monsee/internal/domain"
	sqlcdb "github.com/majcek210/monsee/db/sqlc"
)

type SettingsRepo struct {
	q *sqlcdb.Queries
}

func NewSettingsRepo(pool *pgxpool.Pool) *SettingsRepo {
	return &SettingsRepo{q: sqlcdb.New(pool)}
}

func (r *SettingsRepo) Get(ctx context.Context) (*domain.Settings, error) {
	row, err := r.q.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	return settingsToDomain(row), nil
}

func (r *SettingsRepo) Update(ctx context.Context, p domain.UpdateSettingsParams) (*domain.Settings, error) {
	row, err := r.q.UpdateSettings(ctx, sqlcdb.UpdateSettingsParams{
		SiteTitle:           p.SiteTitle,
		LogoUrl:             p.LogoURL,
		PublicStatusEnabled: p.PublicStatusEnabled,
	})
	if err != nil {
		return nil, err
	}
	return settingsToDomain(row), nil
}

func settingsToDomain(s sqlcdb.Setting) *domain.Settings {
	return &domain.Settings{
		ID:                  int(s.ID),
		SiteTitle:           s.SiteTitle,
		LogoURL:             s.LogoUrl,
		PublicStatusEnabled: s.PublicStatusEnabled,
		UpdatedAt:           tsToTime(s.UpdatedAt),
	}
}
