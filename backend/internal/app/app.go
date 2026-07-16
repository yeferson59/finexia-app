// Package app is the composition root: the only place that wires
// infrastructure into modules, the legacy services/handlers/routes, and the
// schedulers. Adding a domain module means registering it here and nowhere
// else.
package app

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bytedance/sonic"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/auth"
	"github.com/yeferson59/finexia-app/internal/handlers"
	"github.com/yeferson59/finexia-app/internal/marketing"
	"github.com/yeferson59/finexia-app/internal/middlewares"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/geoip"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata/alphavantage"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata/finnhub"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata/yahoo"
	"github.com/yeferson59/finexia-app/internal/repositories"
	"github.com/yeferson59/finexia-app/internal/routes"
	"github.com/yeferson59/finexia-app/internal/scheduler"
	"github.com/yeferson59/finexia-app/internal/services"
)

// Deps carries the already-connected infrastructure the App composes. main
// owns creating (and closing) these; App only wires them.
type Deps struct {
	Envs    *config.Env
	DB      *pgxpool.Pool
	Storage fiber.Storage
	S3      *s3.Client
	Mail    *mail.Service
	Log     logger.Logger
}

type App struct {
	fiber *fiber.App
	deps  Deps
}

type structValidator struct {
	validate *validator.Validate
}

func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

// New builds the Fiber application with the HTTP-level configuration that
// used to live in cmd/api/main.go.
func New(deps Deps) *App {
	fiberApp := fiber.New(fiber.Config{
		JSONEncoder:        sonic.ConfigFastest.Marshal,
		JSONDecoder:        sonic.ConfigFastest.Unmarshal,
		StructValidator:    new(structValidator{validate: validator.New()}),
		ProxyHeader:        fiber.HeaderXForwardedFor,
		TrustProxy:         deps.Envs.TrustProxy,
		EnableIPValidation: true,
		BodyLimit:          10 * 1024 * 1024,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Loopback:  true,
			LinkLocal: true,
			Private:   true,
			Proxies:   deps.Envs.TrustedProxies,
		},
	})

	return new(App{fiber: fiberApp, deps: deps})
}

// Run wires modules, legacy layers and schedulers, then serves HTTP until
// the listener stops.
func (a *App) Run(ctx context.Context) error {
	a.wire(ctx)

	if err := a.fiber.Listen(":" + a.deps.Envs.Port); err != nil {
		return errors.New("failed to listen: " + err.Error())
	}

	return nil
}

// wire composes every layer of the application; separated from Run so tests
// can exercise the composed router without opening a listener.
func (a *App) wire(ctx context.Context) {
	d := a.deps

	repos := repositories.New(d.DB)
	priceProvider := marketdata.NewFallback(
		alphavantage.New(d.Envs.AlphaVantageAPIKey),
		finnhub.New(d.Envs.FinnhubAPIKey),
		yahoo.New(),
	)

	// Migrated domain modules.
	marketingModule := marketing.New(marketing.NewPostgresRepository(d.DB), d.Mail)
	authModule := auth.New(auth.Deps{
		Ctx:     ctx,
		DB:      d.DB,
		Cfg:     d.Envs,
		Storage: d.Storage,
		Mail:    d.Mail,
		Geo:     geoip.New(),
		Log:     d.Log,
	})

	// Legacy wiring: shrinks phase by phase until Fase 8 deletes it.
	svc := services.New(&repos, d.Envs, d.S3, d.Storage, d.Mail, geoip.New(), d.Log, priceProvider, authModule.Service())
	handl, middl := handlers.New(svc, d.Envs), middlewares.New(d.Envs, d.Storage)

	routes.New(a.fiber, middl, handl, authModule, marketingModule).Init()

	a.startSchedulers(ctx, svc, authModule)
}

// startSchedulers is the single place background jobs come to life; Fase 7
// replaces these ad-hoc schedulers with a generic Job runner.
func (a *App) startSchedulers(ctx context.Context, svc services.Services, authModule *auth.Module) {
	go scheduler.NewExchangeRateScheduler(svc, 6, a.deps.Log).Start(ctx)
	go scheduler.NewAssetPriceScheduler(svc, 14, 90*time.Second, a.deps.Log).Start(ctx)
	go scheduler.NewPortfolioSnapshotScheduler(svc, 15, 120*time.Second, a.deps.Log).Start(ctx)
	go scheduler.NewWeeklySummaryScheduler(svc, 9, a.deps.Log).Start(ctx)
	go auth.NewCleanupJob(authModule.Service(), 3, a.deps.Log).Start(ctx)
}
