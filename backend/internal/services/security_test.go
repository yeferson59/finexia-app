package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/logger"
)

// newLockoutServices builds Services with the login lockout enabled, which
// testConfig leaves disabled so unrelated tests never trip it.
func newLockoutServices(repo Repository, storage *memStorage, maxAttempts int) *Services {
	cfg := testConfig()
	cfg.MaxLoginAttempts = maxAttempts
	cfg.LoginLockout = time.Minute
	svc := New(repo, cfg, nil, storage, nil, nil, logger.Noop(), nil)
	return &svc
}

func TestLoginLockout(t *testing.T) {
	const password = "s3cret-password"
	user := verifiedUser(t, password)

	sessionRepo := func() *fakeRepository {
		return &fakeRepository{
			getAccountByEmail: func(_ context.Context, email string) (entities.User, error) {
				if email != user.Email {
					return entities.User{}, errors.New("no rows")
				}
				return user, nil
			},
			createSession: func(context.Context, uuid.UUID, string, *string, *string, time.Time) (uuid.UUID, error) {
				return uuid.New(), nil
			},
			createRefreshToken: func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, *string, *string, time.Time) (uuid.UUID, error) {
				return uuid.New(), nil
			},
		}
	}

	t.Run("locks after max failed attempts even with the right password", func(t *testing.T) {
		svc := newLockoutServices(sessionRepo(), newMemStorage(), 3)

		for range 3 {
			if _, err := svc.Login(context.Background(), user.Email, "wrong-password", "", ""); err == nil || err.Error() != "invalid credentials" {
				t.Fatalf("Login error = %v, want %q", err, "invalid credentials")
			}
		}

		_, err := svc.Login(context.Background(), user.Email, password, "", "")
		if err == nil || err.Error() != "too many failed login attempts" {
			t.Fatalf("Login error = %v, want lockout", err)
		}
	})

	t.Run("unknown emails also accumulate failures", func(t *testing.T) {
		svc := newLockoutServices(sessionRepo(), newMemStorage(), 2)

		for range 2 {
			if _, err := svc.Login(context.Background(), "ghost@example.com", "whatever-pass", "", ""); err == nil || err.Error() != "invalid credentials" {
				t.Fatalf("Login error = %v, want %q", err, "invalid credentials")
			}
		}

		_, err := svc.Login(context.Background(), "ghost@example.com", "whatever-pass", "", "")
		if err == nil || err.Error() != "too many failed login attempts" {
			t.Fatalf("Login error = %v, want lockout", err)
		}
	})

	t.Run("successful login clears the counter", func(t *testing.T) {
		svc := newLockoutServices(sessionRepo(), newMemStorage(), 3)

		for range 2 {
			if _, err := svc.Login(context.Background(), user.Email, "wrong-password", "", ""); err == nil {
				t.Fatal("expected failed login")
			}
		}

		if _, err := svc.Login(context.Background(), user.Email, password, "", ""); err != nil {
			t.Fatalf("Login after 2 failures should succeed: %v", err)
		}

		// The counter was reset: two more failures must not lock yet.
		for range 2 {
			if _, err := svc.Login(context.Background(), user.Email, "wrong-password", "", ""); err == nil || err.Error() != "invalid credentials" {
				t.Fatalf("Login error = %v, want %q (counter should have been reset)", err, "invalid credentials")
			}
		}
	})

	t.Run("disabled when max attempts is zero", func(t *testing.T) {
		svc := newLockoutServices(sessionRepo(), newMemStorage(), 0)

		for range 10 {
			if _, err := svc.Login(context.Background(), user.Email, "wrong-password", "", ""); err == nil || err.Error() != "invalid credentials" {
				t.Fatalf("Login error = %v, want %q", err, "invalid credentials")
			}
		}
	})
}

