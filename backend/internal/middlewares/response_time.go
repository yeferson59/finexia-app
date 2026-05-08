package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/responsetime"
)

func (Middlewares) ResponseTime() fiber.Handler {
	return responsetime.New(responsetime.Config{
		Next: func(c fiber.Ctx) bool {
			return strings.Contains(c.Path(), "/health")
		},
	})
}
