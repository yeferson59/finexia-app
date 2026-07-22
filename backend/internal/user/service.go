package user

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/auth"
	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/marketing"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/internal/platform/objectstore"
	"github.com/yeferson59/finexia-app/pkg/helpers"
)

type mailer interface {
	SendSecurityAlert(email string, data mail.SecurityAlertData) error
}

type authService interface {
	VerifyPassword(ctx context.Context, userID uuid.UUID, currentPassword string) error
	RevokeOtherSessions(ctx context.Context, userID uuid.UUID, currentToken string) (int64, error)
	CreateInvitation(ctx context.Context, email, name, role string, invitedBy uuid.UUID) (auth.Invitation, error)
	ListInvitations(ctx context.Context, offset, limit uint) ([]auth.Invitation, uint, error)
	ResendInvitation(ctx context.Context, id, invitedBy uuid.UUID) (auth.Invitation, error)
	RevokeInvitation(ctx context.Context, id uuid.UUID) error
}

// marketingService is the slice of marketing.Service the admin waitlist
// dashboard reads from.
type marketingService interface {
	ListWaitlist(ctx context.Context, offset, limit uint) ([]marketing.Waitlist, uint, error)
}

type geoService interface {
	Locate(ctx context.Context, ip string) string
}

type Service struct {
	repo      Repository
	mail      mailer
	auth      authService
	marketing marketingService
	store     objectstore.Store
	geo       geoService
	log       logger.Logger
	cfg       *config.Env
}

func NewService(repo Repository, mail mailer, auth authService, marketing marketingService, store objectstore.Store, geo geoService, log logger.Logger, cfg *config.Env) *Service {
	return new(Service{
		repo:      repo,
		mail:      mail,
		auth:      auth,
		marketing: marketing,
		store:     store,
		geo:       geo,
		log:       log,
		cfg:       cfg,
	})
}

// truncate and sanitizeIP back sendPasswordChangedAlert below; temporary
// copies of the auth module's helpers until the user domain migrates in
// Fase 5 and the alert moves with it.

// truncate keeps a string within the column limits (ip VARCHAR(45),
// user_agent VARCHAR(255)) so an oversized header can never fail the insert.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen]
}

// sanitizeIP discards anything that isn't a real IP literal. c.IP() is
// fed by a client-influenced header (X-Forwarded-For); a malformed or
// spoofed value must never be shown back to the user in a security alert as
// if it were their real address.
func sanitizeIP(ipAddress string) string {
	if net.ParseIP(strings.TrimSpace(ipAddress)) == nil {
		return ""
	}

	return ipAddress
}

func (s *Service) GetListUsers(ctx context.Context, offset, limit uint) ([]identity.User, uint, error) {
	return s.repo.List(ctx, offset, limit)
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (identity.User, error) {
	return s.repo.GetByID(ctx, id)
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
		return identity.User{}, errors.New("not found user")
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
	if strings.TrimSpace(preferredCurrency) != "" {
		existing.PreferredCurrency = strings.ToUpper(strings.TrimSpace(preferredCurrency))
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
		return identity.User{}, errors.New("failed to read file")
	}

	key := fmt.Sprintf("avatars/%s/avatar", userID.String())

	err = s.store.Put(ctx, key, contentType, data)
	if err != nil {
		return identity.User{}, fmt.Errorf("failed to upload to S3: %w", err)
	}

	imageURL := fmt.Sprintf("%s/users/%s/avatar", s.cfg.PublicURL, userID.String())

	return s.repo.UpdateImage(ctx, userID, imageURL)
}

func (s *Service) GetAvatarFromS3(ctx context.Context, userID uuid.UUID) (io.ReadCloser, string, error) {
	key := fmt.Sprintf("avatars/%s/avatar", userID.String())

	return s.store.Get(ctx, key)
}

func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, currentToken, currentPassword, newPassword, ipAddress, userAgent string) error {
	if err := s.auth.VerifyPassword(ctx, userID, currentPassword); err != nil {
		return err
	}

	if currentPassword == newPassword {
		return errors.New("invalid new password: must differ from current password")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.repo.UpdatePassword(ctx, userID, string(hashed)); err != nil {
		return err
	}

	// Whoever else holds a session (a stolen token, a forgotten shared
	// computer) must not survive a password change: only the session that
	// performed the change stays alive.
	if _, err := s.auth.RevokeOtherSessions(ctx, userID, currentToken); err != nil {
		s.log.Error(ctx, "change password: failed to revoke other sessions", logger.Err(err))
	}

	go s.sendPasswordChangedAlert(userID, ipAddress, userAgent)

	return nil
}

// sendPasswordChangedAlert notifies the user their password changed. Like the
// login alert, it bypasses email preferences: if the change wasn't theirs,
// this email is their only chance to react. Best-effort.
func (s *Service) sendPasswordChangedAlert(userID uuid.UUID, ipAddress, userAgent string) {
	if s.mail == nil {
		return
	}

	ctx := context.Background()

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return
	}

	ipAddress = sanitizeIP(ipAddress)
	location := s.locateIP(ipAddress)
	if location == "" {
		location = "desconocida"
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
		Location:    location,
		When:        time.Now().UTC().Format("02 Jan 2006 15:04 UTC"),
		SecurityURL: s.cfg.FrontendURL + "/dashboard/settings",
	})
}

// locateIP resolves the approximate location of an IP for security alert
// emails. Bounded by its own timeout so a slow lookup can only delay the
// (already asynchronous) email, never the request that triggered it.
func (s *Service) locateIP(ipAddress string) string {
	if s.geo == nil {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.geo.Locate(ctx, ipAddress)
}

func (s *Service) GetUsersWithWeeklySummary(ctx context.Context) ([]identity.User, error) {
	return s.repo.GetWeeklySummary(ctx)
}

// CreateInvitation, ListInvitations, ResendInvitation and RevokeInvitation
// delegate to the auth module: invitations remain auth's data, but the
// admin-facing HTTP surface lives under /users alongside the rest of user
// management.
func (s *Service) CreateInvitation(ctx context.Context, email, name, role string, invitedBy uuid.UUID) (auth.Invitation, error) {
	return s.auth.CreateInvitation(ctx, email, name, role, invitedBy)
}

func (s *Service) ListInvitations(ctx context.Context, offset, limit uint) ([]auth.Invitation, uint, error) {
	return s.auth.ListInvitations(ctx, offset, limit)
}

func (s *Service) ResendInvitation(ctx context.Context, id, invitedBy uuid.UUID) (auth.Invitation, error) {
	return s.auth.ResendInvitation(ctx, id, invitedBy)
}

func (s *Service) RevokeInvitation(ctx context.Context, id uuid.UUID) error {
	return s.auth.RevokeInvitation(ctx, id)
}

// ListWaitlist delegates to the marketing module: the waitlist is marketing's
// data, but admins manage it from the same /users dashboard as invitations.
func (s *Service) ListWaitlist(ctx context.Context, offset, limit uint) ([]marketing.Waitlist, uint, error) {
	return s.marketing.ListWaitlist(ctx, offset, limit)
}
