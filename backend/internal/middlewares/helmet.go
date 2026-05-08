package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/helmet"
)

func (Middlewares) Helmet() fiber.Handler {
	return helmet.New()
}
