package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
)

// The persistence surface is split into per-sub-area stores (consumer-defined
// interfaces, none above ~11 methods) instead of one god interface. A single
// *PostgresRepository implements all of them; tests fake only the stores they
// touch.

// AccountStore covers account/user lookups and registration. GetUserByID and
// GetUserByEmail belong to the user domain's tables but are consumed by auth
// (2FA alerts, registration duplicate check); they migrate to interfaces over
// the user module in Fase 5.
type AccountStore interface {
	GetAccountByUserID(ctx context.Context, userID uuid.UUID) (identity.Account, error)
	GetAccountByEmail(ctx context.Context, email string) (identity.User, error)
	Register(ctx context.Context, name, email, password string) (identity.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (identity.User, error)
	GetUserByEmail(ctx context.Context, email string) (identity.User, error)
}

type SessionStore interface {
	CreateSession(ctx context.Context, userID uuid.UUID, token string, ip, ua *string, expiresAt time.Time) (uuid.UUID, error)
	UpdateSessionToken(ctx context.Context, sessionID uuid.UUID, newToken string, expiresAt time.Time) (string, error)
	UpdateSessionLocation(ctx context.Context, sessionID uuid.UUID, location string) error
	ListSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]identity.Session, error)
	DeleteSessionsByIDs(ctx context.Context, userID uuid.UUID, sessionIDs []uuid.UUID) (int64, error)
	GetSessionByUserIDToken(ctx context.Context, userID uuid.UUID, token string) (identity.User, error)
	GetSessionByToken(ctx context.Context, token string) (identity.User, error)
	DeleteSessionByUserIDToken(ctx context.Context, userID uuid.UUID, token string) error
	DeleteExpiredSessions(ctx context.Context) (int64, error)
	HasKnownLoginIP(ctx context.Context, userID uuid.UUID, ip string) (bool, error)
	RecordKnownLoginIP(ctx context.Context, userID uuid.UUID, ip string) error
}

type RefreshTokenStore interface {
	CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, familyID, sessionID uuid.UUID, ip, ua *string, expiresAt time.Time) (uuid.UUID, error)
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (identity.RefreshToken, error)
	MarkRefreshTokenUsed(ctx context.Context, id uuid.UUID) error
	RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) ([]string, error)
	GetRefreshTokenFamiliesBySession(ctx context.Context, userID uuid.UUID, sessionToken string) ([]string, []uuid.UUID, error)
	GetRefreshTokensBySessionIDs(ctx context.Context, userID uuid.UUID, sessionIDs []uuid.UUID) ([]string, []uuid.UUID, error)
	DeleteExpiredRefreshTokens(ctx context.Context) (int64, error)
}

type TwoFactorStore interface {
	GetTwoFactor(ctx context.Context, userID uuid.UUID) (TwoFactor, error)
	UpsertTwoFactorSecret(ctx context.Context, userID uuid.UUID, secret string) error
	EnableTwoFactor(ctx context.Context, userID uuid.UUID) error
	DeleteTwoFactor(ctx context.Context, userID uuid.UUID) error
	ReplaceTwoFactorRecoveryCodes(ctx context.Context, userID uuid.UUID, codeHashes []string) error
	ConsumeTwoFactorRecoveryCode(ctx context.Context, userID uuid.UUID, codeHash string) error
	CountUnusedTwoFactorRecoveryCodes(ctx context.Context, userID uuid.UUID) (int, error)
}

type VerificationStore interface {
	CreateEmailVerification(ctx context.Context, email, tokenHash string, expiresAt time.Time) (Verification, error)
	GetEmailVerificationByHash(ctx context.Context, tokenHash string) (Verification, error)
	ConsumeEmailVerification(ctx context.Context, id uuid.UUID, email string) error
}

type PasswordResetStore interface {
	CreatePasswordReset(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (PasswordReset, error)
	GetPasswordResetByHash(ctx context.Context, tokenHash string) (PasswordReset, error)
	ConsumePasswordReset(ctx context.Context, resetID, userID uuid.UUID, hashedPassword string) error
}

// Stores groups the module's persistence interfaces. The composition root
// fills every field with the same *PostgresRepository; tests fill only what
// the case under test touches (a nil store panics loudly, mirroring the old
// embed-the-interface fakes).
type Stores struct {
	Accounts       AccountStore
	Sessions       SessionStore
	RefreshTokens  RefreshTokenStore
	TwoFactor      TwoFactorStore
	Verifications  VerificationStore
	PasswordResets PasswordResetStore
}

// Mailer abstracts the outbound email service so tests can replace the
// Resend-backed implementation with a fake. It declares only what auth sends.
type Mailer interface {
	SendSecurityAlert(email string, data mail.SecurityAlertData) error
	SendEmailVerification(email string, data mail.EmailVerificationData) error
	SendPasswordReset(email string, data mail.PasswordResetData) error
}

var _ Mailer = (*mail.Service)(nil)

// GeoLocator resolves an IP address to a human-readable approximate location
// for security notifications. Implementations return "" when the location is
// unknown (private IP, lookup failure), never an error.
type GeoLocator interface {
	Locate(ctx context.Context, ip string) string
}
