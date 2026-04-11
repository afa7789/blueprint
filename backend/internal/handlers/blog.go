package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

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

type blogAIResponse struct {
	Title   string `json:"title"`
	Slug    string `json:"slug"`
	Excerpt string `json:"excerpt"`
	Content string `json:"content"`
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
	return c.JSON(fiber.Map{"data": posts, "posts": posts, "total": total, "page": page, "limit": limit})
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
		Title      string  `json:"title"`
		Slug       string  `json:"slug"`
		Content    *string `json:"content"`
		Excerpt    *string `json:"excerpt"`
		CoverImage *string `json:"cover_image"`
		Status     string  `json:"status"`
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
		Title:      body.Title,
		Slug:       body.Slug,
		Content:    body.Content,
		Excerpt:    body.Excerpt,
		CoverImage: body.CoverImage,
		Status:     body.Status,
		AuthorID:   authorID,
	}
	applyPublishedState(post)

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
		Title      *string `json:"title"`
		Slug       *string `json:"slug"`
		Content    *string `json:"content"`
		Excerpt    *string `json:"excerpt"`
		CoverImage *string `json:"cover_image"`
		Status     *string `json:"status"`
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
	if body.CoverImage != nil {
		post.CoverImage = body.CoverImage
	}
	if body.Status != nil {
		post.Status = *body.Status
	}
	applyPublishedState(post)

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
	if h.cfg.OpenAIKey == "" {
		return fiber.NewError(fiber.StatusServiceUnavailable, "OPENAI_KEY is not configured")
	}

	var body struct {
		Prompt string `json:"prompt"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if body.Prompt == "" {
		return fiber.NewError(fiber.StatusBadRequest, "prompt is required")
	}

	payload := map[string]interface{}{
		"model": "gpt-5.4-mini",
		"reasoning": map[string]interface{}{
			"effort": "low",
		},
		"input": []map[string]interface{}{
			{
				"role": "developer",
				"content": []map[string]string{
					{
						"type": "input_text",
						"text": "You write polished blog drafts for a SaaS product. Return valid JSON only.",
					},
				},
			},
			{
				"role": "user",
				"content": []map[string]string{
					{
						"type": "input_text",
						"text": "Write a blog draft based on this prompt:\n\n" + body.Prompt + "\n\nReturn JSON with title, slug, excerpt, and content. The content should be HTML with headings and paragraphs.",
					},
				},
			},
		},
		"text": map[string]interface{}{
			"format": map[string]interface{}{
				"type":   "json_schema",
				"name":   "blog_post_draft",
				"strict": true,
				"schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title": map[string]string{
							"type": "string",
						},
						"slug": map[string]string{
							"type": "string",
						},
						"excerpt": map[string]string{
							"type": "string",
						},
						"content": map[string]string{
							"type": "string",
						},
					},
					"required":             []string{"title", "slug", "excerpt", "content"},
					"additionalProperties": false,
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, "https://api.openai.com/v1/responses", bytes.NewReader(bodyBytes))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	req.Header.Set("Authorization", "Bearer "+h.cfg.OpenAIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "failed to reach OpenAI")
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "failed to read OpenAI response")
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return fiber.NewError(fiber.StatusBadGateway, string(raw))
	}

	var response struct {
		OutputText string `json:"output_text"`
		Output     []struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "invalid OpenAI response")
	}

	text := strings.TrimSpace(response.OutputText)
	if text == "" {
		for _, item := range response.Output {
			for _, content := range item.Content {
				if content.Type == "output_text" && strings.TrimSpace(content.Text) != "" {
					text = strings.TrimSpace(content.Text)
					break
				}
			}
			if text != "" {
				break
			}
		}
	}
	if text == "" {
		return fiber.NewError(fiber.StatusBadGateway, "empty OpenAI response")
	}

	var generated blogAIResponse
	if err := json.Unmarshal([]byte(text), &generated); err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "OpenAI returned invalid JSON")
	}

	if generated.Title == "" {
		generated.Title = "Generated Draft"
	}
	if generated.Slug == "" {
		generated.Slug = slugify(generated.Title)
	}
	generated.Slug = slugify(generated.Slug)

	return c.JSON(generated)
}

func applyPublishedState(post *domain.BlogPost) {
	if post.Status == "published" {
		if post.PublishedAt == nil {
			now := time.Now()
			post.PublishedAt = &now
		}
		return
	}

	post.PublishedAt = nil
}
