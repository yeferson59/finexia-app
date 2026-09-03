package portfolio

import (
	"context"

	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type Deps struct {
	DB      *pgxpool.Pool
	Cfg     Config
	Storage fiber.Storage
	Mail    mailer
	User    userReader
	Log     logger.Logger
	// Assets serves the /portfolios/assets endpoints (catalog + manual price),
	// whose domain lives in the market module. Satisfied by *market.Service.
	Assets AssetReader
	// AuthMiddl provides the route guards. The module registers in the public
	// zone, so it applies them itself rather than inheriting them from a
	// global protected group.
	AuthMiddl authMiddleware
	// Limiter is the per-user rate limiter applied to the /portfolios routes,
	// injected so every module shares one budget per user.
	Limiter fiber.Handler
}

type authMiddleware interface {
	RequireAuth() fiber.Handler
}

// AssetReader is the slice of the market module the portfolio HTTP handlers
// need to serve the /portfolios/assets catalog and the manual price update.
// Satisfied by *market.Service.
// The catalog reads take a market.CatalogView: since assets can be contributed
// by users, "the catalog" is no longer one list. Each caller gets the curated
// rows plus their own, and an admin gets everything.
type AssetReader interface {
	GetAssets(ctx context.Context, view market.CatalogView, offset, limit uint) ([]market.Asset, error)
	SearchAssets(ctx context.Context, view market.CatalogView, search string, offset, limit uint) ([]market.Asset, error)
	UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (market.Asset, error)
}

type Module struct {
	service   *service
	handler   *handler
	authMiddl authMiddleware
	limiter   fiber.Handler
}

// deprecatedAlias marks a legacy path as superseded by successor, following
// RFC 8594: clients (and anything scraping access logs) can spot callers still
// on the old route before it is removed. The response is otherwise unchanged —
// the alias keeps serving the same handler.
func deprecatedAlias(successor string) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("Deprecation", "true")
		c.Set("Link", `<`+successor+`>; rel="successor-version"`)

		return c.Next()
	}
}

func New(deps Deps) *Module {
	pg := NewPostgresRepository(deps.DB)
	service := newService(pg, deps.Cfg, deps.Storage, deps.Mail, deps.User, deps.Log)

	return newModule(deps, service)
}

// newModule finishes construction from an already-built service; split out so
// tests can inject a fake repository.
//
// A missing guard panics here rather than at the first request: it is wiring,
// so the composition root is the only thing that can get it wrong, and failing
// at boot is what keeps a misconfigured build from reaching production
// quietly.
func newModule(deps Deps, service *service) *Module {
	if deps.AuthMiddl == nil {
		panic("portfolio.New: Deps.AuthMiddl is required — every /portfolios route is guarded by it")
	}

	return new(Module{
		service:   service,
		handler:   newHandler(service, deps.Assets),
		authMiddl: deps.AuthMiddl,
		limiter:   httpx.OrPassThrough(deps.Limiter),
	})
}

// Service exposes the module's use cases to the composition root and other
// modules (always consumed through interfaces declared by the consumer).
func (m *Module) Service() *service {
	return m.service
}

func (m *Module) Routes(router fiber.Router) {
	portfolios := router.Group("/portfolios")

	portfolios.Use(m.authMiddl.RequireAuth(), m.limiter)

	// Static routes first — Fiber matches in registration order, so they must
	// register before the parametric "/:id" family below.
	portfolios.Get("/risks", m.handler.GetPortfoliosRisks)
	// The user's portfolio list is canonically GET /portfolios. It historically
	// answered at the atypical "/id" path, kept as an alias that flags itself
	// deprecated so existing clients keep working. Drop the alias once the
	// Deprecation header stops showing traffic in the logs.
	portfolios.Get("", m.handler.GetPortfolios)
	portfolios.Get("/id", deprecatedAlias("/portfolios"), m.handler.GetPortfolios)
	portfolios.Get("/summary", m.handler.GetPortfoliosSummary)
	portfolios.Get("/transactions", m.handler.GetUserTransactions)
	portfolios.Post("/transactions/import/preview", m.handler.PreviewTransactionsImport)
	portfolios.Post("/transactions/import", m.handler.ImportTransactions)
	portfolios.Get("/allocation", m.handler.GetAssetAllocation)
	// Consolidated holdings: the same positions as /allocation, one row per
	// asset instead of one per category. Not "/assets" — that path is the
	// catalog, registered below.
	portfolios.Get("/holdings", m.handler.GetAssetHoldings)
	portfolios.Post("", m.handler.CreatePortfolio)
	portfolios.Post("/sources", m.handler.CreatePlatform)
	portfolios.Post("/entries", m.handler.CreatePortfolioEntry)
	portfolios.Get("/entries/:entryId/transactions", m.handler.GetTransactions)
	portfolios.Post("/entries/:entryId/transactions", m.handler.CreateTransaction)
	portfolios.Put("/transactions/:txnId", m.handler.UpdateTransaction)
	portfolios.Delete("/transactions/:txnId", m.handler.DeleteTransaction)
	portfolios.Get("/sources", m.handler.GetPlatforms)
	portfolios.Patch("/sources/:id", m.handler.UpdatePlatform)
	portfolios.Delete("/sources/:id", m.handler.DeletePlatform)
	portfolios.Get("/assets", paginate.New(), m.handler.GetAssets)
	// Admin guard inline per route (never group.Use) so unmatched paths under
	// the group fall through to a 404 instead of a 403.
	portfolios.Patch("/assets/:id/price", httpx.RequireAdmin(), m.handler.UpdateAssetPrice)
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
