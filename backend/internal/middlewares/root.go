package middlewares

import (
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/config"
)

type Middlewares struct {
	envs    *config.Env
	storage fiber.Storage
}

func New(envs *config.Env, storage fiber.Storage) Middlewares {
	return Middlewares{
		envs:    envs,
		storage: storage,
	}
}
