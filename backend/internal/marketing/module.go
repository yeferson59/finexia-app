package marketing

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
)

// authMiddleware is the shared route-guard surface every module consumes,
// satisfied by *auth.Module.
type authMiddleware interface {
	RequireAuth() fiber.Handler
	RequireAdmin() fiber.Handler
}

// Module is the marketing domain module: construction via New, HTTP surface
// via Routes. It receives only the dependencies it uses.
type Module struct {
	service   *Service
	handler   *handler
	authMiddl authMiddleware
	limiter   fiber.Handler
}

func New(repo Repository, mail Mailer) *Module {
	service := NewService(repo, mail)
	return &Module{
		service: service,
		handler: &handler{service: service},
	}
}

// Service exposes the module's use cases to the composition root and other
// modules (always consumed through interfaces declared by the consumer).
func (m *Module) Service() *Service {
	return m.service
}

// SetAdminGuard supplies the guards and per-user rate limiter the admin
// waitlist listing needs. It arrives after construction because the auth
// module is built with marketing's service (the invitation flow advances the
// waitlist), so marketing cannot receive auth through New. Routes runs later
// still, once every module exists, so the guard is always in place by then.
func (m *Module) SetAdminGuard(guard authMiddleware, limiter fiber.Handler) {
	if guard == nil {
		panic("marketing: SetAdminGuard requires a non-nil authMiddleware")
	}

	m.authMiddl = guard
	m.limiter = limiter
}

// Routes registers the module's endpoints: the public signup and the admin
// listing of the waitlist.
//
// The listing answers at GET /users/waitlist — the path the invitation
// dashboard has always called (docs/API.md §2.6), kept unchanged so the
// contract does not move even though the data is marketing's
// (docs/TECH_DEBT.md #10). It is a terminal route registered outside the user
// module's /users group, so it carries its own guards instead of inheriting
// that group's. The composition root mounts marketing before user, which is
// what keeps the static path from being captured by GET /users/:id.
func (m *Module) Routes(router fiber.Router) {
	waitlists := router.Group("/marketing")

	waitlists.Post("/waitlists", m.handler.createWaitlist)

	if m.authMiddl == nil {
		return
	}

	limiter := m.limiter
	if limiter == nil {
		limiter = func(c fiber.Ctx) error { return c.Next() }
	}

	router.Get("/users/waitlist",
		m.authMiddl.RequireAuth(), limiter, m.authMiddl.RequireAdmin(),
		paginate.New(), m.handler.listWaitlist)
}
