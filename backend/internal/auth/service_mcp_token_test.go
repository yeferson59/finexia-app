package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"uuid"
)

// storedMCPToken captures what the service handed the store, which is where
// every guarantee about the secret has to be checked: the raw token exists in
// exactly one place (the response) and its hash in exactly one other.
type storedMCPToken struct {
	name      string
	hash      string
	last4     string
	expiresAt *time.Time
}

// TestCreateMCPTokenStoresOnlyTheHash is the central guarantee: the secret is
// returned once and never written down. A future refactor that persisted the
// raw token — for convenience, for a "show again" button — has to fail here.
func TestCreateMCPTokenStoresOnlyTheHash(t *testing.T) {
	var stored storedMCPToken

	userID := uuid.New()
	repo := new(fakeRepository{
		countMCPTokens: func(context.Context, uuid.UUID) (int, error) { return 0, nil },
		createMCPToken: func(_ context.Context, _ uuid.UUID, name, hash, last4 string, expiresAt *time.Time) (MCPToken, error) {
			stored = storedMCPToken{name: name, hash: hash, last4: last4, expiresAt: expiresAt}

			return MCPToken{ID: uuid.New(), Name: name, Last4: last4, ExpiresAt: expiresAt, CreatedAt: time.Now().UTC()}, nil
		},
	})

	token, err := newTestService(repo, newMemStorage()).
		CreateMCPToken(context.Background(), userID, "  Claude Desktop  ", DefaultMCPTokenExpiryDays)
	if err != nil {
		t.Fatalf("CreateMCPToken: %v", err)
	}

	if !strings.HasPrefix(token.Token, mcpTokenPrefix) {
		t.Errorf("token = %q, want the %q prefix so the guard can recognise it", token.Token, mcpTokenPrefix)
	}

	if stored.hash != hashMCPToken(token.Token) {
		t.Error("the stored hash is not the hash of the returned token")
	}

	if strings.Contains(stored.hash, token.Token) || stored.hash == token.Token {
		t.Fatal("the raw token reached the store")
	}

	if stored.name != "Claude Desktop" {
		t.Errorf("stored name = %q, want it trimmed", stored.name)
	}

	if !strings.HasSuffix(token.Token, stored.last4) {
		t.Errorf("last4 %q is not the end of the token", stored.last4)
	}

	if stored.expiresAt == nil {
		t.Fatal("the default lifetime produced a token that never expires")
	}

	if days := time.Until(*stored.expiresAt).Hours() / 24; days < 89 || days > 90.1 {
		t.Errorf("expiry in %.1f days, want the %d-day default", days, DefaultMCPTokenExpiryDays)
	}
}

// TestCreateMCPTokenLifetimes pins the three cases the pointer in the DTO
// exists to keep apart, at the level where they become an instant.
func TestCreateMCPTokenLifetimes(t *testing.T) {
	for _, tc := range []struct {
		name       string
		days       int
		wantExpiry bool
		wantErr    bool
		wantErrIs  error
	}{
		{name: "explicit zero never expires", days: 0, wantExpiry: false},
		{name: "explicit days", days: 30, wantExpiry: true},
		{name: "the maximum is allowed", days: MaxMCPTokenExpiryDays, wantExpiry: true},
		{name: "beyond the maximum is refused", days: MaxMCPTokenExpiryDays + 1, wantErr: true, wantErrIs: ErrMCPTokenExpiryTooLong},
		{name: "negative is refused", days: -1, wantErr: true, wantErrIs: ErrMCPTokenExpiryInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stored storedMCPToken
			repo := new(fakeRepository{
				countMCPTokens: func(context.Context, uuid.UUID) (int, error) { return 0, nil },
				createMCPToken: func(_ context.Context, _ uuid.UUID, name, hash, last4 string, expiresAt *time.Time) (MCPToken, error) {
					stored = storedMCPToken{name: name, hash: hash, last4: last4, expiresAt: expiresAt}

					return MCPToken{ExpiresAt: expiresAt}, nil
				},
			})

			_, err := newTestService(repo, newMemStorage()).
				CreateMCPToken(context.Background(), uuid.New(), "token", tc.days)

			if tc.wantErr {
				if !errors.Is(err, tc.wantErrIs) {
					t.Fatalf("error = %v, want %v", err, tc.wantErrIs)
				}

				return
			}

			if err != nil {
				t.Fatalf("CreateMCPToken: %v", err)
			}

			if (stored.expiresAt != nil) != tc.wantExpiry {
				t.Errorf("expiresAt = %v, wantExpiry = %v", stored.expiresAt, tc.wantExpiry)
			}
		})
	}
}

// TestCreateMCPTokenRefusesAnEmptyName covers the other input the column
// cannot hold; the name is how a token is recognised when revoking it.
func TestCreateMCPTokenRefusesAnEmptyName(t *testing.T) {
	repo := new(fakeRepository{})

	_, err := newTestService(repo, newMemStorage()).
		CreateMCPToken(context.Background(), uuid.New(), "   ", DefaultMCPTokenExpiryDays)
	if !errors.Is(err, ErrMCPTokenNameRequired) {
		t.Fatalf("error = %v, want ErrMCPTokenNameRequired", err)
	}
}

