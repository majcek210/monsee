package v1

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/majcek210/monsee/internal/service"
)

type BadgeHandler struct {
	services *service.MonitoringService
}

func NewBadgeHandler(services *service.MonitoringService) *BadgeHandler {
	return &BadgeHandler{services: services}
}

// GetBadge returns an SVG status badge for a service.
func (h *BadgeHandler) GetBadge(c fiber.Ctx) error {
	svc, err := h.services.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	if !svc.PublicVisible {
		return fiber.NewError(fiber.StatusNotFound, "service not found")
	}

	status := svc.EffectiveStatus()
	color := statusColor(status)
	label := "status"
	value := status

	svg := buildBadgeSVG(label, value, color)
	c.Set("Content-Type", "image/svg+xml")
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	return c.SendString(svg)
}

func statusColor(status string) string {
	switch status {
	case "operational":
		return "#4c1"
	case "degraded":
		return "#fe7d37"
	case "outage", "down":
		return "#e05d44"
	default:
		return "#9f9f9f"
	}
}

func buildBadgeSVG(label, value, color string) string {
	lw := len(label)*7 + 10
	vw := len(value)*7 + 10
	tw := lw + vw
	lx := lw/2 + 1
	vx := lw + vw/2 - 1
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%d" height="20">
  <linearGradient id="s" x2="0" y2="100%%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <clipPath id="r"><rect width="%d" height="20" rx="3" fill="#fff"/></clipPath>
  <g clip-path="url(#r)">
    <rect width="%d" height="20" fill="#555"/>
    <rect x="%d" width="%d" height="20" fill="%s"/>
    <rect width="%d" height="20" fill="url(#s)"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
    <text x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="%d" y="14">%s</text>
    <text x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="%d" y="14">%s</text>
  </g>
</svg>`, tw, tw, lw, lw, vw, color, tw, lx, label, lx, label, vx, value, vx, value)
}
