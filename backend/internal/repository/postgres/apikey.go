package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majcek210/monsee/internal/domain"
	sqlcdb "github.com/majcek210/monsee/db/sqlc"
)

type APIKeyRepo struct {
	q *sqlcdb.Queries
}

func NewAPIKeyRepo(pool *pgxpool.Pool) *APIKeyRepo {
	return &APIKeyRepo{q: sqlcdb.New(pool)}
}

func (r *APIKeyRepo) Create(ctx context.Context, p domain.CreateAPIKeyParams) (*domain.APIKey, error) {
	uid, err := parseUUID(p.UserID)
	if err != nil {
		return nil, domain.ValidationErr("user_id", "invalid user_id")
	}
	row, err := r.q.CreateAPIKey(ctx, sqlcdb.CreateAPIKeyParams{
		UserID:  uid,
		Name:    p.Name,
		KeyHash: p.KeyHash,
		Prefix:  p.Prefix,
	})
	if err != nil {
		return nil, err
	}
	return apiKeyToDomain(row), nil
}

func (r *APIKeyRepo) GetByID(ctx context.Context, id string) (*domain.APIKey, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("api key not found")
	}
	row, err := r.q.GetAPIKeyByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("api key not found")
		}
		return nil, err
	}
	return apiKeyToDomain(row), nil
}

func (r *APIKeyRepo) GetByHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	row, err := r.q.GetAPIKeyByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("api key not found")
		}
		return nil, err
	}
	return apiKeyToDomain(row), nil
}

func (r *APIKeyRepo) ListByUser(ctx context.Context, userID string) ([]*domain.APIKey, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, domain.ValidationErr("user_id", "invalid user_id")
	}
	rows, err := r.q.ListAPIKeysByUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.APIKey, len(rows))
	for i, row := range rows {
		out[i] = apiKeyToDomain(row)
	}
	return out, nil
}

func (r *APIKeyRepo) UpdateLastUsed(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return domain.NotFound("api key not found")
	}
	return r.q.UpdateAPIKeyLastUsed(ctx, uid)
}

func (r *APIKeyRepo) Archive(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return domain.NotFound("api key not found")
	}
	return r.q.ArchiveAPIKey(ctx, uid)
}

func apiKeyToDomain(k sqlcdb.ApiKey) *domain.APIKey {
	return &domain.APIKey{
		ID:         uuidStr(k.ID),
		UserID:     uuidStr(k.UserID),
		Name:       k.Name,
		KeyHash:    k.KeyHash,
		Prefix:     k.Prefix,
		CreatedAt:  tsToTime(k.CreatedAt),
		LastUsed:   tsToTimePtr(k.LastUsed),
		ArchivedAt: tsToTimePtr(k.ArchivedAt),
	}
}
