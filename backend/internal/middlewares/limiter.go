package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func (m *Middlewares) Limiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Storage: m.storage,
		Max:     10,
	})
}
