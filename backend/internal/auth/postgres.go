package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/user"
)

// This file holds the shared repository type plus the AccountStore
// implementation. The other stores live in siblings named after them
// (postgres_session.go, postgres_refresh_token.go, postgres_twofactor.go,
// postgres_verification.go, postgres_password_reset.go, postgres_invitation.go)
// so no single file carries the whole persistence surface.

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
	return new(PostgresRepository{db})
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

// UpdatePassword replaces the local account's password hash. The
// password-reset flow has its own transactional variant (ConsumePasswordReset)
// because it must also burn the token; this is the plain write used when an
// authenticated user changes their own password.
func (r *PostgresRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE accounts SET password = $1, updated_at = NOW() WHERE user_id = $2 AND provider_id = 'local'",
		hashedPassword, userID.String(),
	)

	return err
}

// Register creates the pair a sign-up needs — the users row and its local
// credentials row — in a single transaction. The two writes must stand or fall
// together: an account insert that failed after the user was already committed
// would leave someone who can neither log in (no credentials) nor sign up again
// (the email is taken), with no way out from the outside.
//
// The users row is written through user.InsertUser rather than a query of our
// own: that table belongs to the user module, and this transaction is the only
// place auth writes it. Reads of users/roles go through UserReader — see
// Deps.Users.
func (r *PostgresRepository) Register(ctx context.Context, name, email, password string) (identity.User, error) {
	contextTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(contextTimeout, pgx.TxOptions{AccessMode: pgx.ReadWrite})
	if err != nil {
		return identity.User{}, errors.New("error create new user")
	}
	// A rollback after a successful commit is a no-op, so this covers every
	// early return without a flag.
	defer func() { _ = tx.Rollback(contextTimeout) }()

	created, err := user.InsertUser(contextTimeout, tx, name, email)
	if err != nil {
		return identity.User{}, errors.New("error create new user")
	}

	if _, err := tx.Exec(contextTimeout,
		"INSERT INTO accounts(user_id, account_id, provider_id, password) VALUES($1, $2, $3, $4)",
		created.ID, "credentials", "local", password,
	); err != nil {
		return identity.User{}, err
	}

	if err := tx.Commit(contextTimeout); err != nil {
		return identity.User{}, err
	}

	return created, nil
}
