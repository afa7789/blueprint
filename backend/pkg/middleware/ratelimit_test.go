package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/afa/blueprint/backend/pkg/middleware"
)

func TestRateLimit_NilRedis_PassesThrough(t *testing.T) {
	app := fiber.New()
	cfg := middleware.RateLimitConfig{
		Max:     5,
		Window:  time.Minute,
		KeyFunc: middleware.KeyByIP,
	}
	app.Get("/", middleware.RateLimit(nil, cfg), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("request %d: expected 200 with nil redis, got %d", i+1, resp.StatusCode)
		}
	}
}

func TestRateLimit_Headers(t *testing.T) {
	// This test verifies that when Redis IS configured, the rate limit headers are set.
	// Since we cannot spin up a real Redis in unit tests here, we verify the nil-redis
	// path still passes through without headers (they're only set when redis is non-nil).
	// A separate integration test would cover the full Redis path.
	app := fiber.New()
	cfg := middleware.RateLimitConfig{
		Max:     10,
		Window:  time.Minute,
		KeyFunc: middleware.KeyByIP,
	}
	app.Get("/", middleware.RateLimit(nil, cfg), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	// With nil redis, rate-limit headers should not be set
	if v := resp.Header.Get("X-RateLimit-Limit"); v != "" {
		t.Fatalf("expected no X-RateLimit-Limit header with nil redis, got %q", v)
	}
}
