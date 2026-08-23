package user

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/currency"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/objectstore"
	"github.com/yeferson59/finexia-app/pkg/helpers"
)

// Service holds the user domain use cases: the admin CRUD, the self-service
// profile and preferences, and the avatar object storage. Credentials are not
// here — password verification, change and reset are auth's, which reads this
// service through its own UserReader interface.
type Service struct {
	repo  Repository
	store objectstore.Store
	log   logger.Logger
	cfg   Config
}

// newService is the fake-friendly constructor: the module's NewService wraps
// it with the real Postgres repository.
func newService(repo Repository, store objectstore.Store, log logger.Logger, cfg Config) *Service {
	return new(Service{
		repo:  repo,
		store: store,
		log:   log,
		cfg:   cfg,
	})
}

func (s *Service) GetListUsers(ctx context.Context, offset, limit uint) ([]identity.User, uint, error) {
	return s.repo.List(ctx, offset, limit)
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (identity.User, error) {
	return s.repo.GetByID(ctx, id)
}

// GetUserByEmail resolves a user by address. It is the read the auth module
// consumes through its own UserReader interface: the users/roles tables belong
// here, so auth asks instead of querying them.
func (s *Service) GetUserByEmail(ctx context.Context, email string) (identity.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *Service) CreateUser(ctx context.Context, name, email string) (identity.User, error) {
	name = helpers.NormalizateNames(name)

	return s.repo.Create(ctx, name, email)
}

func (s *Service) UpdateUser(ctx context.Context, id uuid.UUID, name, email, image string) (identity.User, error) {
	existUser, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return identity.User{}, err
	}

	if existUser.DeletedAt != nil {
		return identity.User{}, ErrUserNotFound
	}

	if strings.TrimSpace(name) != "" && existUser.Name != name {
		existUser.Name = helpers.NormalizateNames(name)
	}

	if strings.TrimSpace(email) != "" && existUser.Email != email {
		existUser.Email = email
	}

	if strings.TrimSpace(image) != "" && existUser.Image != image {
		existUser.Image = image
	}

	return s.repo.Update(ctx, existUser.ID, existUser.Name, existUser.Email, existUser.Image)
}

func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) BanUser(ctx context.Context, id uuid.UUID, ban bool) error {
	return s.repo.Ban(ctx, id, ban)
}

func (s *Service) GetCurrentUser(ctx context.Context, userID uuid.UUID) (identity.User, error) {
	return s.repo.GetByID(ctx, userID)
}

func (s *Service) UpdateCurrentUser(ctx context.Context, userID uuid.UUID, name, preferredCurrency, image string) (identity.User, error) {
	existing, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return identity.User{}, err
	}

	if strings.TrimSpace(name) != "" {
		existing.Name = helpers.NormalizateNames(name)
	}
	// The preferred currency is not a label: it is what the dashboard converts
	// every total into, and what the market-data sync fetches pairs for. A code
	// outside the supported set has no rate behind it, so accepting one would
	// leave the account with figures that cannot be converted and a sync asking
	// providers for a pair that does not exist.
	if strings.TrimSpace(preferredCurrency) != "" {
		code := currency.Normalize(preferredCurrency)
		if !currency.IsSupported(code) {
			return identity.User{}, fmt.Errorf("%w %q: must be one of %s", ErrUnsupportedCurrency, code, currency.List())
		}
		existing.PreferredCurrency = code
	}
	if strings.TrimSpace(image) != "" {
		existing.Image = image
	}

	return s.repo.UpdateProfile(ctx, userID, existing.Name, existing.PreferredCurrency, existing.Image)
}

func (s *Service) GetUserPreferences(ctx context.Context, userID uuid.UUID) (UserPreferences, error) {
	return s.repo.GetPreferences(ctx, userID)
}

func (s *Service) UpdateUserPreferences(ctx context.Context, userID uuid.UUID, emailAlerts, weeklySummary bool) (UserPreferences, error) {
	return s.repo.UpsertPreferences(ctx, userID, emailAlerts, weeklySummary)
}

func (s *Service) UploadAvatarToS3(ctx context.Context, userID uuid.UUID, file io.Reader, contentType string) (identity.User, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return identity.User{}, httpx.AsBadRequest(errors.New("failed to read file"))
	}

	key := fmt.Sprintf("avatars/%s/avatar", userID.String())

	err = s.store.Put(ctx, key, contentType, data)
	if err != nil {
		return identity.User{}, httpx.AsBadRequest(fmt.Errorf("failed to upload to S3: %w", err))
	}

	imageURL := fmt.Sprintf("%s/users/%s/avatar", s.cfg.PublicURL, userID.String())

	return s.repo.UpdateImage(ctx, userID, imageURL)
}

func (s *Service) GetAvatarFromS3(ctx context.Context, userID uuid.UUID) (io.ReadCloser, string, error) {
	key := fmt.Sprintf("avatars/%s/avatar", userID.String())

	return s.store.Get(ctx, key)
}

func (s *Service) GetUsersWithWeeklySummary(ctx context.Context) ([]identity.User, error) {
	return s.repo.GetWeeklySummary(ctx)
}
