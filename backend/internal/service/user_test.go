package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/pkg/hash"
)

// mockUserRepo is an in-memory domain.UserRepository for unit tests.
type mockUserRepo struct {
	mu    sync.Mutex
	items map[string]*domain.User
	seq   int
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{items: map[string]*domain.User{}}
}

func (m *mockUserRepo) Create(ctx context.Context, p domain.CreateUserParams) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	u := &domain.User{
		ID:           fmt.Sprintf("user-%d", m.seq),
		Email:        p.Email,
		PasswordHash: p.PasswordHash,
		Role:         p.Role,
		CreatedAt:    time.Now(),
	}
	m.items[u.ID] = u
	cp := *u
	return &cp, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.items[id]
	if !ok {
		return nil, domain.NotFound("user not found")
	}
	cp := *u
	return &cp, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.items {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, domain.NotFound("user not found")
}

func (m *mockUserRepo) List(ctx context.Context) ([]*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*domain.User
	for _, u := range m.items {
		cp := *u
		out = append(out, &cp)
	}
	return out, nil
}

func (m *mockUserRepo) UpdateRole(ctx context.Context, id, role string) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.items[id]
	if !ok {
		return nil, domain.NotFound("user not found")
	}
	u.Role = role
	cp := *u
	return &cp, nil
}

func (m *mockUserRepo) CountActiveAdmins(ctx context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for _, u := range m.items {
		if u.Role == "admin" && u.ArchivedAt == nil {
			n++
		}
	}
	return n, nil
}

func (m *mockUserRepo) Archive(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.items[id]
	if !ok {
		return domain.NotFound("user not found")
	}
	now := time.Now()
	u.ArchivedAt = &now
	return nil
}

func TestRegisterValidation(t *testing.T) {
	repo := newMockUserRepo()
	s := NewUserService(repo)

	if _, err := s.Register(context.Background(), "", "password123", "admin"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error for empty email, got %v", err)
	}
	if _, err := s.Register(context.Background(), "a@b.com", "short", "admin"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error for short password, got %v", err)
	}
}

func TestRegisterDefaultsInvalidRoleToViewer(t *testing.T) {
	repo := newMockUserRepo()
	s := NewUserService(repo)

	u, err := s.Register(context.Background(), "a@b.com", "password123", "superuser")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if u.Role != "viewer" {
		t.Fatalf("role = %q, want %q", u.Role, "viewer")
	}
}

func TestRegisterHashesPassword(t *testing.T) {
	repo := newMockUserRepo()
	s := NewUserService(repo)

	u, err := s.Register(context.Background(), "a@b.com", "password123", "admin")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if u.PasswordHash == "password123" {
		t.Fatal("password must be hashed, not stored in plaintext")
	}
	if !hash.CheckPassword(u.PasswordHash, "password123") {
		t.Fatal("stored hash does not verify against the original password")
	}
}

func TestLoginSuccess(t *testing.T) {
	repo := newMockUserRepo()
	s := NewUserService(repo)

	if _, err := s.Register(context.Background(), "a@b.com", "password123", "admin"); err != nil {
		t.Fatalf("Register: %v", err)
	}

	u, err := s.Login(context.Background(), "a@b.com", "password123")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if u.Email != "a@b.com" {
		t.Fatalf("email = %q, want %q", u.Email, "a@b.com")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	s := NewUserService(repo)

	if _, err := s.Register(context.Background(), "a@b.com", "password123", "admin"); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if _, err := s.Login(context.Background(), "a@b.com", "wrong-password"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized for wrong password, got %v", err)
	}
}

// TestLoginUnknownEmailDoesNotLeakExistence ensures an unknown email returns
// the same generic "unauthorized" error as a wrong password, so the login
// endpoint can't be used to enumerate registered accounts.
func TestLoginUnknownEmailDoesNotLeakExistence(t *testing.T) {
	repo := newMockUserRepo()
	s := NewUserService(repo)

	if _, err := s.Login(context.Background(), "nobody@example.com", "password123"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized for unknown email, got %v", err)
	}
}

func TestUpdateRoleValidation(t *testing.T) {
	repo := newMockUserRepo()
	s := NewUserService(repo)

	u, err := s.Register(context.Background(), "a@b.com", "password123", "admin")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if _, err := s.UpdateRole(context.Background(), u.ID, "superuser"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error for invalid role, got %v", err)
	}
}

// TestUpdateRoleLastAdminLockout is the regression test for Security.md S5a:
// demoting the only remaining admin must be rejected so the team can't lock
// itself out of /admin/*.
func TestUpdateRoleLastAdminLockout(t *testing.T) {
	repo := newMockUserRepo()
	s := NewUserService(repo)

	admin, err := s.Register(context.Background(), "admin@b.com", "password123", "admin")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if _, err := s.UpdateRole(context.Background(), admin.ID, "viewer"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict demoting last admin, got %v", err)
	}
}

func TestUpdateRoleAllowsDemoteWhenAnotherAdminExists(t *testing.T) {
	repo := newMockUserRepo()
	s := NewUserService(repo)

	admin1, err := s.Register(context.Background(), "admin1@b.com", "password123", "admin")
	if err != nil {
		t.Fatalf("Register admin1: %v", err)
	}
	if _, err := s.Register(context.Background(), "admin2@b.com", "password123", "admin"); err != nil {
		t.Fatalf("Register admin2: %v", err)
	}

	u, err := s.UpdateRole(context.Background(), admin1.ID, "viewer")
	if err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}
	if u.Role != "viewer" {
		t.Fatalf("role = %q, want %q", u.Role, "viewer")
	}
}

// TestArchiveLastAdminLockout is the regression test for Security.md S5a:
// archiving the only remaining admin must be rejected.
func TestArchiveLastAdminLockout(t *testing.T) {
	repo := newMockUserRepo()
	s := NewUserService(repo)

	admin, err := s.Register(context.Background(), "admin@b.com", "password123", "admin")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if err := s.Archive(context.Background(), admin.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict archiving last admin, got %v", err)
	}
}

func TestArchiveAllowsWhenAnotherAdminExists(t *testing.T) {
	repo := newMockUserRepo()
	s := NewUserService(repo)

	admin1, err := s.Register(context.Background(), "admin1@b.com", "password123", "admin")
	if err != nil {
		t.Fatalf("Register admin1: %v", err)
	}
	if _, err := s.Register(context.Background(), "admin2@b.com", "password123", "admin"); err != nil {
		t.Fatalf("Register admin2: %v", err)
	}

	if err := s.Archive(context.Background(), admin1.ID); err != nil {
		t.Fatalf("Archive: %v", err)
	}
}

func TestArchiveViewerAlwaysAllowed(t *testing.T) {
	repo := newMockUserRepo()
	s := NewUserService(repo)

	if _, err := s.Register(context.Background(), "admin@b.com", "password123", "admin"); err != nil {
		t.Fatalf("Register admin: %v", err)
	}
	viewer, err := s.Register(context.Background(), "viewer@b.com", "password123", "viewer")
	if err != nil {
		t.Fatalf("Register viewer: %v", err)
	}

	if err := s.Archive(context.Background(), viewer.ID); err != nil {
		t.Fatalf("Archive: %v", err)
	}
}
