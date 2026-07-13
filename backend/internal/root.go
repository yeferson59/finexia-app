package internal

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/handlers"
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

type Bootstrap struct {
	app         *fiber.App
	db          *pgxpool.Pool
	envs        *config.Env
	storage     fiber.Storage
	s3Client    *s3.Client
	mailService *mail.Service
	log         logger.Logger
}

func New(app *fiber.App, db *pgxpool.Pool, envs *config.Env, storage fiber.Storage, s3Client *s3.Client, mailService *mail.Service, log logger.Logger) *Bootstrap {
	return new(Bootstrap{
		app:         app,
		db:          db,
		envs:        envs,
		storage:     storage,
		s3Client:    s3Client,
		mailService: mailService,
		log:         log,
	})
}

func (b *Bootstrap) Init(ctx context.Context) error {
	repos := repositories.New(b.db)
	priceProvider := marketdata.NewFallback(
		alphavantage.New(b.envs.AlphaVantageAPIKey),
		finnhub.New(b.envs.FinnhubAPIKey),
		yahoo.New(),
	)
	services := services.New(&repos, b.envs, b.s3Client, b.storage, b.mailService, geoip.New(), b.log, priceProvider)
	handlers, middlewares := handlers.New(services, b.envs), middlewares.New(ctx, b.envs, b.storage, services)
	routes := routes.New(b.app, middlewares, handlers)

	routes.Init()

	sched := scheduler.NewExchangeRateScheduler(services, 6, b.log)
	go sched.Start(ctx)

	assetSched := scheduler.NewAssetPriceScheduler(services, 14, 90*time.Second, b.log)
	go assetSched.Start(ctx)

	snapshotSched := scheduler.NewPortfolioSnapshotScheduler(services, 15, 120*time.Second, b.log)
	go snapshotSched.Start(ctx)

	weeklySched := scheduler.NewWeeklySummaryScheduler(services, 9, b.log)
	go weeklySched.Start(ctx)

	authCleanupSched := scheduler.NewAuthCleanupScheduler(services, 3, b.log)
	go authCleanupSched.Start(ctx)

	return nil
}
