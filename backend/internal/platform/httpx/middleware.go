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
		Logger: new(logger),
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

// OrPassThrough returns h, or a middleware that does nothing when h is nil.
//
// Modules use it for the rate limiters the composition root injects, and only
// for those: a missing limiter costs a route its per-user budget but leaves it
// correct and still guarded, so degrading beats refusing to start. Route
// guards get the opposite treatment — each module's New panics when one is
// missing, because a route that answers unguarded (or silently stops
// answering) is not a degradation.
func OrPassThrough(h fiber.Handler) fiber.Handler {
	if h != nil {
		return h
	}

	return func(c fiber.Ctx) error { return c.Next() }
}
