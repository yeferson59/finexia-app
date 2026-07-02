package handlers

import (
	"github.com/yeferson59/finexia-app/internal/config"
	"github.com/yeferson59/finexia-app/internal/services"
)

type Handlers struct {
	services services.Services
	cfg      *config.Env
}

func New(services services.Services, cfg *config.Env) Handlers {
	return Handlers{
		services: services,
		cfg:      cfg,
	}
}
