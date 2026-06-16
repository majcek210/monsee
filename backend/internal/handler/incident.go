package handler

import (
	"github.com/gofiber/fiber/v3"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/internal/service"
)

type IncidentHandler struct {
	incidents *service.IncidentService
}

func (h *IncidentHandler) List(c fiber.Ctx) error {
	serviceID := c.Query("service_id")
	incs, err := h.incidents.List(c.Context(), serviceID)
	if err != nil {
		return err
	}
	return c.JSON(incs)
}

func (h *IncidentHandler) Create(c fiber.Ctx) error {
	var body struct {
		ServiceID string  `json:"service_id"`
		MonitorID *string `json:"monitor_id"`
		Title     string  `json:"title"`
		Severity  string  `json:"severity"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	inc, err := h.incidents.Create(c.Context(), domain.CreateIncidentParams{
		ServiceID: body.ServiceID,
		MonitorID: body.MonitorID,
		Title:     body.Title,
		Severity:  body.Severity,
	})
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(inc)
}

func (h *IncidentHandler) Get(c fiber.Ctx) error {
	inc, err := h.incidents.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(inc)
}

func (h *IncidentHandler) Update(c fiber.Ctx) error {
	var body struct {
		Title    string `json:"title"`
		Severity string `json:"severity"`
		Status   string `json:"status"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	inc, err := h.incidents.Update(c.Context(), c.Params("id"), domain.UpdateIncidentParams{
		Title:    body.Title,
		Severity: body.Severity,
		Status:   body.Status,
	})
	if err != nil {
		return err
	}
	return c.JSON(inc)
}

func (h *IncidentHandler) Resolve(c fiber.Ctx) error {
	inc, err := h.incidents.Resolve(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(inc)
}

func (h *IncidentHandler) PostUpdate(c fiber.Ctx) error {
	var body struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	update, err := h.incidents.PostUpdate(c.Context(), c.Params("id"), body.Status, body.Message)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(update)
}

func (h *IncidentHandler) ListUpdates(c fiber.Ctx) error {
	updates, err := h.incidents.ListUpdates(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(updates)
}
