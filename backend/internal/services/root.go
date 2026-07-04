package services

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/config"
	"github.com/yeferson59/finexia-app/internal/logger"
	"github.com/yeferson59/finexia-app/internal/mail"
	"github.com/yeferson59/finexia-app/internal/prices"
)

// Mailer abstracts the outbound email service so tests can replace the
// Resend-backed implementation with a fake.
type Mailer interface {
	SendWaitlistConfirmation(email string) error
	SendActivityAlert(email string, data mail.ActivityAlertData) error
	SendWeeklySummary(email string, data mail.WeeklySummaryData) error
}

var _ Mailer = (*mail.Service)(nil)

type Services struct {
	repos         Repository
	cfg           *config.Env
	s3Client      *s3.Client
	storage       fiber.Storage
	mail          Mailer
	log           logger.Logger
	priceProvider prices.Provider
	// Pointer so every copy of Services shares the same cache (Services is
	// passed around by value).
	risksCache *risksCache
}

func New(repos Repository, cfg *config.Env, s3Client *s3.Client, storage fiber.Storage, mailService Mailer, log logger.Logger, priceProvider prices.Provider) Services {
	return Services{
		repos:         repos,
		cfg:           cfg,
		s3Client:      s3Client,
		storage:       storage,
		mail:          mailService,
		log:           log,
		priceProvider: priceProvider,
		risksCache:    &risksCache{},
	}
}
