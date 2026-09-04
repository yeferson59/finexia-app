package auth

import (
	"strings"

	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// RequireAuth gates a route behind a live session: jwtware verifies the bearer
// token's signature and parses its claims, then the success handler checks the
// token against the session store (via Service.ValidateToken) and exposes the
// identity through the httpx.Local* keys, which every handler reads back with
// httpx.Identity.
//
// The identity lives in the request context only: it is fully derived from the
// (already validated) bearer token, so persisting it in a server-side session
// store would just add a storage round-trip per request.
//
// The session check runs in the success handler, not in a TokenProcessorFunc,
// so it can use the request's own context (c) for the cache/database lookups
// instead of a context captured at startup — cancelling and deadlining the
// lookup with the request.
func (m *Module) RequireAuth() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(m.cfg.JWTSecret)},
		// jwtware's default answers a *missing* Authorization header with 400
		// and a present-but-bad one with 401, in plain text. Both are wrong
		// here: no credential is not a malformed request, it is an
		// unauthenticated one, and every other error in this API travels in the
		// httpx envelope.
		//
		// The status is what an MCP client acts on. A client's first call to
		// /mcp carries no credential by design — it expects the 401 that names
		// the scheme and starts the authorization flow — and reads a 400 as a
		// broken server instead, which is exactly what "MCP server returned
		// 400" in Claude's connector is.
		ErrorHandler: func(c fiber.Ctx, err error) error {
			c.Set(fiber.HeaderWWWAuthenticate, `Bearer realm="finexia"`)

			return httpx.Unauthorized(c, "Unauthorized", err.Error())
		},
		SuccessHandler: func(c fiber.Ctx) error {
			jwtToken := jwtware.FromContext(c)
			claims, ok := jwtToken.Claims.(jwt.MapClaims)
			if !ok {
				return c.SendStatus(fiber.StatusUnauthorized)
			}

			if _, err := m.service.ValidateToken(c, jwtToken.Raw); err != nil {
				return c.SendStatus(fiber.StatusUnauthorized)
			}

			userID, _ := claims["id"].(string)
			role, _ := claims["role"].(string)
			if userID == "" || role == "" {
				return c.SendStatus(fiber.StatusUnauthorized)
			}

			c.Locals(httpx.LocalUserID, userID)
			c.Locals(httpx.LocalToken, jwtToken.Raw)
			c.Locals(httpx.LocalRole, role)

			return c.Next()
		},
	})
}

// RequireMCPAuth gates /mcp, which is reached by two kinds of caller: a browser
// carrying the session's access token, and an MCP client carrying one of the
// personal access tokens the settings screen mints.
//
// The credential is chosen by prefix rather than by trying both. A token that
// says fnx_mcp_ is answered by the store that owns those, and a JWT by the
// existing guard, so a rejected request fails against the scheme it actually
// claimed — otherwise every bad MCP token would be reported as a bad JWT, and
// the logs would name the wrong subsystem for every one of them.
//
// Both paths leave the same three locals behind, which is what lets everything
// downstream — bindCaller, the rate limiter, httpx.Identity — stay unaware that
// there is more than one way to authenticate here.
func (m *Module) RequireMCPAuth() fiber.Handler {
	jwtGuard := m.RequireAuth()

	return func(c fiber.Ctx) error {
		raw := bearerToken(c)
		if !looksLikeMCPToken(raw) {
			return jwtGuard(c)
		}

		userID, role, err := m.service.AuthenticateMCPToken(c, raw)
		if err != nil {
			// Deliberately one answer for every failure: unknown, expired,
			// revoked and belonging-to-a-banned-account are the same 401, so
			// the response cannot be used to tell live tokens from dead ones.
			return httpx.Unauthorized(c, "Invalid MCP token", "the token is unknown, expired or was revoked")
		}

		c.Locals(httpx.LocalUserID, userID.String())
		c.Locals(httpx.LocalToken, raw)
		c.Locals(httpx.LocalRole, role)

		return c.Next()
	}
}

// MCPGuard adapts this module to the one-method guard dependency the mcp module
// declares, substituting RequireMCPAuth for RequireAuth.
//
// It exists so mcp keeps depending on "something that guards a route" and
// nothing else: which credentials satisfy that guard is an auth decision, and
// this is where it stays.
type MCPGuard struct{ module *Module }

// MCPAuth returns the guard the composition root passes to the mcp module.
func (m *Module) MCPAuth() MCPGuard {
	return MCPGuard{module: m}
}

func (g MCPGuard) RequireAuth() fiber.Handler {
	return g.module.RequireMCPAuth()
}

// bearerToken returns the credential presented in the Authorization header,
// stripped of its scheme. An absent or non-Bearer header yields "", which no
// prefix check will claim.
func bearerToken(c fiber.Ctx) string {
	const scheme = "Bearer "

	header := c.Get(fiber.HeaderAuthorization)
	if len(header) <= len(scheme) || !strings.EqualFold(header[:len(scheme)], scheme) {
		return ""
	}

	return strings.TrimSpace(header[len(scheme):])
}
