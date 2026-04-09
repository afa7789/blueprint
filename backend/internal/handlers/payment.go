package handlers

import (
	"fmt"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/afa/blueprint/backend/pkg/config"
	"github.com/gofiber/fiber/v2"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/webhook"
)

type PaymentHandler struct {
	orders domain.OrderRepository
	cfg    *config.Config
}

func NewPaymentHandler(orders domain.OrderRepository, cfg *config.Config) *PaymentHandler {
	return &PaymentHandler{orders: orders, cfg: cfg}
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

	txID := fmt.Sprintf("TX_%s", order.ID)
	paymentMethod := "pix_manual"

	if err := h.orders.UpdatePayment(c.Context(), order.ID, paymentMethod, txID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"qr_code": fmt.Sprintf("PIX_QR_PLACEHOLDER_%s", order.ID),
		"tx_id":   txID,
		"message": "PIX integration pending - use manual approval",
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
