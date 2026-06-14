package handler

import (
	"github.com/gofiber/fiber/v3"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/internal/service"
)

type NotificationHandler struct {
	notifs *service.NotificationService
}

func (h *NotificationHandler) List(c fiber.Ctx) error {
	channels, err := h.notifs.List(c.Context())
	if err != nil {
		return err
	}
	safe := make([]fiber.Map, len(channels))
	for i, ch := range channels {
		safe[i] = safeNotifChannel(ch)
	}
	return c.JSON(safe)
}

func (h *NotificationHandler) Create(c fiber.Ctx) error {
	var body struct {
		Name   string         `json:"name"`
		Type   string         `json:"type"`
		Config map[string]any `json:"config"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	ch, err := h.notifs.Create(c.Context(), body.Name, body.Type, body.Config)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(safeNotifChannel(ch))
}

func (h *NotificationHandler) Get(c fiber.Ctx) error {
	ch, err := h.notifs.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(safeNotifChannel(ch))
}

func (h *NotificationHandler) Update(c fiber.Ctx) error {
	var body struct {
		Name    *string        `json:"name"`
		Config  map[string]any `json:"config"`
		Enabled *bool          `json:"enabled"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	ch, err := h.notifs.Update(c.Context(), c.Params("id"), body.Name, body.Config, body.Enabled)
	if err != nil {
		return err
	}
	return c.JSON(safeNotifChannel(ch))
}

func (h *NotificationHandler) Archive(c fiber.Ctx) error {
	if err := h.notifs.Archive(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// safeNotifChannel strips the encrypted config from API responses.
func safeNotifChannel(ch *domain.NotificationChannel) fiber.Map {
	return fiber.Map{
		"id":         ch.ID,
		"name":       ch.Name,
		"type":       ch.Type,
		"enabled":    ch.Enabled,
		"created_at": ch.CreatedAt,
	}
}
