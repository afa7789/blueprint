package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/afa/blueprint/backend/pkg/config"
	"github.com/afa/blueprint/backend/pkg/middleware"
)

func testCfg() *config.Config {
	return &config.Config{
		JWTSecret:     "test-secret-key",
		JWTExpiry:     15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Env:           "development",
	}
}

func makeToken(cfg *config.Config, userID, role string, expiry time.Duration) string {
	claims := middleware.Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(cfg.JWTSecret))
	return signed
}

func TestRequireAuth_ValidToken(t *testing.T) {
	cfg := testCfg()
	app := fiber.New()
	app.Get("/", middleware.RequireAuth(cfg), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	token := makeToken(cfg, "user-1", "user", time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	cfg := testCfg()
	app := fiber.New()
	app.Get("/", middleware.RequireAuth(cfg), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	token := makeToken(cfg, "user-1", "user", -time.Hour) // already expired
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	cfg := testCfg()
	app := fiber.New()
	app.Get("/", middleware.RequireAuth(cfg), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer this.is.garbage")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRequireAuth_NoToken(t *testing.T) {
	cfg := testCfg()
	app := fiber.New()
	app.Get("/", middleware.RequireAuth(cfg), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	cfg := testCfg()
	app := fiber.New()
	app.Get("/", middleware.RequireAuth(cfg), middleware.RequireRole("admin"), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	token := makeToken(cfg, "admin-1", "admin", time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	cfg := testCfg()
	app := fiber.New()
	app.Get("/", middleware.RequireAuth(cfg), middleware.RequireRole("admin"), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	token := makeToken(cfg, "user-1", "user", time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 403 {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}
