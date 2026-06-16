package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/internal/middleware"
	"github.com/majcek210/monsee/internal/service"
)

type TwoFactorHandler struct {
	tf *service.TwoFactorService
}

// InitiateSetup starts 2FA enrollment — returns TOTP secret + otpauth URI.
func (h *TwoFactorHandler) InitiateSetup(c fiber.Ctx) error {
	userID := middleware.UserIDFromCtx(c)
	secret, uri, err := h.tf.InitiateSetup(c.Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"secret": secret, "otpauth_uri": uri})
}

// ConfirmSetup verifies the first TOTP code and enables 2FA, returning backup codes.
func (h *TwoFactorHandler) ConfirmSetup(c fiber.Ctx) error {
	var body struct {
		Code string `json:"code"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return domain.ValidationErr("body", "invalid JSON")
	}
	userID := middleware.UserIDFromCtx(c)
	codes, err := h.tf.ConfirmSetup(c.Context(), userID, body.Code)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"backup_codes": codes})
}

// Disable turns off 2FA after verifying the user's password.
func (h *TwoFactorHandler) Disable(c fiber.Ctx) error {
	var body struct {
		Password string `json:"password"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return domain.ValidationErr("body", "invalid JSON")
	}
	userID := middleware.UserIDFromCtx(c)
	if err := h.tf.Disable(c.Context(), userID, body.Password); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Verify checks a TOTP code or backup code during login when 2FA is required.
func (h *TwoFactorHandler) Verify(c fiber.Ctx) error {
	var body struct {
		UserID string `json:"user_id"`
		Code   string `json:"code"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return domain.ValidationErr("body", "invalid JSON")
	}
	if body.UserID == "" || body.Code == "" {
		return domain.ValidationErr("code", "user_id and code are required")
	}
	ok, err := h.tf.Verify(c.Context(), body.UserID, body.Code)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ValidationErr("code", "invalid 2FA code")
	}
	return c.JSON(fiber.Map{"ok": true})
}
