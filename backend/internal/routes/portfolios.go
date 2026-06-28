package routes

import "github.com/gofiber/fiber/v3/middleware/paginate"

func (r *Routes) Portfolios() {
	portfolios := r.router.Group("/portfolios")

	portfolios.Get("/risks", r.handlers.GetPortfoliosRisks)
	portfolios.Get("/id", r.handlers.GetPortfolios)
	portfolios.Get("/summary", r.handlers.GetPortfoliosSummary)
	portfolios.Get("/transactions", r.handlers.GetUserTransactions)
	portfolios.Get("/allocation", r.handlers.GetAssetAllocation)
	portfolios.Post("", r.handlers.CreatePortfolio)
	portfolios.Post("/sources", r.handlers.CreatePlatform)
	portfolios.Post("/entries", r.handlers.CreatePortfolioEntry)
	portfolios.Get("/entries/:entryId/transactions", r.handlers.GetTransactions)
	portfolios.Post("/entries/:entryId/transactions", r.handlers.CreateTransaction)
	portfolios.Put("/transactions/:txnId", r.handlers.UpdateTransaction)
	portfolios.Get("/sources", r.handlers.GetPlatforms)
	portfolios.Patch("/sources/:id", r.handlers.UpdatePlatform)
	portfolios.Delete("/sources/:id", r.handlers.DeletePlatform)
	portfolios.Get("/assets", paginate.New(), r.handlers.GetAssets)
	portfolios.Patch("/assets/:id/price", r.middlewares.RequireAdmin(), r.handlers.UpdateAssetPrice)
	portfolios.Get("/growth", r.handlers.GetPortfolioGrowth)
	portfolios.Get("/export/summary", r.handlers.ExportSummary)
	portfolios.Get("/export/transactions", r.handlers.ExportTransactions)
	portfolios.Get("/export/risk", r.handlers.ExportRiskMetrics)
	// Parametric routes registered last so they don't shadow the static ones above.
	portfolios.Patch("/:id", r.handlers.UpdatePortfolio)
	portfolios.Get("/:id/top-transaction", r.handlers.GetPortfolioTopTransaction)
	portfolios.Get("/:id/assets/:symbol/transactions", paginate.New(), r.handlers.GetAssetTransactions)
	portfolios.Get("/:id", r.handlers.GetPortfolio)
}
