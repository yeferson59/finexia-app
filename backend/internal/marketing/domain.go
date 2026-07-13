// Package marketing is the first domain module of the modular-monolith
// migration (Fase 2 pilot): it owns the public waitlist sign-up.
package marketing

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
	InvitedAt *time.Time     `json:"invitedAt,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}
