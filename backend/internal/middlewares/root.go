package middlewares

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/yeferson59/finexia-app/internal/config"
)

type Middlewares struct {
	ctx     context.Context
	envs    *config.Env
	storage fiber.Storage
}

func New(ctx context.Context, envs *config.Env, storage fiber.Storage) Middlewares {
	return Middlewares{
		ctx:     ctx,
		envs:    envs,
		storage: storage,
	}
}
