package auth

import (
	"context"
	"errors"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
)

// OAuthStore: the authorization server's persistence.
//
// Two things here are deliberately not the obvious statement. Codes are claimed
// with a conditional UPDATE rather than read-then-write, because two token
// calls racing on a stolen code must not both succeed and only the database can
// arbitrate that. And grants are upserted on (user, client, scope) rather than
// inserted, because re-approving a client the user already approved is the same
// consent, not a second one.

// oauthClientColumns is the projection oauthClient is scanned from, in one
// place so the statements returning a client cannot drift apart.
const oauthClientColumns = `client_id, COALESCE(client_secret_hash, ''), client_name, redirect_uris,
	grant_types, response_types, scope, token_endpoint_auth_method,
	COALESCE(client_uri, ''), COALESCE(logo_uri, ''), created_at`

func (r *PostgresRepository) CreateOAuthClient(ctx context.Context, c oauthClient) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO oauth_clients
		   (client_id, client_secret_hash, client_name, redirect_uris, grant_types,
		    response_types, scope, token_endpoint_auth_method, client_uri, logo_uri)
		 VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7, $8, NULLIF($9, ''), NULLIF($10, ''))`,
		c.ClientID, c.SecretHash, c.Name, c.RedirectURIs, c.GrantTypes,
		c.ResponseTypes, c.Scope, c.TokenEndpointAuth, c.ClientURI, c.LogoURI,
	)

	return err
}

func (r *PostgresRepository) GetOAuthClient(ctx context.Context, clientID string) (oauthClient, error) {
	var c oauthClient
	err := r.db.QueryRow(ctx,
		`SELECT `+oauthClientColumns+` FROM oauth_clients WHERE client_id = $1`,
		clientID,
	).Scan(&c.ClientID, &c.SecretHash, &c.Name, &c.RedirectURIs, &c.GrantTypes,
		&c.ResponseTypes, &c.Scope, &c.TokenEndpointAuth, &c.ClientURI, &c.LogoURI, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return oauthClient{}, ErrOAuthClientNotFound
		}

		return oauthClient{}, err
	}

	return c, nil
}

func (r *PostgresRepository) CreateAuthorizationRequest(ctx context.Context, req pendingAuthorization) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx,
		`INSERT INTO oauth_authorization_requests
		   (client_id, redirect_uri, scope, state, code_challenge, code_challenge_method, resource, expires_at)
		 VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, NULLIF($7, ''), $8)
		 RETURNING id`,
		req.ClientID, req.RedirectURI, req.Scope, req.State,
		req.CodeChallenge, req.CodeChallengeMethod, req.Resource, req.ExpiresAt,
	).Scan(&id)

	return id, err
}

// GetAuthorizationRequest reads a parked consent. Expiry is filtered in the
// statement, unlike the token lookups: an expired request and an unknown one
// are the same answer to the consent screen — "this is no longer valid" — so
// there is nothing for the caller to tell apart.
func (r *PostgresRepository) GetAuthorizationRequest(ctx context.Context, id uuid.UUID) (pendingAuthorization, error) {
	var p pendingAuthorization
	err := r.db.QueryRow(ctx,
		`SELECT id, client_id, redirect_uri, scope, COALESCE(state, ''),
		        code_challenge, code_challenge_method, COALESCE(resource, ''), expires_at
		 FROM oauth_authorization_requests
		 WHERE id = $1 AND expires_at > NOW()`,
		id,
	).Scan(&p.ID, &p.ClientID, &p.RedirectURI, &p.Scope, &p.State,
		&p.CodeChallenge, &p.CodeChallengeMethod, &p.Resource, &p.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pendingAuthorization{}, ErrOAuthRequestNotFound
		}

		return pendingAuthorization{}, err
	}

	return p, nil
}

// DeleteAuthorizationRequest removes a consent once it has been answered,
// approved or denied. It is what makes the request single-use: a second POST
// with the same id finds nothing.
func (r *PostgresRepository) DeleteAuthorizationRequest(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM oauth_authorization_requests WHERE id = $1`, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrOAuthRequestNotFound
	}

	return nil
}

