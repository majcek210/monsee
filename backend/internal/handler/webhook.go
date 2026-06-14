package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/internal/service"
)

type WebhookHandler struct {
	webhooks *service.WebhookService
}

func (h *WebhookHandler) List(c fiber.Ctx) error {
	hooks, err := h.webhooks.List(c.Context())
	if err != nil {
		return err
	}
	safe := make([]fiber.Map, len(hooks))
	for i, hook := range hooks {
		safe[i] = safeWebhook(hook)
	}
	return c.JSON(safe)
}

func (h *WebhookHandler) Create(c fiber.Ctx) error {
	var body struct {
		Name   string   `json:"name"`
		URL    string   `json:"url"`
		Secret string   `json:"secret"`
		Events []string `json:"events"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	hook, err := h.webhooks.Create(c.Context(), body.Name, body.URL, body.Secret, body.Events)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(safeWebhook(hook))
}

func (h *WebhookHandler) Get(c fiber.Ctx) error {
	hook, err := h.webhooks.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(safeWebhook(hook))
}

func (h *WebhookHandler) Update(c fiber.Ctx) error {
	var body struct {
		Name    *string   `json:"name"`
		URL     *string   `json:"url"`
		Secret  *string   `json:"secret"`
		Events  *[]string `json:"events"`
		Enabled *bool     `json:"enabled"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	var events []string
	if body.Events != nil {
		events = *body.Events
	}

	hook, err := h.webhooks.Update(c.Context(), c.Params("id"), body.Name, body.URL, body.Secret, events, body.Enabled)
	if err != nil {
		return err
	}
	return c.JSON(safeWebhook(hook))
}

func (h *WebhookHandler) Archive(c fiber.Ctx) error {
	if err := h.webhooks.Archive(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *WebhookHandler) ListLogs(c fiber.Ctx) error {
	limitStr := c.Query("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	logs, err := h.webhooks.ListLogs(c.Context(), c.Params("id"), int32(limit))
	if err != nil {
		return err
	}
	return c.JSON(logs)
}

// safeWebhook omits the encrypted URL and secret from API responses.
func safeWebhook(w *domain.Webhook) fiber.Map {
	return fiber.Map{
		"id":         w.ID,
		"name":       w.Name,
		"events":     w.Events,
		"enabled":    w.Enabled,
		"created_at": w.CreatedAt,
	}
}
