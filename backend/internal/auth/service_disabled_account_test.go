package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/identity"
)

// Banning and soft-deleting a user used to write a column and nothing more:
// no path in this module consulted banned_at, GetSessionByToken filtered
// neither flag, and RefreshToken never read the users table at all. The result
// was that a banned account could log in as usual, that every access token it
// already held kept validating, and that its refresh cookie kept minting new
// ones for the whole 30-day family lifetime. These tests pin each of the three
// paths that now enforce the state.

func bannedUser(t *testing.T, password string) identity.User {
	t.Helper()

	user := verifiedUser(t, password)
	bannedAt := time.Now().UTC().Add(-time.Hour)
	user.BannedAt = &bannedAt

	return user
}

// TestLoginRejectsBannedAccount covers the way in.
func TestLoginRejectsBannedAccount(t *testing.T) {
	const password = "s3cret-password"
	user := bannedUser(t, password)

	var sessionCreated bool
	repo := new(fakeRepository{
		getAccountByEmail: func(_ context.Context, email string) (identity.User, error) {
			if email != user.Email {
				return identity.User{}, errors.New("not found")
			}
			return user, nil
		},
		createSession: func(context.Context, uuid.UUID, string, *string, *string, time.Time) (uuid.UUID, error) {
			sessionCreated = true
			return uuid.New(), nil
		},
	})

	svc := newTestService(repo, newMemStorage())

	_, err := svc.Login(context.Background(), user.Email, password, "", "")
	if !errors.Is(err, ErrAccountDisabled) {
		t.Fatalf("Login of a banned account = %v, want ErrAccountDisabled", err)
	}
	if sessionCreated {
		t.Error("a banned account must not be issued a session")
	}
}

// TestLoginRejectsBannedAccountOnlyAfterThePassword keeps the ban from
// becoming an oracle: an anonymous caller with the wrong password must get the
// same generic failure whether or not the address is banned, so the endpoint
// never confirms which accounts exist or which have been suspended.
func TestLoginRejectsBannedAccountOnlyAfterThePassword(t *testing.T) {
	const password = "s3cret-password"
	user := bannedUser(t, password)

	repo := new(fakeRepository{
		getAccountByEmail: func(context.Context, string) (identity.User, error) { return user, nil },
	})
	svc := newTestService(repo, newMemStorage())

	_, err := svc.Login(context.Background(), user.Email, "wrong-password", "", "")
	if errors.Is(err, ErrAccountDisabled) {
		t.Fatal("a wrong password must not reveal that the account is banned")
	}
	if err == nil {
		t.Fatal("expected a wrong password to be rejected")
	}
}

// TestLoginAllowsActiveAccount is the counterpart: the new check must not
// reject an ordinary user.
func TestLoginAllowsActiveAccount(t *testing.T) {
	const password = "s3cret-password"
	user := verifiedUser(t, password)

	repo := new(fakeRepository{
		getAccountByEmail: func(context.Context, string) (identity.User, error) { return user, nil },
		createSession: func(context.Context, uuid.UUID, string, *string, *string, time.Time) (uuid.UUID, error) {
			return uuid.New(), nil
		},
		createRefreshToken: func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, *string, *string, time.Time) (uuid.UUID, error) {
			return uuid.New(), nil
		},
	})
	svc := newTestService(repo, newMemStorage())

	if _, err := svc.Login(context.Background(), user.Email, password, "", ""); err != nil {
		t.Fatalf("Login of an active account: %v", err)
	}
}

// TestRefreshTokenRejectsDisabledAccount is the important one. The cached
// branch of RefreshToken performs no database work at all, so before the fix a
// banned or deleted user's cookie kept rotating into fresh access tokens for
// JWT_REFRESH_DURATION — 30 days by default — no matter what the other paths
// did.
func TestRefreshTokenRejectsDisabledAccount(t *testing.T) {
	disabled := map[string]func(u *identity.User){
		"banned":       func(u *identity.User) { at := time.Now().UTC(); u.BannedAt = &at },
		"soft-deleted": func(u *identity.User) { at := time.Now().UTC(); u.DeletedAt = &at },
	}

	for name, disable := range disabled {
		t.Run(name, func(t *testing.T) {
			f := newRefreshFixture(t)
			f.primeCache(t, time.Now().UTC().Add(time.Hour))

			f.repo().getUserByID = func(_ context.Context, id uuid.UUID) (identity.User, error) {
				user := identity.User{ID: id}
				disable(&user)
				return user, nil
			}

			if _, err := f.svc.RefreshToken(context.Background(), f.raw, "", ""); err == nil {
				t.Fatal("expected the refresh token of a disabled account to be rejected")
			}

			if len(f.markedUsed) != 0 {
				t.Error("a rejected refresh must not rotate the token")
			}
			if len(f.createdHashes) != 0 {
				t.Error("a rejected refresh must not issue a new refresh token")
			}
			if f.sessionUpdated {
				t.Error("a rejected refresh must not hand the session a new access token")
			}

			// Revoking the family is what stops the cookie being re-presented
			// every fifteen minutes for the rest of its lifetime.
			if len(f.revoked) != 1 || f.revoked[0] != f.familyID {
				t.Errorf("revoked families = %v, want the token's own family %s", f.revoked, f.familyID)
			}
		})
	}
}

