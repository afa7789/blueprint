package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/afa/blueprint/backend/pkg/metrics"
)

func PrometheusMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())
		method := c.Method()

		// Normalize path to avoid high cardinality (replace UUIDs with :id)
		path := normalizePath(c.Route().Path)

		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)

		return err
	}
}

func normalizePath(path string) string {
	if path == "" {
		return "unknown"
	}
	return path
}
