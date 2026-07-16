package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/pkg/helpers"
)

// twoFactorState is a tiny in-memory stand-in for the two 2FA tables.
type twoFactorState struct {
	row   *TwoFactor
	codes map[string]bool // code hash → used
}

// wireTwoFactorRepo attaches stateful 2FA behaviour to a fakeRepository.
func wireTwoFactorRepo(repo *fakeRepository, state *twoFactorState) {
	repo.getTwoFactor = func(_ context.Context, _ uuid.UUID) (TwoFactor, error) {
		if state.row == nil {
			return TwoFactor{}, ErrTwoFactorNotFound
		}
		return *state.row, nil
	}
	repo.upsertTwoFactorSecret = func(_ context.Context, userID uuid.UUID, secret string) error {
		if state.row != nil && state.row.Enabled {
			return errors.New("two-factor already enabled")
		}
		state.row = &TwoFactor{UserID: userID, Secret: secret}
		return nil
	}
	repo.enableTwoFactor = func(_ context.Context, _ uuid.UUID) error {
		if state.row == nil || state.row.Enabled {
			return ErrTwoFactorNotFound
		}
		state.row.Enabled = true
		now := time.Now()
		state.row.ConfirmedAt = &now
		return nil
	}
	repo.deleteTwoFactor = func(_ context.Context, _ uuid.UUID) error {
		state.row = nil
		state.codes = map[string]bool{}
		return nil
	}
	repo.replaceTwoFactorRecoveryCodes = func(_ context.Context, _ uuid.UUID, hashes []string) error {
		state.codes = map[string]bool{}
		for _, h := range hashes {
			state.codes[h] = false
		}
		return nil
	}
	repo.consumeTwoFactorRecoveryCode = func(_ context.Context, _ uuid.UUID, hash string) error {
		used, ok := state.codes[hash]
		if !ok || used {
			return errors.New("invalid recovery code")
		}
		state.codes[hash] = true
		return nil
	}
	repo.countUnusedTwoFactorRecoveryCodes = func(_ context.Context, _ uuid.UUID) (int, error) {
		count := 0
		for _, used := range state.codes {
			if !used {
				count++
			}
		}
		return count, nil
	}
}

func TestTwoFactorSetupAndEnableFlow(t *testing.T) {
	const password = "s3cret-password"
	user := verifiedUser(t, password)
	state := &twoFactorState{codes: map[string]bool{}}

	repo := &fakeRepository{
		getAccountByUserID: func(_ context.Context, _ uuid.UUID) (identity.Account, error) {
			return user.Accounts[0], nil
		},
		getUserByID: func(_ context.Context, _ uuid.UUID) (identity.User, error) {
			return user, nil
		},
	}
	wireTwoFactorRepo(repo, state)
	svc := newTestService(repo, newMemStorage())
	ctx := context.Background()

	// Default state: disabled.
	status, err := svc.TwoFactorStatus(ctx, user.ID)
	if err != nil {
		t.Fatalf("TwoFactorStatus: %v", err)
	}
	if status.Enabled || status.PendingSetup {
		t.Fatalf("expected 2FA off by default, got %+v", status)
	}

	// Wrong password cannot start a setup.
	if _, err := svc.BeginTwoFactorSetup(ctx, user.ID, "wrong-password"); err == nil {
		t.Fatal("expected BeginTwoFactorSetup to reject a wrong password")
	}

	setup, err := svc.BeginTwoFactorSetup(ctx, user.ID, password)
	if err != nil {
		t.Fatalf("BeginTwoFactorSetup: %v", err)
	}
	if setup.Secret == "" || setup.OtpauthURL == "" {
		t.Fatalf("expected secret and otpauth URL, got %+v", setup)
	}

	// A pending setup must NOT gate login yet.
	if status, _ = svc.TwoFactorStatus(ctx, user.ID); status.Enabled || !status.PendingSetup {
		t.Fatalf("expected pending setup, got %+v", status)
	}

	// A wrong code cannot confirm the setup.
	if _, err := svc.ConfirmTwoFactorSetup(ctx, user.ID, "000000", "203.0.113.9", "test-agent"); !errors.Is(err, ErrTwoFactorInvalidCode) {
		t.Fatalf("ConfirmTwoFactorSetup(wrong code) err = %v, want ErrTwoFactorInvalidCode", err)
	}

	code, err := helpers.TOTPCode(setup.Secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("TOTPCode: %v", err)
	}
	enabled, err := svc.ConfirmTwoFactorSetup(ctx, user.ID, code, "203.0.113.9", "test-agent")
	if err != nil {
		t.Fatalf("ConfirmTwoFactorSetup: %v", err)
	}
	if len(enabled.RecoveryCodes) != twoFactorRecoveryCodeCount {
		t.Fatalf("recovery codes = %d, want %d", len(enabled.RecoveryCodes), twoFactorRecoveryCodeCount)
	}

	status, err = svc.TwoFactorStatus(ctx, user.ID)
	if err != nil {
		t.Fatalf("TwoFactorStatus: %v", err)
	}
	if !status.Enabled || status.RecoveryCodesLeft != twoFactorRecoveryCodeCount {
		t.Fatalf("expected enabled with %d codes, got %+v", twoFactorRecoveryCodeCount, status)
	}

	// Re-running setup while enabled must be refused.
	if _, err := svc.BeginTwoFactorSetup(ctx, user.ID, password); !errors.Is(err, ErrTwoFactorAlreadyEnabled) {
		t.Fatalf("BeginTwoFactorSetup(enabled) err = %v, want ErrTwoFactorAlreadyEnabled", err)
	}
}

