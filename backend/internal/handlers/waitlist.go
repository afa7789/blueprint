package handlers

import (
	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/gofiber/fiber/v2"
)

type WaitlistHandler struct {
	waitlist domain.WaitlistRepository
}

func NewWaitlistHandler(waitlist domain.WaitlistRepository) *WaitlistHandler {
	return &WaitlistHandler{waitlist: waitlist}
}

// AddToWaitlist godoc
// @Summary     Join the waitlist
// @Tags        Waitlist
// @Accept      json
// @Produce     json
// @Param       body body object{email=string,name=string} true "Email and optional name"
// @Success     201 {object} map[string]interface{}
// @Failure     409 {object} map[string]interface{}
// @Router      /waitlist [post]
func (h *WaitlistHandler) AddToWaitlist(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil || req.Email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "email is required"})
	}

	exists, err := h.waitlist.ExistsByEmail(c.Context(), req.Email)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "internal error"})
	}
	if exists {
		return c.Status(409).JSON(fiber.Map{"error": "email already on waitlist"})
	}

	entry := &domain.WaitlistEntry{Email: req.Email, Name: &req.Name}
	if err := h.waitlist.Add(c.Context(), entry); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to add to waitlist"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "added to waitlist", "email": req.Email})
}

// GetWaitlist godoc
// @Summary     List waitlist entries (admin)
// @Tags        Admin
// @Produce     json
// @Success     200 {array} domain.WaitlistEntry
// @Security    BearerAuth
// @Router      /waitlist [get]
func (h *WaitlistHandler) GetWaitlist(c *fiber.Ctx) error {
	entries, err := h.waitlist.List(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch waitlist"})
	}
	return c.JSON(entries)
}
