package marketing

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// Like the user module, this one is built in two steps: NewService first (auth
// consumes it — the invitation flow advances the waitlist), then New with
// auth's route guards. Every dependency stays a constructor argument.

// authMiddleware is the shared route-guard surface every module consumes,
// satisfied by *auth.Module.
type authMiddleware interface {
	RequireAuth() fiber.Handler
}

// ServiceDeps is the infrastructure the use cases need; no domain module
// appears here, which is what lets the service be built before auth.
type ServiceDeps struct {
	DB   *pgxpool.Pool
	Mail Mailer
}

// Deps is the routing half: the service NewService returned, plus the guards
// the admin waitlist listing needs.
type Deps struct {
	Service   *Service
	AuthMiddl authMiddleware
	// Limiter is the per-user rate limiter shared with the other /users routes.
	Limiter fiber.Handler
}

// Module is the marketing domain module: construction via NewService + New,
// HTTP surface via Routes. It receives only the dependencies it uses.
type Module struct {
	service   *Service
	handler   *handler
	authMiddl authMiddleware
	limiter   fiber.Handler
}

// NewService builds the module's use cases, before auth.
func NewService(deps ServiceDeps) *Service {
	return newService(NewPostgresRepository(deps.DB), deps.Mail)
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
		panic("marketing.New: Deps.Service is required — pass the value NewService returned")
	}
	if deps.AuthMiddl == nil {
		panic("marketing.New: Deps.AuthMiddl is required — the waitlist listing is admin-only")
	}

	return new(Module{
		service:   deps.Service,
		handler:   newHandler(deps.Service),
		authMiddl: deps.AuthMiddl,
		limiter:   httpx.OrPassThrough(deps.Limiter),
	})
}

// Service exposes the module's use cases to the composition root and other
// modules (always consumed through interfaces declared by the consumer).
func (m *Module) Service() *Service {
	return m.service
}

// Routes registers the module's endpoints: the public signup and the admin
// listing of the waitlist.
//
// The listing answers at GET /users/waitlist — the path the invitation
// dashboard has always called (docs/API.md §2.6), kept unchanged so the
// contract does not move even though the data is marketing's. It is a terminal
// route registered outside the user module's /users group, so it carries its
// own guards instead of inheriting that group's. The composition root mounts marketing before user, which is
// what keeps the static path from being captured by GET /users/:id.
func (m *Module) Routes(router fiber.Router) {
	waitlists := router.Group("/marketing")

	// The one public, unauthenticated write in this module, and it sends an
	// email to whatever address it is given — so it carries the same tight
	// limiter as the auth endpoints rather than only the global 60/min. Without
	// it, a single client can point the confirmation mail at an arbitrary
	// inbox 60 times a minute and burn the mail provider's quota doing it.
	waitlists.Post("/waitlists", httpx.RateLimiter(5, 15*time.Minute, true), m.handler.createWaitlist)

	// The admin guard is httpx's, not the injected middleware's: the role is
	// already in the locals the JWT gate wrote, so checking it needs nothing
	// from auth and every other module reads it the same way. Only RequireAuth
	// has to be injected, because only it can validate a token.
	router.Get("/users/waitlist", m.authMiddl.RequireAuth(), m.limiter, httpx.RequireAdmin(), paginate.New(), m.handler.listWaitlist)
}
