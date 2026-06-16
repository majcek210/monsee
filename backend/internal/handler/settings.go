package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/internal/service"
)

type SettingsHandler struct {
	settings *service.SettingsService
}

func (h *SettingsHandler) Get(c fiber.Ctx) error {
	s, err := h.settings.Get(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(s)
}

func (h *SettingsHandler) Update(c fiber.Ctx) error {
	var body struct {
		SiteTitle           *string `json:"site_title"`
		LogoURL             *string `json:"logo_url"`
		PublicStatusEnabled *bool   `json:"public_status_enabled"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return domain.ValidationErr("body", "invalid JSON")
	}
	s, err := h.settings.Update(c.Context(), domain.UpdateSettingsParams{
		SiteTitle:           body.SiteTitle,
		LogoURL:             body.LogoURL,
		PublicStatusEnabled: body.PublicStatusEnabled,
	})
	if err != nil {
		return err
	}
	return c.JSON(s)
}
