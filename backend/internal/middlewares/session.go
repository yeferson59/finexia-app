package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

func (m *Middlewares) Session() fiber.Handler {
	return session.New(session.Config{
		Storage:        m.storage,
		CookieHTTPOnly: true,
		CookieSecure:   m.envs.Environment == "production",
	})
}
