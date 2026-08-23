package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

func TestUpdateCurrentUser(t *testing.T) {
	userID := uuid.New()
	existing := identity.User{
		ID:                userID,
		Name:              "Old Name",
		Email:             "user@example.com",
		PreferredCurrency: "USD",
		Image:             "old.png",
	}

	newSvc := func(t *testing.T) (*Service, *identity.User) {
		t.Helper()
		saved := new(identity.User{})

		repo := new(fakeRepository{
			getUserByID: func(context.Context, uuid.UUID) (identity.User, error) {
				return existing, nil
			},
			updateUserProfile: func(_ context.Context, id uuid.UUID, name, preferredCurrency, image string) (identity.User, error) {
				*saved = identity.User{ID: id, Name: name, PreferredCurrency: preferredCurrency, Image: image}

				return *saved, nil
			},
		})
		return newService(repo, nil, logger.Noop(), Config{}), saved
	}

	t.Run("normalizes name and currency", func(t *testing.T) {
		svc, saved := newSvc(t)

		_, err := svc.UpdateCurrentUser(context.Background(), userID, "  jane DOE ", " eur ", "new.png")
		if err != nil {
			t.Fatalf("UpdateCurrentUser: %v", err)
		}
		if saved.Name != "Jane Doe" {
			t.Errorf("saved name = %q, want %q", saved.Name, "Jane Doe")
		}
		if saved.PreferredCurrency != "EUR" {
			t.Errorf("saved currency = %q, want EUR", saved.PreferredCurrency)
		}
		if saved.Image != "new.png" {
			t.Errorf("saved image = %q, want new.png", saved.Image)
		}
	})

	// The preferred currency drives what the dashboard converts totals into and
	// which pairs the market-data sync fetches, so a code with no rate behind it
	// has to be refused here rather than stored and quietly ignored later.
	t.Run("rejects a currency the app cannot convert to", func(t *testing.T) {
		svc, saved := newSvc(t)

		_, err := svc.UpdateCurrentUser(context.Background(), userID, "Jane", "ARS", "")
		if !errors.Is(err, ErrUnsupportedCurrency) {
			t.Fatalf("err = %v, want ErrUnsupportedCurrency", err)
		}
		if saved.PreferredCurrency != "" {
			t.Errorf("saved currency = %q, want nothing written", saved.PreferredCurrency)
		}
	})

	t.Run("blank fields keep existing values", func(t *testing.T) {
		svc, saved := newSvc(t)

		_, err := svc.UpdateCurrentUser(context.Background(), userID, "   ", "", "")
		if err != nil {
			t.Fatalf("UpdateCurrentUser: %v", err)
		}
		if saved.Name != existing.Name {
			t.Errorf("saved name = %q, want existing %q", saved.Name, existing.Name)
		}
		if saved.PreferredCurrency != existing.PreferredCurrency {
			t.Errorf("saved currency = %q, want existing %q", saved.PreferredCurrency, existing.PreferredCurrency)
		}
		if saved.Image != existing.Image {
			t.Errorf("saved image = %q, want existing %q", saved.Image, existing.Image)
		}
	})
}

func TestUpdateUserRejectsDeletedUser(t *testing.T) {
	repo := new(fakeRepository{
		getUserByID: func(context.Context, uuid.UUID) (identity.User, error) {
			return identity.User{ID: uuid.New(), DeletedAt: new(time.Now())}, nil
		},
	})

	svc := newService(repo, nil, logger.Noop(), Config{})

	_, err := svc.UpdateUser(context.Background(), uuid.New(), "Name", "mail@example.com", "")
	if err == nil || err.Error() != "not found user" {
		t.Fatalf("UpdateUser error = %v, want %q", err, "not found user")
	}
}
