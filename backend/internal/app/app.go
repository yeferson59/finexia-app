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
	"github.com/yeferson59/finexia-app/internal/scheduler/fiberstore"
	"github.com/yeferson59/finexia-app/internal/user"
)

const (
	// shutdownTimeout bounds the graceful shutdown once the parent context
	// is cancelled: HTTP draining plus stopping the schedulers.
	shutdownTimeout = 30 * time.Second

	// bodyLimit caps request bodies at 10 MiB.
	bodyLimit = 10 * 1024 * 1024
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

// validate reports the first required dependency that is missing, so New
// fails fast at the composition root with a clear message instead of a nil
// dereference deep inside wire(). S3 is intentionally optional: the app
// wires without it (e.g. in tests) and only object-store routes need it.
func (d Deps) validate() error {
	switch {
	case d.Envs == nil:
		return errors.New("app: Deps.Envs is required")
	case d.DB == nil:
		return errors.New("app: Deps.DB is required")
	case d.Storage == nil:
		return errors.New("app: Deps.Storage is required")
	case d.Mail == nil:
		return errors.New("app: Deps.Mail is required")
	case d.Log == nil:
		return errors.New("app: Deps.Log is required")
	default:
		return nil
	}
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

// New validates the dependencies and builds the Fiber application with the
// HTTP-level configuration that used to live in cmd/api/main.go. It returns
// an error (rather than panicking) when a required dependency is missing.
func New(deps Deps) (*App, error) {
	if err := deps.validate(); err != nil {
		return nil, err
	}

	fiberApp := fiber.New(fiber.Config{
		JSONEncoder:     sonic.ConfigFastest.Marshal,
		JSONDecoder:     sonic.ConfigFastest.Unmarshal,
		StructValidator: new(structValidator{validate: validator.New()}),
		ProxyHeader:     fiber.HeaderXForwardedFor,
		TrustProxy:      deps.Envs.TrustProxy,
		BodyLimit:       bodyLimit,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Loopback:  true,
			LinkLocal: true,
			Private:   true,
			Proxies:   deps.Envs.TrustedProxies,
		},
	})

	return new(App{fiber: fiberApp, deps: deps}), nil
}

// Run wires modules, legacy layers and schedulers, then serves HTTP until
// the listener stops or ctx is cancelled (e.g. on SIGINT/SIGTERM), in which
// case it shuts down the HTTP server and the schedulers cleanly.
func (a *App) Run(ctx context.Context) error {
	a.wire(ctx)

	// stopped signals that Listen has returned (e.g. a bind error), so the
	// shutdown watcher exits instead of leaking while blocked on ctx.Done().
	stopped := make(chan struct{})
	defer close(stopped)

	go func() {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancel()

			if err := a.Shutdown(shutdownCtx); err != nil {
				a.deps.Log.Error(shutdownCtx, "error during shutdown", logger.Err(err))
			}
		case <-stopped:
			// Listen already returned; nothing to gracefully shut down.
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

// modules holds every composed domain module so the wiring steps
// (route mounting, scheduler registration) can pass them around as one value
// instead of a long parameter list.
type modules struct {
	health       *health.Module
	marketing    *marketing.Module
	auth         *auth.Module
	user         *user.Module
	market       *market.Module
	portfolio    *portfolio.Module
	notification *notification.Service
}

// wire composes every layer of the application; separated from Run so tests
// can exercise the composed router without opening a listener. It runs the
// three ordered steps of the composition root: build the modules, mount
// their routes, then start the schedulers.
func (a *App) wire(ctx context.Context) {
	mods := a.buildModules(ctx)
	a.mountRoutes(mods)
	a.startScheduler(ctx, mods)
}

// buildModules constructs the shared infrastructure (price provider, geoip,
// per-user rate limiter) and every domain module, respecting their
// dependency order.
func (a *App) buildModules(ctx context.Context) *modules {
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

	marketingModule := marketing.New(marketing.NewPostgresRepository(a.deps.DB), a.deps.Mail)
	authModule := auth.New(auth.Deps{
		Ctx:      ctx,
		DB:       a.deps.DB,
		Cfg:      authConfig(a.deps.Envs),
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

	return &modules{
		health:       health.New(),
		marketing:    marketingModule,
		auth:         authModule,
		user:         userModule,
		market:       marketModule,
		portfolio:    portfolioModule,
		notification: notification.NewService(userModule.Service(), portfolioModule.Service(), a.deps.Mail, a.deps.Envs),
	}
}

// mountRoutes installs the global middleware chain and each module's routes.
func (a *App) mountRoutes(mods *modules) {
	a.fiber.Use(httpx.Recovery(), httpx.CORS(a.deps.Envs.CORSOrigin, true), httpx.Helmet(), httpx.RequestID(), httpx.ResponseTime(), httpx.Logger(), httpx.RateLimiter(60, 1*time.Minute, false))

	mods.health.Routes(a.fiber)
	mods.market.Routes(a.fiber)
	mods.auth.Routes(a.fiber)
	mods.marketing.Routes(a.fiber)
	mods.user.Routes(a.fiber)
	mods.portfolio.Routes(a.fiber)
}

// startScheduler builds the job runner and scheduler, registers every job and
// starts the loops. The Scheduler defaults jobs to in-memory state; the jobs
// that must survive restarts/deploys are registered with a Redis-backed store
// so they resume from their persisted next-run time (catching up on runs
// missed while the process was down) instead of resetting their cadence.
func (a *App) startScheduler(ctx context.Context, mods *modules) {
	runner := scheduler.NewRunner(scheduler.RunnerOptions{
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		BackoffBase: 500 * time.Millisecond,
		BackoffMax:  10 * time.Second,
		OnError: func(name string, err error) {
			a.deps.Log.Error(ctx, "scheduler: ALERTA, job falló definitivamente",
				logger.Str("job", name), logger.Err(err))
		},
		Log: a.deps.Log,
	})

	// Default: in-memory cadence, recomputed at each start.
	a.schedule = scheduler.NewScheduler(runner, scheduler.SchedulerOptions{
		Store: scheduler.NewMemoryStore(),
	})

	// Redis-backed store, opted into per job via WithStore below.
	persistent := fiberstore.New(a.deps.Storage)

	a.registerJobs(a.schedule, mods, persistent)
}

func (a *App) registerJobs(sched *scheduler.Scheduler, mods *modules, persistent scheduler.StateStore) {
	// The daily market jobs all key off the same 09:30 local market open;
	// the price/snapshot jobs run staggered after it.
	marketOpen := scheduler.DailyAt{Hour: 9, Minute: 30}

	// Ephemeral (default in-memory store): a missed daily run is simply
	// skipped; the next day recomputes fresh.
	sched.Register(market.NewExchangeRateScheduler(mods.market.Service(), a.deps.Log), marketOpen)
	sched.Register(market.NewAssetPriceScheduler(mods.market.Service(), a.deps.Log), scheduler.Delayed{Schedule: marketOpen, Delay: 90 * time.Second})

	// Persistent (Redis): resume across restarts and catch up on runs missed
	// while the process was down.
	sched.Register(portfolio.NewSnapshotJob(mods.portfolio.Service(), a.deps.Log), scheduler.Delayed{Schedule: marketOpen, Delay: 120 * time.Second}, scheduler.WithStore(persistent))
	sched.Register(notification.NewWeeklySummaryScheduler(mods.notification, a.deps.Log), scheduler.WeeklyAt{Day: time.Monday, Hour: 8, Minute: 30}, scheduler.WithStore(persistent))
	sched.Register(auth.NewCleanupJob(mods.auth.Service(), a.deps.Log), scheduler.Every{Interval: 5 * time.Hour}, scheduler.WithStore(persistent))

	sched.Start()
}

// authConfig projects the platform-wide environment onto the auth module's own
// Config, keeping the module decoupled from *config.Env (docs/TECH_DEBT.md #8).
func authConfig(env *config.Env) auth.Config {
	return auth.Config{
		JWTSecret:               env.JWTSecret,
		JWTAccessDuration:       env.JWTAccessDuration,
		JWTRefreshDuration:      env.JWTRefreshDuration,
		RefreshGracePeriod:      env.RefreshGracePeriod,
		MaxLoginAttempts:        env.MaxLoginAttempts,
		LoginLockout:            env.LoginLockout,
		Environment:             env.Environment,
		FrontendURL:             env.FrontendURL,
		InvitationExpiry:        env.InvitationExpiry,
		PasswordResetExpiry:     env.PasswordResetExpiry,
		EmailVerificationExpiry: env.EmailVerificationExpiry,
		SelfRegistrationEnabled: env.SelfRegistrationEnabled,
		TwoFactorPendingExpiry:  env.TwoFactorPendingExpiry,
	}
}
