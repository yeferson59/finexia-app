package entities

import (
	"time"

	"github.com/google/uuid"
)

type WaitlistStatus string

const (
	WaitlistStatusPending  WaitlistStatus = "pending"
	WaitlistStatusAccepted WaitlistStatus = "invited"
	WaitlistStatusRejected WaitlistStatus = "registered"
)

type Waitlist struct {
	ID        uuid.UUID      `json:"id"`
	Email     string         `json:"email"`
	Status    WaitlistStatus `json:"status"`
	InvitedAt time.Time      `json:"invitedAt"`
	CreatedAt time.Time      `json:"createdAt"`
}
