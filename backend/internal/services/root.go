package services

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/yeferson59/finexia-app/internal/config"
	"github.com/yeferson59/finexia-app/internal/repositories"
)

type Services struct {
	repos    repositories.Repository
	cfg      *config.Env
	s3Client *s3.Client
}

func New(repos repositories.Repository, cfg *config.Env, s3Client *s3.Client) Services {
	return Services{
		repos:    repos,
		cfg:      cfg,
		s3Client: s3Client,
	}
}
