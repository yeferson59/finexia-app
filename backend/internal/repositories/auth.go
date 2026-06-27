package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/entities"
)

func (r *Repository) GetAccountByUserID(ctx context.Context, userID uuid.UUID) (entities.Account, error) {
	var account entities.Account
	if err := r.db.QueryRow(ctx,
		"SELECT id, user_id, account_id, provider_id, password FROM accounts WHERE user_id = $1 AND provider_id = 'local'",
		userID.String(),
	).Scan(&account.ID, &account.UserID, &account.AccountID, &account.ProviderID, &account.Password); err != nil {
		return entities.Account{}, err
	}

	return account, nil
}

func (r *Repository) GetAccountByEmail(ctx context.Context, email string) (entities.User, error) {
	var account entities.Account
	var user entities.User

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
		return entities.User{}, err
	}

	user.Accounts = append(user.Accounts, account)

	return user, nil
}

func (r *Repository) CreateSession(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, "INSERT INTO sessions(user_id, token, expires_at) VALUES($1, $2, $3) RETURNING id", userID.String(), token, expiresAt).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (r *Repository) UpdateSessionToken(ctx context.Context, sessionID uuid.UUID, newToken string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, "UPDATE sessions SET token=$1, expires_at=$2, updated_at=NOW() WHERE id=$3", newToken, expiresAt, sessionID)
	return err
}

func (r *Repository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, familyID, sessionID uuid.UUID, ip, ua *string, expiresAt time.Time) (uuid.UUID, error) {
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

func (r *Repository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (entities.RefreshToken, error) {
	var rt entities.RefreshToken
	err := r.db.QueryRow(ctx,
		`SELECT rt.id, rt.user_id, rt.family_id, rt.session_id, rt.expires_at, rt.used_at, rt.revoked_at, r.name
		 FROM refresh_tokens rt
		 JOIN users u ON u.id = rt.user_id
		 JOIN roles r ON r.id = u.role_id
		 WHERE rt.token_hash = $1 AND u.deleted_at IS NULL`,
		tokenHash,
	).Scan(&rt.ID, &rt.UserID, &rt.FamilyID, &rt.SessionID, &rt.ExpiresAt, &rt.UsedAt, &rt.RevokedAt, &rt.Role)
	if err != nil {
		return entities.RefreshToken{}, err
	}
	return rt, nil
}

func (r *Repository) MarkRefreshTokenUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "UPDATE refresh_tokens SET used_at=NOW() WHERE id=$1", id)
	return err
}

func (r *Repository) RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) error {
	_, err := r.db.Exec(ctx, "UPDATE refresh_tokens SET revoked_at=NOW() WHERE family_id=$1 AND revoked_at IS NULL", familyID)
	return err
}

func (r *Repository) Register(ctx context.Context, name, email, password string) (entities.User, error) {
	user, err := r.CreateUser(ctx, name, email)
	if err != nil {
		return entities.User{}, errors.New("error create new user")
	}

	_, err = r.db.Exec(ctx, "INSERT INTO accounts(user_id, account_id, provider_id, password) VALUES($1, $2, $3, $4)", user.ID, "credentials", "local", password)
	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *Repository) GetSessionByUserIDToken(ctx context.Context, userID uuid.UUID, token string) (entities.User, error) {
	var user entities.User
	var session entities.Session

	if err := r.db.QueryRow(ctx,
		`SELECT u.id, u.name, u.email, u.email_verified, u.image, u.role_id, u.preferred_currency,
		        u.created_at, u.updated_at, u.deleted_at, u.banned_at,
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
		&user.BannedAt,
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
		return entities.User{}, err
	}

	user.Sessions = append(user.Sessions, session)

	return user, nil
}

func (r *Repository) GetSessionByToken(ctx context.Context, token string) (entities.User, error) {
	var user entities.User
	var session entities.Session

	if err := r.db.QueryRow(ctx, "SELECT u.id, u.email_verified, r.name, s.expires_at, s.token FROM users u JOIN sessions s ON s.user_id = u.id JOIN roles r ON u.role_id = r.id WHERE s.token = $1", token).Scan(
		&user.ID,
		&user.EmailVerified,
		&user.Role.Name,
		&session.ExpiresAt,
		&session.Token,
	); err != nil {
		return entities.User{}, err
	}

	user.Sessions = append(user.Sessions, session)

	return user, nil
}

func (r *Repository) DeleteSessionByUserIDToken(ctx context.Context, userID uuid.UUID, token string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM sessions WHERE user_id = $1 AND token = $2", userID.String(), token)

	return err
}
