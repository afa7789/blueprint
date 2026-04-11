package middleware

import (
	"context"
	"slices"
	"strings"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/gofiber/fiber/v2"
)

func IsFeatureEnabled(ctx context.Context, repo domain.FeatureFlagRepository, key string) (bool, error) {
	if repo == nil {
		return true, nil
	}

	flags, err := repo.GetAll(ctx)
	if err != nil {
		return false, err
	}

	candidates := []string{key}
	if !strings.HasSuffix(key, "_enabled") {
		candidates = append(candidates, key+"_enabled")
	}

	for _, candidate := range candidates {
		idx := slices.IndexFunc(flags, func(flag domain.FeatureFlag) bool {
			return flag.Key == candidate
		})
		if idx >= 0 {
			return flags[idx].Enabled, nil
		}
	}

	return false, nil
}

func RequireFeature(repo domain.FeatureFlagRepository, key string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		enabled, err := IsFeatureEnabled(c.Context(), repo, key)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to check feature flag")
		}
		if !enabled {
			return fiber.NewError(fiber.StatusNotFound, "feature not available")
		}
		return c.Next()
	}
}

func RequireAnyFeature(repo domain.FeatureFlagRepository, keys ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		for _, key := range keys {
			enabled, err := IsFeatureEnabled(c.Context(), repo, key)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "failed to check feature flag")
			}
			if enabled {
				return c.Next()
			}
		}

		return fiber.NewError(fiber.StatusNotFound, "feature not available")
	}
}
