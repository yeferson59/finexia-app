package auth

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// Deps carries everything the auth module needs from the composition root.
type Deps struct {
	DB      *pgxpool.Pool
	Cfg     Config
	Storage fiber.Storage
	Mail    Mailer
	Geo     GeoLocator
	Log     logger.Logger
	// Waitlist is the marketing module's service: the invitation flow reads
	// and advances the waitlist through it instead of touching the table.
	Waitlist WaitlistStore
	// Users is the user module's service: auth reads the users/roles tables
	// through it instead of querying them. The user module's service is built
	// before auth precisely so this can be an ordinary constructor dependency.
	Users UserReader
	// Limiter is the per-user rate limiter applied to the /users routes this
	// module serves (password change, invitation dashboard).
	Limiter fiber.Handler
}

// Module is the auth domain module: construction via New, HTTP surface via
// Routes, and route guards (RequireAuth/RequireRole) for the rest of the app.
type Module struct {
	cfg     Config
	storage fiber.Storage
	limiter fiber.Handler
	service *service
	handler *handler
}

func New(deps Deps) *Module {
	pg := NewPostgresRepository(deps.DB)
	service := newService(Stores{
		Accounts:       pg,
		Sessions:       pg,
		RefreshTokens:  pg,
		TwoFactor:      pg,
		Verifications:  pg,
		PasswordResets: pg,
		Invitations:    pg,
		MCPTokens:      pg,
		OAuth:          pg,
		Waitlist:       deps.Waitlist,
		Users:          deps.Users,
	}, deps.Cfg, deps.Storage, deps.Mail, deps.Geo, deps.Log)

	return newModule(deps, service)
}

// newModule finishes construction from an already-built service; split out so
// tests can inject fake stores through newService.
//
// This module declares no guard dependency of its own — it is the guard — so
// there is nothing here that must fail the build. The limiter is the one
// injected handler, and it degrades rather than refuses; see
// httpx.OrPassThrough.
func newModule(deps Deps, service *service) *Module {
	return new(Module{
		cfg:     deps.Cfg,
		storage: deps.Storage,
		limiter: httpx.OrPassThrough(deps.Limiter),
		service: service,
		handler: newHandler(service, deps.Cfg),
	})
}

// Service exposes the module's use cases to the composition root and other
// modules (always consumed through interfaces declared by the consumer).
func (m *Module) Service() *service {
	return m.service
}

// Routes registers the /auth group, replicating routes/auth.go: the public
// endpoints first (each behind the auth rate limiter), then the group-local
// RequireAuth gate for the session-bound ones.
func (m *Module) Routes(router fiber.Router) {
	auth := router.Group("/auth")

	authLimiter := httpx.AuthLimiter()

	auth.Post("/register", authLimiter, m.handler.register)
	auth.Post("/login", authLimiter, m.handler.login)
	auth.Post("/refresh", authLimiter, m.handler.refresh)

	// Public second step of a 2FA login: exchanges the short-lived pending
	// token plus a TOTP/recovery code for a session. Rate-limited to blunt
	// code guessing on top of the per-token attempt counter.
	auth.Post("/2fa/login", authLimiter, m.handler.twoFactorLogin)

	// Public invitation flow: validate a token, then accept it by setting a
	// password. Rate-limited to blunt token guessing.
	auth.Get("/invitations", authLimiter, m.handler.validateInvitation)
	auth.Post("/invitations/accept", authLimiter, m.handler.acceptInvitation)

	// Public password recovery flow: request a reset link, validate its
	// token, then confirm with a new password. Rate-limited to blunt both
	// mail-bombing an address and token guessing.
	auth.Post("/password-reset", authLimiter, m.handler.requestPasswordReset)
	auth.Get("/password-reset", authLimiter, m.handler.validatePasswordReset)
	auth.Post("/password-reset/confirm", authLimiter, m.handler.confirmPasswordReset)

	// Public email verification flow: (re)send a link, validate its token,
	// then confirm to mark the email verified. Rate-limited to blunt both
	// mail-bombing an address and token guessing.
	auth.Post("/verify-email", authLimiter, m.handler.requestEmailVerification)
	auth.Get("/verify-email", authLimiter, m.handler.validateEmailVerification)
	auth.Post("/verify-email/confirm", authLimiter, m.handler.confirmEmailVerification)

	auth.Use(m.RequireAuth())

	// Two-factor management, always behind a live session. 2FA is off by
	// default; these endpoints let the user opt in, confirm, and opt out.
	auth.Get("/2fa", m.handler.twoFactorStatus)
	auth.Post("/2fa/setup", m.handler.twoFactorSetup)
	auth.Post("/2fa/enable", m.handler.twoFactorEnable)
	auth.Post("/2fa/disable", m.handler.twoFactorDisable)
	auth.Post("/2fa/recovery-codes", m.handler.twoFactorRecoveryCodes)

	// Personal access tokens for /mcp. The writes go through the per-user
	// limiter for the same reason the password change does: they are the
	// endpoints that mint and destroy credentials.
	auth.Get("/mcp-tokens", m.handler.listMCPTokens)
	auth.Post("/mcp-tokens", m.limiter, m.handler.createMCPToken)
	auth.Post("/mcp-tokens/:id/rotate", m.limiter, m.handler.rotateMCPToken)
	auth.Delete("/mcp-tokens/:id", m.limiter, m.handler.deleteMCPToken)

	// The consent screen's two calls. They sit behind the session guard
	// because that is the entire point of them: the authorization code is
	// minted for whoever is logged in here, and nothing the client sent can
	// name a different user.
	auth.Get("/oauth/consent/:id", m.handler.getOAuthConsent)
	auth.Post("/oauth/consent/:id", m.limiter, m.handler.decideOAuthConsent)

	// Connected applications: what the user has approved, and the button that
	// takes it back.
	auth.Get("/oauth-grants", m.handler.listOAuthGrants)
	auth.Delete("/oauth-grants/:id", m.limiter, m.handler.revokeOAuthGrant)

	auth.Get("/session", m.handler.getSession)
	auth.Get("/sessions", m.handler.listSessions)
	auth.Delete("/sessions/:id", m.handler.revokeSession)
	auth.Post("/sessions/revoke-others", m.handler.revokeOtherSessions)
	auth.Post("/logout", m.handler.logout)

	m.oauthRoutes(router)
	m.userRoutes(router)
}

