package domain

import (
	"context"
	"time"
)

type Service struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	ArchivedAt  *time.Time `json:"archived_at"`
}

type CreateServiceParams struct {
	Name        string
	Description *string
}

type UpdateServiceParams struct {
	Name        string
	Description *string
}

type ServiceRepository interface {
	Create(ctx context.Context, p CreateServiceParams) (*Service, error)
	GetByID(ctx context.Context, id string) (*Service, error)
	List(ctx context.Context) ([]*Service, error)
	Update(ctx context.Context, id string, p UpdateServiceParams) (*Service, error)
	UpdateStatus(ctx context.Context, id, status string) error
	Archive(ctx context.Context, id string) error
}
