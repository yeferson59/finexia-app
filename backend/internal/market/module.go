package market

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

type Module struct {
	service   *Service
	authMiddl authMiddleware
	cfg       *config.Env
	storage   fiber.Storage
	handler   *handler
	limiter   fiber.Handler
}

type Deps struct {
	DB               *pgxpool.Pool
	Cfg              *config.Env
	Storage          fiber.Storage
	Log              logger.Logger
	PortfolioService portfolioService
	Provider         marketdata.Provider
	AuthMiddleware   authMiddleware
	Limiter          fiber.Handler
}

type authMiddleware interface {
	RequireAuth() fiber.Handler
	RequireAdmin() fiber.Handler
}

func New(deps Deps) *Module {
	pg := NewPostgresRepository(deps.DB)
	service := NewService(pg, deps.PortfolioService, deps.Storage, deps.Provider, deps.Log)

	return newModule(deps, service)
}

func newModule(deps Deps, service *Service) *Module {
	return new(Module{
		cfg:       deps.Cfg,
		storage:   deps.Storage,
		service:   service,
		handler:   new(handler{service, deps.PortfolioService}),
		authMiddl: deps.AuthMiddleware,
		limiter:   deps.Limiter,
	})
}

func (m *Module) Service() *Service {
	return m.service
}

func (m *Module) Routes(router fiber.Router) {
	assests := router.Group("/assets")

	assests.Use(m.authMiddl.RequireAuth(), m.limiter)

	admin := m.authMiddl.RequireAdmin()

	assests.Post("", admin, m.handler.CreateAsset)
	assests.Post("/import", admin, m.handler.ImportAssets)
	assests.Post("/sync", admin, m.handler.SyncAssetPrices)
	assests.Post("/:id/sync", admin, m.handler.SyncSingleAsset)

	exchangeRates := router.Group("/exchange-rates")
	exchangeRates.Use(m.authMiddl.RequireAuth(), m.limiter)

	exchangeRates.Get("", admin, paginate.New(), m.handler.GetExchangeRates)
	exchangeRates.Post("", admin, m.handler.CreateExchangeRate)
	exchangeRates.Post("/import", admin, m.handler.ImportExchangeRates)
	exchangeRates.Post("/sync", admin, m.handler.SyncExchangeRates)
	exchangeRates.Post("/:id/sync", admin, m.handler.SyncSingleExchangeRate)
	exchangeRates.Patch("/:id", admin, m.handler.UpdateExchangeRate)
}
