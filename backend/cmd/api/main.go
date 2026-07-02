package main

import (
	"context"
	"errors"

	"github.com/bytedance/sonic"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"

	"github.com/yeferson59/finexia-app/internal"
	"github.com/yeferson59/finexia-app/internal/config"
	"github.com/yeferson59/finexia-app/internal/mail"
)

type structValidator struct {
	validate *validator.Validate
}

func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}

func run() error {
	cfg := config.New()
	envs, ctx := cfg.LoadEnvs(), context.Background()

	// The API runs behind the SvelteKit SSR server (and possibly a load
	// balancer), so without trusting X-Forwarded-For every rate limiter would
	// key all end users under the proxy's single IP. Only loopback/link-local/
	// private peers (plus TRUSTED_PROXIES) are trusted, so a directly-connected
	// public client cannot spoof its IP.
	app := fiber.New(fiber.Config{
		JSONEncoder:     sonic.ConfigFastest.Marshal,
		JSONDecoder:     sonic.ConfigFastest.Unmarshal,
		StructValidator: new(structValidator{validate: validator.New()}),
		ProxyHeader:     fiber.HeaderXForwardedFor,
		TrustProxy:      envs.TrustProxy,
		// Without validation, a trusted peer that omits X-Forwarded-For would
		// yield an empty c.IP(); with it, Fiber falls back to the remote IP.
		EnableIPValidation: true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Loopback:  true,
			LinkLocal: true,
			Private:   true,
			Proxies:   envs.TrustedProxies,
		},
	})
	dbPool, err := cfg.ConnectionDB(ctx, envs.DatabaseURL)
	if err != nil {
		return errors.New("failed to connect to database: " + err.Error())
	}
	defer dbPool.Close()

	storageCache := cfg.ConnectionCache(envs.CacheURL)

	defer func() {
		if err := storageCache.Close(); err != nil {
			log.Fatal("failed to close cache store: " + err.Error())
		}
	}()

	s3Client, err := cfg.Storage(ctx, envs.AWSAccessKeyID, envs.AWSDefaultRegion, envs.AWSEndpointURL, envs.AWSSecretAccessKey)
	if err != nil {
		return errors.New("failed to create storage: " + err.Error())
	}

	mailService, err := mail.New(envs.ResendAPIKey, envs.EmailFrom)
	if err != nil {
		return errors.New("failed to init mail service: " + err.Error())
	}

	if err := internal.New(app, dbPool, envs, storageCache, s3Client, mailService).Init(ctx); err != nil {
		return errors.New("failed to initialize app: " + err.Error())
	}

	if err := app.Listen(":" + envs.Port); err != nil {
		return errors.New("failed to listen: " + err.Error())
	}

	return nil
}
