package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/repositories"
)

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	return string(hash)
}

func verifiedUser(t *testing.T, password string) entities.User {
	t.Helper()
	return entities.User{
		ID:            uuid.New(),
		Name:          "Test User",
		Email:         "test@example.com",
		EmailVerified: true,
		Role:          entities.Role{Name: "user"},
		Accounts:      []entities.Account{{Password: mustHashPassword(t, password)}},
	}
}

func TestGenerateAndHashRefreshToken(t *testing.T) {
	raw, hash, err := generateRefreshToken()
	if err != nil {
		t.Fatalf("generateRefreshToken: %v", err)
	}
	if raw == "" || hash == "" {
		t.Fatal("expected non-empty raw token and hash")
	}

	rehashed, err := hashRefreshToken(raw)
	if err != nil {
		t.Fatalf("hashRefreshToken: %v", err)
	}
	if rehashed != hash {
		t.Errorf("hashRefreshToken(raw) = %q, want %q", rehashed, hash)
	}

	if _, err := hashRefreshToken("not-valid-base64!!!"); err == nil {
		t.Error("expected error for malformed raw token")
	}

	raw2, hash2, err := generateRefreshToken()
	if err != nil {
		t.Fatalf("generateRefreshToken: %v", err)
	}
	if raw2 == raw || hash2 == hash {
		t.Error("expected each generated refresh token to be unique")
	}
}

