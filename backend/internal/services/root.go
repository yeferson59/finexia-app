package services

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v3"

	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/logger"
	"github.com/yeferson59/finexia-app/internal/mail"
	"github.com/yeferson59/finexia-app/internal/prices"
)

// GeoLocator resolves an IP address to a human-readable approximate location
// for security notifications. Implementations return "" when the location is
// unknown (private IP, lookup failure), never an error.
type GeoLocator interface {
	Locate(ctx context.Context, ip string) string
}

// Mailer abstracts the outbound email service so tests can replace the
// Resend-backed implementation with a fake.
type Mailer interface {
	SendWaitlistConfirmation(email string) error
	SendActivityAlert(email string, data mail.ActivityAlertData) error
	SendSecurityAlert(email string, data mail.SecurityAlertData) error
	SendWeeklySummary(email string, data mail.WeeklySummaryData) error
	SendInvitation(email string, data mail.InvitationData) error
	SendPasswordReset(email string, data mail.PasswordResetData) error
	SendEmailVerification(email string, data mail.EmailVerificationData) error
}

var _ Mailer = (*mail.Service)(nil)

type Services struct {
	repos         Repository
	cfg           *config.Env
	s3Client      *s3.Client
	storage       fiber.Storage
	mail          Mailer
	geo           GeoLocator
	log           logger.Logger
	priceProvider prices.Provider
	// Pointer so every copy of Services shares the same cache (Services is
	// passed around by value).
	risksCache *risksCache
}

func New(repos Repository, cfg *config.Env, s3Client *s3.Client, storage fiber.Storage, mailService Mailer, geo GeoLocator, log logger.Logger, priceProvider prices.Provider) Services {
	return Services{
		repos:         repos,
		cfg:           cfg,
		s3Client:      s3Client,
		storage:       storage,
		mail:          mailService,
		geo:           geo,
		log:           log,
		priceProvider: priceProvider,
		risksCache:    &risksCache{},
	}
}

// locateIP resolves the approximate location of an IP for security alert
// emails. Bounded by its own timeout so a slow lookup can only delay the
// (already asynchronous) email, never the request that triggered it.
func (s *Services) locateIP(ipAddress string) string {
	if s.geo == nil {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.geo.Locate(ctx, ipAddress)
}
