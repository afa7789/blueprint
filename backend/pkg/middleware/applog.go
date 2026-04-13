package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/afa/blueprint/backend/internal/domain"
	applog "github.com/afa/blueprint/backend/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

func AppLog(repo domain.AppLogRepository) fiber.Handler {
	if repo == nil {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	logger := applog.New(repo, "http")

	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		path := c.Path()
		if path == "/healthz" || path == "/metrics" || strings.HasPrefix(path, "/api/v1/admin/logs/stream") {
			return err
		}

		status := c.Response().StatusCode()
		meta := map[string]interface{}{
			"method":      c.Method(),
			"path":        path,
			"status_code": status,
			"duration_ms": time.Since(start).Milliseconds(),
		}

		message := fmt.Sprintf("%s %s", c.Method(), path)
		switch {
		case status >= fiber.StatusInternalServerError:
			logger.Error(context.Background(), message, meta)
		case status >= fiber.StatusBadRequest:
			logger.Warn(context.Background(), message, meta)
		case c.Method() != fiber.MethodGet:
			logger.Info(context.Background(), message, meta)
		}

		return err
	}
}
