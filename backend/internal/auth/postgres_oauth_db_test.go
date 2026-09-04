package auth

import (
	"context"
	"os"
	"testing"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

// The OAuth repository is the one part of the authorization server no fake can
// check. Its two load-bearing statements are load-bearing *because* they are
// SQL: the code claim is a data-modifying CTE that has to arbitrate a race, and
// the grant write is an upsert on a partial-looking unique index. A store faked
// in Go reproduces the intent of both and none of the syntax — which is exactly
// how the claim shipped once with a column the outer SELECT could not see,
// green suite and all.
//
// It needs a database with the migrations applied:
//
//	TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/finexia go test ./internal/auth/
//
// Without that variable it skips, so `go test ./...` stays a no-setup command.

func oauthTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL no está definida: se omite la prueba contra Postgres")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping: %v", err)
	}

	t.Cleanup(pool.Close)

	return pool
}

// oauthFixture plants a user and a registered client, and cleans both up. The
// client cascades to everything else the flow writes.
func oauthFixture(t *testing.T, r *PostgresRepository) (uuid.UUID, string) {
	t.Helper()

	ctx := context.Background()
	userID := uuid.New()
	clientID := "fnx_client_" + uuid.New().String()

	_, err := r.db.Exec(ctx,
		`INSERT INTO users (id, name, email, role_id, preferred_currency)
		 VALUES ($1, 'oauth probe', $2, (SELECT id FROM roles LIMIT 1), 'USD')`,
		userID, userID.String()+"@probe.test")
	if err != nil {
		t.Fatalf("planting the user: %v", err)
	}

	t.Cleanup(func() {
		_, _ = r.db.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
		_, _ = r.db.Exec(context.Background(), `DELETE FROM oauth_clients WHERE client_id = $1`, clientID)
	})

	if err := r.CreateOAuthClient(ctx, oauthClient{
		ClientID:          clientID,
		Name:              "Probe",
		RedirectURIs:      []string{"https://claude.test/cb"},
		GrantTypes:        []string{"authorization_code", "refresh_token"},
		ResponseTypes:     []string{"code"},
		Scope:             MCPScope,
		TokenEndpointAuth: "none",
	}); err != nil {
		t.Fatalf("CreateOAuthClient: %v", err)
	}

	return userID, clientID
}