// TestRefreshTokenFailsClosedWhenAccountLookupFails pins the direction the
// check errs in: a database hiccup must not be treated as "not disabled".
func TestRefreshTokenFailsClosedWhenAccountLookupFails(t *testing.T) {
	f := newRefreshFixture(t)
	f.primeCache(t, time.Now().UTC().Add(time.Hour))

	f.repo().getUserByID = func(context.Context, uuid.UUID) (identity.User, error) {
		return identity.User{}, errors.New("database unavailable")
	}

	if _, err := f.svc.RefreshToken(context.Background(), f.raw, "", ""); err == nil {
		t.Fatal("expected the refresh to fail when the account state cannot be read")
	}
	if f.sessionUpdated {
		t.Error("no new access token may be issued when the account state is unknown")
	}
}

// TestRevokeAllSessionsClosesEverySession covers the capability the user module
// calls on ban and on delete. Without it the flag would only take effect at the
// next token expiry.
func TestRevokeAllSessionsClosesEverySession(t *testing.T) {
	userID := uuid.New()
	sessions := []identity.Session{
		{ID: uuid.New(), UserID: userID, Token: "access-token-1"},
		{ID: uuid.New(), UserID: userID, Token: "access-token-2"},
	}
	familyID := uuid.New()

	var deleted []uuid.UUID
	repo := new(fakeRepository{
		listSessionsByUserID: func(context.Context, uuid.UUID) ([]identity.Session, error) {
			return sessions, nil
		},
		getRefreshTokensBySessionIDs: func(context.Context, uuid.UUID, []uuid.UUID) ([]string, []uuid.UUID, error) {
			return []string{"refresh-hash-1"}, []uuid.UUID{familyID}, nil
		},
		deleteSessionsByIDs: func(_ context.Context, _ uuid.UUID, ids []uuid.UUID) (int64, error) {
			deleted = ids
			return int64(len(ids)), nil
		},
	})

	storage := newMemStorage()
	svc := newTestService(repo, storage)

	// Both sessions' access tokens are cached as valid, and the refresh token
	// is cached too — all of which must be gone afterwards.
	for _, sess := range sessions {
		_ = storage.Set(validateTokenCacheKey(sess.Token), []byte("true"), time.Hour)
	}
	_ = storage.Set(refreshCacheKey("refresh-hash-1"), []byte("cached"), time.Hour)

	count, err := svc.RevokeAllSessions(context.Background(), userID)
	if err != nil {
		t.Fatalf("RevokeAllSessions: %v", err)
	}
	if count != int64(len(sessions)) {
		t.Errorf("RevokeAllSessions = %d, want %d", count, len(sessions))
	}
	if len(deleted) != len(sessions) {
		t.Errorf("deleted %d sessions, want %d", len(deleted), len(sessions))
	}

	for _, sess := range sessions {
		if storage.has(validateTokenCacheKey(sess.Token)) {
			t.Errorf("access token %q is still cached as valid after revocation", sess.Token)
		}
	}
	if storage.has(refreshCacheKey("refresh-hash-1")) {
		t.Error("the refresh token is still cached after revocation")
	}
	if !storage.has(revokedFamilyCacheKey(familyID)) {
		t.Error("the refresh family must be marked revoked so a cached token cannot outlive the session")
	}
}

func TestRevokeAllSessionsWithNoSessions(t *testing.T) {
	repo := new(fakeRepository{
		listSessionsByUserID: func(context.Context, uuid.UUID) ([]identity.Session, error) {
			return nil, nil
		},
	})

	count, err := newTestService(repo, newMemStorage()).RevokeAllSessions(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("RevokeAllSessions: %v", err)
	}
	if count != 0 {
		t.Errorf("RevokeAllSessions = %d, want 0", count)
	}
}
