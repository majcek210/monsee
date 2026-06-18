package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/majcek210/monsee/internal/config"
	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/internal/middleware"
	"github.com/majcek210/monsee/internal/service"
)

type UserHandler struct {
	users *service.UserService
	cfg   *config.Config
}

func (h *UserHandler) Login(c fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	u, err := h.users.Login(c.Context(), body.Email, body.Password)
	if err != nil {
		return err
	}

	// 2FA accounts must complete /auth/2fa/verify before receiving a session.
	// Return user_id so the frontend can submit the second-step code.
	if u.TOTPEnabled {
		return c.JSON(fiber.Map{"id": u.ID, "totp_required": true})
	}

	token, err := middleware.IssueToken(u.ID, u.Role, h.cfg.JWTSecret)
	if err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:     "session",
		Value:    token,
		HTTPOnly: true,
		Secure:   h.cfg.IsProd(),
		SameSite: "Lax",
		Expires:  time.Now().Add(middleware.SessionTTL),
	})

	return c.JSON(fiber.Map{"id": u.ID, "email": u.Email, "role": u.Role})
}

// Me returns the currently authenticated user, used by the frontend to
// determine the session's role (e.g. to hide admin-only nav links).
func (h *UserHandler) Me(c fiber.Ctx) error {
	u, err := h.users.GetByID(c.Context(), middleware.UserIDFromCtx(c))
	if err != nil {
		return err
	}
	return c.JSON(u)
}

func (h *UserHandler) Logout(c fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:    "session",
		Value:   "",
		Expires: time.Unix(0, 0),
	})
	return c.JSON(fiber.Map{"ok": true})
}

func (h *UserHandler) List(c fiber.Ctx) error {
	users, err := h.users.List(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(users)
}

func (h *UserHandler) Create(c fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	u, err := h.users.Register(c.Context(), body.Email, body.Password, body.Role)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(u)
}

// Update handles role changes (admin-only) and email/password changes
// (admin, or the user themself editing their own account). A non-admin
// request for any user other than themself is rejected; a non-admin request
// that includes "role" has that field silently dropped rather than rejected,
// so a viewer editing their own email/password in one request isn't forced
// to omit a role field they didn't intend to change in the first place.
func (h *UserHandler) Update(c fiber.Ctx) error {
	var body struct {
		Role     *string `json:"role"`
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	targetID := c.Params("id")
	callerID := middleware.UserIDFromCtx(c)
	isAdmin := middleware.RoleFromCtx(c) == "admin"
	isSelf := callerID == targetID

	if !isAdmin && !isSelf {
		return domain.Forbidden("you can only edit your own account")
	}

	var u *domain.User
	var err error

	if body.Email != nil || body.Password != nil {
		u, err = h.users.UpdateProfile(c.Context(), targetID, body.Email, body.Password)
		if err != nil {
			return err
		}
	}

	if body.Role != nil && isAdmin {
		u, err = h.users.UpdateRole(c.Context(), targetID, *body.Role)
		if err != nil {
			return err
		}
	}

	if u == nil {
		u, err = h.users.GetByID(c.Context(), targetID)
		if err != nil {
			return err
		}
	}
	return c.JSON(u)
}

// DisableTOTP force-disables 2FA on another user's account for lockout
// recovery. Admin-only — gated by RequireAdmin in the router.
func (h *UserHandler) DisableTOTP(c fiber.Ctx) error {
	if err := h.users.AdminDisableTOTP(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *UserHandler) Archive(c fiber.Ctx) error {
	if err := h.users.Archive(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
