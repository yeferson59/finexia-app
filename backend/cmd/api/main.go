package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yeferson59/finexia-app/internal/app"
	"github.com/yeferson59/finexia-app/internal/platform/cache"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/database"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/internal/platform/objectstore"
	"github.com/yeferson59/finexia-app/internal/platform/secretbox"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	log := logger.New(logger.Config{
		Level:       logger.LevelInfo,
		Output:      os.Stderr,
		Environment: cfg.Environment,
	})
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg, log); err != nil {
		log.With(logger.Str("cmd", "main")).Fatal(ctx, "application error: "+err.Error())
	}
}

// run creates the infrastructure and hands it to the composition root; all
// application wiring lives in internal/app.
func run(ctx context.Context, envs *config.EnvConfig, log logger.Logger) error {
	// Before anything is connected: a missing or guessable JWT_SECRET means
	// every access token this process would issue is forgeable by anyone, so
	// it stops the boot rather than degrading silently.
	if err := envs.Validate(); err != nil {
		return err
	}

	dbPool, err := database.Connect(ctx, envs.DatabaseURL)
	if err != nil {
		return errors.New("failed to connect to database: " + err.Error())
	}
	defer dbPool.Close()

	client := cache.Conn(envs.RedisHost, envs.RedisPort, envs.RedisPassword, envs.RedisDB)
	defer func() {
		if err := client.Close(); err != nil {
			log.Error(ctx, "failed to close redis: "+err.Error())
		}
	}()

	s3Client, err := objectstore.Connect(ctx, envs.AWSAccessKeyID, envs.AWSDefaultRegion, envs.AWSEndpointURL, envs.AWSSecretAccessKey)
	if err != nil {
		return errors.New("failed to create storage: " + err.Error())
	}

	mailService, err := mail.New(envs.ResendAPIKey, envs.EmailFrom)
	if err != nil {
		return errors.New("failed to init mail service: " + err.Error())
	}

	// The keyring seals the market-data API keys users bring. Refusing to start
	// without it is deliberate: a default would mean storing those keys under a
	// guessable key, which is exactly the failure this is meant to prevent.
	// Generate one with: openssl rand -base64 32
	keyring, err := secretbox.NewKeyring(envs.MarketKEKKeys[0], envs.MarketKEKActive)
	if err != nil {
		return errors.New("failed to load MARKET_KEK_KEYS: " + err.Error())
	}

	application, err := app.New(app.Deps{
		Envs:    envs,
		DB:      dbPool,
		Cache:   client,
		S3:      s3Client,
		Mail:    mailService,
		Keyring: keyring,
		Log:     log,
	})
	if err != nil {
		return errors.New("failed to build application: " + err.Error())
	}

	return application.Run(ctx)
}
