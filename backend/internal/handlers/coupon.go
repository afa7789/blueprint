package handlers

import (
	"time"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/gofiber/fiber/v2"
)

type CouponHandler struct {
	coupons domain.CouponRepository
}

func NewCouponHandler(coupons domain.CouponRepository) *CouponHandler {
	return &CouponHandler{coupons: coupons}
}

// ValidateCoupon godoc
// @Summary     Validate a coupon code
// @Tags        Coupons
// @Accept      json
// @Produce     json
// @Param       body body object{code=string,subtotal=number} true "Coupon code and subtotal"
// @Success     200 {object} map[string]interface{}
// @Router      /coupons/validate [post]
func (h *CouponHandler) ValidateCoupon(c *fiber.Ctx) error {
	var req struct {
		Code     string  `json:"code"`
		Subtotal float64 `json:"subtotal"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if req.Code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "code required")
	}

	coupon, err := h.coupons.FindByCode(c.Context(), req.Code)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "coupon not found")
	}
	if !coupon.IsActive {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "coupon is inactive")
	}

	now := time.Now()
	if coupon.ValidFrom != nil && now.Before(*coupon.ValidFrom) {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "coupon not yet valid")
	}
	if coupon.ValidUntil != nil && now.After(*coupon.ValidUntil) {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "coupon has expired")
	}
	if coupon.MaxUses != nil && coupon.UsedCount >= *coupon.MaxUses {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "coupon usage limit reached")
	}
	if coupon.MinPurchase != nil && req.Subtotal < *coupon.MinPurchase {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "subtotal below minimum purchase")
	}

	discount := applyDiscount(coupon, req.Subtotal)
	return c.JSON(fiber.Map{
		"coupon":      coupon,
		"discount":    discount,
		"final_total": req.Subtotal - discount,
	})
}

// AdminListCoupons godoc
// @Summary     List all coupons (admin)
// @Tags        Admin
// @Produce     json
// @Security    BearerAuth
// @Router      /admin/coupons [get]
func (h *CouponHandler) AdminListCoupons(c *fiber.Ctx) error {
	coupons, err := h.coupons.List(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": coupons})
}

// AdminCreateCoupon godoc
// @Summary     Create a coupon (admin)
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Router      /admin/coupons [post]
func (h *CouponHandler) AdminCreateCoupon(c *fiber.Ctx) error {
	coupon := &domain.Coupon{}
	if err := c.BodyParser(coupon); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := h.coupons.Create(c.Context(), coupon); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(coupon)
}

// AdminDeleteCoupon godoc
// @Summary     Delete a coupon (admin)
// @Tags        Admin
// @Security    BearerAuth
// @Router      /admin/coupons/{id} [delete]
func (h *CouponHandler) AdminDeleteCoupon(c *fiber.Ctx) error {
	// CouponRepository has no Delete method; return method not allowed.
	return fiber.NewError(fiber.StatusMethodNotAllowed, "coupon deletion not supported")
}
