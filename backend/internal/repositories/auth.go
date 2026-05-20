package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/entities"
)

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

func (r *Repository) CreateSession(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, "INSERT INTO sessions(user_id, token, expires_at) VALUES($1, $2, $3)", userID.String(), token, expiresAt)
	if err != nil {
		return err
	}

	return nil
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

	if err := r.db.QueryRow(ctx, "SELECT u.*, s.* FROM users u JOIN sessions s ON s.user_id = u.id WHERE s.user_id = $1 AND s.token = $2", userID.String(), token).Scan(
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
