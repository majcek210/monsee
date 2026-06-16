package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/internal/service"
)

type MaintenanceHandler struct {
	maintenance *service.MaintenanceService
}

func (h *MaintenanceHandler) List(c fiber.Ctx) error {
	serviceID := c.Query("service_id")
	var windows []*domain.MaintenanceWindow
	var err error
	if serviceID != "" {
		windows, err = h.maintenance.ListByService(c.Context(), serviceID)
	} else {
		windows, err = h.maintenance.ListAll(c.Context())
	}
	if err != nil {
		return err
	}
	return c.JSON(windows)
}

func (h *MaintenanceHandler) Create(c fiber.Ctx) error {
	var body struct {
		ServiceID   string  `json:"service_id"`
		Title       string  `json:"title"`
		Description *string `json:"description"`
		StartsAt    string  `json:"starts_at"`
		EndsAt      string  `json:"ends_at"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return domain.ValidationErr("body", "invalid JSON")
	}
	startsAt, err := time.Parse(time.RFC3339, body.StartsAt)
	if err != nil {
		return domain.ValidationErr("starts_at", "invalid RFC3339 date")
	}
	endsAt, err := time.Parse(time.RFC3339, body.EndsAt)
	if err != nil {
		return domain.ValidationErr("ends_at", "invalid RFC3339 date")
	}
	mw, err := h.maintenance.Create(c.Context(), domain.CreateMaintenanceWindowParams{
		ServiceID:   body.ServiceID,
		Title:       body.Title,
		Description: body.Description,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
	})
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(mw)
}

func (h *MaintenanceHandler) Get(c fiber.Ctx) error {
	mw, err := h.maintenance.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(mw)
}

func (h *MaintenanceHandler) Update(c fiber.Ctx) error {
	var body struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		StartsAt    *string `json:"starts_at"`
		EndsAt      *string `json:"ends_at"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return domain.ValidationErr("body", "invalid JSON")
	}
	p := domain.UpdateMaintenanceWindowParams{
		Title:       body.Title,
		Description: body.Description,
	}
	if body.StartsAt != nil {
		t, err := time.Parse(time.RFC3339, *body.StartsAt)
		if err != nil {
			return domain.ValidationErr("starts_at", "invalid RFC3339 date")
		}
		p.StartsAt = &t
	}
	if body.EndsAt != nil {
		t, err := time.Parse(time.RFC3339, *body.EndsAt)
		if err != nil {
			return domain.ValidationErr("ends_at", "invalid RFC3339 date")
		}
		p.EndsAt = &t
	}
	mw, err := h.maintenance.Update(c.Context(), c.Params("id"), p)
	if err != nil {
		return err
	}
	return c.JSON(mw)
}

func (h *MaintenanceHandler) Archive(c fiber.Ctx) error {
	if err := h.maintenance.Archive(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
