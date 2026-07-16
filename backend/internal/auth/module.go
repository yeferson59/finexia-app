package auth

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// Deps carries everything the auth module needs from the composition root.
type Deps struct {
	// Ctx backs the token validation done inside RequireAuth, mirroring the
	// context the legacy middlewares.New received (TECH_DEBT: use the
	// request's own context instead).
	Ctx     context.Context
	DB      *pgxpool.Pool
	Cfg     *config.Env
	Storage fiber.Storage
	Mail    Mailer
	Geo     GeoLocator
	Log     logger.Logger
}

// Module is the auth domain module: construction via New, HTTP surface via
// Routes, and route guards (RequireAuth/RequireRole) for the rest of the app.
type Module struct {
	ctx     context.Context
	cfg     *config.Env
	storage fiber.Storage
	service *Service
	handler *handler
}

func New(deps Deps) *Module {
	pg := NewPostgresRepository(deps.DB)
	service := NewService(Stores{
		Accounts:      pg,
		Sessions:      pg,
		RefreshTokens: pg,
		TwoFactor:     pg,
		Verifications: pg,
	}, deps.Cfg, deps.Storage, deps.Mail, deps.Geo, deps.Log)

	return newModule(deps, service)
}

// newModule finishes construction from an already-built service; split out so
// tests can inject fake stores through NewService.
func newModule(deps Deps, service *Service) *Module {
	return &Module{
		ctx:     deps.Ctx,
		cfg:     deps.Cfg,
		storage: deps.Storage,
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
}