// TestOAuthRepositoryAgainstPostgres runs every statement in postgres_oauth.go
// against a real database, in the order the flow uses them.
func TestOAuthRepositoryAgainstPostgres(t *testing.T) {
	r := NewPostgresRepository(oauthTestPool(t))
	ctx := context.Background()
	userID, clientID := oauthFixture(t, r)

	t.Run("a client round-trips, including the array columns", func(t *testing.T) {
		got, err := r.GetOAuthClient(ctx, clientID)
		if err != nil {
			t.Fatalf("GetOAuthClient: %v", err)
		}

		if got.Name != "Probe" || len(got.RedirectURIs) != 1 || got.RedirectURIs[0] != "https://claude.test/cb" {
			t.Errorf("client = %+v", got)
		}

		// A public client stores NULL and must read back as empty, not as a
		// hash nothing can match: isPublic() decides whether a secret is even
		// asked for.
		if !got.isPublic() {
			t.Error("a client registered without a secret read back as confidential")
		}
	})

	t.Run("an unknown client is ErrOAuthClientNotFound", func(t *testing.T) {
		if _, err := r.GetOAuthClient(ctx, "fnx_client_nope"); err != ErrOAuthClientNotFound {
			t.Errorf("err = %v, want ErrOAuthClientNotFound", err)
		}
	})

	var requestID uuid.UUID

	t.Run("a parked request round-trips", func(t *testing.T) {
		var err error

		requestID, err = r.CreateAuthorizationRequest(ctx, pendingAuthorization{
			ClientID:            clientID,
			RedirectURI:         "https://claude.test/cb",
			Scope:               MCPScope,
			State:               "st",
			CodeChallenge:       "chal",
			CodeChallengeMethod: oauthCodeChallengeS256,
			ExpiresAt:           time.Now().UTC().Add(oauthRequestTTL),
		})
		if err != nil {
			t.Fatalf("CreateAuthorizationRequest: %v", err)
		}

		got, err := r.GetAuthorizationRequest(ctx, requestID)
		if err != nil {
			t.Fatalf("GetAuthorizationRequest: %v", err)
		}

		// An absent resource is stored NULL by NULLIF and has to come back as
		// "", or normalizeResource would compare a canonical URI against a zero
		// value it never wrote.
		if got.State != "st" || got.Resource != "" || got.CodeChallenge != "chal" {
			t.Errorf("request = %+v", got)
		}
	})

	t.Run("an expired request reads as not found", func(t *testing.T) {
		expired, err := r.CreateAuthorizationRequest(ctx, pendingAuthorization{
			ClientID: clientID, RedirectURI: "https://claude.test/cb", Scope: MCPScope,
			CodeChallenge: "c", CodeChallengeMethod: oauthCodeChallengeS256,
			ExpiresAt: time.Now().UTC().Add(-time.Minute),
		})
		if err != nil {
			t.Fatalf("CreateAuthorizationRequest: %v", err)
		}

		if _, err := r.GetAuthorizationRequest(ctx, expired); err != ErrOAuthRequestNotFound {
			t.Errorf("err = %v, want ErrOAuthRequestNotFound", err)
		}
	})

	t.Run("deleting a request twice is not found the second time", func(t *testing.T) {
		if err := r.DeleteAuthorizationRequest(ctx, requestID); err != nil {
			t.Fatalf("DeleteAuthorizationRequest: %v", err)
		}

		if err := r.DeleteAuthorizationRequest(ctx, requestID); err != ErrOAuthRequestNotFound {
			t.Errorf("err = %v, want ErrOAuthRequestNotFound", err)
		}
	})

	// The statement this whole file exists for.
	t.Run("a code can be claimed exactly once", func(t *testing.T) {
		codeHash := hashOAuthToken("raw-code-" + uuid.New().String())

		if err := r.CreateAuthorizationCode(ctx, codeHash, authorizationCode{
			ClientID: clientID, UserID: userID, RedirectURI: "https://claude.test/cb",
			Scope: MCPScope, CodeChallenge: "chal", CodeChallengeMethod: oauthCodeChallengeS256,
			Resource:  "https://api.finexia.test/mcp",
			ExpiresAt: time.Now().UTC().Add(oauthCodeTTL),
		}); err != nil {
			t.Fatalf("CreateAuthorizationCode: %v", err)
		}

		code, claimed, err := r.ConsumeAuthorizationCode(ctx, codeHash)
		if err != nil {
			t.Fatalf("first claim: %v", err)
		}

		if !claimed {
			t.Fatal("the first claim did not win the code")
		}

		if code.UserID != userID || code.Resource != "https://api.finexia.test/mcp" {
			t.Errorf("claimed code = %+v", code)
		}

		// The replay: the row still has to come back, with claimed false, or
		// the service cannot tell an intercepted code from an invented one.
		replayed, claimed, err := r.ConsumeAuthorizationCode(ctx, codeHash)
		if err != nil {
			t.Fatalf("second claim: %v", err)
		}

		if claimed {
			t.Error("the same code was claimed twice")
		}

		if replayed.ClientID != clientID {
			t.Errorf("the replay did not return the row: %+v", replayed)
		}
	})

	t.Run("an unknown code is ErrOAuthCodeNotFound", func(t *testing.T) {
		if _, _, err := r.ConsumeAuthorizationCode(ctx, hashOAuthToken("never-existed")); err != ErrOAuthCodeNotFound {
			t.Errorf("err = %v, want ErrOAuthCodeNotFound", err)
		}
	})

	t.Run("grants upsert on (user, client, scope) and authenticate", func(t *testing.T) {
		accessHash, refreshHash := hashOAuthToken("a1"), hashOAuthToken("r1")
		accessExpiry := time.Now().UTC().Add(oauthAccessTTL)
		refreshExpiry := time.Now().UTC().Add(oauthRefreshTTL)

		grantID, err := r.UpsertOAuthGrant(ctx, userID, clientID, MCPScope,
			"https://api.finexia.test/mcp", accessHash, accessExpiry, refreshHash, &refreshExpiry)
		if err != nil {
			t.Fatalf("UpsertOAuthGrant: %v", err)
		}

		// Re-approving must land on the same row, or the settings screen grows
		// a duplicate every time the user reconnects.
		secondHash := hashOAuthToken("a2")

		sameID, err := r.UpsertOAuthGrant(ctx, userID, clientID, MCPScope,
			"https://api.finexia.test/mcp", secondHash, accessExpiry, hashOAuthToken("r2"), &refreshExpiry)
		if err != nil {
			t.Fatalf("second UpsertOAuthGrant: %v", err)
		}

		if sameID != grantID {
			t.Errorf("re-approving created a second grant (%v then %v)", grantID, sameID)
		}

		// And the replaced access token has to be gone.
		if _, err := r.GetGrantByAccessToken(ctx, accessHash); err != ErrOAuthGrantNotFound {
			t.Errorf("the superseded access token still resolves: %v", err)
		}

		identity, err := r.GetGrantByAccessToken(ctx, secondHash)
		if err != nil {
			t.Fatalf("GetGrantByAccessToken: %v", err)
		}

		if identity.UserID != userID || identity.Role == "" {
			t.Errorf("identity = %+v, want the user and their role from the join", identity)
		}

		if err := r.TouchOAuthGrant(ctx, grantID); err != nil {
			t.Errorf("TouchOAuthGrant: %v", err)
		}

		refreshed, err := r.GetGrantByRefreshToken(ctx, hashOAuthToken("r2"))
		if err != nil {
			t.Fatalf("GetGrantByRefreshToken: %v", err)
		}

		if refreshed.ClientID != clientID || refreshed.Resource != "https://api.finexia.test/mcp" {
			t.Errorf("refresh row = %+v", refreshed)
		}

		if err := r.RotateGrantTokens(ctx, grantID, hashOAuthToken("a3"), accessExpiry,
			hashOAuthToken("r3"), &refreshExpiry); err != nil {
			t.Fatalf("RotateGrantTokens: %v", err)
		}

		if _, err := r.GetGrantByAccessToken(ctx, secondHash); err != ErrOAuthGrantNotFound {
			t.Error("the pre-rotation access token still resolves")
		}

		listed, err := r.ListOAuthGrants(ctx, userID)
		if err != nil {
			t.Fatalf("ListOAuthGrants: %v", err)
		}

		if len(listed) != 1 || listed[0].ClientName != "Probe" || len(listed[0].Scopes) != 1 {
			t.Errorf("listed = %+v", listed)
		}

		// Not this user's grant: the same answer as one that does not exist.
		if err := r.DeleteOAuthGrant(ctx, uuid.New(), grantID); err != ErrOAuthGrantNotFound {
			t.Errorf("err = %v, want ErrOAuthGrantNotFound for another user's grant", err)
		}

		if err := r.DeleteOAuthGrant(ctx, userID, grantID); err != nil {
			t.Fatalf("DeleteOAuthGrant: %v", err)
		}

		if _, err := r.GetGrantByAccessToken(ctx, hashOAuthToken("a3")); err != ErrOAuthGrantNotFound {
			t.Error("a revoked grant still authenticates")
		}
	})

	t.Run("revoking by client clears every grant it holds", func(t *testing.T) {
		expiry := time.Now().UTC().Add(oauthRefreshTTL)

		if _, err := r.UpsertOAuthGrant(ctx, userID, clientID, MCPScope, "",
			hashOAuthToken("bulk-a"), time.Now().UTC().Add(oauthAccessTTL), hashOAuthToken("bulk-r"), &expiry); err != nil {
			t.Fatalf("UpsertOAuthGrant: %v", err)
		}

		deleted, err := r.DeleteOAuthGrantsForClient(ctx, userID, clientID)
		if err != nil {
			t.Fatalf("DeleteOAuthGrantsForClient: %v", err)
		}

		if deleted != 1 {
			t.Errorf("deleted = %d, want 1", deleted)
		}
	})

	t.Run("the sweep runs and counts", func(t *testing.T) {
		if _, err := r.CreateAuthorizationRequest(ctx, pendingAuthorization{
			ClientID: clientID, RedirectURI: "https://claude.test/cb", Scope: MCPScope,
			CodeChallenge: "c", CodeChallengeMethod: oauthCodeChallengeS256,
			ExpiresAt: time.Now().UTC().Add(-time.Hour),
		}); err != nil {
			t.Fatalf("CreateAuthorizationRequest: %v", err)
		}

		swept, err := r.DeleteExpiredOAuthRows(ctx)
		if err != nil {
			t.Fatalf("DeleteExpiredOAuthRows: %v", err)
		}

		if swept < 1 {
			t.Errorf("swept = %d, want at least the expired request", swept)
		}
	})
}
