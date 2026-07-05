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

func TestRegisterSendsEmailVerification(t *testing.T) {
	var capturedHash, capturedEmail string
	repo := &fakeRepository{
		getUserByEmail: notFound,
		register: func(_ context.Context, name, email, hashed string) (entities.User, error) {
			return entities.User{Name: name, Email: email}, nil
		},
		createEmailVerification: func(_ context.Context, email, tokenHash string, expiresAt time.Time) (entities.Verification, error) {
			capturedEmail = email
			capturedHash = tokenHash
			return entities.Verification{ID: uuid.New(), Identifier: email, Value: tokenHash, ExpiresAt: expiresAt}, nil
		},
	}
	mailer := &fakeMailer{}
	svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)
	svc.cfg.EmailVerificationExpiry = 24 * time.Hour
	svc.cfg.FrontendURL = "https://app.finexia.me"

	if _, err := svc.Register(context.Background(), "Jane Doe", "jane@example.com", "s3cretpass"); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if capturedEmail != "jane@example.com" {
		t.Errorf("CreateEmailVerification called with email %q, want jane@example.com", capturedEmail)
	}
	if capturedHash == "" || len(capturedHash) != 64 {
		t.Errorf("expected 64-char sha256 token hash, got %q", capturedHash)
	}

	waitFor(t, 2*time.Second, func() bool { return mailer.emailVerificationCount() == 1 })
	mailer.mu.Lock()
	got := mailer.emailVerificationTo[0]
	mailer.mu.Unlock()
	if got.To != "jane@example.com" {
		t.Errorf("verification email sent to wrong address: %q", got.To)
	}
	if !strings.HasPrefix(got.Data.VerifyURL, "https://app.finexia.me/auth/verify-email?token=") {
		t.Errorf("unexpected verify URL: %q", got.Data.VerifyURL)
	}
}

func TestRequestEmailVerification_UnknownEmailIsSilentSuccess(t *testing.T) {
	repo := &fakeRepository{
		getUserByEmail: notFound,
	}
	mailer := &fakeMailer{}
	svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)

	if err := svc.RequestEmailVerification(context.Background(), "ghost@example.com"); err != nil {
		t.Fatalf("expected nil error for unknown email, got %v", err)
	}
	if mailer.emailVerificationCount() != 0 {
		t.Errorf("expected no email sent for unknown address")
	}
}

func TestRequestEmailVerification_AlreadyVerifiedIsSilentSuccess(t *testing.T) {
	repo := &fakeRepository{
		getUserByEmail: func(_ context.Context, email string) (entities.User, error) {
			return entities.User{Email: email, EmailVerified: true}, nil
		},
	}
	mailer := &fakeMailer{}
	svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)

	if err := svc.RequestEmailVerification(context.Background(), "verified@example.com"); err != nil {
		t.Fatalf("expected nil error for verified email, got %v", err)
	}
	if mailer.emailVerificationCount() != 0 {
		t.Errorf("expected no email sent for an already-verified address")
	}
}

func TestVerifyEmail_Success(t *testing.T) {
	raw, hash, err := generateRefreshToken()
	if err != nil {
		t.Fatalf("token gen: %v", err)
	}
	verificationID := uuid.New()

	var consumedID uuid.UUID
	var consumedEmail string
	repo := &fakeRepository{
		getEmailVerificationByHash: func(_ context.Context, tokenHash string) (entities.Verification, error) {
			if tokenHash != hash {
				t.Errorf("service hashed token differently: %q != %q", tokenHash, hash)
			}
			return entities.Verification{
				ID: verificationID, Identifier: "jane@example.com",
				ExpiresAt: time.Now().UTC().Add(time.Hour),
			}, nil
		},
		consumeEmailVerification: func(_ context.Context, id uuid.UUID, email string) error {
			consumedID, consumedEmail = id, email
			return nil
		},
	}
	svc := newTestServices(repo, newMemStorage())

	if err := svc.VerifyEmail(context.Background(), raw); err != nil {
		t.Fatalf("VerifyEmail: %v", err)
	}
	if consumedID != verificationID {
		t.Errorf("wrong verification id consumed: %v", consumedID)
	}
	if consumedEmail != "jane@example.com" {
		t.Errorf("wrong email consumed: %v", consumedEmail)
	}
}

func TestVerifyEmail_Expired(t *testing.T) {
	raw, _, _ := generateRefreshToken()
	repo := &fakeRepository{
		getEmailVerificationByHash: func(context.Context, string) (entities.Verification, error) {
			return entities.Verification{
				ID: uuid.New(), Identifier: "jane@example.com",
				ExpiresAt: time.Now().UTC().Add(-time.Hour),
			}, nil
		},
	}
	svc := newTestServices(repo, newMemStorage())

	err := svc.VerifyEmail(context.Background(), raw)
	if !errors.Is(err, ErrEmailVerificationExpired) {
		t.Fatalf("expected ErrEmailVerificationExpired, got %v", err)
	}
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	raw, _, _ := generateRefreshToken()
	repo := &fakeRepository{
		getEmailVerificationByHash: func(context.Context, string) (entities.Verification, error) {
			return entities.Verification{}, errors.New("not found")
		},
	}
	svc := newTestServices(repo, newMemStorage())

	err := svc.VerifyEmail(context.Background(), raw)
	if !errors.Is(err, ErrEmailVerificationInvalid) {
		t.Fatalf("expected ErrEmailVerificationInvalid, got %v", err)
	}
}

func TestVerifyEmail_EmptyToken(t *testing.T) {
	svc := newTestServices(&fakeRepository{}, newMemStorage())

	err := svc.VerifyEmail(context.Background(), "")
	if !errors.Is(err, ErrEmailVerificationInvalid) {
		t.Fatalf("expected ErrEmailVerificationInvalid, got %v", err)
	}
}