// loginCapableRepo wires the minimum needed for Login/issueSession to work.
func loginCapableRepo(t *testing.T, user identity.User, state *twoFactorState) *fakeRepository {
	t.Helper()
	repo := &fakeRepository{
		getAccountByEmail: func(_ context.Context, email string) (identity.User, error) {
			if email != user.Email {
				return identity.User{}, errors.New("not found")
			}
			return user, nil
		},
		getAccountByUserID: func(_ context.Context, _ uuid.UUID) (identity.Account, error) {
			return user.Accounts[0], nil
		},
		getUserByID: func(_ context.Context, id uuid.UUID) (identity.User, error) {
			if id != user.ID {
				return identity.User{}, errors.New("not found")
			}
			return user, nil
		},
		createSession: func(_ context.Context, _ uuid.UUID, _ string, _, _ *string, _ time.Time) (uuid.UUID, error) {
			return uuid.New(), nil
		},
		createRefreshToken: func(_ context.Context, _ uuid.UUID, _ string, _, _ uuid.UUID, _, _ *string, _ time.Time) (uuid.UUID, error) {
			return uuid.New(), nil
		},
	}
	wireTwoFactorRepo(repo, state)
	return repo
}

func enabledTwoFactorState(t *testing.T, userID uuid.UUID) (*twoFactorState, string) {
	t.Helper()
	secret, err := helpers.GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret: %v", err)
	}
	now := time.Now()
	return &twoFactorState{
		row:   &TwoFactor{UserID: userID, Secret: secret, Enabled: true, ConfirmedAt: &now},
		codes: map[string]bool{},
	}, secret
}

