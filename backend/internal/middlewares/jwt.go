package middlewares

import (
	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
)

func (m *Middlewares) JWT() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(m.envs.JWTSecret)},
		TokenProcessorFunc: func(token string) (string, error) {
			return m.svc.ValidateToken(m.ctx, token)
		},
	})
}
