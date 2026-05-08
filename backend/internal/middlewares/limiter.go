package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func (Middlewares) Limiter() fiber.Handler {
	return limiter.New()
}
