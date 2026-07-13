package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/mail"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// Exported so handlers can map each failure to a precise HTTP status and
// message instead of pattern-matching error strings.
var (
	ErrPasswordResetInvalid = errors.New("invalid password reset link")
	ErrPasswordResetExpired = errors.New("password reset link expired")
)

// RequestPasswordReset issues a single-use reset link and emails it to the
// account, if one exists for the address. The response is intentionally the
// same regardless of whether the email is registered, so the endpoint never
// confirms which addresses have an account.
func (s *Services) RequestPasswordReset(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil
	}

	user, err := s.repos.GetUserByEmail(ctx, email)
	if err != nil {
		return nil
	}

	raw, hash, err := generateRefreshToken()
	if err != nil {
		return err
	}

	expiresAt := time.Now().UTC().Add(s.cfg.PasswordResetExpiry)

	pr, err := s.repos.CreatePasswordReset(ctx, user.ID, hash, expiresAt)
	if err != nil {
		return err
	}

	s.sendPasswordResetEmail(ctx, user.Name, user.Email, raw, pr.ExpiresAt)

	return nil
}

// ValidatePasswordResetToken reports whether a token is still redeemable, so
// the reset page can reject dead links before asking for a new password.
func (s *Services) ValidatePasswordResetToken(ctx context.Context, rawToken string) error {
	_, err := s.lookupPasswordReset(ctx, rawToken)
	return err
}

// ResetPassword consumes a valid reset token and sets the new password. Every
// existing session is revoked afterward: proving control of the inbox is not
// the same as proving control of any device that was already logged in.
func (s *Services) ResetPassword(ctx context.Context, rawToken, newPassword, ipAddress, userAgent string) error {
	pr, err := s.lookupPasswordReset(ctx, rawToken)
	if err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.repos.ConsumePasswordReset(ctx, pr.ID, pr.UserID, string(hashed)); err != nil {
		return ErrPasswordResetInvalid
	}

	if _, err := s.RevokeOtherSessions(ctx, pr.UserID, ""); err != nil {
		s.log.Error(ctx, "reset password: failed to revoke sessions", logger.Err(err))
	}

	go s.sendPasswordChangedAlert(pr.UserID, ipAddress, userAgent)

	return nil
}

// lookupPasswordReset hashes the raw token, fetches the reset row, and
// rejects anything not currently redeemable. The hash comparison happens in
// the database via the unique token_hash column, so the raw token is never
// stored.
func (s *Services) lookupPasswordReset(ctx context.Context, rawToken string) (entities.PasswordReset, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return entities.PasswordReset{}, ErrPasswordResetInvalid
	}

	hash, err := hashRefreshToken(rawToken)
	if err != nil {
		return entities.PasswordReset{}, ErrPasswordResetInvalid
	}

	pr, err := s.repos.GetPasswordResetByHash(ctx, hash)
	if err != nil {
		return entities.PasswordReset{}, ErrPasswordResetInvalid
	}

	if pr.UsedAt != nil {
		return entities.PasswordReset{}, ErrPasswordResetInvalid
	}
	if time.Now().UTC().After(pr.ExpiresAt) {
		return entities.PasswordReset{}, ErrPasswordResetExpired
	}

	return pr, nil
}

// sendPasswordResetEmail delivers the reset link. Best-effort and async: a
// mail hiccup must not fail the request, and the user can always ask again.
func (s *Services) sendPasswordResetEmail(ctx context.Context, userName, email, rawToken string, expiresAt time.Time) {
	if s.mail == nil {
		return
	}

	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s",
		strings.TrimRight(s.cfg.FrontendURL, "/"), url.QueryEscape(rawToken))

	data := mail.PasswordResetData{
		UserName:  userName,
		ResetURL:  resetURL,
		ExpiresIn: humanizeExpiry(time.Until(expiresAt)),
	}

	go func() {
		if err := s.mail.SendPasswordReset(email, data); err != nil {
			s.log.Error(ctx, "failed to send password reset email", logger.Err(err))
		}
	}()
}
