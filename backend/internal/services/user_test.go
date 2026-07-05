package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/entities"
)

func TestChangePassword(t *testing.T) {
	userID := uuid.New()
	const currentPassword = "current-password"

	newRepo := func(t *testing.T) (*fakeRepository, *string) {
		t.Helper()
		var storedHash string
		repo := &fakeRepository{
			getAccountByUserID: func(context.Context, uuid.UUID) (entities.Account, error) {
				return entities.Account{Password: mustHashPassword(t, currentPassword)}, nil
			},
			updateUserPassword: func(_ context.Context, uid uuid.UUID, hashed string) error {
				if uid != userID {
					t.Errorf("UpdateUserPassword userID = %s, want %s", uid, userID)
				}
				storedHash = hashed
				return nil
			},
			listSessionsByUserID: func(context.Context, uuid.UUID) ([]entities.Session, error) {
				return nil, nil
			},
		}
		return repo, &storedHash
	}

	t.Run("success stores a new bcrypt hash", func(t *testing.T) {
		repo, storedHash := newRepo(t)
		svc := newTestServices(repo, newMemStorage())

		if err := svc.ChangePassword(context.Background(), userID, "current-token", currentPassword, "new-password"); err != nil {
			t.Fatalf("ChangePassword: %v", err)
		}
		if *storedHash == "" {
			t.Fatal("expected a new password hash to be stored")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(*storedHash), []byte("new-password")); err != nil {
			t.Errorf("stored hash does not match the new password: %v", err)
		}
	})

	t.Run("wrong current password is rejected", func(t *testing.T) {
		repo, storedHash := newRepo(t)
		svc := newTestServices(repo, newMemStorage())

		err := svc.ChangePassword(context.Background(), userID, "current-token", "not-the-password", "new-password")
		if err == nil || err.Error() != "invalid current password" {
			t.Fatalf("error = %v, want %q", err, "invalid current password")
		}
		if *storedHash != "" {
			t.Error("password must not change when the current password is wrong")
		}
	})

	t.Run("new password equal to current is rejected", func(t *testing.T) {
		repo, storedHash := newRepo(t)
		svc := newTestServices(repo, newMemStorage())

		if err := svc.ChangePassword(context.Background(), userID, "current-token", currentPassword, currentPassword); err == nil {
			t.Fatal("expected reusing the current password to be rejected")
		}
		if *storedHash != "" {
			t.Error("password must not be rewritten when unchanged")
		}
	})

	t.Run("missing account", func(t *testing.T) {
		repo := &fakeRepository{
			getAccountByUserID: func(context.Context, uuid.UUID) (entities.Account, error) {
				return entities.Account{}, errors.New("no rows")
			},
		}
		svc := newTestServices(repo, newMemStorage())

		if err := svc.ChangePassword(context.Background(), userID, "current-token", currentPassword, "new-password"); err == nil {
			t.Fatal("expected error when the account does not exist")
		}
	})
}

func TestUpdateCurrentUser(t *testing.T) {
	userID := uuid.New()
	existing := entities.User{
		ID:                userID,
		Name:              "Old Name",
		Email:             "user@example.com",
		PreferredCurrency: "USD",
		Image:             "old.png",
	}

	newSvc := func(t *testing.T) (*Services, *entities.User) {
		t.Helper()
		var saved entities.User
		repo := &fakeRepository{
			getUserByID: func(context.Context, uuid.UUID) (entities.User, error) {
				return existing, nil
			},
			updateUserProfile: func(_ context.Context, id uuid.UUID, name, preferredCurrency, image string) (entities.User, error) {
				saved = entities.User{ID: id, Name: name, PreferredCurrency: preferredCurrency, Image: image}
				return saved, nil
			},
		}
		return newTestServices(repo, newMemStorage()), &saved
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
	deletedAt := time.Now()
	repo := &fakeRepository{
		getUserByID: func(context.Context, uuid.UUID) (entities.User, error) {
			return entities.User{ID: uuid.New(), DeletedAt: &deletedAt}, nil
		},
	}
	svc := newTestServices(repo, newMemStorage())

	_, err := svc.UpdateUser(context.Background(), uuid.New(), "Name", "mail@example.com", "")
	if err == nil || err.Error() != "not found user" {
		t.Fatalf("UpdateUser error = %v, want %q", err, "not found user")
	}
}
