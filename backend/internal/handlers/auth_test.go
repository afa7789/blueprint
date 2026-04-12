package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/afa/blueprint/backend/internal/handlers"
	"github.com/afa/blueprint/backend/internal/testutil"
	"github.com/afa/blueprint/backend/pkg/config"
	"github.com/afa/blueprint/backend/pkg/middleware"
)

func testConfig() *config.Config {
	return &config.Config{
		JWTSecret:     "test-secret-key",
		JWTExpiry:     15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Env:           "development",
		BcryptCost:    4, // bcrypt.MinCost — keep tests fast
	}
}

// testTimeout is passed to (*fiber.App).Test so bcrypt-heavy auth flows don't
// hit the default 1s timeout under `-race`.
const testTimeout = -1

func setupAuthApp() (*fiber.App, *handlers.AuthHandler) {
	app := fiber.New()
	userRepo := testutil.NewMockUserRepo()
	flagRepo := testutil.NewMockFeatureFlagRepo()
	cfg := testConfig()
	h := handlers.NewAuthHandler(userRepo, flagRepo, nil, cfg)

	app.Post("/register", h.Register)
	app.Post("/login", h.Login)
	app.Post("/logout", h.Logout)
	app.Post("/refresh", h.Refresh)
	app.Get("/me", middleware.RequireAuth(cfg), h.Me)

	return app, h
}

func jsonBody(v any) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

func TestRegister_Success(t *testing.T) {
	app, _ := setupAuthApp()

	req := httptest.NewRequest(http.MethodPost, "/register", jsonBody(map[string]string{
		"email":    "alice@example.com",
		"password": "secret123",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, testTimeout)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	if body["access_token"] == nil {
		t.Fatal("expected access_token in response")
	}
	if body["user"] == nil {
		t.Fatal("expected user in response")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	app, _ := setupAuthApp()

	payload := jsonBody(map[string]string{"email": "bob@example.com", "password": "secret123"})
	req := httptest.NewRequest(http.MethodPost, "/register", payload)
	req.Header.Set("Content-Type", "application/json")
	if _, err := app.Test(req, testTimeout); err != nil {
		t.Fatal(err)
	}

	payload2 := jsonBody(map[string]string{"email": "bob@example.com", "password": "other"})
	req2 := httptest.NewRequest(http.MethodPost, "/register", payload2)
	req2.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req2, testTimeout)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 409 {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestRegister_MissingFields(t *testing.T) {
	app, _ := setupAuthApp()

	cases := []map[string]string{
		{"email": "", "password": "secret"},
		{"email": "x@x.com", "password": ""},
		{},
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodPost, "/register", jsonBody(c))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, testTimeout)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 400 {
			t.Fatalf("expected 400 for %v, got %d", c, resp.StatusCode)
		}
	}
}

func TestLogin_Success(t *testing.T) {
	app, _ := setupAuthApp()

	// Register first
	req := httptest.NewRequest(http.MethodPost, "/register", jsonBody(map[string]string{
		"email": "carol@example.com", "password": "pass123",
	}))
	req.Header.Set("Content-Type", "application/json")
	app.Test(req, testTimeout)

	// Login
	req2 := httptest.NewRequest(http.MethodPost, "/login", jsonBody(map[string]string{
		"email": "carol@example.com", "password": "pass123",
	}))
	req2.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req2, testTimeout)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	if body["access_token"] == nil {
		t.Fatal("expected access_token in login response")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	app, _ := setupAuthApp()

	req := httptest.NewRequest(http.MethodPost, "/register", jsonBody(map[string]string{
		"email": "dave@example.com", "password": "correct",
	}))
	req.Header.Set("Content-Type", "application/json")
	app.Test(req, testTimeout)

	req2 := httptest.NewRequest(http.MethodPost, "/login", jsonBody(map[string]string{
		"email": "dave@example.com", "password": "wrong",
	}))
	req2.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req2, testTimeout)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLogin_NonexistentEmail(t *testing.T) {
	app, _ := setupAuthApp()

	req := httptest.NewRequest(http.MethodPost, "/login", jsonBody(map[string]string{
		"email": "ghost@example.com", "password": "pass",
	}))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, testTimeout)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestMe_WithValidToken(t *testing.T) {
	app, _ := setupAuthApp()
	cfg := testConfig()

	// Register to get a real user
	req := httptest.NewRequest(http.MethodPost, "/register", jsonBody(map[string]string{
		"email": "eve@example.com", "password": "pass123",
	}))
	req.Header.Set("Content-Type", "application/json")
	regResp, err := app.Test(req, testTimeout)
	if err != nil {
		t.Fatal(err)
	}
	var regBody map[string]any
	json.NewDecoder(regResp.Body).Decode(&regBody)
	token, ok := regBody["access_token"].(string)
	if !ok || token == "" {
		t.Fatal("no access_token in register response")
	}
	_ = cfg // used for reference

	meReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(meReq)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestMe_WithoutToken(t *testing.T) {
	app, _ := setupAuthApp()

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	resp, err := app.Test(req, testTimeout)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRefresh_Success(t *testing.T) {
	app, _ := setupAuthApp()

	// Register to get cookies
	req := httptest.NewRequest(http.MethodPost, "/register", jsonBody(map[string]string{
		"email": "frank@example.com", "password": "pass123",
	}))
	req.Header.Set("Content-Type", "application/json")
	regResp, err := app.Test(req, testTimeout)
	if err != nil {
		t.Fatal(err)
	}

	// Extract refresh_token cookie
	var refreshCookie string
	for _, c := range regResp.Cookies() {
		if c.Name == "refresh_token" {
			refreshCookie = c.Value
		}
	}
	if refreshCookie == "" {
		t.Fatal("no refresh_token cookie in register response")
	}

	refreshReq := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	refreshReq.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshCookie})
	resp, err := app.Test(refreshReq)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	if body["access_token"] == nil {
		t.Fatal("expected access_token in refresh response")
	}
}

func TestLogout_ClearsCookies(t *testing.T) {
	app, _ := setupAuthApp()

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	resp, err := app.Test(req, testTimeout)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Verify that cookies are cleared.
	// Fiber sets MaxAge=-1 internally; the HTTP response encodes this as Max-Age=0
	// (which instructs browsers to delete the cookie) or an Expires in the past.
	cleared := map[string]bool{}
	for _, c := range resp.Cookies() {
		if c.Name == "access_token" || c.Name == "refresh_token" {
			// MaxAge=0 with empty value signals deletion; also accept negative values
			if c.Value == "" && c.MaxAge <= 0 {
				cleared[c.Name] = true
			}
		}
	}
	for _, name := range []string{"access_token", "refresh_token"} {
		if !cleared[name] {
			t.Fatalf("expected cookie %s to be cleared (empty value + MaxAge<=0)", name)
		}
	}
}
