package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/pkg/encrypt"
	"github.com/majcek210/monsee/pkg/netguard"
)

type NotificationService struct {
	channels domain.NotificationChannelRepository
	encKey   []byte
}

func NewNotificationService(channels domain.NotificationChannelRepository, encKey []byte) *NotificationService {
	return &NotificationService{channels: channels, encKey: encKey}
}

func (s *NotificationService) Create(ctx context.Context, name, typ string, config map[string]any) (*domain.NotificationChannel, error) {
	if name == "" {
		return nil, domain.ValidationErr("name", "name is required")
	}
	if typ != "discord" && typ != "email" {
		return nil, domain.ValidationErr("type", "type must be discord or email")
	}
	if err := validateChannelConfig(typ, config); err != nil {
		return nil, err
	}

	encCfg, err := s.encryptConfig(config)
	if err != nil {
		return nil, fmt.Errorf("encrypt config: %w", err)
	}

	return s.channels.Create(ctx, domain.CreateNotificationChannelParams{
		Name:   name,
		Type:   typ,
		Config: encCfg,
	})
}

func (s *NotificationService) GetByID(ctx context.Context, id string) (*domain.NotificationChannel, error) {
	return s.channels.GetByID(ctx, id)
}

func (s *NotificationService) List(ctx context.Context) ([]*domain.NotificationChannel, error) {
	return s.channels.List(ctx)
}

// Update applies a partial update. name == nil or config == nil leaves the
// corresponding column unchanged; enabled == nil leaves enabled unchanged.
func (s *NotificationService) Update(ctx context.Context, id string, name *string, config map[string]any, enabled *bool) (*domain.NotificationChannel, error) {
	channel, err := s.channels.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	p := domain.UpdateNotificationChannelParams{
		Enabled: enabled,
	}

	if name != nil {
		if *name == "" {
			return nil, domain.ValidationErr("name", "name is required")
		}
		p.Name = name
	}

	if config != nil {
		if err := validateChannelConfig(channel.Type, config); err != nil {
			return nil, err
		}
		encCfg, err := s.encryptConfig(config)
		if err != nil {
			return nil, fmt.Errorf("encrypt config: %w", err)
		}
		p.Config = &encCfg
	}

	return s.channels.Update(ctx, id, p)
}

func (s *NotificationService) Archive(ctx context.Context, id string) error {
	return s.channels.Archive(ctx, id)
}

func (s *NotificationService) encryptConfig(config map[string]any) (string, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return encrypt.Encrypt(s.encKey, string(raw))
}

// validateChannelConfig applies type-specific validation to a channel config.
// For discord channels, webhook_url must be a public http(s) URL (SSRF guard,
// IP-literal check only — no DNS lookup at config time). The full
// DNS-resolving check is re-applied at send time in SendDiscord, which is the
// real security boundary.
func validateChannelConfig(typ string, config map[string]any) error {
	if typ != "discord" {
		return nil
	}
	url, _ := config["webhook_url"].(string)
	if url == "" {
		return domain.ValidationErr("config.webhook_url", "webhook_url is required for discord channels")
	}
	if err := netguard.CheckPublicURLSyntax(url); err != nil {
		return domain.ValidationErr("config.webhook_url", "webhook_url must be a public http(s) address: "+err.Error())
	}
	return nil
}
