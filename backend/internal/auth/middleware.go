package auth

import (
	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// RequireAuth gates a route behind a live session: jwtware verifies the bearer
// token's signature and parses its claims, then the success handler checks the
// token against the session store (via Service.ValidateToken) and exposes the
// identity through the httpx.Local* keys, which every handler reads back with
// httpx.Identity.
//
// The identity lives in the request context only: it is fully derived from the
// (already validated) bearer token, so persisting it in a server-side session
// store would just add a storage round-trip per request.
//
// The session check runs in the success handler, not in a TokenProcessorFunc,
// so it can use the request's own context (c) for the cache/database lookups
// instead of a context captured at startup — cancelling and deadlining the
// lookup with the request.
func (m *Module) RequireAuth() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(m.cfg.JWTSecret)},
		SuccessHandler: func(c fiber.Ctx) error {
			jwtToken := jwtware.FromContext(c)
			claims, ok := jwtToken.Claims.(jwt.MapClaims)
			if !ok {
				return c.SendStatus(fiber.StatusUnauthorized)
			}

			if _, err := m.service.ValidateToken(c, jwtToken.Raw); err != nil {
				return c.SendStatus(fiber.StatusUnauthorized)
			}

			userID, _ := claims["id"].(string)
			role, _ := claims["role"].(string)
			if userID == "" || role == "" {
				return c.SendStatus(fiber.StatusUnauthorized)
			}

			c.Locals(httpx.LocalUserID, userID)
			c.Locals(httpx.LocalToken, jwtToken.Raw)
			c.Locals(httpx.LocalRole, role)

			return c.Next()
		},
	})
}
