package internal

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/alphavantage"
	"github.com/yeferson59/finexia-app/internal/config"
	"github.com/yeferson59/finexia-app/internal/finnhub"
	"github.com/yeferson59/finexia-app/internal/handlers"
	"github.com/yeferson59/finexia-app/internal/logger"
	"github.com/yeferson59/finexia-app/internal/mail"
	"github.com/yeferson59/finexia-app/internal/middlewares"
	"github.com/yeferson59/finexia-app/internal/prices"
	"github.com/yeferson59/finexia-app/internal/repositories"
	"github.com/yeferson59/finexia-app/internal/routes"
	"github.com/yeferson59/finexia-app/internal/scheduler"
	"github.com/yeferson59/finexia-app/internal/services"
	"github.com/yeferson59/finexia-app/internal/yahoo"
)

type Bootstrap struct {
	app         *fiber.App
	db          *pgxpool.Pool
	envs        *config.Env
	storage     fiber.Storage
	s3Client    *s3.Client
	mailService *mail.Service
}

func New(app *fiber.App, db *pgxpool.Pool, envs *config.Env, storage fiber.Storage, s3Client *s3.Client, mailService *mail.Service) *Bootstrap {
	return new(Bootstrap{
		app:         app,
		db:          db,
		envs:        envs,
		storage:     storage,
		s3Client:    s3Client,
		mailService: mailService,
	})
}

func (b *Bootstrap) Init(ctx context.Context) error {
	rootLog := logger.New(logger.Config{
		Level:       logger.LevelInfo,
		Output:      os.Stderr,
		Environment: b.envs.Environment,
	})

	repos := repositories.New(b.db)
	priceProvider := prices.NewFallback(
		alphavantage.New(b.envs.AlphaVantageAPIKey),
		finnhub.New(b.envs.FinnhubAPIKey),
		yahoo.New(),
	)
	services := services.New(repos, b.envs, b.s3Client, b.storage, b.mailService, rootLog, priceProvider)
	handlers, middlewares := handlers.New(ctx, services, b.envs), middlewares.New(ctx, b.envs, b.storage, services)
	routes := routes.New(b.app, middlewares, handlers)

	routes.Init()

	sched := scheduler.NewExchangeRateScheduler(services, 6, rootLog) // 06:00 UTC daily
	go sched.Start(ctx)

	assetSched := scheduler.NewAssetPriceScheduler(services, 14, 90*time.Second, rootLog) // 14:00 UTC, 90s startup delay
	go assetSched.Start(ctx)

	snapshotSched := scheduler.NewPortfolioSnapshotScheduler(services, 15, 120*time.Second, rootLog) // 15:00 UTC, 120s startup delay
	go snapshotSched.Start(ctx)

	weeklySched := scheduler.NewWeeklySummaryScheduler(services, 9, rootLog) // Mondays 09:00 UTC
	go weeklySched.Start(ctx)

	authCleanupSched := scheduler.NewAuthCleanupScheduler(services, 3, rootLog) // 03:00 UTC daily
	go authCleanupSched.Start(ctx)

	return nil
}
