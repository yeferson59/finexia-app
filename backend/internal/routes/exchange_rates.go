package routes

import "github.com/gofiber/fiber/v3/middleware/paginate"

func (r *Routes) ExchangeRates() {
	er := r.router.Group("/exchange-rates")
	er.Get("", paginate.New(), r.handlers.GetExchangeRates)
	er.Post("", r.auth.RequireAdmin(), r.handlers.CreateExchangeRate)
	er.Post("/import", r.auth.RequireAdmin(), r.handlers.ImportExchangeRates)
	er.Post("/sync", r.auth.RequireAdmin(), r.handlers.SyncExchangeRates)
	er.Post("/:id/sync", r.auth.RequireAdmin(), r.handlers.SyncSingleExchangeRate)
	er.Patch("/:id", r.auth.RequireAdmin(), r.handlers.UpdateExchangeRate)
}
