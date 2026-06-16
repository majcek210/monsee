package v1

import (
	"github.com/gofiber/fiber/v3"

	"github.com/majcek210/monsee/internal/service"
)


type IncidentHandler struct {
	incidents *service.IncidentService
}

func NewIncidentHandler(incidents *service.IncidentService) *IncidentHandler {
	return &IncidentHandler{incidents: incidents}
}

func (h *IncidentHandler) List(c fiber.Ctx) error {
	serviceID := c.Query("service_id")
	incs, err := h.incidents.List(c.Context(), serviceID)
	if err != nil {
		return err
	}
	return c.JSON(incs)
}

func (h *IncidentHandler) Get(c fiber.Ctx) error {
	inc, err := h.incidents.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	updates, err := h.incidents.ListUpdates(c.Context(), inc.ID)
	if err != nil {
		updates = nil
	}
	return c.JSON(fiber.Map{"incident": inc, "updates": updates})
}

func (h *IncidentHandler) ListUpdates(c fiber.Ctx) error {
	updates, err := h.incidents.ListUpdates(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(updates)
}
