package entities

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Invitation is a single-use, expiring grant that lets an admin bring a new
// person into the app. Only the SHA-256 hash of the token is stored; the raw
// token travels solely in the emailed link, so a database leak cannot be used
// to accept an invitation.
type Invitation struct {
	ID         uuid.UUID  `json:"id"`
	Email      string     `json:"email"`
	Name       string     `json:"name"`
	Role       string     `json:"role"`
	TokenHash  string     `json:"-"`
	InvitedBy  *uuid.UUID `json:"-"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	AcceptedAt *time.Time `json:"acceptedAt,omitempty"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

// Status derives the lifecycle state shown in the admin dashboard from the
// timestamp columns, so the API never has to store a redundant status field
// that could drift from the source-of-truth timestamps.
func (i Invitation) Status() string {
	switch {
	case i.AcceptedAt != nil:
		return "accepted"
	case i.RevokedAt != nil:
		return "revoked"
	case time.Now().UTC().After(i.ExpiresAt):
		return "expired"
	default:
		return "pending"
	}
}

// MarshalJSON augments the serialized invitation with the derived status so the
// admin dashboard can render lifecycle badges without recomputing the rules.
func (i Invitation) MarshalJSON() ([]byte, error) {
	type alias Invitation
	return json.Marshal(struct {
		alias
		Status string `json:"status"`
	}{alias(i), i.Status()})
}
