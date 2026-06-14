package handler

import (
	"github.com/gofiber/fiber/v3"

	"github.com/majcek210/monsee/internal/middleware"
	"github.com/majcek210/monsee/internal/service"
)

type APIKeyHandler struct {
	apikeys *service.APIKeyService
}

func (h *APIKeyHandler) List(c fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)
	keys, err := h.apikeys.ListByUser(c.Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(keys)
}

func (h *APIKeyHandler) Create(c fiber.Ctx) error {
	var body struct {
		Name string `json:"name"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	userID := middleware.UserIDFromCtx(c)
	plain, key, err := h.apikeys.Generate(c.Context(), userID, body.Name)
	if err != nil {
		return err
	}

	// Return the plain key in this response only — never again
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":     key.ID,
		"name":   key.Name,
		"prefix": key.Prefix,
		"key":    plain,
	})
}

func (h *APIKeyHandler) Revoke(c fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)
	role := middleware.RoleFromCtx(c)
	if err := h.apikeys.Revoke(c.Context(), userID, role, c.Params("id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
