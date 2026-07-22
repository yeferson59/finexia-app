package portfolio

import (
	"context"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/internal/user"
)

// UserReader is the slice of the user module this module needs to build the
// transaction activity alert. Satisfied by *user.Service.
type UserReader interface {
	GetUserPreferences(ctx context.Context, userID uuid.UUID) (user.UserPreferences, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (identity.User, error)
}

// Mailer abstracts the outbound email service so tests can replace the
// Resend-backed implementation with a fake.
type Mailer interface {
	SendActivityAlert(email string, data mail.ActivityAlertData) error
}

var _ Mailer = (*mail.Service)(nil)

// risksCache memoizes the risk catalog: it is seed data shared by every user
// and requested on each portfolio page, so a short TTL avoids one DB
// round-trip per page view without risking staleness for long.
type risksCache struct {
	mu        sync.RWMutex
	risks     []Risk
	expiresAt time.Time
}

const risksCacheTTL = 10 * time.Minute

// Service holds the portfolio use cases. It is exposed by Module.Service()
// and consumed by other areas only through interfaces they declare.
type Service struct {
	repo    Repository
	cfg     *config.Env
	storage fiber.Storage
	mail    Mailer
	user    UserReader
	log     logger.Logger

	risksCache *risksCache
}

func NewService(repo Repository, cfg *config.Env, storage fiber.Storage, mailService Mailer, userReader UserReader, log logger.Logger) *Service {
	return new(Service{
		repo:       repo,
		cfg:        cfg,
		storage:    storage,
		mail:       mailService,
		user:       userReader,
		log:        log,
		risksCache: &risksCache{},
	})
}