func TestLoginRecordsIPAndAlertsOnUnknownIP(t *testing.T) {
	const password = "s3cret-password"
	user := verifiedUser(t, password)

	var gotIP, gotUA *string
	repo := &fakeRepository{
		getAccountByEmail: func(context.Context, string) (entities.User, error) {
			return user, nil
		},
		hasKnownLoginIP: func(_ context.Context, _ uuid.UUID, ip string) (bool, error) {
			return false, nil
		},
		createSession: func(_ context.Context, _ uuid.UUID, _ string, ip, ua *string, _ time.Time) (uuid.UUID, error) {
			gotIP, gotUA = ip, ua
			return uuid.New(), nil
		},
		createRefreshToken: func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, *string, *string, time.Time) (uuid.UUID, error) {
			return uuid.New(), nil
		},
	}
	mailer := &fakeMailer{}
	svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)

	if _, err := svc.Login(context.Background(), user.Email, password, "203.0.113.7", "TestAgent/1.0"); err != nil {
		t.Fatalf("Login: %v", err)
	}

	if gotIP == nil || *gotIP != "203.0.113.7" {
		t.Errorf("CreateSession ip = %v, want 203.0.113.7", gotIP)
	}
	if gotUA == nil || *gotUA != "TestAgent/1.0" {
		t.Errorf("CreateSession ua = %v, want TestAgent/1.0", gotUA)
	}

	// The alert is sent from a goroutine; wait for it.
	ok := waitFor(t, 2*time.Second, func() bool {
		mailer.mu.Lock()
		defer mailer.mu.Unlock()
		return len(mailer.security) == 1
	})
	if !ok {
		t.Fatal("expected a security alert email for a login from an unknown IP")
	}

	mailer.mu.Lock()
	defer mailer.mu.Unlock()
	if mailer.security[0].To != user.Email {
		t.Errorf("alert sent to %s, want %s", mailer.security[0].To, user.Email)
	}
	if mailer.security[0].Data.IPAddress != "203.0.113.7" {
		t.Errorf("alert IP = %s, want 203.0.113.7", mailer.security[0].Data.IPAddress)
	}
}

func TestLoginDoesNotAlertOnKnownIP(t *testing.T) {
	const password = "s3cret-password"
	user := verifiedUser(t, password)

	repo := &fakeRepository{
		getAccountByEmail: func(context.Context, string) (entities.User, error) {
			return user, nil
		},
		hasKnownLoginIP: func(context.Context, uuid.UUID, string) (bool, error) {
			return true, nil
		},
		createSession: func(context.Context, uuid.UUID, string, *string, *string, time.Time) (uuid.UUID, error) {
			return uuid.New(), nil
		},
		createRefreshToken: func(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, *string, *string, time.Time) (uuid.UUID, error) {
			return uuid.New(), nil
		},
	}
	mailer := &fakeMailer{}
	svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)

	if _, err := svc.Login(context.Background(), user.Email, password, "203.0.113.7", "TestAgent/1.0"); err != nil {
		t.Fatalf("Login: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	mailer.mu.Lock()
	defer mailer.mu.Unlock()
	if len(mailer.security) != 0 {
		t.Fatalf("expected no security alert for a known IP, got %d", len(mailer.security))
	}
}

func sessionFixture(userID uuid.UUID, token string) entities.Session {
	ip := "198.51.100.1"
	ua := "Agent/1.0"
	return entities.Session{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour),
		IPAddress: &ip,
		UserAgent: &ua,
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}
}

func TestListSessionsMarksCurrent(t *testing.T) {
	userID := uuid.New()
	current := sessionFixture(userID, "current-token")
	other := sessionFixture(userID, "other-token")

	repo := &fakeRepository{
		listSessionsByUserID: func(context.Context, uuid.UUID) ([]entities.Session, error) {
			return []entities.Session{current, other}, nil
		},
	}
	svc := newTestServices(repo, newMemStorage())

	sessions, err := svc.ListSessions(context.Background(), userID, "current-token")
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("got %d sessions, want 2", len(sessions))
	}
	if !sessions[0].Current || sessions[1].Current {
		t.Errorf("current flags = %v/%v, want true/false", sessions[0].Current, sessions[1].Current)
	}
}

func TestRevokeSession(t *testing.T) {
	userID := uuid.New()
	current := sessionFixture(userID, "current-token")
	other := sessionFixture(userID, "other-token")
	familyID := uuid.New()
	const otherHash = "deadbeef"

	newRepo := func(deleted *[]uuid.UUID) *fakeRepository {
		return &fakeRepository{
			listSessionsByUserID: func(context.Context, uuid.UUID) ([]entities.Session, error) {
				return []entities.Session{current, other}, nil
			},
			getRefreshTokensBySessionIDs: func(_ context.Context, _ uuid.UUID, ids []uuid.UUID) ([]string, []uuid.UUID, error) {
				return []string{otherHash}, []uuid.UUID{familyID}, nil
			},
			deleteSessionsByIDs: func(_ context.Context, _ uuid.UUID, ids []uuid.UUID) (int64, error) {
				*deleted = append(*deleted, ids...)
				return int64(len(ids)), nil
			},
		}
	}

	t.Run("revokes another session and purges its caches", func(t *testing.T) {
		var deleted []uuid.UUID
		storage := newMemStorage()
		_ = storage.Set(validateTokenCacheKey("other-token"), []byte("true"), time.Hour)
		_ = storage.Set(refreshCacheKey(otherHash), []byte("x"), time.Hour)
		svc := newTestServices(newRepo(&deleted), storage)

		if err := svc.RevokeSession(context.Background(), userID, other.ID, "current-token"); err != nil {
			t.Fatalf("RevokeSession: %v", err)
		}

		if len(deleted) != 1 || deleted[0] != other.ID {
			t.Errorf("deleted sessions = %v, want [%s]", deleted, other.ID)
		}
		if storage.has(validateTokenCacheKey("other-token")) {
			t.Error("expected access-token cache entry to be purged")
		}
		if storage.has(refreshCacheKey(otherHash)) {
			t.Error("expected refresh-token cache entry to be purged")
		}
		if !storage.has(revokedFamilyCacheKey(familyID)) {
			t.Error("expected revoked-family marker to be set")
		}
	})

	t.Run("rejects revoking the current session", func(t *testing.T) {
		var deleted []uuid.UUID
		svc := newTestServices(newRepo(&deleted), newMemStorage())

		err := svc.RevokeSession(context.Background(), userID, current.ID, "current-token")
		if err == nil || err.Error() != "invalid session: use logout to close the current session" {
			t.Fatalf("error = %v, want current-session rejection", err)
		}
		if len(deleted) != 0 {
			t.Error("current session must not be deleted")
		}
	})

	t.Run("unknown session id", func(t *testing.T) {
		var deleted []uuid.UUID
		svc := newTestServices(newRepo(&deleted), newMemStorage())

		err := svc.RevokeSession(context.Background(), userID, uuid.New(), "current-token")
		if err == nil || err.Error() != "not found session" {
			t.Fatalf("error = %v, want %q", err, "not found session")
		}
	})
}

