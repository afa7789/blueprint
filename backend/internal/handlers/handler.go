package handlers

import (
	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/gofiber/fiber/v2"
)

type FeatureFlagHandler struct {
	flags domain.FeatureFlagRepository
}

func NewFeatureFlagHandler(flags domain.FeatureFlagRepository) *FeatureFlagHandler {
	return &FeatureFlagHandler{flags: flags}
}

func (h *FeatureFlagHandler) GetAll(c *fiber.Ctx) error {
	flags, err := h.flags.GetAll(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch feature flags"})
	}
	return c.JSON(flags)
}

func (h *FeatureFlagHandler) Toggle(c *fiber.Ctx) error {
	key := c.Params("key")
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := h.flags.Set(c.Context(), key, req.Enabled); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update feature flag"})
	}
	return c.JSON(fiber.Map{"key": key, "enabled": req.Enabled})
}
