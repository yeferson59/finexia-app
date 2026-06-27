package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/yeferson59/finexia-app/internal/entities"
)

// userCols is the explicit column list used for SELECT queries that need a JOIN with roles.
const userCols = `u.id, u.name, u.email, u.email_verified, u.image, u.role_id,
	u.preferred_currency, u.created_at, u.updated_at, u.deleted_at, u.banned_at, r.name`

func scanUserWithRole(row interface {
	Scan(...any) error
}, user *entities.User) error {
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

func (r *Repository) ListUsers(ctx context.Context, offset, limit uint) ([]entities.User, uint, error) {
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

	users := make([]entities.User, 0, limit)
	for rows.Next() {
		var user entities.User
		if err := scanUserWithRole(rows, &user); err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, count, rows.Err()
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	var user entities.User
	row := r.db.QueryRow(ctx, `
		SELECT `+userCols+`
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1
	`, id.String())
	if err := scanUserWithRole(row, &user); err != nil {
		return entities.User{}, err
	}
	return user, nil
}

func (r *Repository) CreateUser(ctx context.Context, name, email string) (entities.User, error) {
	contextTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var user entities.User
	var roleID uuid.UUID

	tx, err := r.db.BeginTx(contextTimeout, pgx.TxOptions{AccessMode: pgx.ReadWrite})
	if err != nil {
		return entities.User{}, errors.New("failed create new user")
	}

	if err := tx.QueryRow(contextTimeout, "SELECT id FROM roles WHERE name = $1", "customer").Scan(&roleID); err != nil {
		_ = tx.Rollback(contextTimeout)
		return entities.User{}, errors.New("failed create new user")
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
		return entities.User{}, errors.New("failed create new user")
	}

	user.Role.Name = "customer"
	return user, tx.Commit(contextTimeout)
}

func (r *Repository) UpdateUser(ctx context.Context, id uuid.UUID, name, email, image string) (entities.User, error) {
	var user entities.User
	if err := r.db.QueryRow(ctx,
		`UPDATE users SET name = $1, email = $2, image = $3, updated_at = $4 WHERE id = $5
		 RETURNING id, name, email, email_verified, image, role_id, preferred_currency, created_at, updated_at, deleted_at, banned_at`,
		name, email, image, time.Now(), id.String(),
	).Scan(
		&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID,
		&user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.BannedAt,
	); err != nil {
		return entities.User{}, err
	}
	return user, nil
}

func (r *Repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
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

func (r *Repository) BanUser(ctx context.Context, id uuid.UUID, ban bool) error {
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

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (entities.User, error) {
	var user entities.User
	if err := r.db.QueryRow(ctx,
		`SELECT id, name, email, email_verified, image, role_id, preferred_currency, created_at, updated_at, deleted_at, banned_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(
		&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID,
		&user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.BannedAt,
	); err != nil {
		return entities.User{}, err
	}
	return user, nil
}

func (r *Repository) UpdateUserProfile(ctx context.Context, id uuid.UUID, name, preferredCurrency, image string) (entities.User, error) {
	var user entities.User
	if err := r.db.QueryRow(ctx,
		`UPDATE users SET name = $1, preferred_currency = $2, image = $3, updated_at = $4 WHERE id = $5
		 RETURNING id, name, email, email_verified, image, role_id, preferred_currency, created_at, updated_at, deleted_at, banned_at`,
		name, preferredCurrency, image, time.Now(), id.String(),
	).Scan(
		&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID,
		&user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.BannedAt,
	); err != nil {
		return entities.User{}, err
	}
	return user, nil
}

func (r *Repository) UpdateUserImage(ctx context.Context, id uuid.UUID, image string) (entities.User, error) {
	var user entities.User
	if err := r.db.QueryRow(ctx,
		`UPDATE users SET image = $1, updated_at = $2 WHERE id = $3
		 RETURNING id, name, email, email_verified, image, role_id, preferred_currency, created_at, updated_at, deleted_at, banned_at`,
		image, time.Now(), id.String(),
	).Scan(
		&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID,
		&user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.BannedAt,
	); err != nil {
		return entities.User{}, err
	}
	return user, nil
}

func (r *Repository) GetUserPreferences(ctx context.Context, userID uuid.UUID) (entities.UserPreferences, error) {
	var prefs entities.UserPreferences
	err := r.db.QueryRow(ctx,
		"SELECT user_id, email_alerts, weekly_summary, created_at, updated_at FROM user_preferences WHERE user_id = $1",
		userID.String(),
	).Scan(&prefs.UserID, &prefs.EmailAlerts, &prefs.WeeklySummary, &prefs.CreatedAt, &prefs.UpdatedAt)
	if err != nil {
		return entities.UserPreferences{
			UserID:        userID,
			EmailAlerts:   true,
			WeeklySummary: true,
		}, nil
	}
	return prefs, nil
}

func (r *Repository) UpsertUserPreferences(ctx context.Context, userID uuid.UUID, emailAlerts, weeklySummary bool) (entities.UserPreferences, error) {
	var prefs entities.UserPreferences
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
		return entities.UserPreferences{}, err
	}
	return prefs, nil
}

func (r *Repository) UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE accounts SET password = $1, updated_at = NOW() WHERE user_id = $2 AND provider_id = 'local'",
		hashedPassword, userID.String(),
	)
	return err
}

func (r *Repository) GetUsersWithWeeklySummary(ctx context.Context) ([]entities.User, error) {
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

	var users []entities.User
	for rows.Next() {
		var u entities.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
