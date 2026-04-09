package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/afa/blueprint/backend/internal/domain"
)

type SecurityHandler struct {
	settings domain.SecuritySettingRepository
}

func NewSecurityHandler(settings domain.SecuritySettingRepository) *SecurityHandler {
	return &SecurityHandler{settings: settings}
}

func (h *SecurityHandler) ListSettings(c *fiber.Ctx) error {
	settings, err := h.settings.GetAll(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch settings"})
	}
	return c.JSON(fiber.Map{"data": settings})
}

func (h *SecurityHandler) UpdateSetting(c *fiber.Ctx) error {
	key := c.Params("key")
	var req struct {
		Value string `json:"value"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := h.settings.Update(c.Context(), key, req.Value); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update setting"})
	}
	return c.JSON(fiber.Map{"key": key, "value": req.Value})
}
