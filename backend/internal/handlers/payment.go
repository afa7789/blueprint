package handlers

import (
	"fmt"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/afa/blueprint/backend/pkg/config"
	"github.com/afa/blueprint/backend/pkg/pix"
	"github.com/gofiber/fiber/v2"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/webhook"
)

type PaymentHandler struct {
	orders  domain.OrderRepository
	pixCfg  domain.PixConfigRepository
	cfg     *config.Config
	storage domain.Storage
}

func NewPaymentHandler(orders domain.OrderRepository, pixCfg domain.PixConfigRepository, cfg *config.Config, storage domain.Storage) *PaymentHandler {
	return &PaymentHandler{orders: orders, pixCfg: pixCfg, cfg: cfg, storage: storage}
}

type createPaymentRequest struct {
	OrderID string `json:"order_id"`
}

func (h *PaymentHandler) CreateStripePayment(c *fiber.Ctx) error {
	if h.cfg.StripeKey == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "stripe not configured"})
	}

	var req createPaymentRequest
	if err := c.BodyParser(&req); err != nil || req.OrderID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "order_id required")
	}

	order, err := h.orders.FindByID(c.Context(), req.OrderID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "order not found")
	}
	if order.Status != "pending" {
		return fiber.NewError(fiber.StatusBadRequest, "order is not pending")
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(order.Total * 100)),
		Currency: stripe.String("brl"),
	}
	pi, err := paymentintent.New(params)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create payment intent: "+err.Error())
	}

	if err := h.orders.UpdatePayment(c.Context(), order.ID, "stripe", pi.ID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"client_secret": pi.ClientSecret})
}

func (h *PaymentHandler) StripeWebhook(c *fiber.Ctx) error {
	payload := c.Body()
	sigHeader := c.Get("Stripe-Signature")

	event, err := webhook.ConstructEvent(payload, sigHeader, h.cfg.StripeWebhookSecret)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid signature"})
	}

	switch event.Type {
	case "payment_intent.succeeded":
		pi, ok := event.Data.Object["id"].(string)
		if !ok {
			return c.SendStatus(fiber.StatusOK)
		}
		order, err := h.orders.FindByPaymentID(c.Context(), pi)
		if err != nil {
			return c.SendStatus(fiber.StatusOK)
		}
		_ = h.orders.UpdateStatus(c.Context(), order.ID, "paid")
	case "payment_intent.payment_failed":
		pi, ok := event.Data.Object["id"].(string)
		if !ok {
			return c.SendStatus(fiber.StatusOK)
		}
		order, err := h.orders.FindByPaymentID(c.Context(), pi)
		if err != nil {
			return c.SendStatus(fiber.StatusOK)
		}
		_ = h.orders.UpdateStatus(c.Context(), order.ID, "cancelled")
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *PaymentHandler) CreatePixPayment(c *fiber.Ctx) error {
	var req createPaymentRequest
	if err := c.BodyParser(&req); err != nil || req.OrderID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "order_id required")
	}

	order, err := h.orders.FindByID(c.Context(), req.OrderID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "order not found")
	}
	if order.Status != "pending" {
		return fiber.NewError(fiber.StatusBadRequest, "order is not pending")
	}

	pixCfg, err := h.pixCfg.Get(c.Context())
	if err != nil || pixCfg.PixKey == "" {
		return c.Status(503).JSON(fiber.Map{
			"error":        "PIX not configured",
			"env_required": "Configure PIX in Admin > Payments",
		})
	}

	txID := fmt.Sprintf("TX_%s", order.ID)
	paymentMethod := "pix_manual"

	if err := h.orders.UpdatePayment(c.Context(), order.ID, paymentMethod, txID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update payment: "+err.Error())
	}

	brcode := pix.GeneratePayload(pixCfg.PixKey, pixCfg.Beneficiary, pixCfg.City, int64(order.Total*100))

	return c.JSON(fiber.Map{
		"brcode":      brcode,
		"tx_id":       txID,
		"pix_key":     pixCfg.PixKey,
		"beneficiary": pixCfg.Beneficiary,
		"amount":      order.Total,
		"message":     "Scan QR code to pay via PIX, then upload receipt",
	})
}

func (h *PaymentHandler) GetPixConfig(c *fiber.Ctx) error {
	cfg, err := h.pixCfg.Get(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get PIX config"})
	}
	return c.JSON(cfg)
}

func (h *PaymentHandler) UpdatePixConfig(c *fiber.Ctx) error {
	var cfg domain.PixConfig
	if err := c.BodyParser(&cfg); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if len(cfg.Beneficiary) > 25 {
		cfg.Beneficiary = cfg.Beneficiary[:25]
	}
	if len(cfg.City) > 15 {
		cfg.City = cfg.City[:15]
	}
	if err := h.pixCfg.Update(c.Context(), &cfg); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update PIX config"})
	}
	return c.JSON(cfg)
}

func (h *PaymentHandler) UploadPixReceipt(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	userID, _ := c.Locals("user_id").(string)

	order, err := h.orders.FindByID(c.Context(), orderID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "order not found")
	}
	if order.UserID != nil && *order.UserID != userID {
		return fiber.NewError(fiber.StatusForbidden, "not your order")
	}
	if order.PaymentMethod == nil || *order.PaymentMethod != "pix_manual" {
		return fiber.NewError(fiber.StatusBadRequest, "not a PIX manual order")
	}

	file, err := c.FormFile("receipt")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "receipt file is required")
	}

	url, err := UploadFormFile(c.Context(), h.storage, file, "receipts")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "upload failed")
	}

	if err := h.orders.UpdateReceiptURL(c.Context(), orderID, url); err != nil {
		return err
	}
	if err := h.orders.UpdateStatus(c.Context(), orderID, "awaiting_approval"); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message":     "receipt uploaded, awaiting admin approval",
		"receipt_url": url,
		"order_id":    orderID,
	})
}

func (h *PaymentHandler) ApprovePixPayment(c *fiber.Ctx) error {
	orderID := c.Params("id")
	order, err := h.orders.FindByID(c.Context(), orderID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "order not found")
	}

	if err := h.orders.UpdateStatus(c.Context(), orderID, "paid"); err != nil {
		return err
	}

	order.Status = "paid"
	return c.JSON(order)
}
