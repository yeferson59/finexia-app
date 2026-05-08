package internal

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/finexia-app/internal/config"
	"github.com/yeferson59/finexia-app/internal/handlers"
	"github.com/yeferson59/finexia-app/internal/middlewares"
	"github.com/yeferson59/finexia-app/internal/repositories"
	"github.com/yeferson59/finexia-app/internal/routes"
	"github.com/yeferson59/finexia-app/internal/services"
)

type Bootstrap struct {
	app      *fiber.App
	db       *pgxpool.Pool
	envs     *config.Env
	storage  fiber.Storage
	s3Client *s3.Client
}

func New(app *fiber.App, db *pgxpool.Pool, envs *config.Env, storage fiber.Storage, s3Client *s3.Client) *Bootstrap {
	return new(Bootstrap{
		app:      app,
		db:       db,
		envs:     envs,
		storage:  storage,
		s3Client: s3Client,
	})
}

func (b *Bootstrap) Init(ctx context.Context) error {
	repos := repositories.New(b.db)
	services := services.New(repos, b.envs, b.s3Client)
	handlers, middlewares := handlers.New(ctx, services), middlewares.New(ctx, b.envs, b.storage)
	routes := routes.New(b.app, middlewares, handlers)

	routes.Init()

	return nil
}
