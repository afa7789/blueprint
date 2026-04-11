package handlers

import (
	"encoding/json"
	"time"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/afa/blueprint/backend/pkg/config"
	"github.com/afa/blueprint/backend/pkg/metrics"
	"github.com/gofiber/fiber/v2"
)

type StoreHandler struct {
	products   domain.ProductRepository
	categories domain.CategoryRepository
	orders     domain.OrderRepository
	coupons    domain.CouponRepository
	cfg        *config.Config
}

func NewStoreHandler(
	products domain.ProductRepository,
	categories domain.CategoryRepository,
	orders domain.OrderRepository,
	coupons domain.CouponRepository,
	cfg *config.Config,
) *StoreHandler {
	return &StoreHandler{
		products:   products,
		categories: categories,
		orders:     orders,
		coupons:    coupons,
		cfg:        cfg,
	}
}

// ---- Public ----

func (h *StoreHandler) ListProducts(c *fiber.Ctx) error {
	_, limit, offset := paginate(c)
	var categoryID *string
	if q := c.Query("category_id"); q != "" {
		categoryID = &q
	}
	products, total, err := h.products.List(c.Context(), categoryID, true, offset, limit)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": products, "total": total})
}

func (h *StoreHandler) GetProduct(c *fiber.Ctx) error {
	p, err := h.products.FindByID(c.Context(), c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "product not found")
	}
	return c.JSON(p)
}

func (h *StoreHandler) ListCategories(c *fiber.Ctx) error {
	cats, err := h.categories.List(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(cats)
}

// ---- Auth-required ----

type createOrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type createOrderRequest struct {
	Items           []createOrderItem      `json:"items"`
	ShippingAddress map[string]interface{} `json:"shipping_address"`
	Shipping        map[string]interface{} `json:"shipping"`
	PaymentMethod   string                 `json:"payment_method"`
	CouponCode      string                 `json:"coupon_code"`
}

func (h *StoreHandler) CreateOrder(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	var req createOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if len(req.Items) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "items required")
	}
	if len(req.ShippingAddress) == 0 && len(req.Shipping) > 0 {
		req.ShippingAddress = req.Shipping
	}

	var orderItems []domain.OrderItem
	var total float64

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "quantity must be > 0")
		}
		p, err := h.products.FindByID(c.Context(), item.ProductID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "product not found: "+item.ProductID)
		}
		if p.Stock <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "product out of stock: "+p.Name)
		}
		lineTotal := p.Price * float64(item.Quantity)
		total += lineTotal
		pid := item.ProductID
		orderItems = append(orderItems, domain.OrderItem{
			ProductID: &pid,
			Quantity:  item.Quantity,
			UnitPrice: p.Price,
		})
	}

	if req.CouponCode != "" {
		coupon, err := h.coupons.FindByCode(c.Context(), req.CouponCode)
		if err == nil && coupon.IsActive {
			now := time.Now()
			valid := true
			if coupon.ValidFrom != nil && now.Before(*coupon.ValidFrom) {
				valid = false
			}
			if coupon.ValidUntil != nil && now.After(*coupon.ValidUntil) {
				valid = false
			}
			if coupon.MaxUses != nil && coupon.UsedCount >= *coupon.MaxUses {
				valid = false
			}
			if coupon.MinPurchase != nil && total < *coupon.MinPurchase {
				valid = false
			}
			if valid {
				discount := applyDiscount(coupon, total)
				total -= discount
				if total < 0 {
					total = 0
				}
				_ = h.coupons.IncrementUsed(c.Context(), coupon.ID)
			}
		}
	}

	addrBytes, err := json.Marshal(req.ShippingAddress)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid shipping address")
	}

	pm := req.PaymentMethod
	order := &domain.Order{
		UserID:          &userID,
		Status:          "pending",
		Total:           total,
		PaymentMethod:   &pm,
		ShippingAddress: addrBytes,
	}

	if err := h.orders.Create(c.Context(), order, orderItems); err != nil {
		return err
	}

	pm = "unknown"
	if order.PaymentMethod != nil {
		pm = *order.PaymentMethod
	}
	metrics.OrdersTotal.WithLabelValues("pending", pm).Inc()
	metrics.OrdersRevenue.WithLabelValues(pm).Add(order.Total)

	return c.Status(fiber.StatusCreated).JSON(order)
}

func (h *StoreHandler) ListMyOrders(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	_, limit, offset := paginate(c)
	orders, total, err := h.orders.FindByUser(c.Context(), userID, offset, limit)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": orders, "total": total})
}

// ---- Admin ----

func (h *StoreHandler) AdminListProducts(c *fiber.Ctx) error {
	_, limit, offset := paginate(c)
	var categoryID *string
	if q := c.Query("category_id"); q != "" {
		categoryID = &q
	}

	products, total, err := h.products.List(c.Context(), categoryID, false, offset, limit)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": products, "total": total})
}

func (h *StoreHandler) AdminListCategories(c *fiber.Ctx) error {
	categories, err := h.categories.List(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": categories})
}

func (h *StoreHandler) AdminCreateProduct(c *fiber.Ctx) error {
	p := &domain.Product{}
	if err := c.BodyParser(p); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := h.products.Create(c.Context(), p); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(p)
}

func (h *StoreHandler) AdminUpdateProduct(c *fiber.Ctx) error {
	existing, err := h.products.FindByID(c.Context(), c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "product not found")
	}
	if err := c.BodyParser(existing); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	existing.ID = c.Params("id")
	if err := h.products.Update(c.Context(), existing); err != nil {
		return err
	}
	return c.JSON(existing)
}

func (h *StoreHandler) AdminDeleteProduct(c *fiber.Ctx) error {
	if err := h.products.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *StoreHandler) AdminCreateCategory(c *fiber.Ctx) error {
	cat := &domain.Category{}
	if err := c.BodyParser(cat); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := h.categories.Create(c.Context(), cat); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(cat)
}

func (h *StoreHandler) AdminUpdateCategory(c *fiber.Ctx) error {
	cat := &domain.Category{}
	if err := c.BodyParser(cat); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	cat.ID = c.Params("id")
	if err := h.categories.Update(c.Context(), cat); err != nil {
		return err
	}
	return c.JSON(cat)
}

func (h *StoreHandler) AdminDeleteCategory(c *fiber.Ctx) error {
	if err := h.categories.Delete(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *StoreHandler) AdminListOrders(c *fiber.Ctx) error {
	status := c.Query("status")
	orders, err := h.orders.ListByStatus(c.Context(), status)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": orders})
}

func (h *StoreHandler) AdminUpdateOrderStatus(c *fiber.Ctx) error {
	var body struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := h.orders.UpdateStatus(c.Context(), c.Params("id"), body.Status); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// applyDiscount computes the discount amount from a coupon given the subtotal.
func applyDiscount(coupon *domain.Coupon, subtotal float64) float64 {
	if coupon.DiscountType == nil {
		return 0
	}
	switch *coupon.DiscountType {
	case "percentage":
		return subtotal * coupon.DiscountValue / 100
	case "fixed":
		return coupon.DiscountValue
	}
	return 0
}
