package v1

import (
	"github.com/gofiber/fiber/v3"
	"github.com/majcek210/monsee/internal/service"
)

type UptimeHandler struct {
	uptime   *service.UptimeService
	services *service.MonitoringService
}

func NewUptimeHandler(uptime *service.UptimeService, services *service.MonitoringService) *UptimeHandler {
	return &UptimeHandler{uptime: uptime, services: services}
}

// GetServiceUptime returns daily uptime for all monitors of a service.
func (h *UptimeHandler) GetServiceUptime(c fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.uptime.GetServiceUptime(c.Context(), id, nil)
	if err != nil {
		return err
	}
	return c.JSON(result)
}

// GetAllUptime returns uptime for every public service.
func (h *UptimeHandler) GetAllUptime(c fiber.Ctx) error {
	results, err := h.uptime.GetAllServicesUptime(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(results)
}

// GetMonitorLatency returns recent response-time data points for sparklines.
func (h *UptimeHandler) GetMonitorLatency(c fiber.Ctx) error {
	monitorID := c.Params("id")
	var hours int32 = 24
	results, err := h.uptime.GetMonitorLatency(c.Context(), monitorID, hours)
	if err != nil {
		return err
	}
	return c.JSON(results)
}

// GetPageBySlug returns the status page for a service by its slug.
func (h *UptimeHandler) GetPageBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	svc, err := h.services.GetBySlug(c.Context(), slug)
	if err != nil {
		return err
	}
	uptime, err := h.uptime.GetServiceUptime(c.Context(), svc.ID, nil)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"service": svc, "uptime": uptime})
}

// GetPageByDomain returns the status page for a custom domain (query param).
func (h *UptimeHandler) GetPageByDomain(c fiber.Ctx) error {
	domain := c.Query("domain")
	if domain == "" {
		return fiber.NewError(fiber.StatusBadRequest, "domain query param is required")
	}
	svc, err := h.services.GetByCustomDomain(c.Context(), domain)
	if err != nil {
		return err
	}
	uptime, err := h.uptime.GetServiceUptime(c.Context(), svc.ID, nil)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"service": svc, "uptime": uptime})
}
