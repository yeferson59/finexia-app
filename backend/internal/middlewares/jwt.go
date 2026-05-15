package middlewares

import (
	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/golang-jwt/jwt/v5"
)

func (m *Middlewares) JWT() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(m.envs.JWTSecret)},
		TokenProcessorFunc: func(token string) (string, error) {
			return m.svc.ValidateToken(m.ctx, token)
		},
		SuccessHandler: func(c fiber.Ctx) error {
			sess := session.FromContext(c)
			jwtToken := jwtware.FromContext(c)
			claims := jwtToken.Claims.(jwt.MapClaims)

			sess.Set("userID", claims["id"].(string))
			sess.Set("token", jwtToken.Raw)
			sess.Set("role", claims["role"].(string))

			return c.Next()
		},
	})
}