// TestCreateMCPTokenEnforcesTheLimit: the cap is answered as a conflict, and
// nothing is minted once it is reached.
func TestCreateMCPTokenEnforcesTheLimit(t *testing.T) {
	repo := new(fakeRepository{
		countMCPTokens: func(context.Context, uuid.UUID) (int, error) { return maxMCPTokensPerUser, nil },
	})

	_, err := newTestService(repo, newMemStorage()).
		CreateMCPToken(context.Background(), uuid.New(), "one more", DefaultMCPTokenExpiryDays)
	if !errors.Is(err, ErrTooManyMCPTokens) {
		t.Fatalf("error = %v, want ErrTooManyMCPTokens", err)
	}
}

// TestCreateMCPTokenMapsADuplicateName turns the one constraint a user can
// trip into an answer about their input rather than a 500.
func TestCreateMCPTokenMapsADuplicateName(t *testing.T) {
	repo := new(fakeRepository{
		countMCPTokens: func(context.Context, uuid.UUID) (int, error) { return 1, nil },
		createMCPToken: func(context.Context, uuid.UUID, string, string, string, *time.Time) (MCPToken, error) {
			return MCPToken{}, ErrMCPTokenNameTaken
		},
	})

	_, err := newTestService(repo, newMemStorage()).
		CreateMCPToken(context.Background(), uuid.New(), "Claude Desktop", DefaultMCPTokenExpiryDays)
	if !errors.Is(err, ErrMCPTokenNameTaken) {
		t.Fatalf("error = %v, want ErrMCPTokenNameTaken", err)
	}
}

// TestRotateMCPTokenReplacesTheSecretInPlace: a rotation is one token whose
// secret changed, not a new token, and the previous secret is gone.
func TestRotateMCPTokenReplacesTheSecretInPlace(t *testing.T) {
	tokenID, userID := uuid.New(), uuid.New()

	var (
		sawUserID uuid.UUID
		sawID     uuid.UUID
		newHash   string
	)

	repo := new(fakeRepository{
		rotateMCPToken: func(_ context.Context, u, id uuid.UUID, hash, last4 string, expiresAt *time.Time) (MCPToken, error) {
			sawUserID, sawID, newHash = u, id, hash

			return MCPToken{ID: id, Name: "Claude Desktop", Last4: last4, ExpiresAt: expiresAt}, nil
		},
	})

	rotated, err := newTestService(repo, newMemStorage()).
		RotateMCPToken(context.Background(), userID, tokenID, DefaultMCPTokenExpiryDays)
	if err != nil {
		t.Fatalf("RotateMCPToken: %v", err)
	}

	if sawUserID != userID || sawID != tokenID {
		t.Errorf("rotated (%s, %s), want the caller's own (%s, %s)", sawUserID, sawID, userID, tokenID)
	}

	if newHash != hashMCPToken(rotated.Token) {
		t.Error("the stored hash is not the hash of the returned token")
	}

	if rotated.ID != tokenID || rotated.Name != "Claude Desktop" {
		t.Errorf("rotation produced a different token: %+v", rotated.MCPToken)
	}
}

// TestMCPTokenNotFound covers both mutations: another user's id is the same
// answer as an id that never existed. The status it maps to is pinned by the
// handler test.
func TestMCPTokenNotFound(t *testing.T) {
	repo := new(fakeRepository{
		deleteMCPToken: func(context.Context, uuid.UUID, uuid.UUID) error { return ErrMCPTokenNotFound },
		rotateMCPToken: func(context.Context, uuid.UUID, uuid.UUID, string, string, *time.Time) (MCPToken, error) {
			return MCPToken{}, ErrMCPTokenNotFound
		},
	})
	svc := newTestService(repo, newMemStorage())

	if err := svc.DeleteMCPToken(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, ErrMCPTokenNotFound) {
		t.Errorf("delete error = %v, want ErrMCPTokenNotFound", err)
	}

	if _, err := svc.RotateMCPToken(context.Background(), uuid.New(), uuid.New(), 0); !errors.Is(err, ErrMCPTokenNotFound) {
		t.Errorf("rotate error = %v, want ErrMCPTokenNotFound", err)
	}
}

