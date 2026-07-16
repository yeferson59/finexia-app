package routes

import (
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/auth"
	"github.com/yeferson59/finexia-app/internal/handlers"
	"github.com/yeferson59/finexia-app/internal/middlewares"
)

// Module is the surface every domain module exposes to register its routes.
// While the migration lasts, the bootstrap passes migrated modules here so
// their public routes register before the JWT middleware, exactly where
// their legacy registration used to happen.
type Module interface {
	Routes(router fiber.Router)
}

type Routes struct {
	app         *fiber.App
	router      fiber.Router
	middlewares middlewares.Middlewares
	handlers    handlers.Handlers
	auth        *auth.Module
	modules     []Module
}

func New(app *fiber.App, middlewares middlewares.Middlewares, handlers handlers.Handlers, authModule *auth.Module, modules ...Module) *Routes {
	return new(Routes{
		app:         app,
		middlewares: middlewares,
		handlers:    handlers,
		auth:        authModule,
		modules:     modules,
	})
}

func (r *Routes) Init() {
	r.app.Use(
		r.middlewares.Recovery(),
		r.middlewares.RequestID(),
		r.middlewares.ResponseTime(),
		r.middlewares.Logger(),
		r.middlewares.CORS(),
		r.middlewares.Helmet(),
		r.middlewares.Limiter(),
	)

	r.Health()
	r.auth.Routes(r.app)
	for _, m := range r.modules {
		m.Routes(r.app)
	}
	// Admin invitation/waitlist dashboard: registered in the public zone with
	// its own inline guards, before the app-wide gate (see AdminRoutes).
	r.auth.AdminRoutes(r.app, r.middlewares.UserLimiter())
	r.app.Get("/users/:id/avatar", r.handlers.GetUserAvatar)

	r.router = r.app.Use(r.auth.RequireAuth(), r.middlewares.UserLimiter())
	r.Users()
	r.Portfolios()
	r.ExchangeRates()
	r.Assets()
}
