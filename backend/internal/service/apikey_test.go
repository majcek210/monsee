package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/pkg/hash"
)

// mockAPIKeyRepo is an in-memory domain.APIKeyRepository for unit tests.
type mockAPIKeyRepo struct {
	mu   sync.Mutex
	keys map[string]*domain.APIKey
	seq  int
}

func newMockAPIKeyRepo() *mockAPIKeyRepo {
	return &mockAPIKeyRepo{keys: map[string]*domain.APIKey{}}
}

func (m *mockAPIKeyRepo) Create(ctx context.Context, p domain.CreateAPIKeyParams) (*domain.APIKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	k := &domain.APIKey{
		ID:        fmt.Sprintf("key-%d", m.seq),
		UserID:    p.UserID,
		Name:      p.Name,
		KeyHash:   p.KeyHash,
		Prefix:    p.Prefix,
		CreatedAt: time.Now(),
	}
	m.keys[k.ID] = k
	cp := *k
	return &cp, nil
}

func (m *mockAPIKeyRepo) GetByID(ctx context.Context, id string) (*domain.APIKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k, ok := m.keys[id]
	if !ok {
		return nil, domain.NotFound("api key not found")
	}
	cp := *k
	return &cp, nil
}

func (m *mockAPIKeyRepo) GetByHash(ctx context.Context, h string) (*domain.APIKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, k := range m.keys {
		if k.KeyHash == h {
			cp := *k
			return &cp, nil
		}
	}
	return nil, domain.NotFound("api key not found")
}

func (m *mockAPIKeyRepo) ListByUser(ctx context.Context, userID string) ([]*domain.APIKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*domain.APIKey
	for _, k := range m.keys {
		if k.UserID == userID {
			cp := *k
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (m *mockAPIKeyRepo) UpdateLastUsed(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if k, ok := m.keys[id]; ok {
		now := time.Now()
		k.LastUsed = &now
	}
	return nil
}

func (m *mockAPIKeyRepo) Archive(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k, ok := m.keys[id]
	if !ok {
		return domain.NotFound("api key not found")
	}
	now := time.Now()
	k.ArchivedAt = &now
	return nil
}

func TestGenerateCreatesKeyWithMatchingHashAndPrefix(t *testing.T) {
	repo := newMockAPIKeyRepo()
	s := NewAPIKeyService(repo)

	plain, key, err := s.Generate(context.Background(), "user-1", "CI key")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if !strings.HasPrefix(plain, "sk_") {
		t.Fatalf("plain key %q missing sk_ prefix", plain)
	}
	if key.Prefix != plain[:8] {
		t.Fatalf("prefix %q does not match plain key prefix %q", key.Prefix, plain[:8])
	}
	if key.KeyHash != hash.SHA256(plain) {
		t.Fatal("stored hash does not match SHA256(plain key)")
	}
	if key.UserID != "user-1" || key.Name != "CI key" {
		t.Fatalf("unexpected key: %+v", key)
	}
}

func TestGenerateEmptyNameRejected(t *testing.T) {
	repo := newMockAPIKeyRepo()
	s := NewAPIKeyService(repo)

	if _, _, err := s.Generate(context.Background(), "user-1", ""); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestValidateAndGetUserIDSuccess(t *testing.T) {
	repo := newMockAPIKeyRepo()
	s := NewAPIKeyService(repo)

	plain, _, err := s.Generate(context.Background(), "user-1", "CI key")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	uid, err := s.ValidateAndGetUserID(context.Background(), plain)
	if err != nil {
		t.Fatalf("ValidateAndGetUserID: %v", err)
	}
	if uid != "user-1" {
		t.Fatalf("uid = %q, want %q", uid, "user-1")
	}
}

func TestValidateAndGetUserIDInvalidKey(t *testing.T) {
	repo := newMockAPIKeyRepo()
	s := NewAPIKeyService(repo)

	if _, err := s.ValidateAndGetUserID(context.Background(), "sk_does_not_exist"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized for unknown key, got %v", err)
	}
}

func TestRevokeOwnerCanRevokeOwnKey(t *testing.T) {
	repo := newMockAPIKeyRepo()
	s := NewAPIKeyService(repo)

	_, key, err := s.Generate(context.Background(), "user-1", "CI key")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if err := s.Revoke(context.Background(), "user-1", "viewer", key.ID); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
}

// TestRevokeNonOwnerNonAdminForbidden is the regression test for the API key
// IDOR (Security.md S2): a non-admin user must not be able to revoke another
// user's key just by knowing its ID.
func TestRevokeNonOwnerNonAdminForbidden(t *testing.T) {
	repo := newMockAPIKeyRepo()
	s := NewAPIKeyService(repo)

	_, key, err := s.Generate(context.Background(), "user-1", "CI key")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	err = s.Revoke(context.Background(), "user-2", "viewer", key.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestRevokeAdminCanRevokeAnyKey(t *testing.T) {
	repo := newMockAPIKeyRepo()
	s := NewAPIKeyService(repo)

	_, key, err := s.Generate(context.Background(), "user-1", "CI key")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if err := s.Revoke(context.Background(), "user-2", "admin", key.ID); err != nil {
		t.Fatalf("Revoke as admin: %v", err)
	}
}

func TestRevokeNonexistentKeyNotFound(t *testing.T) {
	repo := newMockAPIKeyRepo()
	s := NewAPIKeyService(repo)

	err := s.Revoke(context.Background(), "user-1", "viewer", "does-not-exist")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
