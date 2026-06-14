package domain

import (
	"context"
	"time"
)

type APIKey struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Name       string     `json:"name"`
	KeyHash    string     `json:"-"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsed   *time.Time `json:"last_used"`
	ArchivedAt *time.Time `json:"archived_at"`
}

type CreateAPIKeyParams struct {
	UserID  string
	Name    string
	KeyHash string
	Prefix  string
}

type APIKeyRepository interface {
	Create(ctx context.Context, p CreateAPIKeyParams) (*APIKey, error)
	GetByID(ctx context.Context, id string) (*APIKey, error)
	GetByHash(ctx context.Context, hash string) (*APIKey, error)
	ListByUser(ctx context.Context, userID string) ([]*APIKey, error)
	UpdateLastUsed(ctx context.Context, id string) error
	Archive(ctx context.Context, id string) error
}
