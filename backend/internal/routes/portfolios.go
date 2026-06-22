package routes

import "github.com/gofiber/fiber/v3/middleware/paginate"

func (r *Routes) Portfolios() {
	portfolios := r.router.Group("/portfolios")

	portfolios.Get("/risks", r.handlers.GetPortfoliosRisks)
	portfolios.Get("/id", r.handlers.GetPortfolios)
	portfolios.Post("", r.handlers.CreatePortfolio)
	portfolios.Post("/sources", r.handlers.CreatePlatform)
	portfolios.Post("/entries", r.handlers.CreatePortfolioEntry)
	portfolios.Get("/sources", r.handlers.GetPlatforms)
	portfolios.Get("/assets", paginate.New(), r.handlers.GetAssets)
	portfolios.Patch("/assets/:id/price", r.handlers.UpdateAssetPrice)
	// Parametric route registered last so it doesn't shadow the static ones above.
	portfolios.Get("/:id", r.handlers.GetPortfolio)
}
