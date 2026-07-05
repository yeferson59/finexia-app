package routes

import "github.com/gofiber/fiber/v3/middleware/paginate"

func (r *Routes) ExchangeRates() {
	er := r.router.Group("/exchange-rates")
	er.Get("", paginate.New(), r.handlers.GetExchangeRates)
	er.Post("", r.middlewares.RequireAdmin(), r.handlers.CreateExchangeRate)
	er.Post("/sync", r.middlewares.RequireAdmin(), r.handlers.SyncExchangeRates)
	er.Post("/:id/sync", r.middlewares.RequireAdmin(), r.handlers.SyncSingleExchangeRate)
	er.Patch("/:id", r.middlewares.RequireAdmin(), r.handlers.UpdateExchangeRate)
}
