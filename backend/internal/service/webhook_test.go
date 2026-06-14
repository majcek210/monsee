package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/pkg/encrypt"
)

var webhookTestKey = []byte("01234567890123456789012345678901")[:32]

// mockWebhookRepo is an in-memory domain.WebhookRepository for unit tests.
// Update applies the same COALESCE-on-nil semantics as the real SQL query.
type mockWebhookRepo struct {
	mu    sync.Mutex
	items map[string]*domain.Webhook
	seq   int
}

func newMockWebhookRepo() *mockWebhookRepo {
	return &mockWebhookRepo{items: map[string]*domain.Webhook{}}
}

func (m *mockWebhookRepo) Create(ctx context.Context, p domain.CreateWebhookParams) (*domain.Webhook, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	w := &domain.Webhook{
		ID:        fmt.Sprintf("wh-%d", m.seq),
		Name:      p.Name,
		URL:       p.URL,
		Secret:    p.Secret,
		Events:    p.Events,
		Enabled:   true,
		CreatedAt: time.Now(),
	}
	m.items[w.ID] = w
	cp := *w
	return &cp, nil
}

func (m *mockWebhookRepo) GetByID(ctx context.Context, id string) (*domain.Webhook, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.items[id]
	if !ok {
		return nil, domain.NotFound("webhook not found")
	}
	cp := *w
	return &cp, nil
}

func (m *mockWebhookRepo) List(ctx context.Context) ([]*domain.Webhook, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*domain.Webhook
	for _, w := range m.items {
		cp := *w
		out = append(out, &cp)
	}
	return out, nil
}

func (m *mockWebhookRepo) ListByEvent(ctx context.Context, event string) ([]*domain.Webhook, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*domain.Webhook
	for _, w := range m.items {
		if slices.Contains(w.Events, event) {
			cp := *w
			out = append(out, &cp)
		}
	}
	return out, nil
}

// Update mirrors the UpdateWebhook SQL query: any nil field in p leaves the
// existing column value unchanged.
func (m *mockWebhookRepo) Update(ctx context.Context, id string, p domain.UpdateWebhookParams) (*domain.Webhook, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.items[id]
	if !ok {
		return nil, domain.NotFound("webhook not found")
	}
	if p.Name != nil {
		w.Name = *p.Name
	}
	if p.URL != nil {
		w.URL = *p.URL
	}
	if p.Secret != nil {
		w.Secret = p.Secret
	}
	if p.Events != nil {
		w.Events = p.Events
	}
	if p.Enabled != nil {
		w.Enabled = *p.Enabled
	}
	cp := *w
	return &cp, nil
}

func (m *mockWebhookRepo) Archive(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.items[id]
	if !ok {
		return domain.NotFound("webhook not found")
	}
	now := time.Now()
	w.ArchivedAt = &now
	return nil
}

func (m *mockWebhookRepo) InsertLog(ctx context.Context, p domain.InsertWebhookLogParams) (*domain.WebhookLog, error) {
	return &domain.WebhookLog{WebhookID: p.WebhookID, Event: p.Event, DeliveredAt: time.Now()}, nil
}

func (m *mockWebhookRepo) ListLogs(ctx context.Context, webhookID string, limit int32) ([]*domain.WebhookLog, error) {
	return nil, nil
}

func TestWebhookCreateValidation(t *testing.T) {
	repo := newMockWebhookRepo()
	s := NewWebhookService(repo, webhookTestKey)

	if _, err := s.Create(context.Background(), "", "https://example.com/hook", "", nil); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error for empty name, got %v", err)
	}
	if _, err := s.Create(context.Background(), "alerts", "", "", nil); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error for empty url, got %v", err)
	}
}

