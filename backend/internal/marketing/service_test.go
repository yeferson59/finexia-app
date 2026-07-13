package marketing

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	saveWaitlistEmail func(ctx context.Context, email string) error
}

func (f *fakeRepository) SaveWaitlistEmail(ctx context.Context, email string) error {
	return f.saveWaitlistEmail(ctx, email)
}

type fakeMailer struct {
	waitlistTo  []string
	waitlistErr error
}

func (m *fakeMailer) SendWaitlistConfirmation(email string) error {
	if m.waitlistErr != nil {
		return m.waitlistErr
	}
	m.waitlistTo = append(m.waitlistTo, email)
	return nil
}

func TestSaveWaitlistEmail(t *testing.T) {
	t.Run("saves and sends the confirmation", func(t *testing.T) {
		var saved string
		repo := &fakeRepository{
			saveWaitlistEmail: func(_ context.Context, email string) error {
				saved = email
				return nil
			},
		}
		mailer := &fakeMailer{}
		svc := NewService(repo, mailer)

		if err := svc.SaveWaitlistEmail(context.Background(), "new@example.com"); err != nil {
			t.Fatalf("SaveWaitlistEmail: %v", err)
		}
		if saved != "new@example.com" {
			t.Errorf("saved = %q", saved)
		}
		if len(mailer.waitlistTo) != 1 || mailer.waitlistTo[0] != "new@example.com" {
			t.Errorf("confirmations = %v", mailer.waitlistTo)
		}
	})

	t.Run("repository failure skips the confirmation email", func(t *testing.T) {
		repo := &fakeRepository{
			saveWaitlistEmail: func(context.Context, string) error {
				return errors.New("duplicate email")
			},
		}
		mailer := &fakeMailer{}
		svc := NewService(repo, mailer)

		if err := svc.SaveWaitlistEmail(context.Background(), "dup@example.com"); err == nil {
			t.Fatal("expected error")
		}
		if len(mailer.waitlistTo) != 0 {
			t.Errorf("no confirmation should be sent when the save fails")
		}
	})

	t.Run("mail failure is surfaced", func(t *testing.T) {
		repo := &fakeRepository{
			saveWaitlistEmail: func(context.Context, string) error { return nil },
		}
		mailer := &fakeMailer{waitlistErr: errors.New("smtp down")}
		svc := NewService(repo, mailer)

		if err := svc.SaveWaitlistEmail(context.Background(), "x@example.com"); err == nil {
			t.Fatal("expected error when the confirmation email fails")
		}
	})
}
