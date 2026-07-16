// Package identity holds the data-only structs shared by the domain modules
// that deal with people and their credentials (auth, user, portfolio,
// notification). It is the single sanctioned exception to the
// one-domain-per-module rule: no behavior lives here, only types.
package identity

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"-"`
}

type User struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	Email             string     `json:"email"`
	EmailVerified     bool       `json:"emailVerified"`
	Image             string     `json:"image"`
	RoleID            uuid.UUID  `json:"-"`
	PreferredCurrency string     `json:"preferredCurrency"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	DeletedAt         *time.Time `json:"-"`
	BannedAt          *time.Time `json:"bannedAt,omitempty"`
	Role              Role       `json:"role,omitzero"`
	Sessions          []Session  `json:"sessions,omitempty"`
	Accounts          []Account  `json:"accounts,omitempty"`
}

type Session struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Token     string    `json:"-"`
	ExpiresAt time.Time `json:"expiresAt"`
	IPAddress *string   `json:"ipAddress"`
	UserAgent *string   `json:"userAgent"`
	Location  *string   `json:"location"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Account struct {
	ID                    uuid.UUID `json:"id"`
	UserID                uuid.UUID `json:"userId"`
	AccountID             string    `json:"accountId"`
	ProviderID            string    `json:"provider"`
	AccessToken           string    `json:"accessToken"`
	RefreshToken          string    `json:"refreshToken"`
	AccessTokenExpiresAt  time.Time `json:"accessTokenExpiresAt"`
	RefreshTokenExpiresAt time.Time `json:"refreshTokenExpiresAt"`
	Scope                 string    `json:"scope"`
	IDToken               string    `json:"idToken"`
	Password              string    `json:"-"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

type RefreshToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"-"`
	TokenHash string     `json:"-"`
	FamilyID  uuid.UUID  `json:"-"`
	SessionID uuid.UUID  `json:"-"`
	IPAddress *string    `json:"-"`
	UserAgent *string    `json:"-"`
	ExpiresAt time.Time  `json:"expiresAt"`
	UsedAt    *time.Time `json:"-"`
	RevokedAt *time.Time `json:"-"`
	CreatedAt time.Time  `json:"createdAt"`
	Role      string     `json:"-"` // populated from JOIN, not a DB column
}
