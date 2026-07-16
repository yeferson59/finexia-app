// Package auth is the authentication domain module: login, sessions, refresh
// tokens, two-factor authentication and email verification (password reset
// and invitations join it in later PRs of Fase 4). It follows the module
// pattern validated by the marketing pilot: consumer-defined interfaces, a
// single Postgres implementation, and an HTTP surface registered through
// Module.Routes.
package auth

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrAccountUnverified signals a login attempt against an account whose email
// has not been verified yet, letting the handler point the client at the
// resend-verification flow instead of a generic credentials error.
var ErrAccountUnverified = errors.New("invalid account")

// ErrEmailAlreadyExists signals a registration attempt with an email that is
// already tied to an account. Unlike password reset / email verification
// requests (which never confirm whether an address exists), registration is
// already an oracle by nature: the caller is asserting they own the address
// and attempting to create an account with it, so returning a precise
// message here reveals nothing an attacker couldn't already infer from the
// request itself, and the endpoint stays behind the same rate limiter as
// login.
var ErrEmailAlreadyExists = errors.New("email already exists")

// Exported so the login handler can distinguish "password OK, now ask for the
// TOTP code" from a plain failure, and the 2FA handlers can map each case to
// a precise HTTP response.
var (
	ErrTwoFactorRequired       = errors.New("two-factor authentication required")
	ErrTwoFactorInvalidCode    = errors.New("invalid two-factor code")
	ErrTwoFactorPendingInvalid = errors.New("invalid or expired two-factor session")
	ErrTwoFactorAlreadyEnabled = errors.New("two-factor already enabled")
	ErrTwoFactorNotEnabled     = errors.New("two-factor not enabled")
	ErrTwoFactorSetupMissing   = errors.New("two-factor setup not started")
)

// Exported so handlers can map each failure to a precise HTTP status and
// message instead of pattern-matching error strings.
var (
	ErrEmailVerificationInvalid = errors.New("invalid email verification link")
	ErrEmailVerificationExpired = errors.New("email verification link expired")
)

// Exported so handlers can map each failure to a precise HTTP status and
// message instead of pattern-matching error strings.
var (
	ErrPasswordResetInvalid = errors.New("invalid password reset link")
	ErrPasswordResetExpired = errors.New("password reset link expired")
)

// TwoFactor holds a user's TOTP enrollment. A row with Enabled=false is a
// pending setup: the secret was issued but the user has not yet confirmed a
// code, so login is NOT gated until the enrollment is confirmed.
type TwoFactor struct {
	UserID      uuid.UUID  `json:"userId"`
	Secret      string     `json:"-"`
	Enabled     bool       `json:"enabled"`
	ConfirmedAt *time.Time `json:"confirmedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type Verification struct {
	ID         uuid.UUID `json:"id"`
	Identifier string    `json:"identifier"`
	Value      string    `json:"value"`
	ExpiresAt  time.Time `json:"expiresAt"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type PasswordReset struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"userId"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

// Exported so handlers can map each failure to a precise HTTP status and
// message instead of pattern-matching error strings.
var (
	ErrInvitationInvalid = errors.New("invalid invitation")
	ErrInvitationExpired = errors.New("invitation expired")
)

// Invitation is a single-use, expiring grant that lets an admin bring a new
// person into the app. Only the SHA-256 hash of the token is stored; the raw
// token travels solely in the emailed link, so a database leak cannot be used
// to accept an invitation.
type Invitation struct {
	ID         uuid.UUID  `json:"id"`
	Email      string     `json:"email"`
	Name       string     `json:"name"`
	Role       string     `json:"role"`
	TokenHash  string     `json:"-"`
	InvitedBy  *uuid.UUID `json:"-"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	AcceptedAt *time.Time `json:"acceptedAt,omitempty"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

// Status derives the lifecycle state shown in the admin dashboard from the
// timestamp columns, so the API never has to store a redundant status field
// that could drift from the source-of-truth timestamps.
func (i Invitation) Status() string {
	switch {
	case i.AcceptedAt != nil:
		return "accepted"
	case i.RevokedAt != nil:
		return "revoked"
	case time.Now().UTC().After(i.ExpiresAt):
		return "expired"
	default:
		return "pending"
	}
}

// MarshalJSON augments the serialized invitation with the derived status so the
// admin dashboard can render lifecycle badges without recomputing the rules.
func (i Invitation) MarshalJSON() ([]byte, error) {
	type alias Invitation
	return json.Marshal(struct {
		alias
		Status string `json:"status"`
	}{alias(i), i.Status()})
}
