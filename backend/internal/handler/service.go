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
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	svc, err := h.svc.Create(c.Context(), domain.CreateServiceParams{
		Name:        body.Name,
		Description: body.Description,
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
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	svc, err := h.svc.Update(c.Context(), c.Params("id"), domain.UpdateServiceParams{
		Name:        body.Name,
		Description: body.Description,
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