func TestLoginWithTwoFactorEnabled(t *testing.T) {
	const password = "s3cret-password"
	user := verifiedUser(t, password)
	state, secret := enabledTwoFactorState(t, user.ID)
	repo := loginCapableRepo(t, user, state)
	svc := newTestService(repo, newMemStorage())
	ctx := context.Background()

	// Password alone must not produce a session.
	result, err := svc.Login(ctx, user.Email, password, "", "")
	if !errors.Is(err, ErrTwoFactorRequired) {
		t.Fatalf("Login err = %v, want ErrTwoFactorRequired", err)
	}
	if result.AccessToken != "" || result.RawRefreshToken != "" {
		t.Fatal("Login leaked tokens before the 2FA step")
	}
	if result.TwoFactorToken == "" {
		t.Fatal("expected a pending two-factor token")
	}

	// Wrong code: rejected, pending token still alive.
	if _, err := svc.CompleteTwoFactorLogin(ctx, result.TwoFactorToken, "000000", "", ""); !errors.Is(err, ErrTwoFactorInvalidCode) {
		t.Fatalf("CompleteTwoFactorLogin(wrong) err = %v, want ErrTwoFactorInvalidCode", err)
	}

	code, err := helpers.TOTPCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("TOTPCode: %v", err)
	}
	session, err := svc.CompleteTwoFactorLogin(ctx, result.TwoFactorToken, code, "", "")
	if err != nil {
		t.Fatalf("CompleteTwoFactorLogin: %v", err)
	}
	if session.AccessToken == "" || session.RawRefreshToken == "" {
		t.Fatal("expected a full session after the 2FA step")
	}

	// The pending token is single-use.
	if _, err := svc.CompleteTwoFactorLogin(ctx, result.TwoFactorToken, code, "", ""); !errors.Is(err, ErrTwoFactorPendingInvalid) {
		t.Fatalf("second CompleteTwoFactorLogin err = %v, want ErrTwoFactorPendingInvalid", err)
	}
}

func TestTwoFactorCodeReplayRejected(t *testing.T) {
	const password = "s3cret-password"
	user := verifiedUser(t, password)
	state, secret := enabledTwoFactorState(t, user.ID)
	repo := loginCapableRepo(t, user, state)
	svc := newTestService(repo, newMemStorage())
	ctx := context.Background()

	first, err := svc.Login(ctx, user.Email, password, "", "")
	if !errors.Is(err, ErrTwoFactorRequired) {
		t.Fatalf("Login err = %v, want ErrTwoFactorRequired", err)
	}
	code, _ := helpers.TOTPCode(secret, time.Now().UTC())
	if _, err := svc.CompleteTwoFactorLogin(ctx, first.TwoFactorToken, code, "", ""); err != nil {
		t.Fatalf("CompleteTwoFactorLogin: %v", err)
	}

	// Same code on a fresh pending login must be rejected (replay).
	second, err := svc.Login(ctx, user.Email, password, "", "")
	if !errors.Is(err, ErrTwoFactorRequired) {
		t.Fatalf("Login err = %v, want ErrTwoFactorRequired", err)
	}
	if _, err := svc.CompleteTwoFactorLogin(ctx, second.TwoFactorToken, code, "", ""); !errors.Is(err, ErrTwoFactorInvalidCode) {
		t.Fatalf("replayed code err = %v, want ErrTwoFactorInvalidCode", err)
	}
}

func TestTwoFactorAttemptsExhaustPendingToken(t *testing.T) {
	const password = "s3cret-password"
	user := verifiedUser(t, password)
	state, _ := enabledTwoFactorState(t, user.ID)
	repo := loginCapableRepo(t, user, state)
	svc := newTestService(repo, newMemStorage())
	ctx := context.Background()

	result, err := svc.Login(ctx, user.Email, password, "", "")
	if !errors.Is(err, ErrTwoFactorRequired) {
		t.Fatalf("Login err = %v, want ErrTwoFactorRequired", err)
	}

	for i := range twoFactorMaxAttempts {
		_, err := svc.CompleteTwoFactorLogin(ctx, result.TwoFactorToken, "000000", "", "")
		if i < twoFactorMaxAttempts-1 && !errors.Is(err, ErrTwoFactorInvalidCode) {
			t.Fatalf("attempt %d err = %v, want ErrTwoFactorInvalidCode", i+1, err)
		}
		if i == twoFactorMaxAttempts-1 && !errors.Is(err, ErrTwoFactorPendingInvalid) {
			t.Fatalf("final attempt err = %v, want ErrTwoFactorPendingInvalid", err)
		}
	}
}

