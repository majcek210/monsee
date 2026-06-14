package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-at-least-32-bytes-long"

func TestIssueTokenClaims(t *testing.T) {
	before := time.Now()
	tok, err := IssueToken("user-123", "admin", testSecret)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}

	claims := &Claims{}
	_, err = jwt.ParseWithClaims(tok, claims, func(t *jwt.Token) (any, error) {
		return []byte(testSecret), nil
	})
	if err != nil {
		t.Fatalf("ParseWithClaims: %v", err)
	}

	if claims.UserID != "user-123" || claims.Role != "admin" {
		t.Fatalf("unexpected claims: %+v", claims)
	}

	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		t.Fatal("expected IssuedAt and ExpiresAt to be set")
	}

	if diff := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time); diff != SessionTTL {
		t.Fatalf("expiry window = %v, want %v", diff, SessionTTL)
	}

	// jwt.NumericDate truncates to whole seconds, so allow a small tolerance.
	wantExpiry := before.Add(SessionTTL)
	if diff := wantExpiry.Sub(claims.ExpiresAt.Time); diff < -time.Second || diff > time.Second {
		t.Fatalf("ExpiresAt %v too far from expected %v (diff %v)", claims.ExpiresAt.Time, wantExpiry, diff)
	}
}

// newAuthTestApp wires a minimal app exercising RequireAuth + RequireAdmin,
// echoing back the userID/role placed in locals by RequireAuth.
func newAuthTestApp() *fiber.App {
	app := fiber.New()

	app.Get("/whoami", RequireAuth(testSecret), func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id": UserIDFromCtx(c),
			"role":    RoleFromCtx(c),
		})
	})

	app.Get("/admin-only", RequireAuth(testSecret), RequireAdmin, func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	return app
}

func TestRequireAuthValidToken(t *testing.T) {
	app := newAuthTestApp()

	tok, err := IssueToken("user-123", "viewer", testSecret)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: tok})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRequireAuthNoCookie(t *testing.T) {
	app := newAuthTestApp()

	req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestRequireAuthInvalidToken(t *testing.T) {
	app := newAuthTestApp()

	req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "not-a-valid-jwt"})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestRequireAuthWrongSecretRejected(t *testing.T) {
	app := newAuthTestApp()

	tok, err := IssueToken("user-123", "admin", "a-completely-different-secret")
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: tok})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestRequireAuthExpiredTokenRejected(t *testing.T) {
	app := newAuthTestApp()

	now := time.Now()
	claims := &Claims{
		UserID: "user-123",
		Role:   "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * SessionTTL)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
		},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: tok})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 for expired token", resp.StatusCode)
	}
}

func TestRequireAdminAllowsAdmin(t *testing.T) {
	app := newAuthTestApp()

	tok, err := IssueToken("user-123", "admin", testSecret)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: tok})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want 200 for admin", resp.StatusCode)
	}
}

func TestRequireAdminRejectsViewer(t *testing.T) {
	app := newAuthTestApp()

	tok, err := IssueToken("user-123", "viewer", testSecret)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: tok})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status = %d, want 403 for viewer", resp.StatusCode)
	}
}
