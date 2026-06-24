package services

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/config"
	"github.com/yeferson59/finexia-app/internal/logger"
	"github.com/yeferson59/finexia-app/internal/mail"
	"github.com/yeferson59/finexia-app/internal/repositories"
)

type Services struct {
	repos    repositories.Repository
	cfg      *config.Env
	s3Client *s3.Client
	storage  fiber.Storage
	mail     *mail.Service
	log      logger.Logger
}

func New(repos repositories.Repository, cfg *config.Env, s3Client *s3.Client, storage fiber.Storage, mailService *mail.Service, log logger.Logger) Services {
	return Services{
		repos:    repos,
		cfg:      cfg,
		s3Client: s3Client,
		storage:  storage,
		mail:     mailService,
		log:      log,
	}
}
