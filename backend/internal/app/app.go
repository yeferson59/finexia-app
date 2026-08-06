// Package app is the composition root: the only place that wires
// infrastructure into the domain modules and the schedulers, and the only one
// that reads the environment. Adding a domain module means registering it here
// and nowhere else.
package app

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bytedance/sonic"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	goredis "github.com/redis/go-redis/v9"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/auth"
	"github.com/yeferson59/finexia-app/internal/health"
	"github.com/yeferson59/finexia-app/internal/market"
	"github.com/yeferson59/finexia-app/internal/marketing"
	"github.com/yeferson59/finexia-app/internal/notification"
	"github.com/yeferson59/finexia-app/internal/platform/cache"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/geoip"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata/providers"
	s3Store "github.com/yeferson59/finexia-app/internal/platform/objectstore/s3"
	"github.com/yeferson59/finexia-app/internal/platform/secretbox"
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
	Envs  *config.Env
	DB    *pgxpool.Pool
	Cache *goredis.Client
	S3    *s3.Client
	Mail  *mail.Service
	// Keyring seals the market-data API keys users bring. Required: without it
	// the market module cannot store a credential at all.
	Keyring *secretbox.Keyring
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
	case d.Cache == nil:
		return errors.New("app: Deps.Cache is required")
	case d.Mail == nil:
		return errors.New("app: Deps.Mail is required")
	case d.Keyring == nil:
		return errors.New("app: Deps.Keyring is required")
	case d.Log == nil:
		return errors.New("app: Deps.Log is required")
	default:
		return nil
	}
}

type App struct {
	fiber    *fiber.App
	deps     Deps
	storage  fiber.Storage
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
		StructValidator: new(structValidator{validator.New()}),
		ProxyHeader:     fiber.HeaderXForwardedFor,
		TrustProxy:      deps.Envs.TrustProxy,
		BodyLimit:       bodyLimit,
		// Without this, Fiber returns the X-Forwarded-For header *verbatim*
		// from c.IP() whenever the immediate peer is trusted — which, with the
		// private/loopback ranges trusted below, is every request in a
		// containerised deployment. The header is written by the client, so
		// c.IP() would be an attacker-chosen string: rotating it per request
		// gives each one its own bucket in every IP-keyed rate limiter (the
		// global 60/min and the 10/15min on the auth routes), which is what
		// stands between the public login, password-reset and email-verification
		// endpoints and unlimited credential stuffing, mail bombing, and
		// lockout-based denial of service against every account at once.
		//
		// With validation on, Fiber walks the forwarded chain right-to-left,
		// skips the proxies trusted below, and returns the first address that
		// is actually outside them.
		EnableIPValidation: true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Loopback:  true,
			LinkLocal: true,
			Private:   true,
			Proxies:   deps.Envs.TrustedProxies,
		},
	})

	return new(App{fiber: fiberApp, deps: deps}), nil
}

// Run wires the modules, their routes and the schedulers, then serves HTTP until
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
//
// The shared fiber.Storage is adapted from the Redis connection here rather
// than in Run because everything below depends on it — sessions, the market
// cache, the scheduler's persisted next-run times — so deriving it anywhere
// Run-only would leave every other caller of wire composing against a nil
// storage.
func (a *App) wire(ctx context.Context) {
	a.storage = cache.Storage(ctx, a.deps.Cache)

	mods := a.buildModules()
	a.mountRoutes(mods)
	a.startScheduler(ctx, mods)
}

