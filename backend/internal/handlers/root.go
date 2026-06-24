package handlers

import (
	"context"

	"github.com/yeferson59/finexia-app/internal/config"
	"github.com/yeferson59/finexia-app/internal/services"
)

type Handlers struct {
	ctx      context.Context
	services services.Services
	cfg      *config.Env
}

func New(ctx context.Context, services services.Services, cfg *config.Env) Handlers {
	return Handlers{
		ctx:      ctx,
		services: services,
		cfg:      cfg,
	}
}
