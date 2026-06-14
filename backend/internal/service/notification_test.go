package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/pkg/encrypt"
)

var notifTestKey = []byte("01234567890123456789012345678901")[:32]

// mockNotificationRepo is an in-memory domain.NotificationChannelRepository
// for unit tests. Update applies the same COALESCE-on-nil semantics as the
// real SQL query.
type mockNotificationRepo struct {
	mu    sync.Mutex
	items map[string]*domain.NotificationChannel
	seq   int
}

func newMockNotificationRepo() *mockNotificationRepo {
	return &mockNotificationRepo{items: map[string]*domain.NotificationChannel{}}
}

func (m *mockNotificationRepo) Create(ctx context.Context, p domain.CreateNotificationChannelParams) (*domain.NotificationChannel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	c := &domain.NotificationChannel{
		ID:        fmt.Sprintf("ch-%d", m.seq),
		Name:      p.Name,
		Type:      p.Type,
		Config:    p.Config,
		Enabled:   true,
		CreatedAt: time.Now(),
	}
	m.items[c.ID] = c
	cp := *c
	return &cp, nil
}

func (m *mockNotificationRepo) GetByID(ctx context.Context, id string) (*domain.NotificationChannel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.items[id]
	if !ok {
		return nil, domain.NotFound("notification channel not found")
	}
	cp := *c
	return &cp, nil
}

func (m *mockNotificationRepo) List(ctx context.Context) ([]*domain.NotificationChannel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*domain.NotificationChannel
	for _, c := range m.items {
		cp := *c
		out = append(out, &cp)
	}
	return out, nil
}

func (m *mockNotificationRepo) ListEnabled(ctx context.Context) ([]*domain.NotificationChannel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*domain.NotificationChannel
	for _, c := range m.items {
		if c.Enabled {
			cp := *c
			out = append(out, &cp)
		}
	}
	return out, nil
}

// Update mirrors the UpdateNotificationChannel SQL query: any nil field in p
// leaves the existing column value unchanged.
func (m *mockNotificationRepo) Update(ctx context.Context, id string, p domain.UpdateNotificationChannelParams) (*domain.NotificationChannel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.items[id]
	if !ok {
		return nil, domain.NotFound("notification channel not found")
	}
	if p.Name != nil {
		c.Name = *p.Name
	}
	if p.Config != nil {
		c.Config = *p.Config
	}
	if p.Enabled != nil {
		c.Enabled = *p.Enabled
	}
	cp := *c
	return &cp, nil
}

func (m *mockNotificationRepo) Archive(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.items[id]
	if !ok {
		return domain.NotFound("notification channel not found")
	}
	now := time.Now()
	c.ArchivedAt = &now
	return nil
}

func TestNotificationCreateValidation(t *testing.T) {
	repo := newMockNotificationRepo()
	s := NewNotificationService(repo, notifTestKey)

	if _, err := s.Create(context.Background(), "", "discord", map[string]any{"webhook_url": "https://discord.com/x"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error for empty name, got %v", err)
	}
	if _, err := s.Create(context.Background(), "alerts", "slack", map[string]any{}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error for unsupported type, got %v", err)
	}
}

func TestNotificationCreateEncryptsConfig(t *testing.T) {
	repo := newMockNotificationRepo()
	s := NewNotificationService(repo, notifTestKey)

	cfg := map[string]any{"webhook_url": "https://discord.com/api/webhooks/secret"}
	ch, err := s.Create(context.Background(), "discord-alerts", "discord", cfg)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	raw, err := encrypt.Decrypt(notifTestKey, ch.Config)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("unmarshal decrypted config: %v", err)
	}
	if got["webhook_url"] != cfg["webhook_url"] {
		t.Fatalf("config did not round-trip: got %v", got)
	}
}

// TestNotificationUpdateToggleEnabledPreservesConfig is the regression test
// for Tests.md F1 / Security.md S4 for notification channels: toggling
// "enabled" alone must not wipe out the encrypted config (e.g. SMTP creds,
// Discord webhook URL).
func TestNotificationUpdateToggleEnabledPreservesConfig(t *testing.T) {
	repo := newMockNotificationRepo()
	s := NewNotificationService(repo, notifTestKey)

	ch, err := s.Create(context.Background(), "discord-alerts", "discord", map[string]any{"webhook_url": "https://discord.com/api/webhooks/secret"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	disabled := false
	updated, err := s.Update(context.Background(), ch.ID, nil, nil, &disabled)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if updated.Enabled {
		t.Fatal("enabled should now be false")
	}
	if updated.Config != ch.Config {
		t.Fatal("config was lost/changed by an enabled-only update")
	}
	if updated.Name != ch.Name {
		t.Fatal("name changed unexpectedly")
	}
}

func TestNotificationUpdateConfigReEncrypts(t *testing.T) {
	repo := newMockNotificationRepo()
	s := NewNotificationService(repo, notifTestKey)

	ch, err := s.Create(context.Background(), "discord-alerts", "discord", map[string]any{"webhook_url": "https://discord.com/old"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	newCfg := map[string]any{"webhook_url": "https://discord.com/new"}
	updated, err := s.Update(context.Background(), ch.ID, nil, newCfg, nil)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	raw, err := encrypt.Decrypt(notifTestKey, updated.Config)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["webhook_url"] != "https://discord.com/new" {
		t.Fatalf("config not updated: got %v", got)
	}
}

func TestNotificationUpdateEmptyNameRejected(t *testing.T) {
	repo := newMockNotificationRepo()
	s := NewNotificationService(repo, notifTestKey)

	ch, err := s.Create(context.Background(), "discord-alerts", "discord", map[string]any{"webhook_url": "https://discord.com/x"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	empty := ""
	if _, err := s.Update(context.Background(), ch.ID, &empty, nil, nil); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error for empty name, got %v", err)
	}
}

func TestNotificationUpdateNonexistentNotFound(t *testing.T) {
	repo := newMockNotificationRepo()
	s := NewNotificationService(repo, notifTestKey)

	enabled := true
	if _, err := s.Update(context.Background(), "does-not-exist", nil, nil, &enabled); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
