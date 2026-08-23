package auth

import (
	"context"
	"time"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/identity"
)

// RefreshTokenStore: the rotating refresh tokens tied to a session, tracked by
// family so a replayed token can revoke every descendant at once.

func (r *PostgresRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, familyID, sessionID uuid.UUID, ip, ua *string, expiresAt time.Time) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx,
		"INSERT INTO refresh_tokens(user_id, token_hash, family_id, session_id, ip_address, user_agent, expires_at) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id",
		userID, tokenHash, familyID, sessionID, ip, ua, expiresAt,
	).Scan(&id)
	if err != nil {
		return uuid.Nil(), err
	}
	return id, nil
}

func (r *PostgresRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (identity.RefreshToken, error) {
	var rt identity.RefreshToken
	err := r.db.QueryRow(ctx,
		`SELECT rt.id, rt.user_id, rt.family_id, rt.session_id, rt.expires_at, rt.used_at, rt.revoked_at, r.name
		 FROM refresh_tokens rt
		 JOIN users u ON u.id = rt.user_id
		 JOIN roles r ON r.id = u.role_id
		 WHERE rt.token_hash = $1 AND u.deleted_at IS NULL`,
		tokenHash,
	).Scan(&rt.ID, &rt.UserID, &rt.FamilyID, &rt.SessionID, &rt.ExpiresAt, &rt.UsedAt, &rt.RevokedAt, &rt.Role)
	if err != nil {
		return identity.RefreshToken{}, err
	}
	return rt, nil
}

func (r *PostgresRepository) MarkRefreshTokenUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "UPDATE refresh_tokens SET used_at=NOW() WHERE id=$1", id)
	return err
}

// RevokeRefreshTokenFamily revokes every live token in the family and returns
// their hashes so callers can purge the corresponding cache entries.
func (r *PostgresRepository) RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) ([]string, error) {
	rows, err := r.db.Query(ctx, "UPDATE refresh_tokens SET revoked_at=NOW() WHERE family_id=$1 AND revoked_at IS NULL RETURNING token_hash", familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hashes []string
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return nil, err
		}
		hashes = append(hashes, hash)
	}
	return hashes, rows.Err()
}

// GetRefreshTokenFamiliesBySession returns hashes and family IDs of the live
// refresh tokens tied to a session, so logout can purge cache entries and mark
// the families revoked before the session row (and, via ON DELETE CASCADE, the
// token rows) disappears.
func (r *PostgresRepository) GetRefreshTokenFamiliesBySession(ctx context.Context, userID uuid.UUID, sessionToken string) ([]string, []uuid.UUID, error) {
	rows, err := r.db.Query(ctx,
		`SELECT rt.token_hash, rt.family_id
		 FROM refresh_tokens rt
		 JOIN sessions s ON s.id = rt.session_id
		 WHERE s.user_id = $1 AND s.token = $2 AND rt.revoked_at IS NULL`,
		userID.String(), sessionToken,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var (
		hashes    []string
		familyIDs []uuid.UUID
	)
	for rows.Next() {
		var (
			hash     string
			familyID uuid.UUID
		)
		if err := rows.Scan(&hash, &familyID); err != nil {
			return nil, nil, err
		}
		hashes = append(hashes, hash)
		familyIDs = append(familyIDs, familyID)
	}
	return hashes, familyIDs, rows.Err()
}

// GetRefreshTokensBySessionIDs returns the hashes and family IDs of the live
// refresh tokens tied to the given sessions, so callers can purge cache
// entries and mark the families revoked before deleting the session rows.
func (r *PostgresRepository) GetRefreshTokensBySessionIDs(ctx context.Context, userID uuid.UUID, sessionIDs []uuid.UUID) ([]string, []uuid.UUID, error) {
	rows, err := r.db.Query(ctx,
		`SELECT rt.token_hash, rt.family_id
		 FROM refresh_tokens rt
		 JOIN sessions s ON s.id = rt.session_id
		 WHERE s.user_id = $1 AND rt.session_id = ANY($2) AND rt.revoked_at IS NULL`,
		userID.String(), sessionIDs,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var (
		hashes    []string
		familyIDs []uuid.UUID
	)
	for rows.Next() {
		var (
			hash     string
			familyID uuid.UUID
		)
		if err := rows.Scan(&hash, &familyID); err != nil {
			return nil, nil, err
		}
		hashes = append(hashes, hash)
		familyIDs = append(familyIDs, familyID)
	}
	return hashes, familyIDs, rows.Err()
}

// DeleteExpiredRefreshTokens removes tokens that can never be redeemed again:
// past their expiry, revoked, or consumed long ago by rotation.
func (r *PostgresRepository) DeleteExpiredRefreshTokens(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM refresh_tokens
		 WHERE expires_at < NOW()
		    OR revoked_at < NOW() - INTERVAL '7 days'
		    OR (used_at IS NOT NULL AND used_at < NOW() - INTERVAL '7 days')`,
	)
	return tag.RowsAffected(), err
}
