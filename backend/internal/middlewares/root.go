package middlewares

import (
	"context"

	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/services"
)

type Middlewares struct {
	ctx     context.Context
	envs    *config.Env
	storage fiber.Storage
	svc     services.Services
}

func New(ctx context.Context, envs *config.Env, storage fiber.Storage, svc services.Services) Middlewares {
	return Middlewares{
		ctx:     ctx,
		envs:    envs,
		storage: storage,
		svc:     svc,
	}
}
