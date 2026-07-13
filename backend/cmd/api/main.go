package main

import (
	"context"
	"errors"
	"os"

	"github.com/bytedance/sonic"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal"
	"github.com/yeferson59/finexia-app/internal/platform/cache"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/database"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/internal/platform/objectstore"
)

type structValidator struct {
	validate *validator.Validate
}

func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

func main() {
	cfg := config.New()
	envs := cfg.LoadEnvs()
	log := logger.New(logger.Config{
		Level:       logger.LevelInfo,
		Output:      os.Stderr,
		Environment: envs.Environment,
	})
	ctx := context.Background()

	if err := run(ctx, envs); err != nil {
		log.With(logger.Str("cmd", "main")).Fatal(ctx, "application error: "+err.Error())
	}
}

func run(ctx context.Context, envs *config.Env) error {
	app := fiber.New(fiber.Config{
		JSONEncoder:        sonic.ConfigFastest.Marshal,
		JSONDecoder:        sonic.ConfigFastest.Unmarshal,
		StructValidator:    new(structValidator{validate: validator.New()}),
		ProxyHeader:        fiber.HeaderXForwardedFor,
		TrustProxy:         envs.TrustProxy,
		EnableIPValidation: true,
		BodyLimit:          10 * 1024 * 1024,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Loopback:  true,
			LinkLocal: true,
			Private:   true,
			Proxies:   envs.TrustedProxies,
		},
	})
	log := logger.New(logger.Config{
		Level:       logger.LevelInfo,
		Output:      os.Stderr,
		Environment: envs.Environment,
	})
	dbPool, err := database.Connect(ctx, envs.DatabaseURL)
	if err != nil {
		return errors.New("failed to connect to database: " + err.Error())
	}
	defer dbPool.Close()

	storageCache := cache.Connect(envs.CacheURL)
	defer func() {
		if err := storageCache.Close(); err != nil {
			log.With(logger.Str("cmd", "run")).Fatal(ctx, "failed to close cache store: "+err.Error())
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

	if err := internal.New(app, dbPool, envs, storageCache, s3Client, mailService, log).Init(ctx); err != nil {
		return errors.New("failed to initialize app: " + err.Error())
	}

	if err := app.Listen(":" + envs.Port); err != nil {
		return errors.New("failed to listen: " + err.Error())
	}

	return nil
}
