package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
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
		mailer := &fakeMailer{}

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

}
