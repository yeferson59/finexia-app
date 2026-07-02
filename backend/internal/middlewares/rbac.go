package middlewares

import (
	"slices"

	"github.com/gofiber/fiber/v3"
)

// RequireRole allows only requests whose authenticated role matches one of the
// given roles. Must be placed after the JWT middleware in the handler chain.
func (m *Middlewares) RequireRole(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		role, _ := c.Locals(LocalRole).(string)

		if slices.Contains(roles, role) {
			return c.Next()
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Forbidden",
			"details": "insufficient privileges",
		})
	}
}

// RequireAdmin is a convenience wrapper that only allows the "admin" role.
func (m *Middlewares) RequireAdmin() fiber.Handler {
	return m.RequireRole("admin")
}
