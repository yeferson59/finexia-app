package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/identity"
)

// PostgresRepository is the single pgx-backed implementation of every store
// interface the module declares.
type PostgresRepository struct {
	db *pgxpool.Pool
}

var (
	_ AccountStore       = (*PostgresRepository)(nil)
	_ SessionStore       = (*PostgresRepository)(nil)
	_ RefreshTokenStore  = (*PostgresRepository)(nil)
	_ TwoFactorStore     = (*PostgresRepository)(nil)
	_ VerificationStore  = (*PostgresRepository)(nil)
	_ PasswordResetStore = (*PostgresRepository)(nil)
)

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// ErrSessionNotFound indicates the session row no longer exists (e.g. the user
// logged out), so any refresh token still pointing at it must be rejected.
var ErrSessionNotFound = errors.New("session not found")

// userCols is the explicit column list used for SELECT queries that need a JOIN with roles.
const userCols = `u.id, u.name, u.email, u.email_verified, u.image, u.role_id,
	u.preferred_currency, u.created_at, u.updated_at, u.deleted_at, u.banned_at, r.name`

func scanUserWithRole(row interface {
	Scan(...any) error
}, user *identity.User) error {
	var roleName string
	if err := row.Scan(
		&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID,
		&user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.BannedAt,
		&roleName,
	); err != nil {
		return err
	}
	user.Role.Name = roleName
	return nil
}

func (r *PostgresRepository) GetAccountByUserID(ctx context.Context, userID uuid.UUID) (identity.Account, error) {
	var account identity.Account
	if err := r.db.QueryRow(ctx,
		"SELECT id, user_id, account_id, provider_id, password FROM accounts WHERE user_id = $1 AND provider_id = 'local'",
		userID.String(),
	).Scan(&account.ID, &account.UserID, &account.AccountID, &account.ProviderID, &account.Password); err != nil {
		return identity.Account{}, err
	}

	return account, nil
}

func (r *PostgresRepository) GetAccountByEmail(ctx context.Context, email string) (identity.User, error) {
	var account identity.Account
	var user identity.User

	if err := r.db.QueryRow(ctx, "SELECT u.id, u.name, u.email_verified, a.id, a.provider_id, a.account_id, a.password, r.name FROM users u JOIN accounts a ON u.id = a.user_id JOIN roles r ON u.role_id = r.id WHERE u.email = $1 AND u.deleted_at IS NULL", email).Scan(
		&user.ID,
		&user.Name,
		&user.EmailVerified,
		&account.ID,
		&account.ProviderID,
		&account.AccountID,
		&account.Password,
		&user.Role.Name,
	); err != nil {
		return identity.User{}, err
	}

	user.Accounts = append(user.Accounts, account)

	return user, nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id uuid.UUID) (identity.User, error) {
	var user identity.User
	row := r.db.QueryRow(ctx, `
		SELECT `+userCols+`
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1
	`, id.String())
	if err := scanUserWithRole(row, &user); err != nil {
		return identity.User{}, err
	}
	return user, nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (identity.User, error) {
	var user identity.User
	if err := r.db.QueryRow(ctx,
		`SELECT id, name, email, email_verified, image, role_id, preferred_currency, created_at, updated_at, deleted_at, banned_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(
		&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID,
		&user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.BannedAt,
	); err != nil {
		return identity.User{}, err
	}
	return user, nil
}

// createUser inserts a user with the default customer role. Temporary copy of
// the user repository's CreateUser (the admin CRUD keeps its own) so Register
// stays self-contained; unified when the user module is extracted in Fase 5.
func (r *PostgresRepository) createUser(ctx context.Context, name, email string) (identity.User, error) {
	contextTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var user identity.User
	var roleID uuid.UUID

	tx, err := r.db.BeginTx(contextTimeout, pgx.TxOptions{AccessMode: pgx.ReadWrite})
	if err != nil {
		return identity.User{}, errors.New("failed create new user")
	}

	if err := tx.QueryRow(contextTimeout, "SELECT id FROM roles WHERE name = $1", "customer").Scan(&roleID); err != nil {
		_ = tx.Rollback(contextTimeout)
		return identity.User{}, errors.New("failed create new user")
	}

	if err := tx.QueryRow(contextTimeout,
		`INSERT INTO users (name, email, role_id) VALUES ($1, $2, $3)
		 RETURNING id, name, email, email_verified, image, role_id, preferred_currency, created_at, updated_at, deleted_at, banned_at`,
		name, email, roleID,
	).Scan(
		&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID,
		&user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.BannedAt,
	); err != nil {
		_ = tx.Rollback(contextTimeout)
		return identity.User{}, errors.New("failed create new user")
	}

	if err := tx.Commit(contextTimeout); err != nil {
		return identity.User{}, errors.New("failed create new user")
	}

	return user, nil
}

func (r *PostgresRepository) Register(ctx context.Context, name, email, password string) (identity.User, error) {
	user, err := r.createUser(ctx, name, email)
	if err != nil {
		return identity.User{}, errors.New("error create new user")
	}

	_, err = r.db.Exec(ctx, "INSERT INTO accounts(user_id, account_id, provider_id, password) VALUES($1, $2, $3, $4)", user.ID, "credentials", "local", password)
	if err != nil {
		return identity.User{}, err
	}

	return user, nil
}

func (r *PostgresRepository) CreateSession(ctx context.Context, userID uuid.UUID, token string, ip, ua *string, expiresAt time.Time) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx,
		"INSERT INTO sessions(user_id, token, ip_address, user_agent, expires_at) VALUES($1, $2, $3, $4, $5) RETURNING id",
		userID.String(), token, ip, ua, expiresAt,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
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

// DeleteSessionsByIDs deletes the given sessions, scoped to the owner so a
// user can never revoke another user's session. Refresh tokens cascade.
func (r *PostgresRepository) DeleteSessionsByIDs(ctx context.Context, userID uuid.UUID, sessionIDs []uuid.UUID) (int64, error) {
	tag, err := r.db.Exec(ctx,
		"DELETE FROM sessions WHERE user_id = $1 AND id = ANY($2)",
		userID.String(), sessionIDs,
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

func (r *PostgresRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, familyID, sessionID uuid.UUID, ip, ua *string, expiresAt time.Time) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx,
		"INSERT INTO refresh_tokens(user_id, token_hash, family_id, session_id, ip_address, user_agent, expires_at) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id",
		userID, tokenHash, familyID, sessionID, ip, ua, expiresAt,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, err
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

	if err := r.db.QueryRow(ctx, "SELECT u.id, u.email_verified, r.name, s.expires_at, s.token FROM users u JOIN sessions s ON s.user_id = u.id JOIN roles r ON u.role_id = r.id WHERE s.token = $1", token).Scan(
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
