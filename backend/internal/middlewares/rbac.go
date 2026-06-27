package middlewares

import (
	"slices"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

// RequireRole allows only requests whose session role matches one of the given roles.
// Must be placed after the Session and JWT middlewares in the handler chain.
func (m *Middlewares) RequireRole(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		sess := session.FromContext(c)
		role, _ := sess.Get("role").(string)

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
