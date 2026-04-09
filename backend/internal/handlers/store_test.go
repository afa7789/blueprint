package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/afa/blueprint/backend/internal/handlers"
	"github.com/afa/blueprint/backend/internal/testutil"
	"github.com/afa/blueprint/backend/pkg/config"
)

func setupStoreApp() *fiber.App {
	app := fiber.New()
	productRepo := testutil.NewMockProductRepo()
	categoryRepo := testutil.NewMockCategoryRepo()
	orderRepo := testutil.NewMockOrderRepo()
	couponRepo := testutil.NewMockCouponRepo()
	cfg := &config.Config{Env: "development"}

	h := handlers.NewStoreHandler(productRepo, categoryRepo, orderRepo, couponRepo, cfg)

	app.Get("/products", h.ListProducts)
	app.Get("/categories", h.ListCategories)

	return app
}

func TestListProducts_Empty(t *testing.T) {
	app := setupStoreApp()

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	total, ok := body["total"].(float64)
	if !ok {
		t.Fatalf("expected total field in response, got %v", body)
	}
	if total != 0 {
		t.Fatalf("expected total 0, got %v", total)
	}
}

func TestListCategories_Empty(t *testing.T) {
	app := setupStoreApp()

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body []any
	json.NewDecoder(resp.Body).Decode(&body)
	if len(body) != 0 {
		t.Fatalf("expected empty categories array, got %v", body)
	}
}
