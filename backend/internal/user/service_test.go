package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

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
