package market

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/paginate"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/platform/secretbox"
)

// This module is built in two steps, and the order matters at the composition
// root — the same shape user and marketing already use.
//
// NewService depends on no other domain module, so it runs first and its result
// is handed to portfolio, which reads the asset catalog through it. New then
// completes the module with the route guards and with portfolio's holdings,
// which only exist once portfolio is built. Splitting the two is what lets the
// BYO-key sync ask "which assets does this user hold" without market importing
// portfolio and closing the cycle.

// ServiceDeps is the infrastructure the use cases need. No domain module
// appears here — that is the property that lets the service be built first.
type ServiceDeps struct {
	DB      *pgxpool.Pool
	Storage fiber.Storage
	Log     logger.Logger
	// Providers builds a provider chain from a user's own keys. There is no
	// process-wide Provider any more: under BYO-key the application holds no
	// provider credentials of its own.
	Providers marketdata.Factory
	// Keyring seals and opens those keys.
	Keyring *secretbox.Keyring
}

// Deps is the routing half: the service NewService returned, plus the guards
// and the holdings reader, which only exist once the other modules are built.
type Deps struct {
	Service        *Service
	AuthMiddleware authMiddleware
	Limiter        fiber.Handler
	// Holdings answers "which assets does this user own", satisfied by the
	// portfolio module.
	Holdings Holdings
	// CredentialLimiter gates the credential routes more tightly than Limiter:
	// each save and each verification spends the user's own provider quota.
	CredentialLimiter fiber.Handler
}

type Module struct {
	service     *Service
	authMiddl   authMiddleware
	handler     *handler
	limiter     fiber.Handler
	credLimiter fiber.Handler
}

type authMiddleware interface {
	RequireAuth() fiber.Handler
	RequireAdmin() fiber.Handler
}

// NewService builds the module's use cases. It is constructed before portfolio,
// which consumes it.
func NewService(deps ServiceDeps) *Service {
	return newService(NewPostgresRepository(deps.DB), deps.Storage, deps.Providers, deps.Keyring, deps.Log)
}

// New completes the module with its HTTP surface. deps.Service must be the
// value NewService returned, so portfolio and these routes share one service.
func New(deps Deps) *Module {
	return new(Module{
		service:     deps.Service,
		handler:     new(handler{deps.Service, deps.Holdings}),
		authMiddl:   deps.AuthMiddleware,
		limiter:     deps.Limiter,
		credLimiter: deps.CredentialLimiter,
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

	exchangeRates := router.Group("/exchange-rates")
	exchangeRates.Use(m.authMiddl.RequireAuth(), m.limiter)

	exchangeRates.Get("", admin, paginate.New(), m.handler.GetExchangeRates)
	exchangeRates.Post("", admin, m.handler.CreateExchangeRate)
	exchangeRates.Post("/import", admin, m.handler.ImportExchangeRates)
	exchangeRates.Patch("/:id", admin, m.handler.UpdateExchangeRate)

	// BYO-key. Every route here acts on the caller's own keys and holdings —
	// the user id comes from the auth locals, never from the path — so there is
	// no admin variant and no way to name somebody else's credential.
	//
	// The credential routes carry a much tighter limiter than the group's:
	// saving and verifying both spend the user's own provider quota.
	marketData := router.Group("/market")
	marketData.Use(m.authMiddl.RequireAuth(), m.limiter)

	marketData.Get("/credentials", m.handler.ListCredentials)
	marketData.Put("/credentials/:provider", m.credLimiter, m.handler.SaveCredential)
	marketData.Post("/credentials/:provider/verify", m.credLimiter, m.handler.VerifyCredential)
	marketData.Delete("/credentials/:provider", m.credLimiter, m.handler.DeleteCredential)

	marketData.Post("/sync", m.credLimiter, m.handler.SyncMarketData)
}