// oauthRoutes registers the authorization server: the two discovery documents
// and the three endpoints a client drives itself.
//
// All five are public and mounted at the root rather than under /auth, because
// their paths are not this module's to choose. RFC 8414 and RFC 9728 derive the
// well-known paths from the issuer and the resource, and a client builds them
// from the URL it already has — so a document served anywhere else is a
// document no client will ever look for.
//
// None of them is a hole in the guard. /register creates a client that can do
// nothing until a user approves it, /authorize only parks a request and sends
// the browser to a screen that *is* session-guarded, and /token spends a code
// that only that screen can mint.
func (m *Module) oauthRoutes(router fiber.Router) {
	wellKnown := router.Group("/.well-known")
	wellKnown.Use(oauthPublicCORS)

	// Both spellings of each document. The spec derives the path from the
	// resource ("/mcp" → the /mcp suffix), and clients in the wild ask for the
	// bare path as well; serving one and not the other is a discovery failure
	// that looks exactly like an unreachable server.
	wellKnown.Get("/oauth-protected-resource", m.handler.protectedResourceMetadata)
	wellKnown.Get("/oauth-protected-resource/mcp", m.handler.protectedResourceMetadata)
	wellKnown.Get("/oauth-authorization-server", m.handler.authorizationServerMetadata)
	wellKnown.Get("/oauth-authorization-server/mcp", m.handler.authorizationServerMetadata)

	oauth := router.Group("/oauth")
	oauth.Use(oauthPublicCORS)

	// Registration writes a row for an unauthenticated caller, so it gets a
	// limit of its own on top of the global one. Twenty an hour per address is
	// far more than the once-per-install this endpoint exists for, and far less
	// than it takes to fill the table.
	oauth.Post("/register", httpx.RateLimiter(20, time.Hour, true), m.handler.registerOAuthClient)

	oauth.Get("/authorize", m.handler.authorize)
	oauth.Post("/token", m.handler.token)
}

// userRoutes registers this module's endpoints that answer under /users: the
// self-service password change and the admin invitation dashboard. Both are
// auth domain (credentials and invitations) served at the paths the user
// dashboard calls (docs/API.md §2.3 and §2.6), so they stay where they are
// while the logic lives with its data.
//
// They are terminal routes registered outside the user module's /users group,
// so they apply the guards themselves rather than inheriting that group's. The
// composition root mounts auth before user, which is what keeps the static
// paths from being captured by GET/PATCH /users/:id.
func (m *Module) userRoutes(router fiber.Router) {
	requireAuth, requireAdmin := m.RequireAuth(), httpx.RequireAdmin()

	router.Patch("/users/me/password", requireAuth, m.limiter, m.handler.changePassword)

	router.Get("/users/invitations", requireAuth, m.limiter, requireAdmin, paginate.New(), m.handler.listInvitations)
	router.Post("/users/invitations", requireAuth, m.limiter, requireAdmin, m.handler.createInvitation)
	router.Post("/users/invitations/:id/resend", requireAuth, m.limiter, requireAdmin, m.handler.resendInvitation)
	router.Delete("/users/invitations/:id", requireAuth, m.limiter, requireAdmin, m.handler.revokeInvitation)
}
