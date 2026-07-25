package auth

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/jackc/pgx/v5/pgxpool"

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
	// through it instead of querying them (docs/TECH_DEBT.md #9). The user
	// module's service is built before auth precisely so this can be an
	// ordinary constructor dependency.
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
	service *Service
	handler *handler
}

func New(deps Deps) *Module {
	pg := NewPostgresRepository(deps.DB)
	service := NewService(Stores{
		Accounts:       pg,
		Sessions:       pg,
		RefreshTokens:  pg,
		TwoFactor:      pg,
		Verifications:  pg,
		PasswordResets: pg,
		Invitations:    pg,
		Waitlist:       deps.Waitlist,
		Users:          deps.Users,
	}, deps.Cfg, deps.Storage, deps.Mail, deps.Geo, deps.Log)

	return newModule(deps, service)
}

// newModule finishes construction from an already-built service; split out so
// tests can inject fake stores through NewService.
func newModule(deps Deps, service *Service) *Module {
	return &Module{
		cfg:     deps.Cfg,
		storage: deps.Storage,
		limiter: deps.Limiter,
		service: service,
		handler: &handler{service: service, cfg: deps.Cfg},
	}
}

// Service exposes the module's use cases to the composition root and other
// modules (always consumed through interfaces declared by the consumer).
func (m *Module) Service() *Service {
	return m.service
}

// Routes registers the /auth group, replicating routes/auth.go: the public
// endpoints first (each behind the auth rate limiter), then the group-local
// RequireAuth gate for the session-bound ones.
func (m *Module) Routes(router fiber.Router) {
	auth := router.Group("/auth")

	auth.Post("/register", m.authLimiter(), m.handler.register)
	auth.Post("/login", m.authLimiter(), m.handler.login)
	auth.Post("/refresh", m.authLimiter(), m.handler.refresh)

	// Public second step of a 2FA login: exchanges the short-lived pending
	// token plus a TOTP/recovery code for a session. Rate-limited to blunt
	// code guessing on top of the per-token attempt counter.
	auth.Post("/2fa/login", m.authLimiter(), m.handler.twoFactorLogin)

	// Public invitation flow: validate a token, then accept it by setting a
	// password. Rate-limited to blunt token guessing.
	auth.Get("/invitations", m.authLimiter(), m.handler.validateInvitation)
	auth.Post("/invitations/accept", m.authLimiter(), m.handler.acceptInvitation)

	// Public password recovery flow: request a reset link, validate its
	// token, then confirm with a new password. Rate-limited to blunt both
	// mail-bombing an address and token guessing.
	auth.Post("/password-reset", m.authLimiter(), m.handler.requestPasswordReset)
	auth.Get("/password-reset", m.authLimiter(), m.handler.validatePasswordReset)
	auth.Post("/password-reset/confirm", m.authLimiter(), m.handler.confirmPasswordReset)

	// Public email verification flow: (re)send a link, validate its token,
	// then confirm to mark the email verified. Rate-limited to blunt both
	// mail-bombing an address and token guessing.
	auth.Post("/verify-email", m.authLimiter(), m.handler.requestEmailVerification)
	auth.Get("/verify-email", m.authLimiter(), m.handler.validateEmailVerification)
	auth.Post("/verify-email/confirm", m.authLimiter(), m.handler.confirmEmailVerification)

	auth.Use(m.RequireAuth())

	// Two-factor management, always behind a live session. 2FA is off by
	// default; these endpoints let the user opt in, confirm, and opt out.
	auth.Get("/2fa", m.handler.twoFactorStatus)
	auth.Post("/2fa/setup", m.handler.twoFactorSetup)
	auth.Post("/2fa/enable", m.handler.twoFactorEnable)
	auth.Post("/2fa/disable", m.handler.twoFactorDisable)
	auth.Post("/2fa/recovery-codes", m.handler.twoFactorRecoveryCodes)

	auth.Get("/session", m.handler.getSession)
	auth.Get("/sessions", m.handler.listSessions)
	auth.Delete("/sessions/:id", m.handler.revokeSession)
	auth.Post("/sessions/revoke-others", m.handler.revokeOtherSessions)
	auth.Post("/logout", m.handler.logout)

	m.userRoutes(router)
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
	limiter := m.limiter
	if limiter == nil {
		limiter = func(c fiber.Ctx) error { return c.Next() }
	}

	requireAuth, requireAdmin := m.RequireAuth(), m.RequireAdmin()

	router.Patch("/users/me/password", requireAuth, limiter, m.handler.changePassword)

	router.Get("/users/invitations", requireAuth, limiter, requireAdmin, paginate.New(), m.handler.listInvitations)
	router.Post("/users/invitations", requireAuth, limiter, requireAdmin, m.handler.createInvitation)
	router.Post("/users/invitations/:id/resend", requireAuth, limiter, requireAdmin, m.handler.resendInvitation)
	router.Delete("/users/invitations/:id", requireAuth, limiter, requireAdmin, m.handler.revokeInvitation)
}
