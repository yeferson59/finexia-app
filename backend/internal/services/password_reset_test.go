package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/entities"
)

func TestRequestPasswordReset_Success(t *testing.T) {
	userID := uuid.New()
	var capturedHash string
	repo := &fakeRepository{
		getUserByEmail: func(_ context.Context, email string) (entities.User, error) {
			return entities.User{ID: userID, Name: "Jane Doe", Email: email}, nil
		},
		createPasswordReset: func(_ context.Context, uid uuid.UUID, tokenHash string, expiresAt time.Time) (entities.PasswordReset, error) {
			if uid != userID {
				t.Errorf("wrong user id: %v", uid)
			}
			capturedHash = tokenHash
			return entities.PasswordReset{ID: uuid.New(), UserID: uid, TokenHash: tokenHash, ExpiresAt: expiresAt}, nil
		},
	}
	mailer := &fakeMailer{}
	svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)
	svc.cfg.PasswordResetExpiry = time.Hour
	svc.cfg.FrontendURL = "https://app.finexia.me"

	if err := svc.RequestPasswordReset(context.Background(), "  Jane@Example.com "); err != nil {
		t.Fatalf("RequestPasswordReset: unexpected error: %v", err)
	}

	if capturedHash == "" || len(capturedHash) != 64 {
		t.Errorf("expected 64-char sha256 token hash, got %q", capturedHash)
	}

	waitFor(t, 2*time.Second, func() bool { return mailer.passwordResetCount() == 1 })
	mailer.mu.Lock()
	got := mailer.passwordResetTo[0]
	mailer.mu.Unlock()
	if got.To != "jane@example.com" {
		t.Errorf("reset email sent to wrong address: %q", got.To)
	}
	if !strings.HasPrefix(got.Data.ResetURL, "https://app.finexia.me/auth/reset-password?token=") {
		t.Errorf("unexpected reset URL: %q", got.Data.ResetURL)
	}
}

func TestRequestPasswordReset_UnknownEmailIsSilentSuccess(t *testing.T) {
	repo := &fakeRepository{
		getUserByEmail: notFound,
	}
	mailer := &fakeMailer{}
	svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)

	if err := svc.RequestPasswordReset(context.Background(), "ghost@example.com"); err != nil {
		t.Fatalf("expected nil error for unknown email, got %v", err)
	}
	if mailer.passwordResetCount() != 0 {
		t.Errorf("expected no email sent for unknown address")
	}
}

func TestResetPassword_Success(t *testing.T) {
	raw, hash, err := generateRefreshToken()
	if err != nil {
		t.Fatalf("token gen: %v", err)
	}
	resetID := uuid.New()
	userID := uuid.New()

	var consumedResetID, consumedUserID uuid.UUID
	var consumedHash string
	repo := &fakeRepository{
		getPasswordResetByHash: func(_ context.Context, tokenHash string) (entities.PasswordReset, error) {
			if tokenHash != hash {
				t.Errorf("service hashed token differently: %q != %q", tokenHash, hash)
			}
			return entities.PasswordReset{
				ID: resetID, UserID: userID,
				ExpiresAt: time.Now().UTC().Add(time.Hour),
			}, nil
		},
		consumePasswordReset: func(_ context.Context, rID, uID uuid.UUID, hashedPassword string) error {
			consumedResetID, consumedUserID, consumedHash = rID, uID, hashedPassword
			return nil
		},
	}
	var revokedUser uuid.UUID
	authSvc := &fakeAuthService{
		revokeOtherSessions: func(_ context.Context, uid uuid.UUID, _ string) (int64, error) {
			revokedUser = uid
			return 0, nil
		},
	}
	svc := newTestServicesAuth(repo, newMemStorage(), nil, authSvc)

	if err := svc.ResetPassword(context.Background(), raw, "s3cretpass", "203.0.113.9", "test-agent"); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}
	if consumedResetID != resetID {
		t.Errorf("wrong reset id consumed: %v", consumedResetID)
	}
	if consumedUserID != userID {
		t.Errorf("wrong user id: %v", consumedUserID)
	}
	if revokedUser != userID {
		t.Errorf("expected other sessions of %s to be revoked, got %s", userID, revokedUser)
	}
	if consumedHash == "" || consumedHash == "s3cretpass" {
		t.Errorf("password must be hashed before persistence, got %q", consumedHash)
	}
}

func TestResetPassword_Expired(t *testing.T) {
	raw, _, _ := generateRefreshToken()
	repo := &fakeRepository{
		getPasswordResetByHash: func(context.Context, string) (entities.PasswordReset, error) {
			return entities.PasswordReset{
				ID: uuid.New(), UserID: uuid.New(),
				ExpiresAt: time.Now().UTC().Add(-time.Hour),
			}, nil
		},
	}
	svc := newTestServices(repo, newMemStorage())

	err := svc.ResetPassword(context.Background(), raw, "s3cretpass", "203.0.113.9", "test-agent")
	if !errors.Is(err, ErrPasswordResetExpired) {
		t.Fatalf("expected ErrPasswordResetExpired, got %v", err)
	}
}

func TestResetPassword_AlreadyUsed(t *testing.T) {
	raw, _, _ := generateRefreshToken()
	usedAt := time.Now().UTC().Add(-time.Minute)
	repo := &fakeRepository{
		getPasswordResetByHash: func(context.Context, string) (entities.PasswordReset, error) {
			return entities.PasswordReset{
				ID: uuid.New(), UserID: uuid.New(),
				ExpiresAt: time.Now().UTC().Add(time.Hour), UsedAt: &usedAt,
			}, nil
		},
	}
	svc := newTestServices(repo, newMemStorage())

	err := svc.ResetPassword(context.Background(), raw, "s3cretpass", "203.0.113.9", "test-agent")
	if !errors.Is(err, ErrPasswordResetInvalid) {
		t.Fatalf("expected ErrPasswordResetInvalid, got %v", err)
	}
}

func TestResetPassword_InvalidToken(t *testing.T) {
	raw, _, _ := generateRefreshToken()
	repo := &fakeRepository{
		getPasswordResetByHash: func(context.Context, string) (entities.PasswordReset, error) {
			return entities.PasswordReset{}, errors.New("not found")
		},
	}
	svc := newTestServices(repo, newMemStorage())

	err := svc.ResetPassword(context.Background(), raw, "s3cretpass", "203.0.113.9", "test-agent")
	if !errors.Is(err, ErrPasswordResetInvalid) {
		t.Fatalf("expected ErrPasswordResetInvalid, got %v", err)
	}
}

func TestResetPassword_EmptyToken(t *testing.T) {
	svc := newTestServices(&fakeRepository{}, newMemStorage())

	err := svc.ResetPassword(context.Background(), "", "s3cretpass", "203.0.113.9", "test-agent")
	if !errors.Is(err, ErrPasswordResetInvalid) {
		t.Fatalf("expected ErrPasswordResetInvalid, got %v", err)
	}
}
