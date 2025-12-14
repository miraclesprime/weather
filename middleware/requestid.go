package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestID returns middleware that ensures each request has an X-Request-ID header
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
			c.Set("X-Request-ID", id)
		}
		c.Locals("requestid", id)
		return c.Next()
	}
}
