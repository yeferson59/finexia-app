package routes

import (
	"github.com/gofiber/fiber/v3"

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
	modules     []Module
}

func New(app *fiber.App, middlewares middlewares.Middlewares, handlers handlers.Handlers, modules ...Module) *Routes {
	return new(Routes{
		app:         app,
		middlewares: middlewares,
		handlers:    handlers,
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
	r.Auth()
	for _, m := range r.modules {
		m.Routes(r.app)
	}
	r.app.Get("/users/:id/avatar", r.handlers.GetUserAvatar)

	r.router = r.app.Use(r.middlewares.JWT(), r.middlewares.UserLimiter())
	r.Users()
	r.Portfolios()
	r.ExchangeRates()
	r.Assets()
}
