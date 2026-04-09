package handlers

import (
	"net/http"
	"time"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/afa/blueprint/backend/pkg/config"
	"github.com/gofiber/fiber/v2"
)

type ToolsHandler struct {
	tools domain.AdminToolRepository
	cfg   *config.Config
}

func NewToolsHandler(tools domain.AdminToolRepository, cfg *config.Config) *ToolsHandler {
	return &ToolsHandler{tools: tools, cfg: cfg}
}

func (h *ToolsHandler) toolURL(tool domain.AdminTool) string {
	if tool.URL != "" {
		return tool.URL
	}
	switch tool.Name {
	case "pgweb":
		return h.cfg.PgwebURL
	case "redis-commander", "redis_commander":
		return h.cfg.RedisCommanderURL
	case "minio":
		return h.cfg.MinioURL
	case "grafana":
		return h.cfg.GrafanaURL
	case "prometheus":
		return h.cfg.PrometheusURL
	}
	return ""
}

func (h *ToolsHandler) ListTools(c *fiber.Ctx) error {
	tools, err := h.tools.List(c.Context(), true)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	// Substitute URLs from config where empty
	for i := range tools {
		if tools[i].URL == "" {
			tools[i].URL = h.toolURL(tools[i])
		}
	}
	return c.JSON(fiber.Map{"data": tools})
}

func (h *ToolsHandler) CreateTool(c *fiber.Ctx) error {
	var tool domain.AdminTool
	if err := c.BodyParser(&tool); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := h.tools.Create(c.Context(), &tool); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(tool)
}

func (h *ToolsHandler) UpdateTool(c *fiber.Ctx) error {
	var tool domain.AdminTool
	if err := c.BodyParser(&tool); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	tool.ID = c.Params("id")
	if err := h.tools.Update(c.Context(), &tool); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(tool)
}

func (h *ToolsHandler) DeleteTool(c *fiber.Ctx) error {
	if err := h.tools.Delete(c.Context(), c.Params("id")); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ToolsHandler) PingTool(c *fiber.Ctx) error {
	tools, err := h.tools.List(c.Context(), false)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	id := c.Params("id")
	var target *domain.AdminTool
	for i := range tools {
		if tools[i].ID == id {
			target = &tools[i]
			break
		}
	}
	if target == nil {
		return fiber.NewError(fiber.StatusNotFound, "tool not found")
	}

	url := h.toolURL(*target)
	if url == "" {
		return fiber.NewError(fiber.StatusBadRequest, "tool has no URL configured")
	}

	client := &http.Client{Timeout: 3 * time.Second}
	start := time.Now()
	resp, err := client.Get(url)
	latency := time.Since(start).Milliseconds()

	if err != nil || resp.StatusCode >= 500 {
		return c.JSON(fiber.Map{"status": "down", "latency_ms": latency})
	}
	return c.JSON(fiber.Map{"status": "up", "latency_ms": latency})
}
