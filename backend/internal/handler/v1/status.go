package v1

import (
	"github.com/gofiber/fiber/v3"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/internal/service"
)

type StatusHandler struct {
	services *service.MonitoringService
	monitors *service.MonitorService
}

func NewStatusHandler(services *service.MonitoringService, monitors *service.MonitorService) *StatusHandler {
	return &StatusHandler{services: services, monitors: monitors}
}

type serviceStatus struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description *string          `json:"description"`
	Status      string           `json:"status"`
	Monitors    []*monitorStatus `json:"monitors"`
}

type monitorStatus struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

func (h *StatusHandler) GetStatus(c fiber.Ctx) error {
	svcs, err := h.services.ListPublic(c.Context())
	if err != nil {
		return err
	}

	result := make([]*serviceStatus, 0, len(svcs))
	for _, svc := range svcs {
		monitors, err := h.monitors.ListByService(c.Context(), svc.ID)
		if err != nil {
			monitors = []*domain.Monitor{}
		}

		ms := make([]*monitorStatus, len(monitors))
		for i, m := range monitors {
			ms[i] = &monitorStatus{
				ID:      m.ID,
				Name:    m.Name,
				Type:    m.Type,
				Enabled: m.Enabled,
			}
		}

		result = append(result, &serviceStatus{
			ID:          svc.ID,
			Name:        svc.Name,
			Description: svc.Description,
			Status:      svc.EffectiveStatus(),
			Monitors:    ms,
		})
	}

	return c.JSON(result)
}
