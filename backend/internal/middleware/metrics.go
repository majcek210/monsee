package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusMiddleware records HTTP request duration for every route.
func PrometheusMiddleware(hist *prometheus.HistogramVec) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		elapsed := time.Since(start).Seconds()

		status := fmt.Sprintf("%d", c.Response().StatusCode())
		route := c.Route().Path
		if route == "" {
			route = "unknown"
		}

		hist.WithLabelValues(route, c.Method(), status).Observe(elapsed)
		return err
	}
}
