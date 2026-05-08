package main

import (
	"context"

	"github.com/bytedance/sonic"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/yeferson59/finexia-app/internal"
	"github.com/yeferson59/finexia-app/internal/config"
)

type structValidator struct {
	validate *validator.Validate
}

func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

func main() {
	app, cfg := fiber.New(fiber.Config{
		JSONEncoder:     sonic.ConfigFastest.Marshal,
		JSONDecoder:     sonic.ConfigFastest.Unmarshal,
		StructValidator: new(structValidator{validate: validator.New()}),
	}), config.New()
	envs, ctx := cfg.LoadEnvs(), context.Background()
	dbPool, err := cfg.ConnectionDB(ctx, envs.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to database: " + err.Error())
	}
	defer dbPool.Close()

	storage := cfg.ConnectionCache(envs.CacheURL)

	defer func() {
		if err := storage.Close(); err != nil {
			log.Fatal("failed to close cache store: " + err.Error())
		}
	}()

	if err := internal.New(app, dbPool, envs, storage).Init(ctx); err != nil {
		log.Fatal("failed to initialize app: " + err.Error())
	}

	app.Listen(":" + envs.Port)
}
