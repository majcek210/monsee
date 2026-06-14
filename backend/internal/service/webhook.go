package service

import (
	"context"
	"fmt"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/pkg/encrypt"
	"github.com/majcek210/monsee/pkg/netguard"
)

type WebhookService struct {
	webhooks domain.WebhookRepository
	encKey   []byte
}

func NewWebhookService(webhooks domain.WebhookRepository, encKey []byte) *WebhookService {
	return &WebhookService{webhooks: webhooks, encKey: encKey}
}

func (s *WebhookService) Create(ctx context.Context, name, rawURL, rawSecret string, events []string) (*domain.Webhook, error) {
	if name == "" {
		return nil, domain.ValidationErr("name", "name is required")
	}
	if rawURL == "" {
		return nil, domain.ValidationErr("url", "url is required")
	}
	if err := netguard.CheckPublicURLSyntax(rawURL); err != nil {
		return nil, domain.ValidationErr("url", "url must be a public http(s) address: "+err.Error())
	}

	encURL, err := encrypt.Encrypt(s.encKey, rawURL)
	if err != nil {
		return nil, fmt.Errorf("encrypt url: %w", err)
	}

	var encSecret *string
	if rawSecret != "" {
		s, err := encrypt.Encrypt(s.encKey, rawSecret)
		if err != nil {
			return nil, fmt.Errorf("encrypt secret: %w", err)
		}
		encSecret = &s
	}

	return s.webhooks.Create(ctx, domain.CreateWebhookParams{
		Name:   name,
		URL:    encURL,
		Secret: encSecret,
		Events: events,
	})
}

func (s *WebhookService) GetByID(ctx context.Context, id string) (*domain.Webhook, error) {
	return s.webhooks.GetByID(ctx, id)
}

func (s *WebhookService) List(ctx context.Context) ([]*domain.Webhook, error) {
	return s.webhooks.List(ctx)
}

// Update applies a partial update. Any of name/rawURL/rawSecret/enabled that
// is nil leaves the corresponding column unchanged; events == nil leaves the
// event list unchanged (pass an empty, non-nil slice to clear it).
func (s *WebhookService) Update(ctx context.Context, id string, name, rawURL, rawSecret *string, events []string, enabled *bool) (*domain.Webhook, error) {
	if _, err := s.webhooks.GetByID(ctx, id); err != nil {
		return nil, err
	}

	p := domain.UpdateWebhookParams{
		Events:  events,
		Enabled: enabled,
	}

	if name != nil {
		if *name == "" {
			return nil, domain.ValidationErr("name", "name is required")
		}
		p.Name = name
	}

	if rawURL != nil {
		if *rawURL == "" {
			return nil, domain.ValidationErr("url", "url is required")
		}
		if err := netguard.CheckPublicURLSyntax(*rawURL); err != nil {
			return nil, domain.ValidationErr("url", "url must be a public http(s) address: "+err.Error())
		}
		encURL, err := encrypt.Encrypt(s.encKey, *rawURL)
		if err != nil {
			return nil, fmt.Errorf("encrypt url: %w", err)
		}
		p.URL = &encURL
	}

	if rawSecret != nil && *rawSecret != "" {
		encSecret, err := encrypt.Encrypt(s.encKey, *rawSecret)
		if err != nil {
			return nil, fmt.Errorf("encrypt secret: %w", err)
		}
		p.Secret = &encSecret
	}

	return s.webhooks.Update(ctx, id, p)
}

func (s *WebhookService) Archive(ctx context.Context, id string) error {
	return s.webhooks.Archive(ctx, id)
}

func (s *WebhookService) ListLogs(ctx context.Context, webhookID string, limit int32) ([]*domain.WebhookLog, error) {
	return s.webhooks.ListLogs(ctx, webhookID, limit)
}
