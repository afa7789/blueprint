package handlers

import (
	"regexp"
	"strings"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/afa/blueprint/backend/pkg/config"
	"github.com/gofiber/fiber/v2"
)

var nonAlphanumHyphen = regexp.MustCompile(`[^a-z0-9-]`)
var multipleHyphens = regexp.MustCompile(`-+`)

type BlogHandler struct {
	blog domain.BlogRepository
	cfg  *config.Config
}

func NewBlogHandler(blog domain.BlogRepository, cfg *config.Config) *BlogHandler {
	return &BlogHandler{blog: blog, cfg: cfg}
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = nonAlphanumHyphen.ReplaceAllString(s, "")
	s = multipleHyphens.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// ---- Public routes ----

func (h *BlogHandler) ListPublished(c *fiber.Ctx) error {
	page, limit, offset := paginate(c)
	posts, total, err := h.blog.List(c.Context(), "published", offset, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": posts, "total": total, "page": page, "limit": limit})
}

func (h *BlogHandler) GetBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	post, err := h.blog.FindBySlug(c.Context(), slug)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "post not found")
	}
	if post.Status != "published" {
		return fiber.NewError(fiber.StatusNotFound, "post not found")
	}
	return c.JSON(post)
}

// ---- Admin routes ----

func (h *BlogHandler) AdminListPosts(c *fiber.Ctx) error {
	page, limit, offset := paginate(c)
	posts, total, err := h.blog.List(c.Context(), "", offset, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": posts, "total": total, "page": page, "limit": limit})
}

func (h *BlogHandler) AdminCreatePost(c *fiber.Ctx) error {
	var body struct {
		Title   string  `json:"title"`
		Slug    string  `json:"slug"`
		Content *string `json:"content"`
		Excerpt *string `json:"excerpt"`
		Status  string  `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if body.Title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title is required")
	}
	if body.Slug == "" {
		body.Slug = slugify(body.Title)
	}
	if body.Status == "" {
		body.Status = "draft"
	}

	userID, _ := c.Locals("user_id").(string)
	var authorID *string
	if userID != "" {
		authorID = &userID
	}

	post := &domain.BlogPost{
		Title:    body.Title,
		Slug:     body.Slug,
		Content:  body.Content,
		Excerpt:  body.Excerpt,
		Status:   body.Status,
		AuthorID: authorID,
	}

	if err := h.blog.Create(c.Context(), post); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(post)
}

func (h *BlogHandler) AdminUpdatePost(c *fiber.Ctx) error {
	id := c.Params("id")
	post, err := h.blog.FindByID(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "post not found")
	}

	var body struct {
		Title   *string `json:"title"`
		Slug    *string `json:"slug"`
		Content *string `json:"content"`
		Excerpt *string `json:"excerpt"`
		Status  *string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if body.Title != nil {
		post.Title = *body.Title
	}
	if body.Slug != nil {
		post.Slug = *body.Slug
	}
	if body.Content != nil {
		post.Content = body.Content
	}
	if body.Excerpt != nil {
		post.Excerpt = body.Excerpt
	}
	if body.Status != nil {
		post.Status = *body.Status
	}

	if err := h.blog.Update(c.Context(), post); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(post)
}

func (h *BlogHandler) AdminDeletePost(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.blog.Delete(c.Context(), id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *BlogHandler) AdminUploadCover(c *fiber.Ctx) error {
	id := c.Params("id")
	post, err := h.blog.FindByID(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "post not found")
	}

	file, err := c.FormFile("cover")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cover file is required")
	}

	url, err := UploadFile(file, "covers", h.cfg)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	post.CoverImage = &url
	if err := h.blog.Update(c.Context(), post); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"cover_image": url})
}

func (h *BlogHandler) AdminAIGenerate(c *fiber.Ctx) error {
	var body struct {
		Prompt string `json:"prompt"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if body.Prompt == "" {
		return fiber.NewError(fiber.StatusBadRequest, "prompt is required")
	}

	title := "Generated: " + body.Prompt
	content := "AI content generation coming soon. Prompt: " + body.Prompt
	excerpt := "..."

	return c.JSON(fiber.Map{
		"title":   title,
		"content": content,
		"excerpt": excerpt,
	})
}
