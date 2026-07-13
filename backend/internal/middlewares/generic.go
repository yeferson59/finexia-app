package middlewares

// The generic middlewares live in platform/httpx since Fase 1 of the
// architecture migration; these methods only keep the legacy Middlewares
// surface stable until routes are wired per module.

import (
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

func (Middlewares) Recovery() fiber.Handler {
	return httpx.Recovery()
}

func (Middlewares) RequestID() fiber.Handler {
	return httpx.RequestID()
}

func (Middlewares) ResponseTime() fiber.Handler {
	return httpx.ResponseTime()
}

func (Middlewares) Logger() fiber.Handler {
	return httpx.Logger()
}

func (m *Middlewares) CORS() fiber.Handler {
	return httpx.CORS(m.envs.CORSOrigin, m.envs.CORSEnabled)
}

func (Middlewares) Helmet() fiber.Handler {
	return httpx.Helmet()
}
