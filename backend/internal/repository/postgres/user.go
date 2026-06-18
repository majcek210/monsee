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

func (r *UserRepo) UpdateEmail(ctx context.Context, id, email string) (*domain.User, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, domain.NotFound("user not found")
	}
	row, err := r.q.UpdateUserEmail(ctx, sqlcdb.UpdateUserEmailParams{
		ID:    uid,
		Email: email,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("user not found")
		}
		return nil, err
	}
	return userToDomain(row), nil
}

func (r *UserRepo) UpdatePasswordHash(ctx context.Context, id, passwordHash string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return domain.NotFound("user not found")
	}
	return r.q.UpdateUserPassword(ctx, sqlcdb.UpdateUserPasswordParams{
		ID:           uid,
		PasswordHash: passwordHash,
	})
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

func (r *UserRepo) GetTOTP(ctx context.Context, userID string) (*domain.TOTPData, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, domain.NotFound("user not found")
	}
	row, err := r.q.GetTOTPByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return &domain.TOTPData{
		Secret:      row.TotpSecret,
		Enabled:     row.TotpEnabled,
		BackupCodes: row.TotpBackupCodes,
	}, nil
}

func (r *UserRepo) SetTOTPSecret(ctx context.Context, userID, secret string) error {
	uid, err := parseUUID(userID)
	if err != nil {
		return domain.NotFound("user not found")
	}
	return r.q.SetTOTPSecret(ctx, sqlcdb.SetTOTPSecretParams{
		ID:         uid,
		TotpSecret: &secret,
	})
}

func (r *UserRepo) EnableTOTP(ctx context.Context, userID string, backupCodes []string) error {
	uid, err := parseUUID(userID)
	if err != nil {
		return domain.NotFound("user not found")
	}
	return r.q.EnableTOTP(ctx, sqlcdb.EnableTOTPParams{
		ID:              uid,
		TotpBackupCodes: backupCodes,
	})
}

func (r *UserRepo) DisableTOTP(ctx context.Context, userID string) error {
	uid, err := parseUUID(userID)
	if err != nil {
		return domain.NotFound("user not found")
	}
	return r.q.DisableTOTP(ctx, uid)
}

func (r *UserRepo) RemoveBackupCode(ctx context.Context, userID, code string) error {
	uid, err := parseUUID(userID)
	if err != nil {
		return domain.NotFound("user not found")
	}
	return r.q.RemoveBackupCode(ctx, sqlcdb.RemoveBackupCodeParams{
		ID:          uid,
		ArrayRemove: code,
	})
}

func (r *UserRepo) ConsumeBackupCode(ctx context.Context, userID, codeHash string) (int64, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return 0, domain.NotFound("user not found")
	}
	return r.q.ConsumeBackupCode(ctx, sqlcdb.ConsumeBackupCodeParams{
		ID:          uid,
		ArrayRemove: codeHash,
	})
}

func userToDomain(u sqlcdb.User) *domain.User {
	return &domain.User{
		ID:           uuidStr(u.ID),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		TOTPEnabled:  u.TotpEnabled,
		CreatedAt:    tsToTime(u.CreatedAt),
		ArchivedAt:   tsToTimePtr(u.ArchivedAt),
	}
}