func (r *PostgresRepository) CreateAuthorizationCode(ctx context.Context, codeHash string, c authorizationCode) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO oauth_authorization_codes
		   (code_hash, client_id, user_id, redirect_uri, scope, code_challenge,
		    code_challenge_method, resource, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $9)`,
		codeHash, c.ClientID, c.UserID, c.RedirectURI, c.Scope,
		c.CodeChallenge, c.CodeChallengeMethod, c.Resource, c.ExpiresAt,
	)

	return err
}

// The claim below reads the same nine fields out of two different places — the
// UPDATE's RETURNING and the base table — so it needs the projection twice, in
// two forms.
//
// oauthCodeProjection reads from the table and normalises the nullable column;
// the alias is load-bearing, not decoration. Without it the RETURNING yields a
// column named "coalesce", and the SELECT over the CTE fails at run time with
// `column "resource" does not exist` — a query that compiles, passes every test
// with a faked store, and breaks the first real token exchange.
const oauthCodeProjection = `id, client_id, user_id, redirect_uri, scope,
	code_challenge, code_challenge_method, COALESCE(resource, '') AS resource, expires_at`

// oauthCodeFields reads those same nine back out of the CTE, where the
// normalising already happened.
const oauthCodeFields = `id, client_id, user_id, redirect_uri, scope,
	code_challenge, code_challenge_method, resource, expires_at`

// ConsumeAuthorizationCode claims a code and reports whether this call is the
// one that got it.
//
// The claim is a conditional UPDATE, so two token requests arriving with the
// same code cannot both be told yes — Postgres serialises them and the loser
// sees consumed_at already set. That matters because a code presented twice is
// the signature of one that was intercepted, and the service answers it by
// revoking the grant rather than merely refusing the second call. Doing this
// with a SELECT followed by an UPDATE would leave exactly the window the
// attacker needs.
//
// The second arm of the UNION is what makes "already consumed" distinguishable
// from "never existed": without it, both would arrive as no rows and the reuse
// would be invisible.
func (r *PostgresRepository) ConsumeAuthorizationCode(ctx context.Context, codeHash string) (authorizationCode, bool, error) {
	var (
		c       authorizationCode
		claimed bool
	)

	err := r.db.QueryRow(ctx,
		`WITH claimed AS (
		   UPDATE oauth_authorization_codes
		   SET consumed_at = NOW()
		   WHERE code_hash = $1 AND consumed_at IS NULL
		   RETURNING `+oauthCodeProjection+`
		 )
		 SELECT `+oauthCodeFields+`, TRUE FROM claimed
		 UNION ALL
		 SELECT `+oauthCodeProjection+`, FALSE
		 FROM oauth_authorization_codes
		 WHERE code_hash = $1 AND NOT EXISTS (SELECT 1 FROM claimed)`,
		codeHash,
	).Scan(&c.ID, &c.ClientID, &c.UserID, &c.RedirectURI, &c.Scope,
		&c.CodeChallenge, &c.CodeChallengeMethod, &c.Resource, &c.ExpiresAt, &claimed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authorizationCode{}, false, ErrOAuthCodeNotFound
		}

		return authorizationCode{}, false, err
	}

	return c, claimed, nil
}

// UpsertOAuthGrant records the authorization and its first token pair, or
// replaces the tokens of the grant this user already gave this client for this
// scope.
//
// The conflict target is uq_oauth_grants_user_client_scope. Re-approving is
// therefore idempotent from the settings screen's point of view: one row per
// connected application, keeping its created_at, rather than a list that grows
// every time a user reconnects.
func (r *PostgresRepository) UpsertOAuthGrant(
	ctx context.Context,
	userID uuid.UUID,
	clientID, scope, resource, accessHash string,
	accessExpiresAt time.Time,
	refreshHash string,
	refreshExpiresAt *time.Time,
) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx,
		`INSERT INTO oauth_grants
		   (user_id, client_id, scope, resource, access_token_hash, access_expires_at,
		    refresh_token_hash, refresh_expires_at)
		 VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, NULLIF($7, ''), $8)
		 ON CONFLICT (user_id, client_id, scope) DO UPDATE
		 SET access_token_hash  = EXCLUDED.access_token_hash,
		     access_expires_at  = EXCLUDED.access_expires_at,
		     refresh_token_hash = EXCLUDED.refresh_token_hash,
		     refresh_expires_at = EXCLUDED.refresh_expires_at,
		     resource           = EXCLUDED.resource,
		     updated_at         = NOW()
		 RETURNING id`,
		userID, clientID, scope, resource, accessHash, accessExpiresAt, refreshHash, refreshExpiresAt,
	).Scan(&id)

	return id, err
}

// GetGrantByAccessToken resolves a presented access token to its account. It
// joins users and roles for the same reason GetMCPTokenByHash does: the guard
// needs the role, and a deleted or banned account must stop authenticating
// immediately rather than when its tokens happen to expire.
//
// Expiry is not filtered here, also for that reason: an expired token is a
// different answer from an unknown one for the service deciding what to log.
func (r *PostgresRepository) GetGrantByAccessToken(ctx context.Context, tokenHash string) (oauthGrantIdentity, error) {
	var g oauthGrantIdentity
	err := r.db.QueryRow(ctx,
		`SELECT g.id, g.user_id, r.name, g.scope, g.access_expires_at, g.last_used_at
		 FROM oauth_grants g
		 JOIN users u ON u.id = g.user_id
		 JOIN roles r ON r.id = u.role_id
		 WHERE g.access_token_hash = $1 AND u.deleted_at IS NULL AND u.banned_at IS NULL`,
		tokenHash,
	).Scan(&g.ID, &g.UserID, &g.Role, &g.Scope, &g.ExpiresAt, &g.LastUsedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return oauthGrantIdentity{}, ErrOAuthGrantNotFound
		}

		return oauthGrantIdentity{}, err
	}

	return g, nil
}

func (r *PostgresRepository) GetGrantByRefreshToken(ctx context.Context, tokenHash string) (grantRefresh, error) {
	var g grantRefresh
	err := r.db.QueryRow(ctx,
		`SELECT g.id, g.user_id, g.client_id, g.scope, COALESCE(g.resource, ''), g.refresh_expires_at
		 FROM oauth_grants g
		 JOIN users u ON u.id = g.user_id
		 WHERE g.refresh_token_hash = $1 AND u.deleted_at IS NULL AND u.banned_at IS NULL`,
		tokenHash,
	).Scan(&g.ID, &g.UserID, &g.ClientID, &g.Scope, &g.Resource, &g.RefreshExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return grantRefresh{}, ErrOAuthGrantNotFound
		}

		return grantRefresh{}, err
	}

	return g, nil
}

// RotateGrantTokens replaces both hashes of a live grant. Both, always: a
// refresh that left the old access token working would make revocation mean
// nothing for up to its full lifetime.
func (r *PostgresRepository) RotateGrantTokens(
	ctx context.Context,
	id uuid.UUID,
	accessHash string,
	accessExpiresAt time.Time,
	refreshHash string,
	refreshExpiresAt *time.Time,
) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE oauth_grants
		 SET access_token_hash  = $2,
		     access_expires_at  = $3,
		     refresh_token_hash = NULLIF($4, ''),
		     refresh_expires_at = $5,
		     updated_at         = NOW()
		 WHERE id = $1`,
		id, accessHash, accessExpiresAt, refreshHash, refreshExpiresAt,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrOAuthGrantNotFound
	}

	return nil
}

