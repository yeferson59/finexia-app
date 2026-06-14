package services

import "context"

func (s Services) SaveWaitlistEmail(ctx context.Context, email string) error {
	err := s.repos.SaveWaitlistEmail(ctx, email)
	if err != nil {
		return err
	}

	return s.mail.SendWaitlistConfirmation(email)
}
