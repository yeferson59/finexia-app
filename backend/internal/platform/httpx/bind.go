package httpx

import (
	"github.com/gofiber/fiber/v3"
)

func Bind[a any](c fiber.Ctx) (a, error) {
	var result a

	if err := c.Bind().All(&result); err != nil {
		return result, err
	}

	return result, nil
}
