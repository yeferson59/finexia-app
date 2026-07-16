package marketing

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	saveWaitlistEmail  func(ctx context.Context, email string) error
	listWaitlist       func(ctx context.Context, offset, limit uint) ([]Waitlist, uint, error)
	setWaitlistInvited func(ctx context.Context, email string) error
}

func (f *fakeRepository) SaveWaitlistEmail(ctx context.Context, email string) error {
	return f.saveWaitlistEmail(ctx, email)
}

func (f *fakeRepository) ListWaitlist(ctx context.Context, offset, limit uint) ([]Waitlist, uint, error) {
	return f.listWaitlist(ctx, offset, limit)
}

func (f *fakeRepository) SetWaitlistInvited(ctx context.Context, email string) error {
	return f.setWaitlistInvited(ctx, email)
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

func TestWaitlistAdminPassthroughs(t *testing.T) {
	t.Run("ListWaitlist delegates with the same window", func(t *testing.T) {
		var gotOffset, gotLimit uint
		repo := &fakeRepository{
			listWaitlist: func(_ context.Context, offset, limit uint) ([]Waitlist, uint, error) {
				gotOffset, gotLimit = offset, limit
				return []Waitlist{{Email: "a@example.com"}}, 1, nil
			},
		}
		svc := NewService(repo, &fakeMailer{})

		items, count, err := svc.ListWaitlist(context.Background(), 20, 10)
		if err != nil {
			t.Fatalf("ListWaitlist: %v", err)
		}
		if gotOffset != 20 || gotLimit != 10 {
			t.Errorf("window = (%d, %d), want (20, 10)", gotOffset, gotLimit)
		}
		if count != 1 || len(items) != 1 || items[0].Email != "a@example.com" {
			t.Errorf("items = %v, count = %d", items, count)
		}
	})

	t.Run("SetWaitlistInvited delegates the email", func(t *testing.T) {
		var got string
		repo := &fakeRepository{
			setWaitlistInvited: func(_ context.Context, email string) error {
				got = email
				return nil
			},
		}
		svc := NewService(repo, &fakeMailer{})

		if err := svc.SetWaitlistInvited(context.Background(), "b@example.com"); err != nil {
			t.Fatalf("SetWaitlistInvited: %v", err)
		}
		if got != "b@example.com" {
			t.Errorf("email = %q, want b@example.com", got)
		}
	})
}
