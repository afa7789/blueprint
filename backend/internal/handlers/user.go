package handlers

import (
	"encoding/json"
	"errors"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/afa/blueprint/backend/pkg/config"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	stripe "github.com/stripe/stripe-go/v82"
	stripecustomer "github.com/stripe/stripe-go/v82/customer"
	stripepaymentmethod "github.com/stripe/stripe-go/v82/paymentmethod"
	stripesetupintent "github.com/stripe/stripe-go/v82/setupintent"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	users    domain.UserRepository
	profiles domain.UserProfileRepository
	cfg      *config.Config
}

func NewUserHandler(users domain.UserRepository, profiles domain.UserProfileRepository, cfg *config.Config) *UserHandler {
	return &UserHandler{users: users, profiles: profiles, cfg: cfg}
}

type profileResponse struct {
	ID               string          `json:"id"`
	Email            string          `json:"email"`
	Name             *string         `json:"name"`
	Role             string          `json:"role"`
	EmailVerified    bool            `json:"email_verified"`
	Phone            *string         `json:"phone,omitempty"`
	AvatarURL        *string         `json:"avatar_url,omitempty"`
	Address          json.RawMessage `json:"address,omitempty"`
	StripeCustomerID *string         `json:"stripe_customer_id,omitempty"`
}

func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	u, err := h.users.FindByID(c.Context(), userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	resp := profileResponse{
		ID:               u.ID,
		Email:            u.Email,
		Name:             u.Name,
		Role:             u.Role,
		EmailVerified:    u.EmailVerified,
		StripeCustomerID: u.StripeCustomerID,
	}

	p, err := h.profiles.FindByUserID(c.Context(), userID)
	if err == nil {
		resp.Phone = p.Phone
		resp.AvatarURL = p.AvatarURL
		resp.Address = p.Address
	}

	return c.JSON(resp)
}

func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)

	var body struct {
		Name      *string         `json:"name"`
		Phone     *string         `json:"phone"`
		AvatarURL *string         `json:"avatar_url"`
		Address   json.RawMessage `json:"address"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	u, err := h.users.FindByID(c.Context(), userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	if body.Name != nil {
		u.Name = body.Name
	}
	if err := h.users.Update(c.Context(), u); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update user"})
	}

	// Upsert profile
	p, err := h.profiles.FindByUserID(c.Context(), userID)
	if err != nil {
		// No profile yet — create one
		p = &domain.UserProfile{
			ID:     uuid.NewString(),
			UserID: userID,
		}
	}
	if body.Phone != nil {
		p.Phone = body.Phone
	}
	if body.AvatarURL != nil {
		p.AvatarURL = body.AvatarURL
	}
	if body.Address != nil {
		p.Address = body.Address
	}
	if err := h.profiles.Upsert(c.Context(), p); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update profile"})
	}

	resp := profileResponse{
		ID:               u.ID,
		Email:            u.Email,
		Name:             u.Name,
		Role:             u.Role,
		EmailVerified:    u.EmailVerified,
		Phone:            p.Phone,
		AvatarURL:        p.AvatarURL,
		Address:          p.Address,
		StripeCustomerID: u.StripeCustomerID,
	}
	return c.JSON(resp)
}

func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)

	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if body.CurrentPassword == "" || body.NewPassword == "" {
		return c.Status(400).JSON(fiber.Map{"error": "current_password and new_password are required"})
	}

	u, err := h.users.FindByID(c.Context(), userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(body.CurrentPassword)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "current password is incorrect"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to hash password"})
	}
	u.PasswordHash = string(hash)

	if err := h.users.Update(c.Context(), u); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update password"})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *UserHandler) CreateSetupIntent(c *fiber.Ctx) error {
	if h.cfg.StripeKey == "" {
		return c.Status(503).JSON(fiber.Map{"error": "stripe not configured", "env_required": "STRIPE_KEY"})
	}

	userID, _ := c.Locals("user_id").(string)
	u, err := h.users.FindByID(c.Context(), userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	// Create Stripe customer if not exists
	if u.StripeCustomerID == nil {
		customer, err := stripecustomer.New(&stripe.CustomerParams{
			Email: stripe.String(u.Email),
		})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to create stripe customer"})
		}
		if err := h.users.UpdateStripeCustomerID(c.Context(), userID, customer.ID); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to save stripe customer"})
		}
		u.StripeCustomerID = &customer.ID
	}

	si, err := stripesetupintent.New(&stripe.SetupIntentParams{
		Customer:           u.StripeCustomerID,
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create setup intent"})
	}

	return c.JSON(fiber.Map{"client_secret": si.ClientSecret})
}

func (h *UserHandler) ListSavedCards(c *fiber.Ctx) error {
	if h.cfg.StripeKey == "" {
		return c.Status(503).JSON(fiber.Map{"error": "stripe not configured", "env_required": "STRIPE_KEY"})
	}

	userID, _ := c.Locals("user_id").(string)
	u, err := h.users.FindByID(c.Context(), userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	if u.StripeCustomerID == nil {
		return c.JSON([]fiber.Map{})
	}

	params := &stripe.PaymentMethodListParams{
		Customer: u.StripeCustomerID,
		Type:     stripe.String("card"),
	}
	iter := stripepaymentmethod.List(params)

	type cardInfo struct {
		ID       string `json:"id"`
		Brand    string `json:"brand"`
		Last4    string `json:"last4"`
		ExpMonth int64  `json:"exp_month"`
		ExpYear  int64  `json:"exp_year"`
	}
	var cards []cardInfo
	for iter.Next() {
		pm := iter.PaymentMethod()
		if pm.Card != nil {
			cards = append(cards, cardInfo{
				ID:       pm.ID,
				Brand:    string(pm.Card.Brand),
				Last4:    pm.Card.Last4,
				ExpMonth: pm.Card.ExpMonth,
				ExpYear:  pm.Card.ExpYear,
			})
		}
	}
	if err := iter.Err(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to list cards"})
	}
	if cards == nil {
		cards = []cardInfo{}
	}
	return c.JSON(cards)
}

func (h *UserHandler) DeleteSavedCard(c *fiber.Ctx) error {
	if h.cfg.StripeKey == "" {
		return c.Status(503).JSON(fiber.Map{"error": "stripe not configured", "env_required": "STRIPE_KEY"})
	}

	pmID := c.Params("id")
	_, err := stripepaymentmethod.Detach(pmID, nil)
	if err != nil {
		var stripeErr *stripe.Error
		if errors.As(err, &stripeErr) {
			return c.Status(400).JSON(fiber.Map{"error": stripeErr.Msg})
		}
		return c.Status(500).JSON(fiber.Map{"error": "failed to detach card"})
	}

	return c.JSON(fiber.Map{"success": true})
}
