package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/afa/blueprint/backend/internal/domain"
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
	defer func() { _ = resp.Body.Close() }()
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
	defer func() { _ = resp.Body.Close() }()
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
	defer func() { _ = resp.Body.Close() }()
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
	defer func() { _ = resp.Body.Close() }()
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
	defer func() { _ = resp.Body.Close() }()
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
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 403 {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

type mockFeatureFlagRepo struct {
	flags map[string]domain.FeatureFlag
	err   error
}

func (m *mockFeatureFlagRepo) GetAll(ctx context.Context) ([]domain.FeatureFlag, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]domain.FeatureFlag, 0, len(m.flags))
	for _, f := range m.flags {
		result = append(result, f)
	}
	return result, nil
}

func (m *mockFeatureFlagRepo) GetByKey(ctx context.Context, key string) (*domain.FeatureFlag, error) {
	if m.err != nil {
		return nil, m.err
	}
	if f, ok := m.flags[key]; ok {
		return &f, nil
	}
	return nil, nil
}

func (m *mockFeatureFlagRepo) Set(ctx context.Context, key string, enabled bool) error {
	if m.err != nil {
		return m.err
	}
	m.flags[key] = domain.FeatureFlag{Key: key, Enabled: enabled}
	return nil
}

func TestRequireFeature_NilRepo(t *testing.T) {
	app := fiber.New()
	app.Get("/", middleware.RequireFeature(nil, "some_feature"), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 (fail-open), got %d", resp.StatusCode)
	}
}

func TestRequireFeature_RepoError(t *testing.T) {
	repo := &mockFeatureFlagRepo{err: errors.New("db error")}
	app := fiber.New()
	app.Get("/", middleware.RequireFeature(repo, "some_feature"), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 500 {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func TestRequireFeature_Enabled(t *testing.T) {
	repo := &mockFeatureFlagRepo{
		flags: map[string]domain.FeatureFlag{"some_feature": {Key: "some_feature", Enabled: true}},
	}
	app := fiber.New()
	app.Get("/", middleware.RequireFeature(repo, "some_feature"), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRequireFeature_Disabled(t *testing.T) {
	repo := &mockFeatureFlagRepo{
		flags: map[string]domain.FeatureFlag{"some_feature": {Key: "some_feature", Enabled: false}},
	}
	app := fiber.New()
	app.Get("/", middleware.RequireFeature(repo, "some_feature"), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestRequireAnyFeature_OneEnabled(t *testing.T) {
	repo := &mockFeatureFlagRepo{
		flags: map[string]domain.FeatureFlag{
			"feature_a": {Key: "feature_a", Enabled: false},
			"feature_b": {Key: "feature_b", Enabled: true},
		},
	}
	app := fiber.New()
	app.Get("/", middleware.RequireAnyFeature(repo, "feature_a", "feature_b"), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRequireAnyFeature_NoneEnabled(t *testing.T) {
	repo := &mockFeatureFlagRepo{
		flags: map[string]domain.FeatureFlag{
			"feature_a": {Key: "feature_a", Enabled: false},
			"feature_b": {Key: "feature_b", Enabled: false},
		},
	}
	app := fiber.New()
	app.Get("/", middleware.RequireAnyFeature(repo, "feature_a", "feature_b"), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}
