package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// APIKeyValidator is a function that validates a raw API key and returns the owner's user ID.
type APIKeyValidator func(ctx context.Context, rawKey string) (userID string, err error)

// RequireAPIKey validates the Bearer token in the Authorization header.
// On success it sets the "userID" local (same key used by RequireAuth).
func RequireAPIKey(validate APIKeyValidator) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		raw := strings.TrimPrefix(header, "Bearer ")
		if !strings.HasPrefix(raw, "sk_") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		userID, err := validate(c.Context(), raw)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		c.Locals(claimUserID, userID)
		return c.Next()
	}
}
