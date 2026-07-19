package marketing

import "context"

// Mailer declares the single mail capability this module consumes; the
// platform mail service satisfies it.
type Mailer interface {
	SendWaitlistConfirmation(email string) error
}

// Service holds the module's use cases.
type Service struct {
	repo Repository
	mail Mailer
}

func NewService(repo Repository, mail Mailer) *Service {
	return new(Service{repo: repo, mail: mail})
}

func (s *Service) SaveWaitlistEmail(ctx context.Context, email string) error {
	if err := s.repo.SaveWaitlistEmail(ctx, email); err != nil {
		return err
	}

	return s.mail.SendWaitlistConfirmation(email)
}

// ListWaitlist returns the funnel for the admin dashboard. Consumed by the
// auth module's invitation flow through its own WaitlistStore interface.
func (s *Service) ListWaitlist(ctx context.Context, offset, limit uint) ([]Waitlist, uint, error) {
	return s.repo.ListWaitlist(ctx, offset, limit)
}

// SetWaitlistInvited advances a waitlist row to "invited". Consumed by the
// auth module when an admin issues an invitation.
func (s *Service) SetWaitlistInvited(ctx context.Context, email string) error {
	return s.repo.SetWaitlistInvited(ctx, email)
}
