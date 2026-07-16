package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/identity"
)

func TestCreateInvitation_Success(t *testing.T) {
	var capturedHash string
	repo := &fakeRepository{
		getUserByEmail: notFound,
		createInvitation: func(_ context.Context, email, name, role, tokenHash string, invitedBy *uuid.UUID, expiresAt time.Time) (Invitation, error) {
			capturedHash = tokenHash
			return Invitation{
				ID: uuid.New(), Email: email, Name: name, Role: role,
				TokenHash: tokenHash, InvitedBy: invitedBy, ExpiresAt: expiresAt,
			}, nil
		},
		setWaitlistInvited: func(context.Context, string) error { return nil },
	}
	mailer := &fakeMailer{}
	svc := newTestServiceFull(repo, newMemStorage(), mailer)
	svc.cfg.InvitationExpiry = 72 * time.Hour
	svc.cfg.FrontendURL = "https://app.finexia.me"

	inv, err := svc.CreateInvitation(context.Background(), "  New.User@Example.com ", "", "", uuid.New())
	if err != nil {
		t.Fatalf("CreateInvitation: unexpected error: %v", err)
	}

	if inv.Email != "new.user@example.com" {
		t.Errorf("email not normalized: got %q", inv.Email)
	}
	if inv.Role != "customer" {
		t.Errorf("role should default to customer, got %q", inv.Role)
	}
	if inv.Name == "" {
		t.Error("name should be derived from email, got empty")
	}
	if capturedHash == "" || len(capturedHash) != 64 {
		t.Errorf("expected 64-char sha256 token hash, got %q", capturedHash)
	}

	// The email is sent from a goroutine; give it a moment to record.
	waitFor(t, 2*time.Second, func() bool { return mailer.invitationCount() == 1 })
	mailer.mu.Lock()
	got := mailer.invitationTo[0]
	mailer.mu.Unlock()
	if got.To != "new.user@example.com" {
		t.Errorf("invitation sent to wrong address: %q", got.To)
	}
	if !strings.HasPrefix(got.Data.InviteURL, "https://app.finexia.me/auth/accept-invite?token=") {
		t.Errorf("unexpected invite URL: %q", got.Data.InviteURL)
	}
}

func TestCreateInvitation_RejectsExistingUser(t *testing.T) {
	repo := &fakeRepository{
		getUserByEmail: func(context.Context, string) (identity.User, error) {
			return identity.User{ID: uuid.New()}, nil
		},
	}
	svc := newTestServiceFull(repo, newMemStorage(), &fakeMailer{})

	_, err := svc.CreateInvitation(context.Background(), "taken@example.com", "Taken", "customer", uuid.New())
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected 'already exists' error, got %v", err)
	}
}

func TestCreateInvitation_RejectsBadRole(t *testing.T) {
	repo := &fakeRepository{}
	svc := newTestServiceFull(repo, newMemStorage(), &fakeMailer{})

	_, err := svc.CreateInvitation(context.Background(), "x@example.com", "X", "superadmin", uuid.New())
	if err == nil || !strings.Contains(err.Error(), "invalid role") {
		t.Fatalf("expected 'invalid role' error, got %v", err)
	}
}

func TestAcceptInvitation_Success(t *testing.T) {
	raw, hash, err := generateRefreshToken()
	if err != nil {
		t.Fatalf("token gen: %v", err)
	}
	invID := uuid.New()

	var acceptedName, acceptedHash string
	repo := &fakeRepository{
		getInvitationByHash: func(_ context.Context, tokenHash string) (Invitation, error) {
			if tokenHash != hash {
				t.Errorf("service hashed token differently: %q != %q", tokenHash, hash)
			}
			return Invitation{
				ID: invID, Email: "invitee@example.com", Name: "Fallback", Role: "customer",
				ExpiresAt: time.Now().UTC().Add(time.Hour),
			}, nil
		},
		acceptInvitation: func(_ context.Context, id uuid.UUID, name, email, role, passwordHash string) (identity.User, error) {
			if id != invID {
				t.Errorf("wrong invitation id: %v", id)
			}
			acceptedName, acceptedHash = name, passwordHash
			return identity.User{ID: uuid.New(), Email: email, Name: name}, nil
		},
	}
	svc := newTestServiceFull(repo, newMemStorage(), &fakeMailer{})

	u, err := svc.AcceptInvitation(context.Background(), raw, "Jane Doe", "s3cretpass")
	if err != nil {
		t.Fatalf("AcceptInvitation: %v", err)
	}
	if u.Email != "invitee@example.com" {
		t.Errorf("wrong email on created user: %q", u.Email)
	}
	if acceptedName != "Jane Doe" {
		t.Errorf("expected provided name to win, got %q", acceptedName)
	}
	if acceptedHash == "" || acceptedHash == "s3cretpass" {
		t.Errorf("password must be hashed before persistence, got %q", acceptedHash)
	}
}

func TestAcceptInvitation_Expired(t *testing.T) {
	raw, _, _ := generateRefreshToken()
	repo := &fakeRepository{
		getInvitationByHash: func(context.Context, string) (Invitation, error) {
			return Invitation{
				ID: uuid.New(), Email: "old@example.com", Role: "customer",
				ExpiresAt: time.Now().UTC().Add(-time.Hour),
			}, nil
		},
	}
	svc := newTestServiceFull(repo, newMemStorage(), &fakeMailer{})

	_, err := svc.AcceptInvitation(context.Background(), raw, "", "s3cretpass")
	if !errors.Is(err, ErrInvitationExpired) {
		t.Fatalf("expected ErrInvitationExpired, got %v", err)
	}
}

func TestAcceptInvitation_Revoked(t *testing.T) {
	raw, _, _ := generateRefreshToken()
	revoked := time.Now().UTC().Add(-time.Minute)
	repo := &fakeRepository{
		getInvitationByHash: func(context.Context, string) (Invitation, error) {
			return Invitation{
				ID: uuid.New(), Email: "revoked@example.com", Role: "customer",
				ExpiresAt: time.Now().UTC().Add(time.Hour), RevokedAt: &revoked,
			}, nil
		},
	}
	svc := newTestServiceFull(repo, newMemStorage(), &fakeMailer{})

	_, err := svc.AcceptInvitation(context.Background(), raw, "", "s3cretpass")
	if !errors.Is(err, ErrInvitationInvalid) {
		t.Fatalf("expected ErrInvitationInvalid, got %v", err)
	}
}