// buildModules constructs the shared infrastructure (the market-data provider
// factory, geoip, per-user rate limiter) and every domain module, respecting
// their dependency order.
func (a *App) buildModules() *modules {

	// Market data is BYO-key: there is no process-wide provider to build here
	// because the application holds no provider credentials. The factory
	// assembles a chain per sync run from the calling user's own keys.
	priceProviders := providers.New(marketdata.DefaultHTTPClient)

	geo := geoip.New()
	userLimiter := httpx.KeyedRateLimiter(200, 1*time.Minute, func(c fiber.Ctx) string {
		userID := c.Locals(httpx.LocalUserID).(string)

		return "user_limit:" + userID
	})

	// Services first, then modules. The services form a DAG — marketing and
	// user depend on no other domain, auth reads both — so every dependency
	// below is a constructor argument. The modules come after, once auth
	// exists to supply the route guards they all share.
	marketingService := marketing.NewService(marketing.ServiceDeps{
		DB:   a.deps.DB,
		Mail: a.deps.Mail,
	})
	userService := user.NewService(user.ServiceDeps{
		DB:    a.deps.DB,
		Store: s3Store.New(a.deps.S3, a.deps.Envs.AWSS3BucketName),
		Log:   a.deps.Log,
		Cfg:   userConfig(a.deps.Envs),
	})
	authModule := auth.New(auth.Deps{
		DB:      a.deps.DB,
		Cfg:     authConfig(a.deps.Envs),
		Storage: a.storage,
		Mail:    a.deps.Mail,
		Geo:     geo,
		Log:     a.deps.Log,
		// auth owns neither table it needs here: users/roles belong to user and
		// the waitlist to marketing, so it reads both through their services
		// instead of querying them. Both are built above for this reason.
		Users:    userService,
		Waitlist: marketingService,
		Limiter:  userLimiter,
	})
	marketingModule := marketing.New(marketing.Deps{
		Service:   marketingService,
		AuthMiddl: authModule,
		Limiter:   userLimiter,
	})
	userModule := user.New(user.Deps{
		Service:   userService,
		AuthMiddl: authModule,
		Sessions:  authModule.Service(),
		Limiter:   userLimiter,
	})
	// market and portfolio need each other's services, so market is built in
	// two steps like user and marketing: the service first, which portfolio
	// consumes for the asset catalog, then the module with its routes and with
	// portfolio's holdings, which the per-user BYO-key sync needs.
	marketService := market.NewService(market.ServiceDeps{
		DB:        a.deps.DB,
		Storage:   a.storage,
		Log:       a.deps.Log,
		Providers: priceProviders,
		Keyring:   a.deps.Keyring,
	})
	portfolioModule := portfolio.New(portfolio.Deps{
		DB:        a.deps.DB,
		Cfg:       portfolioConfig(a.deps.Envs),
		Storage:   a.storage,
		Mail:      a.deps.Mail,
		User:      userService,
		Assets:    marketService,
		Log:       a.deps.Log,
		AuthMiddl: authModule,
		Limiter:   userLimiter,
	})
	marketModule := market.New(market.Deps{
		Service:        marketService,
		AuthMiddleware: authModule,
		Limiter:        userLimiter,
		Holdings:       portfolioModule.Service(),
		// Writing a credential is a sensitive surface and every verification
		// spends the user's own provider quota, so it gets a much tighter gate
		// than the shared 200/min above.
		CredentialLimiter: httpx.KeyedRateLimiter(10, 1*time.Minute, func(c fiber.Ctx) string {
			return "market_credentials:" + c.Locals(httpx.LocalUserID).(string)
		}),
	})

	return new(modules{
		health:       health.New(),
		marketing:    marketingModule,
		auth:         authModule,
		user:         userModule,
		market:       marketModule,
		portfolio:    portfolioModule,
		notification: notification.NewService(userService, portfolioModule.Service(), a.deps.Mail, notificationConfig(a.deps.Envs)),
	})
}

