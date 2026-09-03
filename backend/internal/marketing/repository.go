package marketing

import (
	"context"

	"uuid"
)

// Repository declares only what this module needs from persistence; the
// consumer defines the interface, the postgres implementation satisfies it.
type Repository interface {
	SaveWaitlistEmail(ctx context.Context, email string) error
	ListWaitlist(ctx context.Context, offset, limit uint) ([]Waitlist, uint, error)
	SetWaitlistInvited(ctx context.Context, email string) error
	DeleteWaitlist(ctx context.Context, id uuid.UUID) error
}