func TestLoginSuccess(t *testing.T) {
	const password = "s3cret-password"
	user := verifiedUser(t, password)
	sessionID := uuid.New()
	rtID := uuid.New()

	var storedHash string
	repo := &fakeRepository{
		getAccountByEmail: func(_ context.Context, email string) (entities.User, error) {
			if email != user.Email {
				return entities.User{}, errors.New("not found")
			}
			return user, nil
		},
		createSession: func(_ context.Context, userID uuid.UUID, token string, _ time.Time) (uuid.UUID, error) {
			if userID != user.ID {
				t.Errorf("CreateSession userID = %s, want %s", userID, user.ID)
			}
			if token == "" {
				t.Error("CreateSession received empty token")
			}
			return sessionID, nil
		},
		createRefreshToken: func(_ context.Context, userID uuid.UUID, tokenHash string, _, sid uuid.UUID, _, _ *string, _ time.Time) (uuid.UUID, error) {
			if userID != user.ID {
				t.Errorf("CreateRefreshToken userID = %s, want %s", userID, user.ID)
			}
			if sid != sessionID {
				t.Errorf("CreateRefreshToken sessionID = %s, want %s", sid, sessionID)
			}
			storedHash = tokenHash
			return rtID, nil
		},
	}
	storage := newMemStorage()
	svc := newTestServices(repo, storage)

	result, err := svc.Login(context.Background(), user.Email, password)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	if result.ID != user.ID {
		t.Errorf("result.ID = %s, want %s", result.ID, user.ID)
	}

	// The access token must be a valid HS256 JWT carrying id and role.
	parsed, err := jwt.Parse(result.AccessToken, func(*jwt.Token) (any, error) {
		return []byte(testConfig().JWTSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !parsed.Valid {
		t.Fatalf("access token invalid: %v", err)
	}
	claims := parsed.Claims.(jwt.MapClaims)
	if claims["id"] != user.ID.String() {
		t.Errorf("token id claim = %v, want %s", claims["id"], user.ID)
	}
	if claims["role"] != "user" {
		t.Errorf("token role claim = %v, want user", claims["role"])
	}

	// The raw refresh token handed to the client must hash to what was stored.
	gotHash, err := hashRefreshToken(result.RawRefreshToken)
	if err != nil {
		t.Fatalf("hashRefreshToken: %v", err)
	}
	if gotHash != storedHash {
		t.Error("raw refresh token does not match the hash persisted in the repository")
	}

	if !storage.has(refreshCacheKey(storedHash)) {
		t.Error("expected refresh token cache entry after login")
	}
	if !result.RefreshExpiresAt.After(time.Now()) {
		t.Error("refresh token expiry should be in the future")
	}
}

func TestLoginFailures(t *testing.T) {
	const password = "s3cret-password"
	user := verifiedUser(t, password)

	cases := []struct {
		name     string
		email    string
		password string
		user     entities.User
		userErr  error
		wantErr  string
	}{
		{
			name:    "unknown email",
			email:   "missing@example.com",
			userErr: errors.New("no rows"),
			wantErr: "invalid credentials",
		},
		{
			name:     "wrong password",
			email:    user.Email,
			password: "wrong-password",
			user:     user,
			wantErr:  "invalid credentials",
		},
		{
			name:     "unverified email",
			email:    user.Email,
			password: password,
			user: func() entities.User {
				u := user
				u.EmailVerified = false
				return u
			}(),
			wantErr: "invalid account",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepository{
				getAccountByEmail: func(context.Context, string) (entities.User, error) {
					return tc.user, tc.userErr
				},
			}
			svc := newTestServices(repo, newMemStorage())

			_, err := svc.Login(context.Background(), tc.email, tc.password)
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("Login error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestRegister(t *testing.T) {
	t.Run("existing user is rejected", func(t *testing.T) {
		repo := &fakeRepository{
			getUserByEmail: func(context.Context, string) (entities.User, error) {
				return entities.User{Email: "taken@example.com"}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		_, err := svc.Register(context.Background(), "Any Name", "taken@example.com", "password")
		if err == nil || err.Error() != "user existing" {
			t.Fatalf("Register error = %v, want %q", err, "user existing")
		}
	})

	t.Run("stores normalized name and hashed password", func(t *testing.T) {
		const password = "plain-password"
		repo := &fakeRepository{
			getUserByEmail: func(context.Context, string) (entities.User, error) {
				return entities.User{}, errors.New("no rows")
			},
			register: func(_ context.Context, name, email, hashed string) (entities.User, error) {
				if name != "John Doe" {
					t.Errorf("register name = %q, want %q", name, "John Doe")
				}
				if hashed == password {
					t.Error("password must not be stored in plain text")
				}
				if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
					t.Errorf("stored hash does not match password: %v", err)
				}
				return entities.User{Name: name, Email: email}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		result, err := svc.Register(context.Background(), "  john DOE ", "john@example.com", password)
		if err != nil {
			t.Fatalf("Register: %v", err)
		}
		if result.Name != "John Doe" || result.Email != "john@example.com" {
			t.Errorf("unexpected response: %+v", result)
		}
	})
}

func TestValidateToken(t *testing.T) {
	user := entities.User{
		ID:   uuid.New(),
		Role: entities.Role{Name: "user"},
	}

	newSvc := func(sessionUser entities.User, sessionErr error) (*Services, *memStorage, *int) {
		calls := 0
		repo := &fakeRepository{
			getSessionByToken: func(context.Context, string) (entities.User, error) {
				calls++
				return sessionUser, sessionErr
			},
		}
		storage := newMemStorage()
		return newTestServices(repo, storage), storage, &calls
	}

	signToken := func(t *testing.T, svc *Services, exp time.Time) string {
		t.Helper()
		token, err := svc.CreateJWToken(user.ID, user.Role.Name, exp)
		if err != nil {
			t.Fatalf("CreateJWToken: %v", err)
		}
		return token
	}

	t.Run("valid token passes and is cached", func(t *testing.T) {
		sessionUser := user
		svc, storage, calls := newSvc(sessionUser, nil)
		token := signToken(t, svc, time.Now().UTC().Add(10*time.Minute))
		sessionUser.Sessions = []entities.Session{{Token: token}}
		svc, storage, calls = newSvc(sessionUser, nil)

		got, err := svc.ValidateToken(context.Background(), token)
		if err != nil {
			t.Fatalf("ValidateToken: %v", err)
		}
		if got != token {
			t.Errorf("ValidateToken returned %q, want the same token", got)
		}
		if !storage.has(validateTokenCacheKey(token)) {
			t.Error("expected token validation to be cached")
		}

		// Second call must be served from cache without hitting the repository.
		if _, err := svc.ValidateToken(context.Background(), token); err != nil {
			t.Fatalf("cached ValidateToken: %v", err)
		}
		if *calls != 1 {
			t.Errorf("repository hit %d times, want 1 (second call should use cache)", *calls)
		}
	})

	t.Run("cached negative result is rejected", func(t *testing.T) {
		svc, storage, _ := newSvc(entities.User{}, errors.New("must not be called"))
		storage.Set(validateTokenCacheKey("some-token"), []byte("false"), time.Minute)

		if _, err := svc.ValidateToken(context.Background(), "some-token"); err == nil {
			t.Fatal("expected cached invalid token to be rejected")
		}
	})

	t.Run("wrong signature is rejected", func(t *testing.T) {
		svc, _, _ := newSvc(user, nil)
		forged := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"id":   user.ID.String(),
			"role": "admin",
			"exp":  time.Now().Add(time.Hour).Unix(),
		})
		token, err := forged.SignedString([]byte("other-secret"))
		if err != nil {
			t.Fatalf("sign: %v", err)
		}

		if _, err := svc.ValidateToken(context.Background(), token); err == nil {
			t.Fatal("expected token signed with a different secret to be rejected")
		}
	})

	t.Run("unexpected signing algorithm is rejected", func(t *testing.T) {
		svc, _, _ := newSvc(user, nil)
		forged := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"id":   user.ID.String(),
			"role": "user",
			"exp":  time.Now().Add(time.Hour).Unix(),
		})
		token, err := forged.SignedString([]byte(testConfig().JWTSecret))
		if err != nil {
			t.Fatalf("sign: %v", err)
		}

		if _, err := svc.ValidateToken(context.Background(), token); err == nil {
			t.Fatal("expected non-HS256 token to be rejected")
		}
	})

	t.Run("token without matching session is rejected", func(t *testing.T) {
		svc, _, _ := newSvc(entities.User{}, errors.New("no session"))
		token := signToken(t, svc, time.Now().UTC().Add(10*time.Minute))

		if _, err := svc.ValidateToken(context.Background(), token); err == nil {
			t.Fatal("expected token without session to be rejected")
		}
	})

	t.Run("session token mismatch is rejected", func(t *testing.T) {
		sessionUser := user
		sessionUser.Sessions = []entities.Session{{Token: "a-different-token"}}
		svc, _, _ := newSvc(sessionUser, nil)
		token := signToken(t, svc, time.Now().UTC().Add(10*time.Minute))

		if _, err := svc.ValidateToken(context.Background(), token); err == nil {
			t.Fatal("expected token that does not match the stored session to be rejected")
		}
	})

	t.Run("role mismatch is rejected", func(t *testing.T) {
		svcForSigning := newTestServices(&fakeRepository{}, newMemStorage())
		token, err := svcForSigning.CreateJWToken(user.ID, "admin", time.Now().UTC().Add(10*time.Minute))
		if err != nil {
			t.Fatalf("CreateJWToken: %v", err)
		}
		sessionUser := user // DB role is "user", token claims "admin"
		sessionUser.Sessions = []entities.Session{{Token: token}}
		svc, _, _ := newSvc(sessionUser, nil)

		if _, err := svc.ValidateToken(context.Background(), token); err == nil {
			t.Fatal("expected role escalation in token to be rejected")
		}
	})
}

// primedRefreshFixture wires a service whose cache already holds a valid
// refresh token entry, mimicking the state right after a login.
type refreshFixture struct {
	svc       *Services
	storage   *memStorage
	raw       string
	hash      string
	userID    uuid.UUID
	familyID  uuid.UUID
	sessionID uuid.UUID

	markedUsed     []uuid.UUID
	revoked        []uuid.UUID
	createdHashes  []string
	sessionUpdated bool
}

func newRefreshFixture(t *testing.T) *refreshFixture {
	t.Helper()

	raw, hash, err := generateRefreshToken()
	if err != nil {
		t.Fatalf("generateRefreshToken: %v", err)
	}

	f := &refreshFixture{
		raw:       raw,
		hash:      hash,
		userID:    uuid.New(),
		familyID:  uuid.New(),
		sessionID: uuid.New(),
		storage:   newMemStorage(),
	}

	repo := &fakeRepository{
		markRefreshTokenUsed: func(_ context.Context, id uuid.UUID) error {
			f.markedUsed = append(f.markedUsed, id)
			return nil
		},
		updateSessionToken: func(_ context.Context, sessionID uuid.UUID, _ string, _ time.Time) (string, error) {
			if sessionID != f.sessionID {
				t.Errorf("UpdateSessionToken sessionID = %s, want %s", sessionID, f.sessionID)
			}
			f.sessionUpdated = true
			return "old-access-token", nil
		},
		createRefreshToken: func(_ context.Context, _ uuid.UUID, tokenHash string, familyID, _ uuid.UUID, _, _ *string, _ time.Time) (uuid.UUID, error) {
			if familyID != f.familyID {
				t.Errorf("CreateRefreshToken familyID = %s, want %s (rotation must stay in the family)", familyID, f.familyID)
			}
			f.createdHashes = append(f.createdHashes, tokenHash)
			return uuid.New(), nil
		},
		revokeRefreshTokenFamily: func(_ context.Context, familyID uuid.UUID) ([]string, error) {
			f.revoked = append(f.revoked, familyID)
			return nil, nil
		},
	}
	f.svc = newTestServices(repo, f.storage)
	return f
}

func (f *refreshFixture) primeCache(t *testing.T, expiresAt time.Time) {
	t.Helper()
	tokenID := uuid.New()
	value := fmt.Sprintf("%s|%s|user|%s|%s|%d", tokenID, f.userID, f.familyID, f.sessionID, expiresAt.Unix())
	if err := f.storage.Set(refreshCacheKey(f.hash), []byte(value), time.Hour); err != nil {
		t.Fatalf("prime cache: %v", err)
	}
}

func (f *refreshFixture) repo() *fakeRepository {
	return f.svc.repos.(*fakeRepository)
}

func TestRefreshTokenRotationFromCache(t *testing.T) {
	f := newRefreshFixture(t)
	f.primeCache(t, time.Now().UTC().Add(time.Hour))

	result, err := f.svc.RefreshToken(context.Background(), f.raw, "1.2.3.4", "test-agent")
	if err != nil {
		t.Fatalf("RefreshToken: %v", err)
	}

	if result.AccessToken == "" || result.RawRefreshToken == "" {
		t.Fatal("expected new access and refresh tokens")
	}
	if result.RawRefreshToken == f.raw {
		t.Error("refresh token must rotate, got the same raw token back")
	}
	if len(f.markedUsed) != 1 {
		t.Errorf("MarkRefreshTokenUsed called %d times, want 1", len(f.markedUsed))
	}
	if !f.sessionUpdated {
		t.Error("expected session to be updated with the new access token")
	}
	if f.storage.has(refreshCacheKey(f.hash)) {
		t.Error("old refresh cache entry must be deleted after rotation")
	}
	if len(f.createdHashes) != 1 {
		t.Fatalf("CreateRefreshToken called %d times, want 1", len(f.createdHashes))
	}
	if !f.storage.has(refreshCacheKey(f.createdHashes[0])) {
		t.Error("new refresh token must be cached")
	}
	newHash, err := hashRefreshToken(result.RawRefreshToken)
	if err != nil {
		t.Fatalf("hashRefreshToken: %v", err)
	}
	if newHash != f.createdHashes[0] {
		t.Error("returned raw refresh token does not match the persisted hash")
	}
	if len(f.revoked) != 0 {
		t.Errorf("no family should be revoked on a clean rotation, got %v", f.revoked)
	}
}

func TestRefreshTokenRejectsMalformedAndExpired(t *testing.T) {
	t.Run("malformed raw token", func(t *testing.T) {
		f := newRefreshFixture(t)
		if _, err := f.svc.RefreshToken(context.Background(), "!!!not-base64!!!", "", ""); err == nil {
			t.Fatal("expected malformed refresh token to be rejected")
		}
	})

	t.Run("expired cache entry", func(t *testing.T) {
		f := newRefreshFixture(t)
		f.primeCache(t, time.Now().UTC().Add(-time.Minute))

		if _, err := f.svc.RefreshToken(context.Background(), f.raw, "", ""); err == nil {
			t.Fatal("expected expired refresh token to be rejected")
		}
		if len(f.markedUsed) != 0 {
			t.Error("expired token must not be marked as used")
		}
	})

	t.Run("revoked family marker blocks cached token", func(t *testing.T) {
		f := newRefreshFixture(t)
		f.primeCache(t, time.Now().UTC().Add(time.Hour))
		f.storage.Set(revokedFamilyCacheKey(f.familyID), []byte("1"), time.Hour)

		if _, err := f.svc.RefreshToken(context.Background(), f.raw, "", ""); err == nil {
			t.Fatal("expected token of a revoked family to be rejected")
		}
		if len(f.markedUsed) != 0 {
			t.Error("revoked-family token must not be marked as used")
		}
	})
}

func TestRefreshTokenReuseDetection(t *testing.T) {
	t.Run("reuse outside grace period revokes the family", func(t *testing.T) {
		f := newRefreshFixture(t)
		usedAt := time.Now().UTC().Add(-time.Hour) // way past the 30s grace period
		f.repo().getRefreshTokenByHash = func(context.Context, string) (entities.RefreshToken, error) {
			return entities.RefreshToken{
				ID:        uuid.New(),
				UserID:    f.userID,
				FamilyID:  f.familyID,
				SessionID: f.sessionID,
				Role:      "user",
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				UsedAt:    &usedAt,
			}, nil
		}

		if _, err := f.svc.RefreshToken(context.Background(), f.raw, "", ""); err == nil {
			t.Fatal("expected reused refresh token to be rejected")
		}
		if len(f.revoked) != 1 || f.revoked[0] != f.familyID {
			t.Fatalf("expected family %s to be revoked, got %v", f.familyID, f.revoked)
		}
		if !f.storage.has(revokedFamilyCacheKey(f.familyID)) {
			t.Error("expected revoked-family marker in cache")
		}
	})

	t.Run("reuse within grace period is treated as benign", func(t *testing.T) {
		f := newRefreshFixture(t)
		usedAt := time.Now().UTC().Add(-time.Second)
		f.repo().getRefreshTokenByHash = func(context.Context, string) (entities.RefreshToken, error) {
			return entities.RefreshToken{
				ID:        uuid.New(),
				UserID:    f.userID,
				FamilyID:  f.familyID,
				SessionID: f.sessionID,
				Role:      "user",
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				UsedAt:    &usedAt,
			}, nil
		}

		if _, err := f.svc.RefreshToken(context.Background(), f.raw, "", ""); err != nil {
			t.Fatalf("RefreshToken within grace period: %v", err)
		}
		if len(f.revoked) != 0 {
			t.Errorf("benign concurrent refresh must not revoke the family, got %v", f.revoked)
		}
	})

	t.Run("revoked token from the database is rejected", func(t *testing.T) {
		f := newRefreshFixture(t)
		revokedAt := time.Now().UTC()
		f.repo().getRefreshTokenByHash = func(context.Context, string) (entities.RefreshToken, error) {
			return entities.RefreshToken{
				ID:        uuid.New(),
				FamilyID:  f.familyID,
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				RevokedAt: &revokedAt,
			}, nil
		}

		if _, err := f.svc.RefreshToken(context.Background(), f.raw, "", ""); err == nil {
			t.Fatal("expected revoked refresh token to be rejected")
		}
	})

	t.Run("expired token from the database is rejected", func(t *testing.T) {
		f := newRefreshFixture(t)
		f.repo().getRefreshTokenByHash = func(context.Context, string) (entities.RefreshToken, error) {
			return entities.RefreshToken{
				ID:        uuid.New(),
				FamilyID:  f.familyID,
				ExpiresAt: time.Now().UTC().Add(-time.Minute),
			}, nil
		}

		if _, err := f.svc.RefreshToken(context.Background(), f.raw, "", ""); err == nil {
			t.Fatal("expected expired refresh token to be rejected")
		}
	})
}

func TestRefreshTokenRevokesOrphanedFamilyWhenSessionGone(t *testing.T) {
	f := newRefreshFixture(t)
	f.primeCache(t, time.Now().UTC().Add(time.Hour))
	f.repo().updateSessionToken = func(context.Context, uuid.UUID, string, time.Time) (string, error) {
		return "", repositories.ErrSessionNotFound
	}

	if _, err := f.svc.RefreshToken(context.Background(), f.raw, "", ""); err == nil {
		t.Fatal("expected refresh with a deleted session to fail")
	}
	if len(f.revoked) != 1 || f.revoked[0] != f.familyID {
		t.Fatalf("expected orphaned family %s to be revoked, got %v", f.familyID, f.revoked)
	}
}

func TestLogout(t *testing.T) {
	userID := uuid.New()
	familyID := uuid.New()
	accessToken := "access-token"
	raw, hash, err := generateRefreshToken()
	if err != nil {
		t.Fatalf("generateRefreshToken: %v", err)
	}

	var deletedSession bool
	repo := &fakeRepository{
		getRefreshTokenFamiliesBySession: func(_ context.Context, uid uuid.UUID, token string) ([]string, []uuid.UUID, error) {
			if uid != userID || token != accessToken {
				t.Errorf("unexpected session lookup: %s %s", uid, token)
			}
			return []string{hash}, []uuid.UUID{familyID, familyID}, nil
		},
		deleteSessionByUserIDToken: func(_ context.Context, uid uuid.UUID, token string) error {
			deletedSession = uid == userID && token == accessToken
			return nil
		},
	}
	storage := newMemStorage()
	storage.Set(validateTokenCacheKey(accessToken), []byte("true"), time.Hour)
	storage.Set(refreshCacheKey(hash), []byte("cached"), time.Hour)
	svc := newTestServices(repo, storage)

	if err := svc.Logout(context.Background(), userID, accessToken, raw); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	if !deletedSession {
		t.Error("expected the session row to be deleted")
	}
	if storage.has(validateTokenCacheKey(accessToken)) {
		t.Error("access token must no longer be cached as valid after logout")
	}
	if storage.has(refreshCacheKey(hash)) {
		t.Error("refresh token cache entry must be purged on logout")
	}
	if !storage.has(revokedFamilyCacheKey(familyID)) {
		t.Error("refresh token family must be marked revoked on logout")
	}
}

func TestLogoutPropagatesSessionDeleteError(t *testing.T) {
	repo := &fakeRepository{
		getRefreshTokenFamiliesBySession: func(context.Context, uuid.UUID, string) ([]string, []uuid.UUID, error) {
			return nil, nil, errors.New("db down")
		},
		deleteSessionByUserIDToken: func(context.Context, uuid.UUID, string) error {
			return errors.New("delete failed")
		},
	}
	svc := newTestServices(repo, newMemStorage())

	err := svc.Logout(context.Background(), uuid.New(), "token", "")
	if err == nil || !strings.Contains(err.Error(), "delete failed") {
		t.Fatalf("Logout error = %v, want delete failure to propagate", err)
	}
}

func TestCleanupExpiredAuth(t *testing.T) {
	t.Run("returns both counters", func(t *testing.T) {
		repo := &fakeRepository{
			deleteExpiredRefreshTokens: func(context.Context) (int64, error) { return 7, nil },
			deleteExpiredSessions:      func(context.Context) (int64, error) { return 3, nil },
		}
		svc := newTestServices(repo, newMemStorage())

		sessions, refreshTokens, err := svc.CleanupExpiredAuth(context.Background())
		if err != nil {
			t.Fatalf("CleanupExpiredAuth: %v", err)
		}
		if sessions != 3 || refreshTokens != 7 {
			t.Errorf("got sessions=%d refreshTokens=%d, want 3 and 7", sessions, refreshTokens)
		}
	})

	t.Run("propagates refresh token cleanup error", func(t *testing.T) {
		repo := &fakeRepository{
			deleteExpiredRefreshTokens: func(context.Context) (int64, error) { return 0, errors.New("boom") },
		}
		svc := newTestServices(repo, newMemStorage())

		if _, _, err := svc.CleanupExpiredAuth(context.Background()); err == nil {
			t.Fatal("expected error from refresh token cleanup")
		}
	})
}
