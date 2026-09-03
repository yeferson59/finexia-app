package marketing

import (
	"context"

	"uuid"
)

// Mailer declares the single mail capability this module consumes; the
// platform mail service satisfies it.
type mailer interface {
	SendWaitlistConfirmation(email string) error
}

// Service holds the module's use cases.
type service struct {
	repo Repository
	mail mailer
}

// newService is the fake-friendly constructor: the module's NewService wraps
// it with the real Postgres repository.
func newService(repo Repository, mail mailer) *service {
	return new(service{repo: repo, mail: mail})
}

func (s *service) SaveWaitlistEmail(ctx context.Context, email string) error {
	if err := s.repo.SaveWaitlistEmail(ctx, email); err != nil {
		return err
	}

	return s.mail.SendWaitlistConfirmation(email)
}

// ListWaitlist returns the funnel for the admin dashboard. Consumed by the
// auth module's invitation flow through its own WaitlistStore interface.
func (s *service) ListWaitlist(ctx context.Context, offset, limit uint) ([]Waitlist, uint, error) {
	return s.repo.ListWaitlist(ctx, offset, limit)
}

// SetWaitlistInvited advances a waitlist row to "invited". Consumed by the
// auth module when an admin issues an invitation.
func (s *service) SetWaitlistInvited(ctx context.Context, email string) error {
	return s.repo.SetWaitlistInvited(ctx, email)
}

// DeleteWaitlist removes an entry from the waitlist. It is the admin's way of
// clearing a typo or a duplicate sign-up out of the funnel; the email can
// register again afterwards, since the row is what holds the unique address.
func (s *service) DeleteWaitlist(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteWaitlist(ctx, id)
}
