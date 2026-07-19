package auth

import (
	"slices"
	"time"

	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// Locals keys under which the JWT middleware exposes the authenticated
// identity to downstream middlewares and handlers. The identity lives in the
// request context only: it is fully derived from the (already validated)
// bearer token, so persisting it in a server-side session store would just add
// a storage round-trip per request. Exported because legacy consumers (the
// per-user rate limiter, the legacy handlers) read them until their domains
// migrate.
const (
	LocalUserID = "auth_user_id"
	LocalToken  = "auth_token"
	LocalRole   = "auth_role"
)

// RequireAuth gates a route behind a live session: it verifies the bearer
// token's signature and checks it against the session store (via
// Service.ValidateToken), then exposes the identity through the Local* keys.
func (m *Module) RequireAuth() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(m.cfg.JWTSecret)},
		TokenProcessorFunc: func(token string) (string, error) {
			return m.service.ValidateToken(m.ctx, token)
		},
		SuccessHandler: func(c fiber.Ctx) error {
			jwtToken := jwtware.FromContext(c)
			claims, ok := jwtToken.Claims.(jwt.MapClaims)
			if !ok {
				return c.SendStatus(fiber.StatusUnauthorized)
			}

			userID, _ := claims["id"].(string)
			role, _ := claims["role"].(string)
			if userID == "" || role == "" {
				return c.SendStatus(fiber.StatusUnauthorized)
			}

			c.Locals(LocalUserID, userID)
			c.Locals(LocalToken, jwtToken.Raw)
			c.Locals(LocalRole, role)

			return c.Next()
		},
	})
}

// RequireRole allows only requests whose authenticated role matches one of the
// given roles. Must be placed after RequireAuth in the handler chain.
func (m *Module) RequireRole(roles ...string) fiber.Handler {
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
func (m *Module) RequireAdmin() fiber.Handler {
	return m.RequireRole("admin")
}

// authLimiter rate-limits the public auth endpoints (credential guessing,
// token guessing, mail bombing). Same policy as the legacy AuthLimiter.
func (m *Module) authLimiter() fiber.Handler {
	return httpx.RateLimiter(10, 15*time.Minute, true)
}
