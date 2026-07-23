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
	"github.com/yeferson59/finexia-app/internal/health"
	"github.com/yeferson59/finexia-app/internal/market"
	"github.com/yeferson59/finexia-app/internal/marketing"
	"github.com/yeferson59/finexia-app/internal/notification"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/geoip"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata/alphavantage"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata/finnhub"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata/yahoo"
	"github.com/yeferson59/finexia-app/internal/platform/objectstore"
	"github.com/yeferson59/finexia-app/internal/portfolio"
	"github.com/yeferson59/finexia-app/internal/scheduler"
	"github.com/yeferson59/finexia-app/internal/user"
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
	fiber    *fiber.App
	deps     Deps
	schedule *scheduler.Scheduler
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
		JSONEncoder:     sonic.ConfigFastest.Marshal,
		JSONDecoder:     sonic.ConfigFastest.Unmarshal,
		StructValidator: new(structValidator{validate: validator.New()}),
		ProxyHeader:     fiber.HeaderXForwardedFor,
		TrustProxy:      deps.Envs.TrustProxy,
		BodyLimit:       10 * 1024 * 1024,
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
// the listener stops or ctx is cancelled (e.g. on SIGINT/SIGTERM), in which
// case it shuts down the HTTP server and the schedulers cleanly.
func (a *App) Run(ctx context.Context) error {
	a.wire(ctx)

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := a.Shutdown(shutdownCtx); err != nil {
			a.deps.Log.Error(shutdownCtx, "error during shutdown: "+err.Error())
		}
	}()

	if err := a.fiber.Listen(":" + a.deps.Envs.Port); err != nil {
		return errors.New("failed to listen: " + err.Error())
	}

	return nil
}

// Shutdown stops accepting new HTTP requests and drains in-flight ones,
// then stops the schedulers: it cancels their loops so no new job fires,
// without aborting a job already in progress inside Runner.Execute.
func (a *App) Shutdown(ctx context.Context) error {
	if err := a.fiber.ShutdownWithContext(ctx); err != nil {
		return err
	}

	if a.schedule != nil {
		a.schedule.Stop()
	}

	return nil
}

const localUserID = "auth_user_id"

// wire composes every layer of the application; separated from Run so tests
// can exercise the composed router without opening a listener.
func (a *App) wire(ctx context.Context) {
	priceProvider := marketdata.NewFallback(
		alphavantage.New(a.deps.Envs.AlphaVantageAPIKey),
		finnhub.New(a.deps.Envs.FinnhubAPIKey),
		yahoo.New(),
	)

	geo := geoip.New()
	userLimiter := httpx.KeyedRateLimiter(200, 1*time.Minute, func(c fiber.Ctx) string {
		userID := c.Locals(localUserID).(string)

		return "user_limit:" + userID
	})

	// Migrated domain modules.
	healthModule := health.New()
	marketingModule := marketing.New(marketing.NewPostgresRepository(a.deps.DB), a.deps.Mail)
	authModule := auth.New(auth.Deps{
		Ctx:      ctx,
		DB:       a.deps.DB,
		Cfg:      a.deps.Envs,
		Storage:  a.deps.Storage,
		Mail:     a.deps.Mail,
		Geo:      geo,
		Log:      a.deps.Log,
		Waitlist: marketingModule.Service(),
	})
	userModule := user.New(user.Deps{
		DB:        a.deps.DB,
		Cfg:       a.deps.Envs,
		Store:     objectstore.NewS3Store(a.deps.S3, a.deps.Envs.AWSS3BucketName),
		Mail:      a.deps.Mail,
		Geo:       geo,
		Log:       a.deps.Log,
		Auth:      authModule.Service(),
		Marketing: marketingModule.Service(),
		AuthMiddl: authModule,
		Limiter:   userLimiter,
	})
	// market owns the asset catalog; portfolio consumes it (portfolio → market),
	// so market is built first and injected as portfolio's AssetReader.
	marketModule := market.New(market.Deps{
		DB:             a.deps.DB,
		Cfg:            a.deps.Envs,
		Storage:        a.deps.Storage,
		Log:            a.deps.Log,
		Provider:       priceProvider,
		AuthMiddleware: authModule,
		Limiter:        userLimiter,
	})
	portfolioModule := portfolio.New(portfolio.Deps{
		DB:        a.deps.DB,
		Cfg:       a.deps.Envs,
		Storage:   a.deps.Storage,
		Mail:      a.deps.Mail,
		User:      userModule.Service(),
		Assets:    marketModule.Service(),
		Log:       a.deps.Log,
		AuthMiddl: authModule,
		Limiter:   userLimiter,
	})
	notificationService := notification.NewService(userModule.Service(), portfolioModule.Service(), a.deps.Mail, a.deps.Envs)

	a.fiber.Use(httpx.Recovery(), httpx.CORS(a.deps.Envs.CORSOrigin, true), httpx.Helmet(), httpx.RequestID(), httpx.ResponseTime(), httpx.Logger(), httpx.RateLimiter(60, 1*time.Minute, false))

	healthModule.Routes(a.fiber)
	marketModule.Routes(a.fiber)
	authModule.Routes(a.fiber)
	marketingModule.Routes(a.fiber)
	userModule.Routes(a.fiber)
	portfolioModule.Routes(a.fiber)

	runner := scheduler.NewRunner(scheduler.RunnerOptions{
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		BackoffBase: 500 * time.Millisecond,
		BackoffMax:  10 * time.Second,
		OnError: func(name string, err error) {
			a.deps.Log.Error(ctx, "ALERTA: job "+name+" falló definitivamente "+err.Error())
		},
		Log: a.deps.Log,
	})

	schedule := scheduler.NewScheduler(runner)
	a.schedule = schedule

	a.registerJobs(schedule, authModule, portfolioModule, marketModule, notificationService)
}

func (a *App) registerJobs(sched *scheduler.Scheduler, authModule *auth.Module, portfolioModule *portfolio.Module, marketModule *market.Module, notificationService *notification.Service) {
	sched.Register(market.NewExchangeRateScheduler(marketModule.Service(), a.deps.Log), scheduler.DailyAt{Hour: 9, Minute: 30})
	sched.Register(market.NewAssetPriceScheduler(marketModule.Service(), a.deps.Log), scheduler.Delayed{Schedule: scheduler.DailyAt{Hour: 9, Minute: 30}, Delay: 90 * time.Second})
	sched.Register(portfolio.NewSnapshotJob(portfolioModule.Service(), a.deps.Log), scheduler.Delayed{Schedule: scheduler.DailyAt{Hour: 9, Minute: 30}, Delay: 120 * time.Second})
	sched.Register(notification.NewWeeklySummaryScheduler(notificationService, a.deps.Log), scheduler.WeeklyAt{Day: time.Monday, Hour: 8, Minute: 30})
	sched.Register(auth.NewCleanupJob(authModule.Service(), a.deps.Log), scheduler.Every{Interval: 5 * time.Hour})

	sched.Start()
}
