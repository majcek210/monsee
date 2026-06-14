package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/majcek210/monsee/internal/config"
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

func (h *UserHandler) UpdateRole(c fiber.Ctx) error {
	var body struct {
		Role string `json:"role"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	u, err := h.users.UpdateRole(c.Context(), c.Params("id"), body.Role)
	if err != nil {
		return err
	}
	return c.JSON(u)
}

func (h *UserHandler) Archive(c fiber.Ctx) error {
	if err := h.users.Archive(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
