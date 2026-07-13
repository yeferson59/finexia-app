package marketing

import "context"

// Repository declares only what this module needs from persistence; the
// consumer defines the interface, the postgres implementation satisfies it.
type Repository interface {
	SaveWaitlistEmail(ctx context.Context, email string) error
}
