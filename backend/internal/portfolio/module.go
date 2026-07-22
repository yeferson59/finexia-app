package portfolio

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type Deps struct {
	DB      *pgxpool.Pool
	Cfg     *config.Env
	Storage fiber.Storage
	Mail    Mailer
	User    UserReader
	Log     logger.Logger
	// AuthMiddl provides the route guards; the module registers in the public
	// zone and applies them itself (Fase 5 retro: port the gate's middlewares
	// explicitly when leaving the global protected zone).
	AuthMiddl authMiddleware
	// Limiter is the per-user rate limiter the legacy /portfolios routes had
	// via the app-wide gate.
	Limiter fiber.Handler
}

type authMiddleware interface {
	RequireAuth() fiber.Handler
	RequireAdmin() fiber.Handler
}

type Module struct {
	service   *Service
	handler   *handler
	authMiddl authMiddleware
	limiter   fiber.Handler
}

func New(deps Deps) *Module {
	pg := NewPostgresRepository(deps.DB)
	service := NewService(pg, deps.Cfg, deps.Storage, deps.Mail, deps.User, deps.Log)

	return newModule(deps, service)
}

func newModule(deps Deps, service *Service) *Module {
	return new(Module{
		service:   service,
		handler:   new(handler{service}),
		authMiddl: deps.AuthMiddl,
		limiter:   deps.Limiter,
	})
}

// Service exposes the module's use cases to the composition root and other
// modules (always consumed through interfaces declared by the consumer).
func (m *Module) Service() *Service {
	return m.service
}

func (m *Module) Routes(router fiber.Router) {
	portfolios := router.Group("/portfolios")

	portfolios.Use(m.authMiddl.RequireAuth(), m.limiter)

	// Static routes first — Fiber matches in registration order, so they must
	// register before the parametric "/:id" family below.
	portfolios.Get("/risks", m.handler.GetPortfoliosRisks)
	portfolios.Get("/id", m.handler.GetPortfolios)
	portfolios.Get("/summary", m.handler.GetPortfoliosSummary)
	portfolios.Get("/transactions", m.handler.GetUserTransactions)
	portfolios.Post("/transactions/import/preview", m.handler.PreviewTransactionsImport)
	portfolios.Post("/transactions/import", m.handler.ImportTransactions)
	portfolios.Get("/allocation", m.handler.GetAssetAllocation)
	portfolios.Post("", m.handler.CreatePortfolio)
	portfolios.Post("/sources", m.handler.CreatePlatform)
	portfolios.Post("/entries", m.handler.CreatePortfolioEntry)
	portfolios.Get("/entries/:entryId/transactions", m.handler.GetTransactions)
	portfolios.Post("/entries/:entryId/transactions", m.handler.CreateTransaction)
	portfolios.Put("/transactions/:txnId", m.handler.UpdateTransaction)
	portfolios.Get("/sources", m.handler.GetPlatforms)
	portfolios.Patch("/sources/:id", m.handler.UpdatePlatform)
	portfolios.Delete("/sources/:id", m.handler.DeletePlatform)
	portfolios.Get("/assets", paginate.New(), m.handler.GetAssets)
	// Admin guard inline per route (never group.Use) so unmatched paths under
	// the group fall through to a 404 instead of a 403.
	portfolios.Patch("/assets/:id/price", m.authMiddl.RequireAdmin(), m.handler.UpdateAssetPrice)
	portfolios.Get("/growth", m.handler.GetPortfolioGrowth)
	portfolios.Get("/export/summary", m.handler.ExportSummary)
	portfolios.Get("/export/transactions", m.handler.ExportTransactions)
	portfolios.Get("/export/risk", m.handler.ExportRiskMetrics)
	// Parametric routes registered last so they don't shadow the static ones above.
	portfolios.Patch("/:id", m.handler.UpdatePortfolio)
	portfolios.Get("/:id/top-transaction", m.handler.GetPortfolioTopTransaction)
	portfolios.Get("/:id/growth", m.handler.GetPortfolioGrowthByID)
	portfolios.Get("/:id/assets/:symbol/transactions", paginate.New(), m.handler.GetAssetTransactions)
	portfolios.Get("/:id", m.handler.GetPortfolio)
}
