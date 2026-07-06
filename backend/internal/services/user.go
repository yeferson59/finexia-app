package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/logger"
	"github.com/yeferson59/finexia-app/internal/mail"
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

func (s *Services) BanUser(ctx context.Context, id uuid.UUID, ban bool) error {
	return s.repos.BanUser(ctx, id, ban)
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

func (s *Services) UploadAvatarToS3(ctx context.Context, userID uuid.UUID, file io.Reader, contentType string) (entities.User, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return entities.User{}, errors.New("failed to read file")
	}

	key := fmt.Sprintf("avatars/%s/avatar", userID.String())

	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.AWSS3BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return entities.User{}, fmt.Errorf("failed to upload to S3: %w", err)
	}

	imageURL := fmt.Sprintf("%s/users/%s/avatar", s.cfg.PublicURL, userID.String())

	return s.repos.UpdateUserImage(ctx, userID, imageURL)
}

func (s *Services) GetAvatarFromS3(ctx context.Context, userID uuid.UUID) (io.ReadCloser, string, error) {
	key := fmt.Sprintf("avatars/%s/avatar", userID.String())

	result, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.AWSS3BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", err
	}

	return result.Body, aws.ToString(result.ContentType), nil
}

func (s *Services) ChangePassword(ctx context.Context, userID uuid.UUID, currentToken, currentPassword, newPassword, ipAddress, userAgent string) error {
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

	if err := s.repos.UpdateUserPassword(ctx, userID, string(hashed)); err != nil {
		return err
	}

	// Whoever else holds a session (a stolen token, a forgotten shared
	// computer) must not survive a password change: only the session that
	// performed the change stays alive.
	if _, err := s.RevokeOtherSessions(ctx, userID, currentToken); err != nil {
		s.log.Error("change password: failed to revoke other sessions", logger.Err(err))
	}

	go s.sendPasswordChangedAlert(userID, ipAddress, userAgent)

	return nil
}

// sendPasswordChangedAlert notifies the user their password changed. Like the
// login alert, it bypasses email preferences: if the change wasn't theirs,
// this email is their only chance to react. Best-effort.
func (s *Services) sendPasswordChangedAlert(userID uuid.UUID, ipAddress, userAgent string) {
	if s.mail == nil {
		return
	}

	ctx := context.Background()

	user, err := s.repos.GetUserByID(ctx, userID)
	if err != nil {
		return
	}

	if ipAddress == "" {
		ipAddress = "desconocida"
	}
	if userAgent == "" {
		userAgent = "desconocido"
	}

	_ = s.mail.SendSecurityAlert(user.Email, mail.SecurityAlertData{
		UserName:    user.Name,
		Event:       "cambio de contraseña",
		Detail:      "La contraseña de tu cuenta fue cambiada y se cerraron las demás sesiones activas.",
		IPAddress:   truncate(ipAddress, 45),
		UserAgent:   truncate(userAgent, 255),
		When:        time.Now().UTC().Format("02 Jan 2006 15:04 UTC"),
		SecurityURL: s.cfg.FrontendURL + "/dashboard/settings",
	})
}
