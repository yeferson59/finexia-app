package auth

import "time"

// Config is the auth module's own configuration surface: exactly the settings
// this domain reads, decoupled from the platform-wide *config.Env. The
// composition root (internal/app) populates it from the environment, so the
// module — and its tests — depend on a small, explicit struct instead of the
// full ~35-field Env (see docs/TECH_DEBT.md #8).
type Config struct {
	// JWTSecret signs and verifies access tokens.
	JWTSecret string
	// JWTAccessDuration is the lifetime of an access token.
	JWTAccessDuration time.Duration
	// JWTRefreshDuration is the lifetime of a refresh token (and its cookie).
	JWTRefreshDuration time.Duration
	// RefreshGracePeriod lets a just-rotated refresh token still validate
	// briefly, absorbing races between concurrent refreshes.
	RefreshGracePeriod time.Duration
	// MaxLoginAttempts is how many consecutive failures lock an account.
	MaxLoginAttempts int
	// LoginLockout is how long that lock lasts.
	LoginLockout time.Duration
	// Environment gates production-only behavior (e.g. Secure cookies).
	Environment string
	// FrontendURL is the base URL used to build links in emails.
	FrontendURL string
	// InvitationExpiry is how long an invitation token stays valid.
	InvitationExpiry time.Duration
	// PasswordResetExpiry is how long a password-reset token stays valid.
	PasswordResetExpiry time.Duration
	// EmailVerificationExpiry is how long an email-verification token stays valid.
	EmailVerificationExpiry time.Duration
	// SelfRegistrationEnabled opens public registration; off during the beta.
	SelfRegistrationEnabled bool
	// TwoFactorPendingExpiry is how long a password-validated login may wait
	// for its TOTP code before the user must start over.
	TwoFactorPendingExpiry time.Duration
}
