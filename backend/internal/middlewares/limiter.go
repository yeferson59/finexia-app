package middlewares

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/session"
)

func (m *Middlewares) Limiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Storage:    m.storage,
		Max:        60,
		Expiration: 1 * time.Minute,
	})
}

func (m *Middlewares) AuthLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Storage:    m.storage,
		Max:        10,
		Expiration: 15 * time.Minute,
	})
}

func (m *Middlewares) UserLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Storage:    m.storage,
		Max:        200,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			sess := session.FromContext(c)
			if userID, ok := sess.Get("userID").(string); ok && userID != "" {
				return "user_limit:" + userID
			}
			return c.IP()
		},
	})
}
