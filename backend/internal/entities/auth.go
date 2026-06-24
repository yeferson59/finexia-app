package entities

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Token     string    `json:"-"`
	ExpiresAt time.Time `json:"expiresAt"`
	IPAddress *string   `json:"ipAddress"`
	UserAgent *string   `json:"userAgent"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	User      User      `json:"user,omitzero"`
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
	User                  User      `json:"user,omitzero"`
}

func (a *Account) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
}

type Verification struct {
	ID         uuid.UUID `json:"id"`
	Identifier string    `json:"identifier"`
	Value      string    `json:"value"`
	ExpiresAt  time.Time `json:"expiresAt"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
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
