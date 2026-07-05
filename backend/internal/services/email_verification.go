package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/logger"
	"github.com/yeferson59/finexia-app/internal/mail"
)

// Exported so handlers can map each failure to a precise HTTP status and
// message instead of pattern-matching error strings.
var (
	ErrEmailVerificationInvalid = errors.New("invalid email verification link")
	ErrEmailVerificationExpired = errors.New("email verification link expired")
)

// issueEmailVerification creates a single-use verification token for the
// email and emails it. Errors are logged, not returned: it runs right after
// account creation, and a mail hiccup must not fail the registration since the
// user can always request a new link.
func (s *Services) issueEmailVerification(ctx context.Context, name, email string) {
	raw, hash, err := generateRefreshToken()
	if err != nil {
		s.log.Error("email verification: failed to generate token", logger.Err(err))
		return
	}

	expiresAt := time.Now().UTC().Add(s.cfg.EmailVerificationExpiry)

	v, err := s.repos.CreateEmailVerification(ctx, email, hash, expiresAt)
	if err != nil {
		s.log.Error("email verification: failed to create token", logger.Err(err))
		return
	}

	s.sendEmailVerificationMail(name, email, raw, v.ExpiresAt)
}

// RequestEmailVerification (re)issues a verification link for an email that
// is registered but not yet verified. The response is intentionally the same
// regardless of whether the email exists or is already verified, so the
// endpoint never confirms which addresses are registered.
func (s *Services) RequestEmailVerification(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil
	}

	user, err := s.repos.GetUserByEmail(ctx, email)
	if err != nil || user.EmailVerified {
		return nil
	}

	s.issueEmailVerification(ctx, user.Name, user.Email)

	return nil
}

// ValidateEmailVerification reports whether a token is still redeemable, so
// the verify page can reject dead links before confirming.
func (s *Services) ValidateEmailVerification(ctx context.Context, rawToken string) error {
	_, err := s.lookupEmailVerification(ctx, rawToken)
	return err
}

// VerifyEmail consumes a valid verification token and marks the account's
// email as verified.
func (s *Services) VerifyEmail(ctx context.Context, rawToken string) error {
	v, err := s.lookupEmailVerification(ctx, rawToken)
	if err != nil {
		return err
	}

	if err := s.repos.ConsumeEmailVerification(ctx, v.ID, v.Identifier); err != nil {
		return ErrEmailVerificationInvalid
	}

	return nil
}

// lookupEmailVerification hashes the raw token, fetches the verification row,
// and rejects anything not currently redeemable. The hash comparison happens
// in the database via the value column, so the raw token is never stored.
func (s *Services) lookupEmailVerification(ctx context.Context, rawToken string) (entities.Verification, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return entities.Verification{}, ErrEmailVerificationInvalid
	}

	hash, err := hashRefreshToken(rawToken)
	if err != nil {
		return entities.Verification{}, ErrEmailVerificationInvalid
	}

	v, err := s.repos.GetEmailVerificationByHash(ctx, hash)
	if err != nil {
		return entities.Verification{}, ErrEmailVerificationInvalid
	}

	if time.Now().UTC().After(v.ExpiresAt) {
		return entities.Verification{}, ErrEmailVerificationExpired
	}

	return v, nil
}

// sendEmailVerificationMail delivers the verification link. Best-effort and
// async: a mail hiccup must not fail the caller, and a new link can always be
// requested.
func (s *Services) sendEmailVerificationMail(userName, email, rawToken string, expiresAt time.Time) {
	if s.mail == nil {
		return
	}

	verifyURL := fmt.Sprintf("%s/auth/verify-email?token=%s",
		strings.TrimRight(s.cfg.FrontendURL, "/"), url.QueryEscape(rawToken))

	data := mail.EmailVerificationData{
		UserName:  userName,
		VerifyURL: verifyURL,
		ExpiresIn: humanizeExpiry(time.Until(expiresAt)),
	}

	go func() {
		if err := s.mail.SendEmailVerification(email, data); err != nil {
			s.log.Error("failed to send email verification", logger.Err(err))
		}
	}()
}
