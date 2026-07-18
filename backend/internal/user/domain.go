package user

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
	Users       []User     `json:"users,omitempty"`
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
}

type UserPreferences struct {
	UserID        uuid.UUID `json:"userId"`
	EmailAlerts   bool      `json:"emailAlerts"`
	WeeklySummary bool      `json:"weeklySummary"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
