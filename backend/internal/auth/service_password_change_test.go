package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// ChangePassword moved here from the user module: the credentials live in
// accounts, and the verification, the session revocation and the alert email
// are all this module's. These cases came with it.
func TestChangePassword(t *testing.T) {
	userID := uuid.New()
	const currentPassword = "current-password"

	currentHash, err := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash current password: %v", err)
	}

	user := identity.User{ID: userID, Name: "Ada", Email: "test@example.com"}

	// baseRepo answers the reads every case needs: the stored credentials and
	// the profile the alert email is addressed to.
	baseRepo := func() *fakeRepository {
		return &fakeRepository{
			getAccountByUserID: func(context.Context, uuid.UUID) (identity.Account, error) {
				return identity.Account{UserID: userID, Password: string(currentHash)}, nil
			},
			getUserByID: func(context.Context, uuid.UUID) (identity.User, error) {
				return user, nil
			},
			updatePassword: func(context.Context, uuid.UUID, string) error { return nil },
			listSessionsByUserID: func(context.Context, uuid.UUID) ([]identity.Session, error) {
				return nil, nil
			},
		}
	}

	t.Run("revokes the other sessions and alerts", func(t *testing.T) {
		otherSession := identity.Session{ID: uuid.New(), Token: "other-token"}
		repo := baseRepo()
		repo.listSessionsByUserID = func(context.Context, uuid.UUID) ([]identity.Session, error) {
			// The caller's own session must survive; only the other one goes.
			return []identity.Session{{ID: uuid.New(), Token: "current-token"}, otherSession}, nil
		}
		repo.getRefreshTokensBySessionIDs = func(context.Context, uuid.UUID, []uuid.UUID) ([]string, []uuid.UUID, error) {
			return nil, nil, nil
		}
		var deleted []uuid.UUID
		repo.deleteSessionsByIDs = func(_ context.Context, _ uuid.UUID, ids []uuid.UUID) (int64, error) {
			deleted = ids
			return int64(len(ids)), nil
		}
		mailer := &fakeMailer{}
		svc := newTestServiceFull(repo, newMemStorage(), mailer)

		if err := svc.ChangePassword(context.Background(), userID, "current-token", currentPassword, "new-password", "203.0.113.5", "test-agent"); err != nil {
			t.Fatalf("ChangePassword: %v", err)
		}

		if len(deleted) != 1 || deleted[0] != otherSession.ID {
			t.Errorf("revoked sessions = %v, want only %s", deleted, otherSession.ID)
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
		if mailer.security[0].To != user.Email {
			t.Errorf("alert sent to %s, want %s", mailer.security[0].To, user.Email)
		}
		if mailer.security[0].Data.IPAddress != "203.0.113.5" {
			t.Errorf("alert IP = %s, want 203.0.113.5", mailer.security[0].Data.IPAddress)
		}
	})

	t.Run("stores a new bcrypt hash", func(t *testing.T) {
		repo := baseRepo()
		var storedHash string
		repo.updatePassword = func(_ context.Context, uid uuid.UUID, hashed string) error {
			if uid != userID {
				t.Errorf("UpdatePassword userID = %s, want %s", uid, userID)
			}
			storedHash = hashed
			return nil
		}
		svc := newTestServiceFull(repo, newMemStorage(), &fakeMailer{})

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
		repo := baseRepo()
		// A nil updatePassword hook panics if the write is reached, so a plain
		// error proves the check short-circuits.
		repo.updatePassword = nil
		svc := newTestServiceFull(repo, newMemStorage(), &fakeMailer{})

		if err := svc.ChangePassword(context.Background(), userID, "current-token", "wrong-password", "new-password", "", ""); err == nil {
			t.Fatal("expected an error for an incorrect current password")
		}
	})

	t.Run("rejects a new password identical to the current one", func(t *testing.T) {
		repo := baseRepo()
		repo.updatePassword = nil
		svc := newTestServiceFull(repo, newMemStorage(), &fakeMailer{})

		if err := svc.ChangePassword(context.Background(), userID, "current-token", currentPassword, currentPassword, "", ""); err == nil {
			t.Fatal("expected an error when the new password matches the current one")
		}
	})
}

// TestChangePasswordRoute covers the moved HTTP surface: the path stays
// PATCH /users/me/password even though the handler is now this module's.
func TestChangePasswordRoute(t *testing.T) {
	userID := uuid.New()
	cfg := testConfig()

	signer := newService(testStores(&fakeRepository{}), cfg, newMemStorage(), nil, nil, logger.Noop())
	token, err := signer.CreateJWToken(userID, "user", time.Now().UTC().Add(time.Hour))
	if err != nil {
		t.Fatalf("CreateJWToken: %v", err)
	}

	app := newTestApp(&fakeRepository{
		getSessionByToken: func(context.Context, string) (identity.User, error) {
			return identity.User{
				ID:       userID,
				Role:     identity.Role{Name: "user"},
				Sessions: []identity.Session{{Token: token}},
			}, nil
		},
	}, cfg)

	patch := func(t *testing.T, body, bearer string) int {
		t.Helper()
		req := httptest.NewRequest(http.MethodPatch, "/users/me/password", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode
	}

	// An unknown token must be stopped by the module's own RequireAuth.
	if status := patch(t, `{}`, "bogus-token"); status != fiber.StatusUnauthorized {
		t.Errorf("status = %d, want 401 with an invalid token", status)
	}

	// With one, the request reaches the handler, which rejects a new password
	// under the minimum length. That is enough to prove the route is mounted
	// at the old path and wired to this module's handler.
	status := patch(t, `{"currentPassword":"current-password","newPassword":"short"}`, token)
	if status != fiber.StatusBadRequest {
		t.Errorf("status = %d, want 400 for a too-short new password", status)
	}
}
