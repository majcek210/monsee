package middleware

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Limiter is the interface the rate-limit middleware requires.
type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

// RateLimit returns a middleware that limits requests per IP.
//
// If the limiter backend (Redis) is unavailable, the request is allowed
// through (fail open) rather than taking down the public status page during
// a transient outage — but the condition is logged and counted via
// failOpenCounter so it stays visible/alertable. log and failOpenCounter may
// be nil.
func RateLimit(limiter Limiter, log *zap.Logger, failOpenCounter prometheus.Counter) fiber.Handler {
	return func(c fiber.Ctx) error {
		key := "rl:" + c.IP()
		allowed, err := limiter.Allow(c.Context(), key)
		if err != nil {
			if log != nil {
				log.Warn("rate limiter backend unavailable, failing open",
					zap.Error(err),
					zap.String("ip", c.IP()),
				)
			}
			if failOpenCounter != nil {
				failOpenCounter.Inc()
			}
			return c.Next()
		}
		if !allowed {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate limit exceeded",
			})
		}
		return c.Next()
	}
}