func TestWebhookCreateEncryptsURLAndSecret(t *testing.T) {
	repo := newMockWebhookRepo()
	s := NewWebhookService(repo, webhookTestKey)

	wh, err := s.Create(context.Background(), "alerts", "https://example.com/hook", "shh-secret", []string{"monitor.down"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if wh.URL == "https://example.com/hook" {
		t.Fatal("URL must be stored encrypted, not in plaintext")
	}
	gotURL, err := encrypt.Decrypt(webhookTestKey, wh.URL)
	if err != nil || gotURL != "https://example.com/hook" {
		t.Fatalf("URL did not round-trip: got %q, err %v", gotURL, err)
	}

	if wh.Secret == nil {
		t.Fatal("expected secret to be set")
	}
	gotSecret, err := encrypt.Decrypt(webhookTestKey, *wh.Secret)
	if err != nil || gotSecret != "shh-secret" {
		t.Fatalf("secret did not round-trip: got %q, err %v", gotSecret, err)
	}
}

func TestWebhookCreateWithoutSecretLeavesSecretNil(t *testing.T) {
	repo := newMockWebhookRepo()
	s := NewWebhookService(repo, webhookTestKey)

	wh, err := s.Create(context.Background(), "alerts", "https://example.com/hook", "", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if wh.Secret != nil {
		t.Fatal("expected nil secret when none provided")
	}
}

// TestWebhookUpdateToggleEnabledPreservesSecretsAndURL is the regression test
// for Tests.md F1 / Security.md S4: toggling "enabled" alone must not wipe
// out the encrypted URL or secret.
func TestWebhookUpdateToggleEnabledPreservesSecretsAndURL(t *testing.T) {
	repo := newMockWebhookRepo()
	s := NewWebhookService(repo, webhookTestKey)

	wh, err := s.Create(context.Background(), "alerts", "https://example.com/hook", "shh-secret", []string{"monitor.down"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	disabled := false
	updated, err := s.Update(context.Background(), wh.ID, nil, nil, nil, nil, &disabled)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if updated.Enabled {
		t.Fatal("enabled should now be false")
	}
	if updated.URL != wh.URL {
		t.Fatalf("URL changed unexpectedly: %q -> %q", wh.URL, updated.URL)
	}
	if updated.Secret == nil || *updated.Secret != *wh.Secret {
		t.Fatal("secret was lost/changed by an enabled-only update")
	}
	if updated.Name != wh.Name {
		t.Fatal("name changed unexpectedly")
	}
}

// TestWebhookUpdateEmptySecretDoesNotOverwriteExisting verifies the frontend's
// "leave blank to keep existing secret" UX: an explicit empty-string secret
// must not clear the previously stored secret.
func TestWebhookUpdateEmptySecretDoesNotOverwriteExisting(t *testing.T) {
	repo := newMockWebhookRepo()
	s := NewWebhookService(repo, webhookTestKey)

	wh, err := s.Create(context.Background(), "alerts", "https://example.com/hook", "original-secret", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	newName := "renamed"
	emptySecret := ""
	updated, err := s.Update(context.Background(), wh.ID, &newName, nil, &emptySecret, nil, nil)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if updated.Name != "renamed" {
		t.Fatalf("name = %q, want %q", updated.Name, "renamed")
	}
	if updated.Secret == nil {
		t.Fatal("secret should be preserved, not cleared")
	}
	got, err := encrypt.Decrypt(webhookTestKey, *updated.Secret)
	if err != nil || got != "original-secret" {
		t.Fatalf("secret changed unexpectedly: got %q, err %v", got, err)
	}
}

func TestWebhookUpdateEmptyNameRejected(t *testing.T) {
	repo := newMockWebhookRepo()
	s := NewWebhookService(repo, webhookTestKey)

	wh, err := s.Create(context.Background(), "alerts", "https://example.com/hook", "", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	empty := ""
	if _, err := s.Update(context.Background(), wh.ID, &empty, nil, nil, nil, nil); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error for empty name, got %v", err)
	}
}

func TestWebhookUpdateNonexistentNotFound(t *testing.T) {
	repo := newMockWebhookRepo()
	s := NewWebhookService(repo, webhookTestKey)

	enabled := true
	if _, err := s.Update(context.Background(), "does-not-exist", nil, nil, nil, nil, &enabled); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
