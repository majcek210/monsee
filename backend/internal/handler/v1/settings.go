package v1

import (
	"github.com/gofiber/fiber/v3"
	"github.com/majcek210/monsee/internal/service"
)

type PublicSettingsHandler struct {
	settings *service.SettingsService
}

func NewPublicSettingsHandler(settings *service.SettingsService) *PublicSettingsHandler {
	return &PublicSettingsHandler{settings: settings}
}

// GetSettings returns non-sensitive public settings (site title, logo, etc.).
func (h *PublicSettingsHandler) GetSettings(c fiber.Ctx) error {
	s, err := h.settings.Get(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"site_title":             s.SiteTitle,
		"logo_url":               s.LogoURL,
		"custom_domains_enabled": s.CustomDomainsEnabled,
	})
}