// mountRoutes installs the global middleware chain and each module's routes.
func (a *App) mountRoutes(mods *modules) {
	a.fiber.Use(httpx.Recovery())

	// CORS_ENABLED was read from the environment and then ignored, so an
	// operator who set it to false still got cross-origin responses with
	// credentials allowed. Honour it: off means no Access-Control-* headers at
	// all, which is the correct posture for a same-origin deployment.
	if a.deps.Envs.CORSEnabled {
		a.fiber.Use(httpx.CORS(a.deps.Envs.CORSOrigin, true))
	}

	a.fiber.Use(httpx.Helmet(), httpx.RequestID(), httpx.ResponseTime(), httpx.Logger(), httpx.RateLimiter(60, 1*time.Minute, false))

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
			a.deps.Log.Error(ctx, "scheduler: ALERT, failed finalizate job", logger.Str("job", name), logger.Err(err))
		},
		Log: a.deps.Log,
	})

	// Default: in-memory cadence, recomputed at each start.
	a.schedule = scheduler.NewScheduler(runner)

	// Redis-backed store, opted into per job via WithStore below.
	persistent := fiberstore.New(a.storage)

	a.registerJobs(a.schedule, mods, persistent)
}

func (a *App) registerJobs(sched *scheduler.Scheduler, mods *modules, persistent scheduler.StateStore) {
	// The market sync starts at the local market open.
	marketOpen := scheduler.DailyAt{Hour: 9, Minute: 30}

	// Market data is BYO-key, so this walks the users who configured a key and
	// syncs each with their own. It is persistent, unlike the two global jobs it
	// replaces: personal quotas are small, and a run missed over a restart is
	// worth catching up rather than silently skipping to tomorrow.
	//
	// It also needs its own retry policy. The runner's 30s default is a
	// per-attempt deadline, and this job paces its calls to fit personal
	// free-tier quotas — 13s between two Alpha Vantage requests — so a single
	// user with a handful of holdings already runs past it. Under the default it
	// would be cancelled mid-run every morning. Retries are off for the same
	// reason: a second attempt re-spends quota the first one already burned, and
	// per-user failures are logged and counted inside the job rather than raised.
	sched.Register(
		market.NewSyncJob(mods.market.Service(), mods.portfolio.Service(), a.deps.Log),
		marketOpen,
		scheduler.WithStore(persistent),
		scheduler.WithRetry(scheduler.JobOptions{
			Timeout:    2 * time.Hour,
			MaxRetries: scheduler.Retries(0),
		}),
	)

	// Persistent (Redis): resume across restarts and catch up on runs missed
	// while the process was down.
	//
	// The snapshot runs in the evening rather than staggered minutes behind the
	// market open. Under the old model one global job refreshed every price in
	// seconds; the BYO-key sync walks every user at their own pace and can take
	// hours, so a snapshot taken two minutes in would record yesterday's values.
	sched.Register(portfolio.NewSnapshotJob(mods.portfolio.Service(), a.deps.Log), scheduler.DailyAt{Hour: 22, Minute: 0}, scheduler.WithStore(persistent))
	sched.Register(notification.NewWeeklySummaryScheduler(mods.notification, a.deps.Log), scheduler.WeeklyAt{Day: time.Monday, Hour: 8, Minute: 30}, scheduler.WithStore(persistent))
	sched.Register(auth.NewCleanupJob(mods.auth.Service(), a.deps.Log), scheduler.Every{Interval: 5 * time.Hour}, scheduler.WithStore(persistent))

	sched.Start()
}

// The *Config helpers below project the platform-wide environment onto each
// module's own Config. Reading the environment is the composition root's job:
// a module declares the handful of settings it actually consumes and stays
// decoupled from *config.Env. market needs none, so it takes no config at all.

// authConfig projects the platform-wide environment onto the auth module's own
// Config, keeping the module decoupled from *config.Env.
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

// userConfig projects the environment onto the user module's Config.
func userConfig(env *config.Env) user.Config {
	return user.Config{
		PublicURL:   env.PublicURL,
		FrontendURL: env.FrontendURL,
	}
}

// portfolioConfig projects the environment onto the portfolio module's Config.
func portfolioConfig(env *config.Env) portfolio.Config {
	return portfolio.Config{
		FrontendURL: env.FrontendURL,
	}
}

// notificationConfig projects the environment onto the notification module's Config.
func notificationConfig(env *config.Env) notification.Config {
	return notification.Config{
		PublicURL: env.PublicURL,
	}
}