func TestTwoFactorLoginWithRecoveryCode(t *testing.T) {
	const password = "s3cret-password"
	user := verifiedUser(t, password)
	state, _ := enabledTwoFactorState(t, user.ID)
	repo := loginCapableRepo(t, user, state)
	svc := newTestService(repo, newMemStorage())
	ctx := context.Background()

	raws, hashes, err := generateTwoFactorRecoveryCodes()
	if err != nil {
		t.Fatalf("generateTwoFactorRecoveryCodes: %v", err)
	}
	for _, h := range hashes {
		state.codes[h] = false
	}

	result, err := svc.Login(ctx, user.Email, password, "", "")
	if !errors.Is(err, ErrTwoFactorRequired) {
		t.Fatalf("Login err = %v, want ErrTwoFactorRequired", err)
	}

	session, err := svc.CompleteTwoFactorLogin(ctx, result.TwoFactorToken, raws[0], "", "")
	if err != nil {
		t.Fatalf("CompleteTwoFactorLogin(recovery) err = %v", err)
	}
	if session.AccessToken == "" {
		t.Fatal("expected a session from the recovery code")
	}

	// The recovery code is single-use.
	again, err := svc.Login(ctx, user.Email, password, "", "")
	if !errors.Is(err, ErrTwoFactorRequired) {
		t.Fatalf("Login err = %v, want ErrTwoFactorRequired", err)
	}
	if _, err := svc.CompleteTwoFactorLogin(ctx, again.TwoFactorToken, raws[0], "", ""); !errors.Is(err, ErrTwoFactorInvalidCode) {
		t.Fatalf("reused recovery code err = %v, want ErrTwoFactorInvalidCode", err)
	}
}

func TestDisableTwoFactor(t *testing.T) {
	const password = "s3cret-password"
	user := verifiedUser(t, password)
	state, secret := enabledTwoFactorState(t, user.ID)
	repo := loginCapableRepo(t, user, state)
	svc := newTestService(repo, newMemStorage())
	ctx := context.Background()

	if err := svc.DisableTwoFactor(ctx, user.ID, "wrong-password", "whatever", "203.0.113.9", "test-agent"); err == nil {
		t.Fatal("expected DisableTwoFactor to reject a wrong password")
	}
	if err := svc.DisableTwoFactor(ctx, user.ID, password, "000000", "203.0.113.9", "test-agent"); !errors.Is(err, ErrTwoFactorInvalidCode) {
		t.Fatalf("DisableTwoFactor(wrong code) err = %v, want ErrTwoFactorInvalidCode", err)
	}

	code, _ := helpers.TOTPCode(secret, time.Now().UTC())
	if err := svc.DisableTwoFactor(ctx, user.ID, password, code, "203.0.113.9", "test-agent"); err != nil {
		t.Fatalf("DisableTwoFactor: %v", err)
	}
	if state.row != nil {
		t.Fatal("expected the 2FA enrollment to be removed")
	}

	// Back to the default: login no longer asks for a code.
	if _, err := svc.Login(ctx, user.Email, password, "", ""); err != nil {
		t.Fatalf("Login after disable: %v", err)
	}
}

func TestLoginWithPendingSetupDoesNotRequireCode(t *testing.T) {
	const password = "s3cret-password"
	user := verifiedUser(t, password)
	secret, _ := helpers.GenerateTOTPSecret()
	// Row exists but was never confirmed: enabled = false.
	state := &twoFactorState{
		row:   &TwoFactor{UserID: user.ID, Secret: secret},
		codes: map[string]bool{},
	}
	repo := loginCapableRepo(t, user, state)
	svc := newTestService(repo, newMemStorage())

	result, err := svc.Login(context.Background(), user.Email, password, "", "")
	if err != nil {
		t.Fatalf("Login with pending (unconfirmed) setup: %v", err)
	}
	if result.AccessToken == "" {
		t.Fatal("expected a normal session while setup is unconfirmed")
	}
}
