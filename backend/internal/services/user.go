package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/pkg/helpers"
)

func (s *Services) GetListUsers(ctx context.Context, offset, limit uint) ([]entities.User, uint, error) {
	return s.repos.ListUsers(ctx, offset, limit)
}

func (s *Services) GetUserByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	return s.repos.GetUserByID(ctx, id)
}

func (s *Services) CreateUser(ctx context.Context, name, email string) (entities.User, error) {
	name = helpers.NormalizateNames(name)

	return s.repos.CreateUser(ctx, name, email)
}

func (s *Services) UpdateUser(ctx context.Context, id uuid.UUID, name, email, image string) (entities.User, error) {
	existUser, err := s.repos.GetUserByID(ctx, id)
	if err != nil {
		return entities.User{}, err
	}

	if existUser.DeletedAt != nil {
		return entities.User{}, errors.New("not found user")
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

	return s.repos.UpdateUser(ctx, existUser.ID, existUser.Name, existUser.Email, existUser.Image)
}

func (s *Services) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repos.DeleteUser(ctx, id)
}

func (s *Services) GetCurrentUser(ctx context.Context, userID uuid.UUID) (entities.User, error) {
	return s.repos.GetUserByID(ctx, userID)
}

func (s *Services) UpdateCurrentUser(ctx context.Context, userID uuid.UUID, name, preferredCurrency, image string) (entities.User, error) {
	existing, err := s.repos.GetUserByID(ctx, userID)
	if err != nil {
		return entities.User{}, err
	}

	if strings.TrimSpace(name) != "" {
		existing.Name = helpers.NormalizateNames(name)
	}
	if strings.TrimSpace(preferredCurrency) != "" {
		existing.PreferredCurrency = strings.ToUpper(strings.TrimSpace(preferredCurrency))
	}
	if strings.TrimSpace(image) != "" {
		existing.Image = image
	}

	return s.repos.UpdateUserProfile(ctx, userID, existing.Name, existing.PreferredCurrency, existing.Image)
}

func (s *Services) GetUserPreferences(ctx context.Context, userID uuid.UUID) (entities.UserPreferences, error) {
	return s.repos.GetUserPreferences(ctx, userID)
}

func (s *Services) UpdateUserPreferences(ctx context.Context, userID uuid.UUID, emailAlerts, weeklySummary bool) (entities.UserPreferences, error) {
	return s.repos.UpsertUserPreferences(ctx, userID, emailAlerts, weeklySummary)
}

func (s *Services) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	account, err := s.repos.GetAccountByUserID(ctx, userID)
	if err != nil {
		return errors.New("not found account")
	}

	if err := account.ComparePassword(currentPassword); err != nil {
		return errors.New("invalid current password")
	}

	if currentPassword == newPassword {
		return errors.New("invalid new password: must differ from current password")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repos.UpdateUserPassword(ctx, userID, string(hashed))
}
