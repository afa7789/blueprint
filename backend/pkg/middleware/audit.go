package middleware

import (
	"encoding/json"
	"strings"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/gofiber/fiber/v2"
)

func AuditLog(repo domain.AuditLogRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Execute the handler first
		err := c.Next()

		// Only log mutating actions (POST, PUT, DELETE)
		method := c.Method()
		if method == "GET" {
			return err
		}

		userID, _ := c.Locals("user_id").(string)

		// Parse action from method + path
		path := c.Path()
		action := method + " " + path

		// Extract resource and resource ID from path
		// e.g., /api/v1/admin/users/123 → resource="users", resource_id="123"
		parts := strings.Split(strings.Trim(path, "/"), "/")
		var resource, resourceID *string
		if len(parts) >= 4 {
			r := parts[3] // after api/v1/admin/
			resource = &r
		}
		if len(parts) >= 5 {
			rid := parts[4]
			resourceID = &rid
		}

		// Build details
		details := map[string]interface{}{
			"status_code": c.Response().StatusCode(),
		}
		if len(c.Body()) > 0 && len(c.Body()) < 10000 {
			var body interface{}
			if json.Unmarshal(c.Body(), &body) == nil {
				details["request_body"] = body
			}
		}
		detailsJSON, _ := json.Marshal(details)

		ip := c.IP()
		entry := &domain.AuditLog{
			UserID:     &userID,
			Action:     action,
			Resource:   resource,
			ResourceID: resourceID,
			Details:    detailsJSON,
			IPAddress:  &ip,
		}

		// Non-blocking
		go func() {
			_ = repo.Create(c.Context(), entry)
		}()

		return err
	}
}
