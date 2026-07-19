package httpx

import (
	"os"
	"strings"
	"time"

	zerologmw "github.com/gofiber/contrib/v3/zerolog"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/responsetime"
	"github.com/rs/zerolog"
	ratelimit "github.com/yeferson59/goratelimit"
)

// Recovery converts panics into 500 responses.
func Recovery() fiber.Handler {
	return recover.New()
}

// RequestID stamps every request with an X-Request-ID header.
func RequestID() fiber.Handler {
	return requestid.New()
}

// ResponseTime adds the X-Response-Time header, skipping health checks.
func ResponseTime() fiber.Handler {
	return responsetime.New(responsetime.Config{
		Next: func(c fiber.Ctx) bool {
			return strings.Contains(c.Path(), "/health")
		},
	})
}

// Logger logs every request through zerolog to stderr.
func Logger() fiber.Handler {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	return zerologmw.New(zerologmw.Config{
		Logger: &logger,
	})
}

// CORS configures cross-origin access from the given origins.
func CORS(allowOrigins []string, allowCredentials bool) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowCredentials: allowCredentials,
	})
}

// Helmet sets the standard security headers.
func Helmet() fiber.Handler {
	return helmet.New()
}

// RateLimiter limits each client (by IP) to max requests per window.
func RateLimiter(max int, window time.Duration, disableHeaders bool) fiber.Handler {
	return ratelimit.New(ratelimit.Config{
		Max:            max,
		Expiration:     window,
		DisableHeaders: disableHeaders,
	})
}

// KeyedRateLimiter is RateLimiter with a custom key function, for callers
// that limit by something other than the client IP (e.g. the user ID).
func KeyedRateLimiter(max int, window time.Duration, key func(fiber.Ctx) string) fiber.Handler {
	return ratelimit.New(ratelimit.Config{
		Max:          max,
		Expiration:   window,
		KeyGenerator: key,
	})
}
