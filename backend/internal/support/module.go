package support

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/platform/bold"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// authMiddleware is the shared route-guard surface every module consumes,
// satisfied by *auth.Module. Only the admin listing needs it.
type authMiddleware interface {
	RequireAuth() fiber.Handler
}

// Deps is everything the module uses. It needs no other domain module.
type Deps struct {
	DB  *pgxpool.Pool
	Cfg Config
	Log logger.Logger
	// AuthMiddl guards the admin listing; the payer's routes are public.
	AuthMiddl authMiddleware
	// Limiter is the per-user rate limiter shared with the other admin routes.
	Limiter fiber.Handler
}

// Module is the support domain module.
type Module struct {
	service   *service
	handler   *handler
	authMiddl authMiddleware
	limiter   fiber.Handler
}

// New builds the module. The Bold client is built here from the identity key
// rather than injected, because the key is configuration and the client has
// nothing else to it.
func New(deps Deps) *Module {
	if deps.DB == nil {
		panic("support.New: Deps.DB is required")
	}
	if deps.AuthMiddl == nil {
		panic("support.New: Deps.AuthMiddl is required — the contributions listing is admin-only")
	}

	svc := newService(NewPostgresRepository(deps.DB), bold.New(nil, deps.Cfg.APIKey), deps.Cfg, deps.Log)

	return new(Module{
		service:   svc,
		handler:   newHandler(svc),
		authMiddl: deps.AuthMiddl,
		limiter:   httpx.OrPassThrough(deps.Limiter),
	})
}

// Service exposes the use cases to the composition root, for the reconcile job.
func (m *Module) Service() *service {
	return m.service
}

// Routes registers the module's endpoints (docs/API.md §2.13).
//
// The payer's four are public: the app's server calls the first three on the
// payer's behalf, and Bold calls the webhook through the app's public origin.
// The two under /support/contributions are the admin's.
func (m *Module) Routes(router fiber.Router) {
	group := router.Group("/support")

	group.Get("/", m.handler.config)

	// Each call writes a row, so it gets a limit tighter than the global
	// 60/min: enough for someone changing their mind about the amount a few
	// times, not for filling the table.
	group.Post("/checkouts", httpx.RateLimiter(20, 15*time.Minute, true), m.handler.createCheckout)
	group.Get("/checkouts/:orderId", m.handler.getCheckout)

	group.Post("/webhooks/bold", m.handler.boldWebhook)

	admin := group.Group("/contributions", m.authMiddl.RequireAuth(), m.limiter, httpx.RequireAdmin())
	admin.Get("/", paginate.New(), m.handler.listContributions)
	admin.Get("/summary", m.handler.summary)
}
