package auth

import (
	"context"
	"errors"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// SessionStore: the live sessions a login creates, plus the known-login-IP
// trail the new-device alert reads.

// ErrSessionNotFound indicates the session row no longer exists (e.g. the user
// logged out), so any refresh token still pointing at it must be rejected.
var ErrSessionNotFound = httpx.AsNotFound(errors.New("session not found"))

func (r *PostgresRepository) CreateSession(ctx context.Context, userID uuid.UUID, token string, ip, ua *string, expiresAt time.Time) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx,
		"INSERT INTO sessions(user_id, token, ip_address, user_agent, expires_at) VALUES($1, $2, $3, $4, $5) RETURNING id",
		userID.String(), token, ip, ua, expiresAt,
	).Scan(&id)
	if err != nil {
		return uuid.Nil(), err
	}

	return id, nil
}

// UpdateSessionToken swaps the session's access token and returns the previous
// one so callers can invalidate its cache entry.
func (r *PostgresRepository) UpdateSessionToken(ctx context.Context, sessionID uuid.UUID, newToken string, expiresAt time.Time) (string, error) {
	var oldToken string
	err := r.db.QueryRow(ctx,
		`UPDATE sessions s SET token=$1, expires_at=$2, updated_at=NOW()
		 FROM sessions old
		 WHERE s.id=$3 AND old.id=s.id
		 RETURNING old.token`,
		newToken, expiresAt, sessionID,
	).Scan(&oldToken)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrSessionNotFound
	}
	if err != nil {
		return "", err
	}
	return oldToken, nil
}

// UpdateSessionLocation stamps the approximate location resolved from the
// session's IP. Runs after session creation (the lookup is asynchronous), so
// a missing row — session already revoked — is not an error.
func (r *PostgresRepository) UpdateSessionLocation(ctx context.Context, sessionID uuid.UUID, location string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE sessions SET location = $1 WHERE id = $2",
		location, sessionID.String(),
	)
	return err
}

// ListSessionsByUserID returns the user's live sessions: those whose access
// token has not expired, or that still hold a redeemable refresh token (the
// cleanup job uses the same liveness rule before deleting a session).
func (r *PostgresRepository) ListSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]identity.Session, error) {
	rows, err := r.db.Query(ctx,
		`SELECT s.id, s.user_id, s.token, s.expires_at, s.ip_address, s.user_agent, s.location, s.created_at, s.updated_at
		 FROM sessions s
		 WHERE s.user_id = $1
		   AND (s.expires_at > NOW() OR EXISTS (
		     SELECT 1 FROM refresh_tokens rt
		     WHERE rt.session_id = s.id
		       AND rt.revoked_at IS NULL
		       AND rt.expires_at > NOW()
		   ))
		 ORDER BY s.updated_at DESC`,
		userID.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []identity.Session
	for rows.Next() {
		var s identity.Session
		if err := rows.Scan(&s.ID, &s.UserID, &s.Token, &s.ExpiresAt, &s.IPAddress, &s.UserAgent, &s.Location, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (r *PostgresRepository) GetSessionByUserIDToken(ctx context.Context, userID uuid.UUID, token string) (identity.User, error) {
	var user identity.User
	var session identity.Session

	if err := r.db.QueryRow(ctx,
		`SELECT u.id, u.name, u.email, u.email_verified, u.image, u.role_id, u.preferred_currency,
		        u.created_at, u.updated_at, u.deleted_at,
		        r.name,
		        s.id, s.user_id, s.token, s.expires_at, s.ip_address, s.user_agent, s.created_at, s.updated_at
		 FROM users u
		 JOIN sessions s ON s.user_id = u.id
		 JOIN roles r ON r.id = u.role_id
		 WHERE s.user_id = $1 AND s.token = $2`,
		userID.String(), token,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.EmailVerified,
		&user.Image,
		&user.RoleID,
		&user.PreferredCurrency,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&user.Role.Name,
		&session.ID,
		&session.UserID,
		&session.Token,
		&session.ExpiresAt,
		&session.IPAddress,
		&session.UserAgent,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		return identity.User{}, err
	}

	user.Sessions = append(user.Sessions, session)

	return user, nil
}

func (r *PostgresRepository) GetSessionByToken(ctx context.Context, token string) (identity.User, error) {
	var user identity.User
	var session identity.Session

	// The user-state filter is the point of this query, not decoration: it is
	// the only database read on the access-token validation path, so a session
	// belonging to a banned or soft-deleted user has to stop resolving here or
	// that user keeps every privilege they had until the token expires.
	if err := r.db.QueryRow(ctx, "SELECT u.id, u.email_verified, r.name, s.expires_at, s.token FROM users u JOIN sessions s ON s.user_id = u.id JOIN roles r ON u.role_id = r.id WHERE s.token = $1 AND u.deleted_at IS NULL AND u.banned_at IS NULL", token).Scan(
		&user.ID,
		&user.EmailVerified,
		&user.Role.Name,
		&session.ExpiresAt,
		&session.Token,
	); err != nil {
		return identity.User{}, err
	}

	user.Sessions = append(user.Sessions, session)

	return user, nil
}

func (r *PostgresRepository) DeleteSessionByUserIDToken(ctx context.Context, userID uuid.UUID, token string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM sessions WHERE user_id = $1 AND token = $2", userID.String(), token)

	return err
}

// DeleteSessionsByIDs deletes the given sessions, scoped to the owner so a
// user can never revoke another user's session. Refresh tokens cascade.
func (r *PostgresRepository) DeleteSessionsByIDs(ctx context.Context, userID uuid.UUID, sessionIDs []uuid.UUID) (int64, error) {
	tag, err := r.db.Exec(ctx,
		"DELETE FROM sessions WHERE user_id = $1 AND id = ANY($2)",
		userID.String(), sessionIDs,
	)
	return tag.RowsAffected(), err
}

// DeleteExpiredSessions removes sessions whose access token expired and that
// have no live refresh token left. Sessions with a live refresh token must be
// kept even if expires_at is in the past: refresh_tokens.session_id cascades on
// delete, so removing the session would silently log the user out.
func (r *PostgresRepository) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM sessions s
		 WHERE s.expires_at < NOW()
		   AND NOT EXISTS (
		     SELECT 1 FROM refresh_tokens rt
		     WHERE rt.session_id = s.id
		       AND rt.revoked_at IS NULL
		       AND rt.expires_at > NOW()
		   )`,
	)
	return tag.RowsAffected(), err
}

// HasKnownLoginIP reports whether the user has ever logged in from the given
// IP. Backed by known_login_ips rather than sessions: sessions are deleted on
// logout and swept on expiry, so that table can't answer "have we seen this
// device before" once the session that recorded it is gone.
func (r *PostgresRepository) HasKnownLoginIP(ctx context.Context, userID uuid.UUID, ip string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM known_login_ips WHERE user_id = $1 AND ip_address = $2)",
		userID.String(), ip,
	).Scan(&exists)
	return exists, err
}

// RecordKnownLoginIP remembers that the user has logged in from ip, so a
// later login from the same address is not flagged as a new device even
// after this session is logged out or expires.
func (r *PostgresRepository) RecordKnownLoginIP(ctx context.Context, userID uuid.UUID, ip string) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO known_login_ips(user_id, ip_address) VALUES ($1, $2) ON CONFLICT (user_id, ip_address) DO NOTHING",
		userID.String(), ip,
	)
	return err
}
