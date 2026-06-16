package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/majcek210/monsee/internal/repository/postgres"
)

type AuditHandler struct {
	repo *postgres.AuditRepo
}

func (h *AuditHandler) List(c fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	var resource *string
	if r := c.Query("resource"); r != "" {
		resource = &r
	}
	var userID *string
	if u := c.Query("user_id"); u != "" {
		userID = &u
	}

	entries, total, err := h.repo.List(c.Context(), postgres.AuditListParams{
		Resource: resource,
		UserID:   userID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"entries": entries,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}
