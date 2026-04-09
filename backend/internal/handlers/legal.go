package handlers

import (
	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/gofiber/fiber/v2"
)

type LegalHandler struct {
	pages domain.LegalPageRepository
}

func NewLegalHandler(pages domain.LegalPageRepository) *LegalHandler {
	return &LegalHandler{pages: pages}
}

// Public: get active page by slug
func (h *LegalHandler) GetBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	page, err := h.pages.FindBySlug(c.Context(), slug)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "page not found"})
	}
	if !page.IsActive {
		return c.Status(404).JSON(fiber.Map{"error": "page not found"})
	}
	return c.JSON(page)
}

// Public: list active pages (slug + title only, for footer)
func (h *LegalHandler) ListActive(c *fiber.Ctx) error {
	pages, err := h.pages.List(c.Context(), true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch pages"})
	}
	// Return only slug and title for footer links
	type pageLink struct {
		Slug  string `json:"slug"`
		Title string `json:"title"`
	}
	links := make([]pageLink, len(pages))
	for i, p := range pages {
		links[i] = pageLink{Slug: p.Slug, Title: p.Title}
	}
	return c.JSON(links)
}

// Admin: list all pages (including inactive)
func (h *LegalHandler) AdminList(c *fiber.Ctx) error {
	pages, err := h.pages.List(c.Context(), false)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch pages"})
	}
	return c.JSON(fiber.Map{"data": pages})
}

// Admin: create page
func (h *LegalHandler) AdminCreate(c *fiber.Ctx) error {
	var page domain.LegalPage
	if err := c.BodyParser(&page); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if page.Slug == "" || page.Title == "" {
		return c.Status(400).JSON(fiber.Map{"error": "slug and title are required"})
	}
	if err := h.pages.Create(c.Context(), &page); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create page"})
	}
	return c.Status(201).JSON(page)
}

// Admin: update page
func (h *LegalHandler) AdminUpdate(c *fiber.Ctx) error {
	id := c.Params("id")
	var page domain.LegalPage
	if err := c.BodyParser(&page); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	page.ID = id
	if err := h.pages.Update(c.Context(), &page); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update page"})
	}
	return c.JSON(page)
}

// Admin: delete page
func (h *LegalHandler) AdminDelete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.pages.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete page"})
	}
	return c.JSON(fiber.Map{"deleted": true})
}
