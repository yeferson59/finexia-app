package handlers

import (
	"context"

	"github.com/yeferson59/finexia-app/internal/services"
)

type Handlers struct {
	ctx      context.Context
	services services.Services
}

func New(ctx context.Context, services services.Services) Handlers {
	return Handlers{
		ctx:      ctx,
		services: services,
	}
}
