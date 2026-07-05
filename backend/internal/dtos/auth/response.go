package auth

import (
	"time"

	"github.com/google/uuid"
)

type RegisterResponseDTO struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Image string `json:"image"`
}

type LoginResponseDTO struct {
	ID          uuid.UUID `json:"id"`
	AccessToken string    `json:"accessToken"`
}

type UserResponseDTO struct {
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	EmailVerified     bool      `json:"emailVerified"`
	Image             string    `json:"image"`
	Role              string    `json:"role"`
	PreferredCurrency string    `json:"preferredCurrency"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type SessionResponseDTO struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
	IPAddress *string   `json:"ipAddress"`
	UserAgent *string   `json:"userAgent"`
	CreatedAt time.Time `json:"createdAt"`
}

// ActiveSessionDTO describes a live session for the "active devices" list.
// It deliberately excludes the session token.
type ActiveSessionDTO struct {
	ID           uuid.UUID `json:"id"`
	IPAddress    *string   `json:"ipAddress"`
	UserAgent    *string   `json:"userAgent"`
	CreatedAt    time.Time `json:"createdAt"`
	LastActiveAt time.Time `json:"lastActiveAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
	Current      bool      `json:"current"`
}

type UserSessionResponseDTO struct {
	User    UserResponseDTO    `json:"user"`
	Session SessionResponseDTO `json:"session"`
}

// LoginInternalDTO is used only between service and handler; never serialized to JSON.
type LoginInternalDTO struct {
	ID               uuid.UUID
	AccessToken      string
	RawRefreshToken  string
	RefreshExpiresAt time.Time
}
