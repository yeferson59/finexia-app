package auth

import (
	"context"
	"errors"

	"uuid"

	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// ChangePassword updates the password of an already-authenticated user, who
// proves intent with their current one. It lives here, not in the user module,
// because every step is this module's: the credentials sit in accounts, the
// verification and the session revocation are auth use cases, and the alert
// email is the same one the reset flow sends.
//
// currentToken is the session token of the caller, kept alive so the user is
// not logged out of the device they just changed the password on.
func (s *service) ChangePassword(ctx context.Context, userID uuid.UUID, currentToken, currentPassword, newPassword, ipAddress, userAgent string) error {
	if err := s.VerifyPassword(ctx, userID, currentPassword); err != nil {
		return err
	}

	if currentPassword == newPassword {
		return httpx.AsBadRequest(errors.New("invalid new password: must differ from current password"))
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.stores.Accounts.UpdatePassword(ctx, userID, string(hashed)); err != nil {
		return err
	}

	// Whoever else holds a session (a stolen token, a forgotten shared
	// computer) must not survive a password change: only the session that
	// performed the change stays alive.
	if _, err := s.RevokeOtherSessions(ctx, userID, currentToken); err != nil {
		s.log.Error(ctx, "change password: failed to revoke other sessions", logger.Err(err))
	}

	go s.sendPasswordChangedAlert(userID, ipAddress, userAgent)

	return nil
}
