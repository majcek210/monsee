package service

import (
	"context"
	"sync"
	"time"

	"github.com/majcek210/monsee/internal/domain"
)

type SettingsService struct {
	repo     domain.SettingsRepository
	mu       sync.RWMutex
	cached   *domain.Settings
	cachedAt time.Time
	ttl      time.Duration
}

func NewSettingsService(repo domain.SettingsRepository) *SettingsService {
	return &SettingsService{repo: repo, ttl: 30 * time.Second}
}

func (s *SettingsService) Get(ctx context.Context) (*domain.Settings, error) {
	s.mu.RLock()
	if s.cached != nil && time.Since(s.cachedAt) < s.ttl {
		cached := *s.cached
		s.mu.RUnlock()
		return &cached, nil
	}
	s.mu.RUnlock()

	settings, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.cached = settings
	s.cachedAt = time.Now()
	s.mu.Unlock()

	return settings, nil
}

func (s *SettingsService) Update(ctx context.Context, p domain.UpdateSettingsParams) (*domain.Settings, error) {
	if p.SiteTitle != nil && *p.SiteTitle == "" {
		return nil, domain.ValidationErr("site_title", "site_title cannot be empty")
	}
	settings, err := s.repo.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.cached = nil
	s.mu.Unlock()
	return settings, nil
}
