package entities

import (
	"time"

	"github.com/google/uuid"
)

// PasswordReset is the last auth entity still living here: it migrates to the
// auth module together with the password-reset flow (Fase 4, PR B), deleting
// this file. Session, Account, RefreshToken and the 2FA types now live in
// internal/identity and internal/auth.
type PasswordReset struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"userId"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}
