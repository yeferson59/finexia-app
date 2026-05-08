package services

import (
	"github.com/yeferson59/finexia-app/internal/config"
	"github.com/yeferson59/finexia-app/internal/repositories"
)

type Services struct {
	repos repositories.Repository
	cfg   *config.Env
}

func New(repos repositories.Repository, cfg *config.Env) Services {
	return Services{
		repos: repos,
		cfg:   cfg,
	}
}