func (r *PostgresRepository) TouchOAuthGrant(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE oauth_grants SET last_used_at = NOW() WHERE id = $1`, id)

	return err
}

// ListOAuthGrants returns the user's connected applications, named by the
// client rather than by an id nobody recognises.
func (r *PostgresRepository) ListOAuthGrants(ctx context.Context, userID uuid.UUID) ([]OAuthGrant, error) {
	rows, err := r.db.Query(ctx,
		`SELECT g.id, c.client_name, COALESCE(c.client_uri, ''), g.scope, g.last_used_at, g.created_at
		 FROM oauth_grants g
		 JOIN oauth_clients c ON c.client_id = g.client_id
		 WHERE g.user_id = $1
		 ORDER BY g.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grants := make([]OAuthGrant, 0)

	for rows.Next() {
		var (
			g     OAuthGrant
			scope string
		)

		if err := rows.Scan(&g.ID, &g.ClientName, &g.ClientURI, &scope, &g.LastUsedAt, &g.CreatedAt); err != nil {
			return nil, err
		}

		g.Scopes = scopeList(scope)
		grants = append(grants, g)
	}

	return grants, rows.Err()
}

// DeleteOAuthGrant revokes one connected application. Scoped to the owner in
// the WHERE clause, so "not yours" and "does not exist" are the same answer at
// the only layer that could tell them apart.
func (r *PostgresRepository) DeleteOAuthGrant(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM oauth_grants WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrOAuthGrantNotFound
	}

	return nil
}

// DeleteOAuthGrantsForClient revokes everything one client holds for one user.
// It is the response to a replayed authorization code: the code was intercepted,
// so whatever was minted from it is suspect, not just the second attempt.
func (r *PostgresRepository) DeleteOAuthGrantsForClient(ctx context.Context, userID uuid.UUID, clientID string) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM oauth_grants WHERE user_id = $1 AND client_id = $2`,
		userID, clientID,
	)

	return tag.RowsAffected(), err
}

// DeleteExpiredOAuthRows sweeps the two short-lived tables.
//
// Codes are kept a day past expiry rather than deleted on sight: replay
// detection needs the consumed row to still be there, and a code replayed hours
// later is exactly the case worth catching. Grants are not swept at all — an
// expired one is what the settings screen shows as a connection that lapsed,
// and it is the user who decides it is gone.
func (r *PostgresRepository) DeleteExpiredOAuthRows(ctx context.Context) (int64, error) {
	requests, err := r.db.Exec(ctx, `DELETE FROM oauth_authorization_requests WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}

	codes, err := r.db.Exec(ctx,
		`DELETE FROM oauth_authorization_codes WHERE expires_at < NOW() - INTERVAL '1 day'`,
	)
	if err != nil {
		return requests.RowsAffected(), err
	}

	return requests.RowsAffected() + codes.RowsAffected(), nil
}
