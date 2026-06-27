package routes

import (
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/handlers"
	"github.com/yeferson59/finexia-app/internal/middlewares"
)

type Routes struct {
	app         *fiber.App
	router      fiber.Router
	middlewares middlewares.Middlewares
	handlers    handlers.Handlers
}

func New(app *fiber.App, middlewares middlewares.Middlewares, handlers handlers.Handlers) *Routes {
	return new(Routes{
		app:         app,
		middlewares: middlewares,
		handlers:    handlers,
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
	r.Marketing()
	r.app.Get("/users/:id/avatar", r.handlers.GetUserAvatar)

	r.router = r.app.Use(r.middlewares.Session(), r.middlewares.JWT(), r.middlewares.UserLimiter())
	r.Users()
	r.Portfolios()
	r.ExchangeRates()
	r.Assets()
}