// TestAuthenticateMCPToken is the guard's half of the contract.
func TestAuthenticateMCPToken(t *testing.T) {
	userID := uuid.New()
	raw, hash, _, err := generateMCPToken()
	if err != nil {
		t.Fatalf("generateMCPToken: %v", err)
	}

	live := func(expiresAt *time.Time, lastUsed *time.Time) *fakeRepository {
		return new(fakeRepository{
			getMCPTokenByHash: func(_ context.Context, presented string) (mcpTokenIdentity, error) {
				if presented != hash {
					return mcpTokenIdentity{}, ErrMCPTokenNotFound
				}

				return mcpTokenIdentity{ID: uuid.New(), UserID: userID, Role: "user", ExpiresAt: expiresAt, LastUsedAt: lastUsed}, nil
			},
			touchMCPToken: func(context.Context, uuid.UUID) error { return nil },
		})
	}

	t.Run("a live token yields its owner and role", func(t *testing.T) {
		future := time.Now().UTC().Add(time.Hour)

		gotID, role, err := newTestService(live(&future, nil), newMemStorage()).
			AuthenticateMCPToken(context.Background(), raw)
		if err != nil {
			t.Fatalf("AuthenticateMCPToken: %v", err)
		}

		if gotID != userID || role != "user" {
			t.Errorf("got (%s, %q), want (%s, \"user\")", gotID, role, userID)
		}
	})

	t.Run("a token with no expiry stays valid", func(t *testing.T) {
		if _, _, err := newTestService(live(nil, nil), newMemStorage()).
			AuthenticateMCPToken(context.Background(), raw); err != nil {
			t.Fatalf("AuthenticateMCPToken: %v", err)
		}
	})

	t.Run("an expired token is refused", func(t *testing.T) {
		past := time.Now().UTC().Add(-time.Minute)

		if _, _, err := newTestService(live(&past, nil), newMemStorage()).
			AuthenticateMCPToken(context.Background(), raw); !errors.Is(err, ErrInvalidMCPToken) {
			t.Fatalf("error = %v, want ErrInvalidMCPToken", err)
		}
	})

	t.Run("an unknown token is refused", func(t *testing.T) {
		future := time.Now().UTC().Add(time.Hour)
		other, _, _, _ := generateMCPToken()

		if _, _, err := newTestService(live(&future, nil), newMemStorage()).
			AuthenticateMCPToken(context.Background(), other); !errors.Is(err, ErrInvalidMCPToken) {
			t.Fatalf("error = %v, want ErrInvalidMCPToken", err)
		}
	})

	// A JWT reaching this path is a wiring mistake, not a credential: it must
	// be refused without a store lookup, since its hash could never match and
	// the round trip would be spent on every unauthenticated request.
	t.Run("a token without the prefix never reaches the store", func(t *testing.T) {
		repo := new(fakeRepository{}) // every hook nil: a lookup panics

		if _, _, err := newTestService(repo, newMemStorage()).
			AuthenticateMCPToken(context.Background(), "eyJhbGciOiJIUzI1NiJ9.e30.sig"); !errors.Is(err, ErrInvalidMCPToken) {
			t.Fatalf("error = %v, want ErrInvalidMCPToken", err)
		}
	})
}

// TestAuthenticateMCPTokenThrottlesTheTouch: last_used_at exists for the
// settings screen, and paying a write for it on every tool call would be the
// wrong trade.
func TestAuthenticateMCPTokenThrottlesTheTouch(t *testing.T) {
	raw, _, _, err := generateMCPToken()
	if err != nil {
		t.Fatalf("generateMCPToken: %v", err)
	}

	for _, tc := range []struct {
		name      string
		lastUsed  time.Duration
		wantTouch bool
	}{
		{name: "just used", lastUsed: time.Minute, wantTouch: false},
		{name: "stale", lastUsed: mcpTokenTouchInterval + time.Minute, wantTouch: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			touched := false
			lastUsed := time.Now().UTC().Add(-tc.lastUsed)

			repo := new(fakeRepository{
				getMCPTokenByHash: func(context.Context, string) (mcpTokenIdentity, error) {
					return mcpTokenIdentity{ID: uuid.New(), UserID: uuid.New(), Role: "user", LastUsedAt: &lastUsed}, nil
				},
				touchMCPToken: func(context.Context, uuid.UUID) error {
					touched = true

					return nil
				},
			})

			if _, _, err := newTestService(repo, newMemStorage()).
				AuthenticateMCPToken(context.Background(), raw); err != nil {
				t.Fatalf("AuthenticateMCPToken: %v", err)
			}

			if touched != tc.wantTouch {
				t.Errorf("touched = %v, want %v", touched, tc.wantTouch)
			}
		})
	}
}

// TestListMCPTokensFlagsTheExpired: an expired token stays listed, because the
// list is where a user finds out why their client stopped working.
func TestListMCPTokensFlagsTheExpired(t *testing.T) {
	past := time.Now().UTC().Add(-time.Hour)
	future := time.Now().UTC().Add(time.Hour)

	repo := new(fakeRepository{
		listMCPTokens: func(context.Context, uuid.UUID) ([]MCPToken, error) {
			return []MCPToken{
				{Name: "expired", ExpiresAt: &past},
				{Name: "live", ExpiresAt: &future},
				{Name: "no expiry"},
			}, nil
		},
	})

	tokens, err := newTestService(repo, newMemStorage()).ListMCPTokens(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ListMCPTokens: %v", err)
	}

	want := map[string]bool{"expired": true, "live": false, "no expiry": false}
	for _, token := range tokens {
		if token.Expired != want[token.Name] {
			t.Errorf("%q: Expired = %v, want %v", token.Name, token.Expired, want[token.Name])
		}
	}
}
