package user

import (
	"time"

	"github.com/google/uuid"
)

type UserPreferences struct {
	UserID        uuid.UUID `json:"userId"`
	EmailAlerts   bool      `json:"emailAlerts"`
	WeeklySummary bool      `json:"weeklySummary"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
