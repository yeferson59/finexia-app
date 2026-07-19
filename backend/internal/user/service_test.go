package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/config"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

func TestChangePassword(t *testing.T) {
	userID := uuid.New()
	const currentPassword = "current-password"

	// verifyPassword mirrors the auth module's real behavior (and its frozen
	// error strings) so the delegation keeps the same HTTP mapping.
	authSvc := func() *fakeAuthService {
		return &fakeAuthService{
			verifyPassword: func(_ context.Context, _ uuid.UUID, password string) error {
				if password != currentPassword {
					return errors.New("invalid current password")
				}
				return nil
			},
		}
	}

	t.Run("revokes other sessions and alerts", func(t *testing.T) {
		auth := authSvc()
		var revokedUser uuid.UUID
		var revokedToken string
		auth.revokeOtherSessions = func(_ context.Context, uid uuid.UUID, token string) (int64, error) {
			revokedUser, revokedToken = uid, token
			return 1, nil
		}

		repo := &fakeRepository{
			getUserByID: func(context.Context, uuid.UUID) (identity.User, error) {
				return identity.User{ID: userID, Name: "Ada", Email: "test@example.com"}, nil
			},
			updateUserPassword: func(context.Context, uuid.UUID, string) error {
				return nil
			},
		}
		mailer := &fakeMailer{}

		svc := NewService(repo, mailer, auth, nil, nil, logger.Noop(), &config.Env{FrontendURL: "https://app.finexia.me"})

		err := svc.ChangePassword(context.Background(), userID, "current-token", currentPassword, "new-password", "203.0.113.5", "test-agent")
		if err != nil {
			t.Fatalf("ChangePassword: %v", err)
		}

		if revokedUser != userID || revokedToken != "current-token" {
			t.Errorf("RevokeOtherSessions called with (%s, %q), want (%s, %q)", revokedUser, revokedToken, userID, "current-token")
		}

		ok := waitFor(t, 2*time.Second, func() bool {
			mailer.mu.Lock()
			defer mailer.mu.Unlock()
			return len(mailer.security) == 1
		})
		if !ok {
			t.Fatal("expected a security alert email after the password change")
		}
		mailer.mu.Lock()
		defer mailer.mu.Unlock()
		if mailer.security[0].To != "test@example.com" {
			t.Errorf("alert sent to %s, want test@example.com", mailer.security[0].To)
		}
	})

	t.Run("stores a new bcrypt hash", func(t *testing.T) {
		var storedHash string
		repo := &fakeRepository{
			getUserByID: func(context.Context, uuid.UUID) (identity.User, error) {
				return identity.User{ID: userID, Email: "test@example.com"}, nil
			},
			updateUserPassword: func(_ context.Context, uid uuid.UUID, hashed string) error {
				if uid != userID {
					t.Errorf("UpdatePassword userID = %s, want %s", uid, userID)
				}
				storedHash = hashed
				return nil
			},
		}
		auth := authSvc()
		auth.revokeOtherSessions = func(context.Context, uuid.UUID, string) (int64, error) {
			return 0, nil
		}
		svc := NewService(repo, &fakeMailer{}, auth, nil, nil, logger.Noop(), &config.Env{})

		if err := svc.ChangePassword(context.Background(), userID, "current-token", currentPassword, "new-password", "", ""); err != nil {
			t.Fatalf("ChangePassword: %v", err)
		}
		if storedHash == "" {
			t.Fatal("expected a new password hash to be stored")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte("new-password")); err != nil {
			t.Errorf("stored hash does not match the new password: %v", err)
		}
	})

	t.Run("rejects an incorrect current password", func(t *testing.T) {
		repo := &fakeRepository{}
		mailer := &fakeMailer{}
		svc := NewService(repo, mailer, authSvc(), nil, nil, logger.Noop(), &config.Env{})

		err := svc.ChangePassword(context.Background(), userID, "current-token", "wrong-password", "new-password", "", "")
		if err == nil {
			t.Fatal("expected an error for an incorrect current password")
		}
	})

	t.Run("rejects a new password identical to the current one", func(t *testing.T) {
		repo := &fakeRepository{}
		mailer := &fakeMailer{}
		svc := NewService(repo, mailer, authSvc(), nil, nil, logger.Noop(), &config.Env{})

		err := svc.ChangePassword(context.Background(), userID, "current-token", currentPassword, currentPassword, "", "")
		if err == nil {
			t.Fatal("expected an error when the new password matches the current one")
		}
	})
}

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
		var saved identity.User
		repo := &fakeRepository{
			getUserByID: func(context.Context, uuid.UUID) (identity.User, error) {
				return existing, nil
			},
			updateUserProfile: func(_ context.Context, id uuid.UUID, name, preferredCurrency, image string) (identity.User, error) {
				saved = identity.User{ID: id, Name: name, PreferredCurrency: preferredCurrency, Image: image}
				return saved, nil
			},
		}
		return NewService(repo, nil, nil, nil, nil, logger.Noop(), &config.Env{}), &saved
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
		getUserByID: func(context.Context, uuid.UUID) (identity.User, error) {
			return identity.User{ID: uuid.New(), DeletedAt: &deletedAt}, nil
		},
	}
	svc := NewService(repo, nil, nil, nil, nil, logger.Noop(), &config.Env{})

	_, err := svc.UpdateUser(context.Background(), uuid.New(), "Name", "mail@example.com", "")
	if err == nil || err.Error() != "not found user" {
		t.Fatalf("UpdateUser error = %v, want %q", err, "not found user")
	}
}
