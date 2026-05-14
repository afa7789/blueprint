package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"
	stripe "github.com/stripe/stripe-go/v82"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/afa/blueprint/backend/internal/handlers"
	"github.com/afa/blueprint/backend/internal/testutil"
	"github.com/afa/blueprint/backend/pkg/config"
)

// stripeMock spins up an httptest server that mimics the minimum Stripe
// REST surface used by the payment handler (customer create + payment
// intent create). It records the form values posted to each endpoint so
// tests can assert what the handler sent.
type stripeMock struct {
	mu               sync.Mutex
	server           *httptest.Server
	createdCustomers []url.Values
	createdPIs       []url.Values
}

func newStripeMock(t *testing.T) *stripeMock {
	t.Helper()
	m := &stripeMock{}
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		m.mu.Lock()
		defer m.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/customers":
			m.createdCustomers = append(m.createdCustomers, r.PostForm)
			_, _ = w.Write([]byte(`{"id":"cus_test_123","object":"customer","email":"u@example.com"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/payment_intents":
			m.createdPIs = append(m.createdPIs, r.PostForm)
			_, _ = w.Write([]byte(`{"id":"pi_test_abc","object":"payment_intent","client_secret":"pi_test_abc_secret_xyz","status":"requires_payment_method"}`))
		default:
			http.Error(w, "stripe mock: unexpected "+r.Method+" "+r.URL.Path, http.StatusNotImplemented)
		}
	}))

	origAPI := stripe.GetBackend(stripe.APIBackend)
	url := m.server.URL
	stripe.SetBackend(stripe.APIBackend, stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{URL: &url}))
	t.Cleanup(func() {
		stripe.SetBackend(stripe.APIBackend, origAPI)
		m.server.Close()
	})
	return m
}

type paymentTestEnv struct {
	app       *fiber.App
	orders    *testutil.MockOrderRepo
	users     *testutil.MockUserRepo
	cfg       *config.Config
	userID    string
	orderID   string
	mock      *stripeMock
}

func setupPaymentApp(t *testing.T, stripeKey, publishableKey string) *paymentTestEnv {
	t.Helper()
	mock := newStripeMock(t)

	app := fiber.New()
	orders := testutil.NewMockOrderRepo()
	users := testutil.NewMockUserRepo()
	cfg := &config.Config{
		StripeKey:            stripeKey,
		StripePublishableKey: publishableKey,
		FrontendURL:          "http://localhost:5173",
	}

	userID := "user-1"
	_ = users.Create(t.Context(), &domain.User{ID: userID, Email: "u@example.com", Role: "user"})

	orderID := "order-1"
	userIDPtr := userID
	_ = orders.Create(t.Context(), &domain.Order{
		ID:     orderID,
		UserID: &userIDPtr,
		Status: "pending",
		Total:  120.50,
	}, nil)

	h := handlers.NewPaymentHandler(orders, users, nil, cfg, nil)

	app.Post("/payments/stripe", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.CreateStripePayment(c)
	}, /* second handler not needed */ )

	return &paymentTestEnv{
		app:     app,
		orders:  orders,
		users:   users,
		cfg:     cfg,
		userID:  userID,
		orderID: orderID,
		mock:    mock,
	}
}

func doPostJSON(t *testing.T, app *fiber.App, path string, body any) (*http.Response, map[string]any) {
	t.Helper()
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	var parsed map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&parsed)
	return resp, parsed
}

func TestCreateStripePayment_StripeNotConfigured(t *testing.T) {
	env := setupPaymentApp(t, "", "")
	resp, body := doPostJSON(t, env.app, "/payments/stripe", map[string]any{"order_id": env.orderID})
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d (%v)", resp.StatusCode, body)
	}
	if got, _ := body["error"].(string); got != "stripe not configured" {
		t.Fatalf("unexpected error: %v", body)
	}
}

func TestCreateStripePayment_PublishableKeyMissing(t *testing.T) {
	env := setupPaymentApp(t, "sk_test", "")
	resp, _ := doPostJSON(t, env.app, "/payments/stripe", map[string]any{"order_id": env.orderID})
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("expected 503 when publishable key missing, got %d", resp.StatusCode)
	}
}

func TestCreateStripePayment_MissingOrderID(t *testing.T) {
	env := setupPaymentApp(t, "sk_test", "pk_test")
	resp, _ := doPostJSON(t, env.app, "/payments/stripe", map[string]any{})
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateStripePayment_OrderNotFound(t *testing.T) {
	env := setupPaymentApp(t, "sk_test", "pk_test")
	resp, _ := doPostJSON(t, env.app, "/payments/stripe", map[string]any{"order_id": "missing"})
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestCreateStripePayment_OrderNotPending(t *testing.T) {
	env := setupPaymentApp(t, "sk_test", "pk_test")
	_ = env.orders.UpdateStatus(t.Context(), env.orderID, "paid")
	resp, _ := doPostJSON(t, env.app, "/payments/stripe", map[string]any{"order_id": env.orderID})
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateStripePayment_HappyPath_CreatesCustomerAndIntent(t *testing.T) {
	env := setupPaymentApp(t, "sk_test", "pk_test_live")

	resp, body := doPostJSON(t, env.app, "/payments/stripe", map[string]any{"order_id": env.orderID})
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d (%v)", resp.StatusCode, body)
	}

	if cs, _ := body["client_secret"].(string); !strings.HasPrefix(cs, "pi_test_abc_secret") {
		t.Fatalf("missing client_secret: %v", body)
	}
	if pk, _ := body["publishable_key"].(string); pk != "pk_test_live" {
		t.Fatalf("publishable_key not echoed: %v", body)
	}

	// customer was created
	if len(env.mock.createdCustomers) != 1 {
		t.Fatalf("expected 1 customer created, got %d", len(env.mock.createdCustomers))
	}
	// payment intent posted with required fields
	if len(env.mock.createdPIs) != 1 {
		t.Fatalf("expected 1 PI created, got %d", len(env.mock.createdPIs))
	}
	form := env.mock.createdPIs[0]
	if form.Get("amount") != "12050" {
		t.Fatalf("expected amount 12050 (cents), got %q", form.Get("amount"))
	}
	if form.Get("currency") != "brl" {
		t.Fatalf("expected currency brl, got %q", form.Get("currency"))
	}
	if form.Get("customer") != "cus_test_123" {
		t.Fatalf("expected customer cus_test_123, got %q", form.Get("customer"))
	}
	if form.Get("automatic_payment_methods[enabled]") != "true" {
		t.Fatalf("automatic_payment_methods not enabled: %v", form)
	}

	// user record was updated with the new stripe customer id
	u, _ := env.users.FindByID(t.Context(), env.userID)
	if u.StripeCustomerID == nil || *u.StripeCustomerID != "cus_test_123" {
		t.Fatalf("user customer id not persisted: %+v", u.StripeCustomerID)
	}
}

func TestCreateStripePayment_SaveCard_SetsFutureUsage(t *testing.T) {
	env := setupPaymentApp(t, "sk_test", "pk_test")
	resp, _ := doPostJSON(t, env.app, "/payments/stripe", map[string]any{
		"order_id":  env.orderID,
		"save_card": true,
	})
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	form := env.mock.createdPIs[0]
	if form.Get("setup_future_usage") != "off_session" {
		t.Fatalf("expected setup_future_usage=off_session, got %q", form.Get("setup_future_usage"))
	}
}

func TestCreateStripePayment_SavedPaymentMethod_ConfirmsImmediately(t *testing.T) {
	env := setupPaymentApp(t, "sk_test", "pk_test")
	resp, _ := doPostJSON(t, env.app, "/payments/stripe", map[string]any{
		"order_id":          env.orderID,
		"payment_method_id": "pm_card_visa",
	})
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	form := env.mock.createdPIs[0]
	if form.Get("payment_method") != "pm_card_visa" {
		t.Fatalf("payment_method not forwarded: %v", form)
	}
	if form.Get("confirm") != "true" {
		t.Fatalf("expected confirm=true, got %q", form.Get("confirm"))
	}
	if !strings.Contains(form.Get("return_url"), env.orderID) {
		t.Fatalf("return_url missing order id: %q", form.Get("return_url"))
	}
}

func TestCreateStripePayment_ForbidsOtherUserOrder(t *testing.T) {
	env := setupPaymentApp(t, "sk_test", "pk_test")
	// re-create order under a different user
	other := "user-2"
	otherPtr := other
	_ = env.orders.UpdateStatus(t.Context(), env.orderID, "pending")
	o, _ := env.orders.FindByID(t.Context(), env.orderID)
	o.UserID = &otherPtr
	// MockOrderRepo doesn't expose direct write; recreate the order with a new id under other user
	_ = env.orders.Create(t.Context(), &domain.Order{ID: "order-2", UserID: &otherPtr, Status: "pending", Total: 10}, nil)

	resp, _ := doPostJSON(t, env.app, "/payments/stripe", map[string]any{"order_id": "order-2"})
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}
