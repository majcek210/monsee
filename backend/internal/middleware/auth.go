package middleware

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/majcek210/monsee/internal/domain"
)

// SessionTTL is how long an issued session token is valid — must match the
// "session" cookie's Expires duration set on login.
const SessionTTL = 24 * time.Hour

const (
	cookieName  = "session"
	claimUserID = "user_id"
	claimRole   = "role"
)

// Claims holds the JWT payload stored in the session cookie.
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// RequireAuth verifies the session cookie and sets "userID" + "role" locals.
func RequireAuth(secret string) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := c.Cookies(cookieName)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		claims := &Claims{}
		_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		c.Locals(claimUserID, claims.UserID)
		c.Locals(claimRole, claims.Role)
		return c.Next()
	}
}

// RequireAdmin allows only admin role through.
func RequireAdmin(c fiber.Ctx) error {
	if c.Locals(claimRole) != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}
	return c.Next()
}

// UserIDFromCtx extracts the user_id local set by RequireAuth.
func UserIDFromCtx(c fiber.Ctx) string {
	id, _ := c.Locals(claimUserID).(string)
	return id
}

// RoleFromCtx extracts the role local.
func RoleFromCtx(c fiber.Ctx) string {
	role, _ := c.Locals(claimRole).(string)
	return role
}

// IssueToken creates a signed JWT for the given user, valid for SessionTTL —
// matching the "session" cookie's expiry so a leaked token can't outlive it.
func IssueToken(userID, role, secret string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(SessionTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ErrForbidden is exposed so handlers can use it without importing domain.
var ErrForbidden = domain.ErrForbidden
