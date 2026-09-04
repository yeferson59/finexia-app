package auth

import (
	"context"
	"errors"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// MCPTokenStore: the personal access tokens that authenticate an MCP client.
//
// Every read here is by hash or by (user, id) — never by id alone. Scoping the
// mutations to the owner in the WHERE clause, rather than reading the row and
// comparing in Go, is what makes "not yours" and "does not exist" the same
// answer at the only layer that could tell them apart.

// pgUniqueViolation is the SQLSTATE Postgres returns for a unique constraint
// violation (23505). Here it can only be uq_mcp_tokens_user_name: token_hash
// carries 256 bits of entropy.
const pgUniqueViolation = "23505"

// mcpTokenColumns is the projection MCPToken is scanned from, kept in one place
// so the three statements returning a token cannot drift apart.
const mcpTokenColumns = `id, name, last4, expires_at, last_used_at, rotated_at, created_at`

func (r *PostgresRepository) CreateMCPToken(ctx context.Context, userID uuid.UUID, name, tokenHash, last4 string, expiresAt *time.Time) (MCPToken, error) {
	var t MCPToken
	err := r.db.QueryRow(ctx,
		`INSERT INTO mcp_tokens (user_id, name, token_hash, last4, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING `+mcpTokenColumns,
		userID, name, tokenHash, last4, expiresAt,
	).Scan(&t.ID, &t.Name, &t.Last4, &t.ExpiresAt, &t.LastUsedAt, &t.RotatedAt, &t.CreatedAt)
	if err != nil {
		return MCPToken{}, mapMCPTokenWriteError(err)
	}

	return t, nil
}

func (r *PostgresRepository) ListMCPTokens(ctx context.Context, userID uuid.UUID) ([]MCPToken, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+mcpTokenColumns+`
		 FROM mcp_tokens
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make([]MCPToken, 0)
	for rows.Next() {
		var t MCPToken
		if err := rows.Scan(&t.ID, &t.Name, &t.Last4, &t.ExpiresAt, &t.LastUsedAt, &t.RotatedAt, &t.CreatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}

	return tokens, rows.Err()
}

func (r *PostgresRepository) CountMCPTokens(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM mcp_tokens WHERE user_id = $1`, userID).Scan(&count)

	return count, err
}

// GetMCPTokenByHash resolves a presented token to its account. It joins users
// and roles for the same reason GetRefreshTokenByHash does: the guard needs the
// role, and a deleted account must stop authenticating immediately rather than
// when its tokens happen to expire.
//
// Expiry is deliberately not filtered here. An expired token is a different
// answer from an unknown one for the service that has to decide what to log,
// and folding both into "no rows" would throw that away.
func (r *PostgresRepository) GetMCPTokenByHash(ctx context.Context, tokenHash string) (mcpTokenIdentity, error) {
	var t mcpTokenIdentity
	err := r.db.QueryRow(ctx,
		`SELECT mt.id, mt.user_id, r.name, mt.expires_at, mt.last_used_at
		 FROM mcp_tokens mt
		 JOIN users u ON u.id = mt.user_id
		 JOIN roles r ON r.id = u.role_id
		 WHERE mt.token_hash = $1 AND u.deleted_at IS NULL AND u.banned_at IS NULL`,
		tokenHash,
	).Scan(&t.ID, &t.UserID, &t.Role, &t.ExpiresAt, &t.LastUsedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return mcpTokenIdentity{}, ErrMCPTokenNotFound
		}

		return mcpTokenIdentity{}, err
	}

	return t, nil
}

// RotateMCPToken replaces the secret of a token the user owns, in place. The
// row keeps its id, name and created_at, so what the settings screen shows is
// one token that was rotated rather than a new one next to a deleted one.
func (r *PostgresRepository) RotateMCPToken(ctx context.Context, userID, id uuid.UUID, tokenHash, last4 string, expiresAt *time.Time) (MCPToken, error) {
	var t MCPToken
	err := r.db.QueryRow(ctx,
		`UPDATE mcp_tokens
		 SET token_hash = $3, last4 = $4, expires_at = $5, rotated_at = NOW(), last_used_at = NULL
		 WHERE id = $1 AND user_id = $2
		 RETURNING `+mcpTokenColumns,
		id, userID, tokenHash, last4, expiresAt,
	).Scan(&t.ID, &t.Name, &t.Last4, &t.ExpiresAt, &t.LastUsedAt, &t.RotatedAt, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MCPToken{}, ErrMCPTokenNotFound
		}

		return MCPToken{}, mapMCPTokenWriteError(err)
	}

	return t, nil
}

func (r *PostgresRepository) DeleteMCPToken(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM mcp_tokens WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrMCPTokenNotFound
	}

	return nil
}

// TouchMCPToken records that a token was just used. The staleness check lives
// in the caller, which already read last_used_at; this statement is reached
// only when the column is actually due for a write.
func (r *PostgresRepository) TouchMCPToken(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE mcp_tokens SET last_used_at = NOW() WHERE id = $1`, id)

	return err
}

// DeleteExpiredMCPTokens removes what can never authenticate again. Tokens with
// no expiry are never swept: their lifetime is the choice the user made.
func (r *PostgresRepository) DeleteExpiredMCPTokens(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM mcp_tokens WHERE expires_at IS NOT NULL AND expires_at < NOW() - INTERVAL '30 days'`,
	)

	return tag.RowsAffected(), err
}

// mapMCPTokenWriteError turns the one constraint a user can trip into an error
// about their input. Anything else is the server's problem and stays a 500.
func mapMCPTokenWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return ErrMCPTokenNameTaken
	}

	return err
}
