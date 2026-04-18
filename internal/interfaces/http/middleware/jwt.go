// ABOUTME: JWT authentication middleware for protected routes using Fiber.
// ABOUTME: Validates Bearer token and writes a refreshed token to X-Refresh-Token header.

package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/amavis442/gasolina-api/auth"
)

func JWT(secret string, ttl time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing bearer token"})
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if _, err := auth.ValidateToken(tokenStr, secret); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}
		refreshed, err := auth.GenerateToken(secret, ttl)
		if err == nil {
			c.Set("X-Refresh-Token", refreshed)
		}
		return c.Next()
	}
}
