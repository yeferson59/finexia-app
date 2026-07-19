package middlewares

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/auth"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

func (m *Middlewares) Limiter() fiber.Handler {
	return httpx.RateLimiter(60, 1*time.Minute, false)
}

func (m *Middlewares) UserLimiter() fiber.Handler {
	// Keyed by user ID (set by the auth module's RequireAuth middleware); that
	// coupling to auth is why this limiter stays here instead of moving
	// wholesale into httpx.
	return httpx.KeyedRateLimiter(200, 1*time.Minute, func(c fiber.Ctx) string {
		if userID, ok := c.Locals(auth.LocalUserID).(string); ok && userID != "" {
			return "user_limit:" + userID
		}

		return c.IP()
	})
}
