package user

import (
	"context"

	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/objectstore"
)

// This module is built in two steps, and the order matters at the composition
// root. NewService needs nothing from another module, so it runs first and its
// result is handed to auth, which reads users/roles through it. New then
// completes the module with auth's route guards. Splitting the two is what
// keeps every dependency a constructor argument: no setter to forget, no slot
// that is briefly nil.

// ServiceDeps is the infrastructure the use cases need. No domain module
// appears here — that is the property that lets the service be built first.
type ServiceDeps struct {
	DB    *pgxpool.Pool
	Store objectstore.Store
	Log   logger.Logger
	Cfg   Config
}

// Deps is the routing half: the service NewService returned, plus the guards,
// which only exist once the auth module is built.
type Deps struct {
	Service   *Service
	AuthMiddl authMiddleware
	// Sessions closes out the live sessions of a user this module bans or
	// deletes. The session and refresh-token tables belong to auth, so the
	// capability is injected the same way the route guards are; satisfied by
	// *auth.Service, which exists by the time this half of the module is built.
	Sessions sessionRevoker
	// Limiter is the per-user rate limiter applied to the /users routes. The
	// module registers in the public zone, so it applies this and its own
	// RequireAuth itself.
	Limiter fiber.Handler
}

type authMiddleware interface {
	RequireAuth() fiber.Handler
}

// sessionRevoker is the slice of the auth module this one needs to make a ban
// or a deletion effective immediately. Without it both operations would only
// write a column: every access token already issued keeps validating, and the
// refresh cookie keeps rotating, for as long as the token lifetimes allow.
type sessionRevoker interface {
	RevokeAllSessions(ctx context.Context, userID uuid.UUID) (int64, error)
}

type Module struct {
	service   *Service
	handler   *handler
	authMiddl authMiddleware
	limiter   fiber.Handler
}

// NewService builds the module's use cases. It is the first domain thing the
// composition root constructs, before auth.
func NewService(deps ServiceDeps) *Service {
	return newService(NewPostgresRepository(deps.DB), deps.Store, deps.Log, deps.Cfg)
}

// New completes the module with its HTTP surface. deps.Service must be the
// value NewService returned, so auth and these routes share one service.
//
// A missing service or guard panics here rather than at the first request:
// both are wiring, so the composition root is the only thing that can get them
// wrong, and failing at boot is what keeps a misconfigured build from reaching
// production quietly.
func New(deps Deps) *Module {
	if deps.Service == nil {
		panic("user.New: Deps.Service is required — pass the value NewService returned")
	}
	if deps.AuthMiddl == nil {
		panic("user.New: Deps.AuthMiddl is required — every /users route is guarded by it")
	}
	if deps.Sessions == nil {
		panic("user.New: Deps.Sessions is required — a ban that leaves live sessions running is not a ban")
	}

	return new(Module{
		service:   deps.Service,
		handler:   newHandler(deps.Service, deps.Sessions),
		authMiddl: deps.AuthMiddl,
		limiter:   httpx.OrPassThrough(deps.Limiter),
	})
}

// Service exposes the module's use cases to the composition root and other
// modules (always consumed through interfaces declared by the consumer).
func (m *Module) Service() *Service {
	return m.service
}

func (m *Module) Routes(router fiber.Router) {
	// Public avatar (docs/API.md §2.3): registered as a terminal handler
	// before the group's RequireAuth so it stays outside the auth gate.
	router.Get("/users/:id/avatar", m.handler.GetUserAvatar)

	users := router.Group("/users")

	users.Use(m.authMiddl.RequireAuth(), m.limiter)

	// Self-service routes — must be registered before /:id to avoid shadowing.
	// The sibling PATCH /users/me/password is served by auth, which owns the
	// credentials and mounts earlier.
	users.Get("/me", m.handler.GetMe)
	users.Patch("/me", m.handler.UpdateMe)
	users.Post("/me/avatar", m.handler.UploadAvatar)
	users.Get("/me/preferences", m.handler.GetMyPreferences)
	users.Patch("/me/preferences", m.handler.UpdateMyPreferences)

	// Admin guards go inline per route (never group.Use) so unmatched
	// /users/* requests fall through to a 404 instead of a 403.
	admin := httpx.RequireAdmin()
	users.Get("", admin, paginate.New(), m.handler.GetListUsers)
	users.Post("", admin, m.handler.CreateUser)

	// The static /users/invitations and /users/waitlist paths belong to auth
	// and marketing respectively; both mount before this module, which is what
	// keeps them from being captured by the "/:id" routes below.
	users.Get("/:id", admin, m.handler.GetUserByID)
	users.Patch("/:id", admin, m.handler.UpdateUser)
	users.Patch("/:id/ban", admin, m.handler.BanUser)
	users.Delete("/:id", admin, m.handler.DeleteUser)
}
