package domain

import (
	"context"
	"time"
)

type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	TOTPEnabled  bool       `json:"totp_enabled"`
	CreatedAt    time.Time  `json:"created_at"`
	ArchivedAt   *time.Time `json:"archived_at"`
}

type CreateUserParams struct {
	Email        string
	PasswordHash string
	Role         string
}

type TOTPData struct {
	Secret      *string
	Enabled     bool
	BackupCodes []string
}

type UserRepository interface {
	Create(ctx context.Context, p CreateUserParams) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]*User, error)
	UpdateRole(ctx context.Context, id, role string) (*User, error)
	CountActiveAdmins(ctx context.Context) (int64, error)
	Archive(ctx context.Context, id string) error
	GetTOTP(ctx context.Context, userID string) (*TOTPData, error)
	SetTOTPSecret(ctx context.Context, userID, secret string) error
	EnableTOTP(ctx context.Context, userID string, backupCodes []string) error
	DisableTOTP(ctx context.Context, userID string) error
	RemoveBackupCode(ctx context.Context, userID, code string) error
}
