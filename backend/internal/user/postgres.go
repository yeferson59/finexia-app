package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yeferson59/finexia-app/internal/identity"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

var _ Repository = (*PostgresRepository)(nil)

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return new(PostgresRepository{db})
}

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

func (r *PostgresRepository) List(ctx context.Context, offset, limit uint) ([]identity.User, uint, error) {
	var count uint
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&count); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT `+userCols+`
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.deleted_at IS NULL
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]identity.User, 0, limit)
	for rows.Next() {
		var user identity.User
		if err := scanUserWithRole(rows, &user); err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, count, rows.Err()
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (identity.User, error) {
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

func (r *PostgresRepository) Create(ctx context.Context, name, email string) (identity.User, error) {
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

	user.Role.Name = "customer"
	return user, tx.Commit(contextTimeout)
}

func (r *PostgresRepository) Update(ctx context.Context, id uuid.UUID, name, email, image string) (identity.User, error) {
	var user identity.User
	if err := r.db.QueryRow(ctx,
		`UPDATE users SET name = $1, email = $2, image = $3, updated_at = $4 WHERE id = $5
		 RETURNING id, name, email, email_verified, image, role_id, preferred_currency, created_at, updated_at, deleted_at, banned_at`,
		name, email, image, time.Now(), id.String(),
	).Scan(
		&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID,
		&user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.BannedAt,
	); err != nil {
		return identity.User{}, err
	}
	return user, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var roleName string
	err := r.db.QueryRow(ctx,
		`SELECT r.name FROM users u JOIN roles r ON r.id = u.role_id WHERE u.id = $1`,
		id.String(),
	).Scan(&roleName)
	if err == nil && roleName == "admin" {
		return errors.New("cannot delete an admin user")
	}

	_, err = r.db.Exec(ctx,
		"UPDATE users SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL",
		time.Now(), id.String(),
	)
	return err
}

func (r *PostgresRepository) Ban(ctx context.Context, id uuid.UUID, ban bool) error {
	var bannedAt any
	if ban {
		bannedAt = time.Now()
	}
	_, err := r.db.Exec(ctx,
		"UPDATE users SET banned_at = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL",
		bannedAt, id.String(),
	)
	return err
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (identity.User, error) {
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

func (r *PostgresRepository) UpdateProfile(ctx context.Context, id uuid.UUID, name, preferredCurrency, image string) (identity.User, error) {
	var user identity.User
	if err := r.db.QueryRow(ctx,
		`UPDATE users SET name = $1, preferred_currency = $2, image = $3, updated_at = $4 WHERE id = $5
		 RETURNING id, name, email, email_verified, image, role_id, preferred_currency, created_at, updated_at, deleted_at, banned_at`,
		name, preferredCurrency, image, time.Now(), id.String(),
	).Scan(
		&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID,
		&user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.BannedAt,
	); err != nil {
		return identity.User{}, err
	}
	return user, nil
}

func (r *PostgresRepository) UpdateImage(ctx context.Context, id uuid.UUID, image string) (identity.User, error) {
	var user identity.User
	if err := r.db.QueryRow(ctx,
		`UPDATE users SET image = $1, updated_at = $2 WHERE id = $3
		 RETURNING id, name, email, email_verified, image, role_id, preferred_currency, created_at, updated_at, deleted_at, banned_at`,
		image, time.Now(), id.String(),
	).Scan(
		&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID,
		&user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.BannedAt,
	); err != nil {
		return identity.User{}, err
	}
	return user, nil
}

func (r *PostgresRepository) GetPreferences(ctx context.Context, userID uuid.UUID) (UserPreferences, error) {
	var prefs UserPreferences
	err := r.db.QueryRow(ctx,
		"SELECT user_id, email_alerts, weekly_summary, created_at, updated_at FROM user_preferences WHERE user_id = $1",
		userID.String(),
	).Scan(&prefs.UserID, &prefs.EmailAlerts, &prefs.WeeklySummary, &prefs.CreatedAt, &prefs.UpdatedAt)
	if err != nil {
		return UserPreferences{
			UserID:        userID,
			EmailAlerts:   true,
			WeeklySummary: true,
		}, nil
	}
	return prefs, nil
}

func (r *PostgresRepository) UpsertPreferences(ctx context.Context, userID uuid.UUID, emailAlerts, weeklySummary bool) (UserPreferences, error) {
	var prefs UserPreferences
	if err := r.db.QueryRow(ctx,
		`INSERT INTO user_preferences (user_id, email_alerts, weekly_summary)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id) DO UPDATE
		   SET email_alerts = EXCLUDED.email_alerts,
		       weekly_summary = EXCLUDED.weekly_summary,
		       updated_at = NOW()
		 RETURNING user_id, email_alerts, weekly_summary, created_at, updated_at`,
		userID.String(), emailAlerts, weeklySummary,
	).Scan(&prefs.UserID, &prefs.EmailAlerts, &prefs.WeeklySummary, &prefs.CreatedAt, &prefs.UpdatedAt); err != nil {
		return UserPreferences{}, err
	}
	return prefs, nil
}

func (r *PostgresRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE accounts SET password = $1, updated_at = NOW() WHERE user_id = $2 AND provider_id = 'local'",
		hashedPassword, userID.String(),
	)
	return err
}

func (r *PostgresRepository) GetWeeklySummary(ctx context.Context) ([]identity.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.name, u.email
		FROM users u
		JOIN user_preferences up ON up.user_id = u.id
		WHERE up.weekly_summary = true AND u.deleted_at IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []identity.User

	for rows.Next() {
		var u identity.User

		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, rows.Err()
}
