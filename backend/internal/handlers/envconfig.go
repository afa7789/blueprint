package handlers

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/afa/blueprint/backend/pkg/config"
	"github.com/gofiber/fiber/v2"
)

type EnvConfigHandler struct {
	settings domain.SecuritySettingRepository
	flags    domain.FeatureFlagRepository
	cfg      *config.Config
}

func NewEnvConfigHandler(settings domain.SecuritySettingRepository, flags domain.FeatureFlagRepository, cfg *config.Config) *EnvConfigHandler {
	return &EnvConfigHandler{settings: settings, flags: flags, cfg: cfg}
}

type envEntry struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	IsSet       bool   `json:"is_set"`
	IsSecret    bool   `json:"is_secret"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// knownEnvVars returns all known ENV vars with metadata
func knownEnvVars() []envEntry {
	return []envEntry{
		// Core
		{Key: "PORT", Category: "core", Description: "API server port", Required: true},
		{Key: "ENV", Category: "core", Description: "Environment (development/production)"},
		{Key: "FRONTEND_URL", Category: "core", Description: "Frontend URL for CORS"},

		// Database
		{Key: "DATABASE_URL", Category: "database", Description: "PostgreSQL connection string", Required: true, IsSecret: true},
		{Key: "DATABASE_MIGRATION_URL", Category: "database", Description: "Migration DB URL (defaults to DATABASE_URL)", IsSecret: true},
		{Key: "REDIS_URL", Category: "database", Description: "Redis connection string", IsSecret: true},

		// Auth
		{Key: "JWT_SECRET", Category: "auth", Description: "Secret for signing JWT tokens", Required: true, IsSecret: true},
		{Key: "JWT_EXPIRY", Category: "auth", Description: "Access token expiry in minutes (default: 15)"},
		{Key: "REFRESH_EXPIRY", Category: "auth", Description: "Refresh token expiry in minutes (default: 10080)"},
		{Key: "EMAIL_VERIFICATION_REQUIRED", Category: "auth", Description: "Require email verification on register (true/false)"},

		// Payments
		{Key: "STRIPE_KEY", Category: "payments", Description: "Stripe secret key", IsSecret: true},
		{Key: "STRIPE_WEBHOOK_SECRET", Category: "payments", Description: "Stripe webhook signing secret", IsSecret: true},

		// Storage
		{Key: "STORAGE_TYPE", Category: "storage", Description: "Storage backend (local/s3)"},
		{Key: "UPLOAD_DIR", Category: "storage", Description: "Local upload directory path"},
		{Key: "AWS_BUCKET", Category: "storage", Description: "S3 bucket name"},
		{Key: "AWS_REGION", Category: "storage", Description: "AWS region"},

		// AI
		{Key: "OPENAI_KEY", Category: "ai", Description: "OpenAI API key for blog generation", IsSecret: true},

		// Notifications
		{Key: "TELEGRAM_BOT_TOKEN", Category: "notifications", Description: "Telegram bot token for alerts", IsSecret: true},
		{Key: "TELEGRAM_CHAT_ID", Category: "notifications", Description: "Telegram chat ID for alerts"},
		{Key: "SMTP_HOST", Category: "notifications", Description: "SMTP server host"},
		{Key: "SMTP_PORT", Category: "notifications", Description: "SMTP server port"},

		// Tools
		{Key: "PGWEB_URL", Category: "tools", Description: "pgweb URL"},
		{Key: "REDIS_COMMANDER_URL", Category: "tools", Description: "Redis Commander URL"},
		{Key: "MINIO_URL", Category: "tools", Description: "MinIO Console URL"},
		{Key: "GRAFANA_URL", Category: "tools", Description: "Grafana URL"},
		{Key: "PROMETHEUS_URL", Category: "tools", Description: "Prometheus URL"},

		// Rate Limiting
		{Key: "RATE_LIMIT_API", Category: "security", Description: "API rate limit per minute per IP (default: 60)"},
		{Key: "RATE_LIMIT_AUTH", Category: "security", Description: "Auth rate limit per minute per email (default: 10)"},
		{Key: "RATE_LIMIT_REGISTER", Category: "security", Description: "Register rate limit per hour per email (default: 5)"},
		{Key: "RATE_LIMIT_FORGOT", Category: "security", Description: "Forgot-password rate limit per hour (default: 3)"},
		{Key: "MAX_REQUEST_BODY_MB", Category: "security", Description: "Max request body size in MB (default: 10)"},
	}
}

// GetEnvStatus returns all known ENV vars with their set/unset status.
// Secret values are masked.
func (h *EnvConfigHandler) GetEnvStatus(c *fiber.Ctx) error {
	vars := knownEnvVars()
	for i := range vars {
		val := os.Getenv(vars[i].Key)
		vars[i].IsSet = val != ""
		if vars[i].IsSecret {
			if val != "" {
				vars[i].Value = "***"
			}
		} else {
			vars[i].Value = val
		}
	}
	return c.JSON(fiber.Map{"data": vars})
}

// ExportEnv exports all DB-configurable settings (security_settings + feature_flags)
// as a .env-compatible text file.
func (h *EnvConfigHandler) ExportEnv(c *fiber.Ctx) error {
	var lines []string

	lines = append(lines, "# Blueprint Configuration Export")
	lines = append(lines, "# Generated from admin panel — DB-backed settings")
	lines = append(lines, "")

	// Security settings
	settings, err := h.settings.GetAll(c.Context())
	if err == nil && len(settings) > 0 {
		lines = append(lines, "# --- Security Settings ---")
		for _, s := range settings {
			desc := ""
			if s.Description != nil {
				desc = " # " + *s.Description
			}
			lines = append(lines, fmt.Sprintf("%s=%s%s", strings.ToUpper(s.Key), s.Value, desc))
		}
		lines = append(lines, "")
	}

	// Feature flags
	flags, err := h.flags.GetAll(c.Context())
	if err == nil && len(flags) > 0 {
		lines = append(lines, "# --- Feature Flags ---")
		for _, f := range flags {
			val := "false"
			if f.Enabled {
				val = "true"
			}
			lines = append(lines, fmt.Sprintf("FLAG_%s=%s", strings.ToUpper(f.Key), val))
		}
		lines = append(lines, "")
	}

	// ENV vars (non-secret current values)
	lines = append(lines, "# --- Environment Variables (current) ---")
	vars := knownEnvVars()
	sort.Slice(vars, func(i, j int) bool { return vars[i].Category < vars[j].Category })
	currentCat := ""
	for _, v := range vars {
		if v.Category != currentCat {
			currentCat = v.Category
			lines = append(lines, fmt.Sprintf("\n# [%s]", currentCat))
		}
		val := os.Getenv(v.Key)
		if v.IsSecret {
			if val != "" {
				val = "CHANGE_ME"
			}
		}
		comment := ""
		if v.Description != "" {
			comment = " # " + v.Description
		}
		lines = append(lines, fmt.Sprintf("%s=%s%s", v.Key, val, comment))
	}

	c.Set("Content-Type", "text/plain")
	c.Set("Content-Disposition", "attachment; filename=blueprint.env")
	return c.SendString(strings.Join(lines, "\n") + "\n")
}

// ImportEnv imports settings from a .env-format text body.
// Only updates DB-backed settings (security_settings + feature_flags).
// ENV vars are NOT updated (require server restart).
func (h *EnvConfigHandler) ImportEnv(c *fiber.Ctx) error {
	body := string(c.Body())
	lines := strings.Split(body, "\n")

	updated := 0
	skipped := 0
	var errors []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Strip inline comments
		if idx := strings.Index(value, " #"); idx > 0 {
			value = strings.TrimSpace(value[:idx])
		}

		// Feature flags: FLAG_xxx=true/false
		if strings.HasPrefix(key, "FLAG_") {
			flagKey := strings.ToLower(strings.TrimPrefix(key, "FLAG_"))
			enabled := value == "true"
			if err := h.flags.Set(c.Context(), flagKey, enabled); err != nil {
				errors = append(errors, fmt.Sprintf("flag %s: %v", flagKey, err))
			} else {
				updated++
			}
			continue
		}

		// Security settings: lowercase keys
		lowerKey := strings.ToLower(key)
		if _, err := h.settings.GetByKey(c.Context(), lowerKey); err == nil {
			if err := h.settings.Update(c.Context(), lowerKey, value); err != nil {
				errors = append(errors, fmt.Sprintf("setting %s: %v", lowerKey, err))
			} else {
				updated++
			}
			continue
		}

		// ENV vars can't be updated at runtime
		skipped++
	}

	result := fiber.Map{
		"updated": updated,
		"skipped": skipped,
		"message": fmt.Sprintf("%d settings updated, %d ENV vars skipped (require restart)", updated, skipped),
	}
	if len(errors) > 0 {
		result["errors"] = errors
	}

	return c.JSON(result)
}
