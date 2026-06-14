package service

import (
	"context"
	"errors"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/pkg/hash"
)

type UserService struct {
	users domain.UserRepository
}

func NewUserService(users domain.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) Register(ctx context.Context, email, password, role string) (*domain.User, error) {
	if email == "" || password == "" {
		return nil, domain.ValidationErr("", "email and password are required")
	}
	if len(password) < 8 {
		return nil, domain.ValidationErr("password", "password must be at least 8 characters")
	}
	if role != "admin" && role != "viewer" {
		role = "viewer"
	}

	h, err := hash.Password(password)
	if err != nil {
		return nil, err
	}

	u, err := s.users.Create(ctx, domain.CreateUserParams{
		Email:        email,
		PasswordHash: h,
		Role:         role,
	})
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*domain.User, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.Unauthorized("invalid credentials")
		}
		return nil, err
	}
	if !hash.CheckPassword(u.PasswordHash, password) {
		return nil, domain.Unauthorized("invalid credentials")
	}
	return u, nil
}

func (s *UserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.users.GetByID(ctx, id)
}

func (s *UserService) List(ctx context.Context) ([]*domain.User, error) {
	return s.users.List(ctx)
}

// UpdateRole changes a user's role. Demoting the last remaining admin is
// rejected so the team can never lock itself out of /admin/*.
func (s *UserService) UpdateRole(ctx context.Context, id, role string) (*domain.User, error) {
	if role != "admin" && role != "viewer" {
		return nil, domain.ValidationErr("role", "role must be 'admin' or 'viewer'")
	}

	target, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if target.Role == "admin" && role != "admin" {
		count, err := s.users.CountActiveAdmins(ctx)
		if err != nil {
			return nil, err
		}
		if count <= 1 {
			return nil, domain.Conflict("cannot demote the last remaining admin")
		}
	}

	return s.users.UpdateRole(ctx, id, role)
}

// Archive deactivates a user. Archiving the last remaining admin is rejected
// for the same reason as UpdateRole.
func (s *UserService) Archive(ctx context.Context, id string) error {
	target, err := s.users.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if target.Role == "admin" {
		count, err := s.users.CountActiveAdmins(ctx)
		if err != nil {
			return err
		}
		if count <= 1 {
			return domain.Conflict("cannot archive the last remaining admin")
		}
	}

	return s.users.Archive(ctx, id)
}
