package handler

import (
	"github.com/gofiber/fiber/v3"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/internal/service"
)

type ServiceHandler struct {
	svc *service.MonitoringService
}

func (h *ServiceHandler) List(c fiber.Ctx) error {
	svcs, err := h.svc.List(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(svcs)
}

func (h *ServiceHandler) Create(c fiber.Ctx) error {
	var body struct {
		Name                 string  `json:"name"`
		Description          *string `json:"description"`
		PublicVisible        *bool   `json:"public_visible"`
		ShowUptime           *bool   `json:"show_uptime"`
		DedicatedPageEnabled *bool   `json:"dedicated_page_enabled"`
		Slug                 *string `json:"slug"`
		CustomDomain         *string `json:"custom_domain"`
		UptimeRangeDays      *int32  `json:"uptime_range_days"`
		StatusOverride       *string `json:"status_override"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	svc, err := h.svc.Create(c.Context(), domain.CreateServiceParams{
		Name:                 body.Name,
		Description:          body.Description,
		PublicVisible:        body.PublicVisible,
		ShowUptime:           body.ShowUptime,
		DedicatedPageEnabled: body.DedicatedPageEnabled,
		Slug:                 body.Slug,
		CustomDomain:         body.CustomDomain,
		UptimeRangeDays:      body.UptimeRangeDays,
		StatusOverride:       body.StatusOverride,
	})
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(svc)
}

func (h *ServiceHandler) Get(c fiber.Ctx) error {
	svc, err := h.svc.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(svc)
}

func (h *ServiceHandler) Update(c fiber.Ctx) error {
	var body struct {
		Name                 string  `json:"name"`
		Description          *string `json:"description"`
		PublicVisible        *bool   `json:"public_visible"`
		ShowUptime           *bool   `json:"show_uptime"`
		DedicatedPageEnabled *bool   `json:"dedicated_page_enabled"`
		Slug                 *string `json:"slug"`
		CustomDomain         *string `json:"custom_domain"`
		UptimeRangeDays      *int32  `json:"uptime_range_days"`
		StatusOverride       *string `json:"status_override"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	svc, err := h.svc.Update(c.Context(), c.Params("id"), domain.UpdateServiceParams{
		Name:                 body.Name,
		Description:          body.Description,
		PublicVisible:        body.PublicVisible,
		ShowUptime:           body.ShowUptime,
		DedicatedPageEnabled: body.DedicatedPageEnabled,
		Slug:                 body.Slug,
		CustomDomain:         body.CustomDomain,
		UptimeRangeDays:      body.UptimeRangeDays,
		StatusOverride:       body.StatusOverride,
	})
	if err != nil {
		return err
	}
	return c.JSON(svc)
}

func (h *ServiceHandler) Archive(c fiber.Ctx) error {
	if err := h.svc.Archive(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
