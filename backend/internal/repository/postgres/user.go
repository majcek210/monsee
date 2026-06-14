package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majcek210/monsee/internal/domain"
	sqlcdb "github.com/majcek210/monsee/db/sqlc"
)

type UserRepo struct {
	q *sqlcdb.Queries
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{q: sqlcdb.New(pool)}
}

func (r *UserRepo) Create(ctx context.Context, p domain.CreateUserParams) (*domain.User, error) {
	row, err := r.q.CreateUser(ctx, sqlcdb.CreateUserParams{
		Email:        p.Email,
		PasswordHash: p.PasswordHash,
		Role:         p.Role,
	})
	if err != nil {
		return nil, err
	}
	return userToDomain(row), nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("user not found")
	}
	row, err := r.q.GetUserByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("user not found")
		}
		return nil, err
	}
	return userToDomain(row), nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("user not found")
		}
		return nil, err
	}
	return userToDomain(row), nil
}

func (r *UserRepo) List(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.q.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.User, len(rows))
	for i, row := range rows {
		out[i] = userToDomain(row)
	}
	return out, nil
}

func (r *UserRepo) UpdateRole(ctx context.Context, id, role string) (*domain.User, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("user not found")
	}
	row, err := r.q.UpdateUserRole(ctx, sqlcdb.UpdateUserRoleParams{
		ID:   uid,
		Role: role,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("user not found")
		}
		return nil, err
	}
	return userToDomain(row), nil
}

func (r *UserRepo) CountActiveAdmins(ctx context.Context) (int64, error) {
	return r.q.CountActiveAdmins(ctx)
}

func (r *UserRepo) Archive(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return domain.NotFound("user not found")
	}
	return r.q.ArchiveUser(ctx, uid)
}

func userToDomain(u sqlcdb.User) *domain.User {
	return &domain.User{
		ID:           uuidStr(u.ID),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		CreatedAt:    tsToTime(u.CreatedAt),
		ArchivedAt:   tsToTimePtr(u.ArchivedAt),
	}
}
