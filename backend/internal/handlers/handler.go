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

// GetAll godoc
// @Summary     List all feature flags
// @Tags        Features
// @Produce     json
// @Success     200 {array} domain.FeatureFlag
// @Router      /features [get]
func (h *FeatureFlagHandler) GetAll(c *fiber.Ctx) error {
	flags, err := h.flags.GetAll(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch feature flags"})
	}
	return c.JSON(flags)
}

// Toggle godoc
// @Summary     Toggle a feature flag
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Param       key path string true "Feature flag key"
// @Param       body body object{enabled=boolean} true "Flag state"
// @Success     200 {object} map[string]interface{}
// @Security    BearerAuth
// @Router      /admin/features/{key} [put]
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
