package user

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
)

// fakeRepository embeds the Repository interface so tests only override the
// methods a scenario needs; calling anything else panics loudly.
type fakeRepository struct {
	Repository

	getUserByEmail            func(ctx context.Context, email string) (identity.User, error)
	getUserByID               func(ctx context.Context, id uuid.UUID) (identity.User, error)
	updateUser                func(ctx context.Context, id uuid.UUID, name, email, image string) (identity.User, error)
	updateUserPassword        func(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	getUserPreferences        func(ctx context.Context, userID uuid.UUID) (UserPreferences, error)
	getUsersWithWeeklySummary func(ctx context.Context) ([]identity.User, error)
}

func (f *fakeRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	return f.updateUserPassword(ctx, userID, hashedPassword)
}

// fakeAuthService stubs the auth module slice Service.ChangePassword
// delegates to (current-password verification, session revocation).
type fakeAuthService struct {
	verifyPassword      func(ctx context.Context, userID uuid.UUID, currentPassword string) error
	revokeOtherSessions func(ctx context.Context, userID uuid.UUID, currentToken string) (int64, error)
}

func (f *fakeAuthService) VerifyPassword(ctx context.Context, userID uuid.UUID, currentPassword string) error {
	return f.verifyPassword(ctx, userID, currentPassword)
}

func (f *fakeAuthService) RevokeOtherSessions(ctx context.Context, userID uuid.UUID, currentToken string) (int64, error) {
	return f.revokeOtherSessions(ctx, userID, currentToken)
}

// fakeMailer records security alerts sent through the mailer slice, guarded
// by a mutex since ChangePassword sends its alert on a background goroutine.
type fakeMailer struct {
	mu sync.Mutex

	security []struct {
		To   string
		Data mail.SecurityAlertData
	}
}

func (m *fakeMailer) SendSecurityAlert(email string, data mail.SecurityAlertData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.security = append(m.security, struct {
		To   string
		Data mail.SecurityAlertData
	}{email, data})
	return nil
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return cond()
}

func (f *fakeRepository) GetPreferences(ctx context.Context, userID uuid.UUID) (UserPreferences, error) {
	return f.getUserPreferences(ctx, userID)
}

func (f *fakeRepository) GetWeeklySummary(ctx context.Context) ([]identity.User, error) {
	return f.getUsersWithWeeklySummary(ctx)
}

func (f *fakeRepository) GetByEmail(ctx context.Context, email string) (identity.User, error) {
	return f.getUserByEmail(ctx, email)
}

func (f *fakeRepository) GetByID(ctx context.Context, id uuid.UUID) (identity.User, error) {
	return f.getUserByID(ctx, id)
}

func (f *fakeRepository) Update(ctx context.Context, id uuid.UUID, name, email, image string) (identity.User, error) {
	return f.updateUser(ctx, id, name, email, image)
}
