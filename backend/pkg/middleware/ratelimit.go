package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	"github.com/afa/blueprint/backend/pkg/metrics"
)

type RateLimitConfig struct {
	Max     int           // max requests
	Window  time.Duration // time window
	KeyFunc func(*fiber.Ctx) string // how to identify the client
}

func RateLimit(rdb *redis.Client, cfg RateLimitConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if rdb == nil {
			return c.Next() // skip if redis not available
		}

		key := fmt.Sprintf("rl:%s", cfg.KeyFunc(c))
		ctx := context.Background()

		count, _ := rdb.Incr(ctx, key).Result()
		if count == 1 {
			rdb.Expire(ctx, key, cfg.Window)
		}

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Max))
		remaining := cfg.Max - int(count)
		if remaining < 0 {
			remaining = 0
		}
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		ttl, _ := rdb.TTL(ctx, key).Result()
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(ttl).Unix()))

		if int(count) > cfg.Max {
			keyType := "ip"
			if strings.HasPrefix(key, "rl:email:") {
				keyType = "email"
			}
			metrics.RateLimitHits.WithLabelValues(keyType).Inc()
			return c.Status(429).JSON(fiber.Map{
				"error":       "too many requests",
				"retry_after": int(ttl.Seconds()),
			})
		}

		return c.Next()
	}
}

// Preset key functions
func KeyByIP(c *fiber.Ctx) string {
	return "ip:" + c.IP()
}

func KeyByUserID(c *fiber.Ctx) string {
	uid, _ := c.Locals("user_id").(string)
	if uid != "" {
		return "uid:" + uid
	}
	return "ip:" + c.IP()
}

func KeyByEmail(c *fiber.Ctx) string {
	// Parse email from request body for login/register
	var body struct {
		Email string `json:"email"`
	}
	_ = c.BodyParser(&body)
	if body.Email != "" {
		return "email:" + body.Email
	}
	return "ip:" + c.IP()
}
