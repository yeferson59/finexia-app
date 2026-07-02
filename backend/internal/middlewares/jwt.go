package middlewares

import (
	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// Locals keys under which the JWT middleware exposes the authenticated
// identity to downstream middlewares and handlers. The identity lives in the
// request context only: it is fully derived from the (already validated)
// bearer token, so persisting it in a server-side session store would just add
// a storage round-trip per request.
const (
	LocalUserID = "auth_user_id"
	LocalToken  = "auth_token"
	LocalRole   = "auth_role"
)

func (m *Middlewares) JWT() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(m.envs.JWTSecret)},
		TokenProcessorFunc: func(token string) (string, error) {
			return m.svc.ValidateToken(m.ctx, token)
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
