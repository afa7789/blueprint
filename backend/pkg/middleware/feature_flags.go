package middleware

import (
	"context"
	"strings"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/gofiber/fiber/v2"
)

func IsFeatureEnabled(ctx context.Context, repo domain.FeatureFlagRepository, key string) (bool, error) {
	if repo == nil {
		return true, nil // fail-open: avoid blocking features in edge/test scenarios
	}

	flag, err := repo.GetByKey(ctx, key)
	if err != nil {
		return false, err
	}
	if flag != nil {
		return flag.Enabled, nil
	}

	if !strings.HasSuffix(key, "_enabled") {
		flag, err := repo.GetByKey(ctx, key+"_enabled")
		if err != nil {
			return false, err
		}
		if flag != nil {
			return flag.Enabled, nil
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
