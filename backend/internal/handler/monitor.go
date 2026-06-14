package handler

import (
	"github.com/gofiber/fiber/v3"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/internal/service"
)

type MonitorHandler struct {
	monitors *service.MonitorService
}

func (h *MonitorHandler) List(c fiber.Ctx) error {
	serviceID := c.Query("service_id")
	if serviceID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "service_id query param required")
	}
	monitors, err := h.monitors.ListByService(c.Context(), serviceID)
	if err != nil {
		return err
	}
	return c.JSON(monitors)
}

func (h *MonitorHandler) Create(c fiber.Ctx) error {
	var body struct {
		ServiceID           string  `json:"service_id"`
		Name                string  `json:"name"`
		Type                string  `json:"type"`
		URL                 *string `json:"url"`
		Host                *string `json:"host"`
		Port                *int32  `json:"port"`
		IntervalSeconds     int32   `json:"interval_seconds"`
		TimeoutMs           int32   `json:"timeout_ms"`
		RetryCount          int32   `json:"retry_count"`
		DegradedThresholdMs *int32  `json:"degraded_threshold_ms"`
		HTTPMethod          *string `json:"http_method"`
		HTTPExpectedStatus  *int32  `json:"http_expected_status"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	mon, err := h.monitors.Create(c.Context(), domain.CreateMonitorParams{
		ServiceID:           body.ServiceID,
		Name:                body.Name,
		Type:                body.Type,
		URL:                 body.URL,
		Host:                body.Host,
		Port:                body.Port,
		IntervalSeconds:     body.IntervalSeconds,
		TimeoutMs:           body.TimeoutMs,
		RetryCount:          body.RetryCount,
		DegradedThresholdMs: body.DegradedThresholdMs,
		HTTPMethod:          body.HTTPMethod,
		HTTPExpectedStatus:  body.HTTPExpectedStatus,
	})
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(mon)
}

func (h *MonitorHandler) Get(c fiber.Ctx) error {
	mon, err := h.monitors.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(mon)
}

func (h *MonitorHandler) Update(c fiber.Ctx) error {
	var body struct {
		Name                string  `json:"name"`
		URL                 *string `json:"url"`
		Host                *string `json:"host"`
		Port                *int32  `json:"port"`
		IntervalSeconds     int32   `json:"interval_seconds"`
		TimeoutMs           int32   `json:"timeout_ms"`
		RetryCount          int32   `json:"retry_count"`
		DegradedThresholdMs *int32  `json:"degraded_threshold_ms"`
		HTTPMethod          *string `json:"http_method"`
		HTTPExpectedStatus  *int32  `json:"http_expected_status"`
		Enabled             bool    `json:"enabled"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	mon, err := h.monitors.Update(c.Context(), c.Params("id"), domain.UpdateMonitorParams{
		Name:                body.Name,
		URL:                 body.URL,
		Host:                body.Host,
		Port:                body.Port,
		IntervalSeconds:     body.IntervalSeconds,
		TimeoutMs:           body.TimeoutMs,
		RetryCount:          body.RetryCount,
		DegradedThresholdMs: body.DegradedThresholdMs,
		HTTPMethod:          body.HTTPMethod,
		HTTPExpectedStatus:  body.HTTPExpectedStatus,
		Enabled:             body.Enabled,
	})
	if err != nil {
		return err
	}
	return c.JSON(mon)
}

func (h *MonitorHandler) Archive(c fiber.Ctx) error {
	if err := h.monitors.Archive(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
