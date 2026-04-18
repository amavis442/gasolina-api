// ABOUTME: Shared response helpers for writing JSON and error responses via Fiber.
// ABOUTME: All handlers must use these — no ad-hoc JSON writing elsewhere.

package response

import (
	"github.com/gofiber/fiber/v2"
)

func WriteJSON(c *fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(data)
}

func WriteError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}
