package middlewares

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

func (m *Middlewares) Limiter() fiber.Handler {
	return httpx.RateLimiter(m.storage, 60, 1*time.Minute)
}

func (m *Middlewares) AuthLimiter() fiber.Handler {
	return httpx.RateLimiter(m.storage, 10, 15*time.Minute)
}

func (m *Middlewares) UserLimiter() fiber.Handler {
	// Keyed by user ID (set by the JWT middleware); that coupling to auth is
	// why this limiter stays here instead of moving wholesale into httpx.
	return httpx.KeyedRateLimiter(m.storage, 200, 1*time.Minute, func(c fiber.Ctx) string {
		if userID, ok := c.Locals(LocalUserID).(string); ok && userID != "" {
			return "user_limit:" + userID
		}
		return c.IP()
	})
}
