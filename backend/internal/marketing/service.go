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
	return &Service{repo: repo, mail: mail}
}

func (s *Service) SaveWaitlistEmail(ctx context.Context, email string) error {
	if err := s.repo.SaveWaitlistEmail(ctx, email); err != nil {
		return err
	}

	return s.mail.SendWaitlistConfirmation(email)
}