func TestRevokeOtherSessions(t *testing.T) {
	userID := uuid.New()
	current := sessionFixture(userID, "current-token")
	otherA := sessionFixture(userID, "token-a")
	otherB := sessionFixture(userID, "token-b")

	var deleted []uuid.UUID
	repo := &fakeRepository{
		listSessionsByUserID: func(context.Context, uuid.UUID) ([]entities.Session, error) {
			return []entities.Session{current, otherA, otherB}, nil
		},
		getRefreshTokensBySessionIDs: func(context.Context, uuid.UUID, []uuid.UUID) ([]string, []uuid.UUID, error) {
			return nil, nil, nil
		},
		deleteSessionsByIDs: func(_ context.Context, _ uuid.UUID, ids []uuid.UUID) (int64, error) {
			deleted = append(deleted, ids...)
			return int64(len(ids)), nil
		},
	}
	storage := newMemStorage()
	_ = storage.Set(validateTokenCacheKey("current-token"), []byte("true"), time.Hour)
	svc := newTestServices(repo, storage)

	revoked, err := svc.RevokeOtherSessions(context.Background(), userID, "current-token")
	if err != nil {
		t.Fatalf("RevokeOtherSessions: %v", err)
	}
	if revoked != 2 {
		t.Errorf("revoked = %d, want 2", revoked)
	}
	if len(deleted) != 2 {
		t.Errorf("deleted %d sessions, want 2", len(deleted))
	}
	for _, id := range deleted {
		if id == current.ID {
			t.Error("current session must survive RevokeOtherSessions")
		}
	}
	if !storage.has(validateTokenCacheKey("current-token")) {
		t.Error("current session's cache entry must not be purged")
	}
}

func TestChangePasswordRevokesOtherSessionsAndAlerts(t *testing.T) {
	userID := uuid.New()
	const currentPassword = "current-password"
	current := sessionFixture(userID, "current-token")
	other := sessionFixture(userID, "other-token")

	var deleted []uuid.UUID
	repo := &fakeRepository{
		getAccountByUserID: func(context.Context, uuid.UUID) (entities.Account, error) {
			return entities.Account{Password: mustHashPassword(t, currentPassword)}, nil
		},
		updateUserPassword: func(context.Context, uuid.UUID, string) error {
			return nil
		},
		listSessionsByUserID: func(context.Context, uuid.UUID) ([]entities.Session, error) {
			return []entities.Session{current, other}, nil
		},
		getRefreshTokensBySessionIDs: func(context.Context, uuid.UUID, []uuid.UUID) ([]string, []uuid.UUID, error) {
			return nil, nil, nil
		},
		deleteSessionsByIDs: func(_ context.Context, _ uuid.UUID, ids []uuid.UUID) (int64, error) {
			deleted = append(deleted, ids...)
			return int64(len(ids)), nil
		},
		getUserByID: func(context.Context, uuid.UUID) (entities.User, error) {
			return entities.User{ID: userID, Name: "Test User", Email: "test@example.com"}, nil
		},
	}
	mailer := &fakeMailer{}
	svc := newTestServicesFull(repo, newMemStorage(), mailer, nil)

	if err := svc.ChangePassword(context.Background(), userID, "current-token", currentPassword, "brand-new-pass", "203.0.113.9", "test-agent"); err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}

	if len(deleted) != 1 || deleted[0] != other.ID {
		t.Errorf("deleted sessions = %v, want only the other session %s", deleted, other.ID)
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
}
