package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/gofiber/fiber/v2"
)

type LogsHandler struct {
	appLogs   domain.AppLogRepository
	auditLogs domain.AuditLogRepository
	logConfig domain.LogConfigRepository
}

type cleanupLogsRequest struct {
	All bool `json:"all"`
}

func parseFlexibleTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, value); err == nil {
			return &parsed
		}
	}

	return nil
}

func NewLogsHandler(
	appLogs domain.AppLogRepository,
	auditLogs domain.AuditLogRepository,
	logConfig domain.LogConfigRepository,
) *LogsHandler {
	return &LogsHandler{
		appLogs:   appLogs,
		auditLogs: auditLogs,
		logConfig: logConfig,
	}
}

func (h *LogsHandler) ListLogs(c *fiber.Ctx) error {
	page, limit, offset := paginate(c)

	var level, source, search *string
	if v := c.Query("level"); v != "" {
		level = &v
	}
	if v := c.Query("source"); v != "" {
		source = &v
	}
	if v := c.Query("search"); v != "" {
		search = &v
	}

	var from, to *time.Time
	if v := c.Query("from"); v != "" {
		from = parseFlexibleTime(v)
	}
	if v := c.Query("to"); v != "" {
		to = parseFlexibleTime(v)
	}

	logs, total, err := h.appLogs.List(c.Context(), level, source, search, from, to, offset, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": logs, "total": total, "page": page, "limit": limit})
}

func (h *LogsHandler) StreamLogs(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		var lastID int64
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			logs, err := h.appLogs.LatestSince(context.Background(), lastID, 50)
			if err != nil || len(logs) == 0 {
				if _, err := fmt.Fprintf(w, ": keepalive\n\n"); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
				continue
			}

			for _, entry := range logs {
				data, err := json.Marshal(entry)
				if err != nil {
					continue
				}
				if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
					return
				}
				if entry.ID > lastID {
					lastID = entry.ID
				}
			}
			if err := w.Flush(); err != nil {
				return
			}
		}
	})
	return nil
}

func (h *LogsHandler) ListAuditLogs(c *fiber.Ctx) error {
	page, limit, offset := paginate(c)

	var userID, action, resource *string
	if v := c.Query("user_id"); v != "" {
		userID = &v
	}
	if v := c.Query("action"); v != "" {
		action = &v
	}
	if v := c.Query("resource"); v != "" {
		resource = &v
	}

	var from, to *time.Time
	if v := c.Query("from"); v != "" {
		from = parseFlexibleTime(v)
	}
	if v := c.Query("to"); v != "" {
		to = parseFlexibleTime(v)
	}

	logs, total, err := h.auditLogs.List(c.Context(), userID, action, resource, from, to, offset, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": logs, "total": total, "page": page, "limit": limit})
}

func (h *LogsHandler) GetLogConfig(c *fiber.Ctx) error {
	cfg, err := h.logConfig.Get(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(cfg)
}

func (h *LogsHandler) UpdateLogConfig(c *fiber.Ctx) error {
	var cfg domain.LogConfig
	if err := c.BodyParser(&cfg); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := h.logConfig.Update(c.Context(), &cfg); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(cfg)
}

func (h *LogsHandler) CleanupLogs(c *fiber.Ctx) error {
	var req cleanupLogsRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}

	if req.All {
		deleted, err := h.appLogs.Cleanup(c.Context(), 0)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(fiber.Map{"deleted": deleted})
	}

	cfg, err := h.logConfig.Get(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	deleted, err := h.appLogs.Cleanup(c.Context(), cfg.RetentionDays)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"deleted": deleted})
}
