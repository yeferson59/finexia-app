package services

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/user"
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
	SendActivityAlert(email string, data mail.ActivityAlertData) error
	SendSecurityAlert(email string, data mail.SecurityAlertData) error
	SendWeeklySummary(email string, data mail.WeeklySummaryData) error
}

var _ Mailer = (*mail.Service)(nil)

// AuthService is the slice of the auth module these legacy services still
// need (change-password verification, session revocation on password
// changes). Consumer-defined, satisfied by *auth.Service; disappears when
// the user domain migrates in Fase 5.
type AuthService interface {
	VerifyPassword(ctx context.Context, userID uuid.UUID, currentPassword string) error
	RevokeOtherSessions(ctx context.Context, userID uuid.UUID, currentToken string) (int64, error)
}

type UserService interface {
	GetUserPreferences(ctx context.Context, userID uuid.UUID) (user.UserPreferences, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (identity.User, error)
	GetUsersWithWeeklySummary(ctx context.Context) ([]identity.User, error)
}

type Services struct {
	repos         Repository
	cfg           *config.Env
	s3Client      *s3.Client
	storage       fiber.Storage
	mail          Mailer
	geo           GeoLocator
	log           logger.Logger
	priceProvider marketdata.Provider
	auth          AuthService
	user          UserService
	// Pointer so every copy of Services shares the same cache (Services is
	// passed around by value).
	risksCache *risksCache
}

func New(repos Repository, cfg *config.Env, s3Client *s3.Client, storage fiber.Storage, mailService Mailer, geo GeoLocator, log logger.Logger, priceProvider marketdata.Provider, authSvc AuthService, userSvc UserService) Services {
	return Services{
		repos:         repos,
		cfg:           cfg,
		s3Client:      s3Client,
		storage:       storage,
		mail:          mailService,
		geo:           geo,
		log:           log,
		priceProvider: priceProvider,
		auth:          authSvc,
		user:          userSvc,
		risksCache:    &risksCache{},
	}
}
