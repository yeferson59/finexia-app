package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/yeferson59/finexia-app/internal/entities"
)

func (r *Repository) ListUsers(ctx context.Context, offset, limit uint) ([]entities.User, uint, error) {
	var count uint

	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, "SELECT * FROM users LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]entities.User, 0, limit)

	for rows.Next() {
		var user entities.User

		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID, &user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
			return nil, 0, err
		}

		users = append(users, user)
	}

	return users, count, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	var user entities.User

	if err := r.db.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", id.String()).Scan(&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID, &user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *Repository) CreateUser(ctx context.Context, name, email string) (entities.User, error) {
	contextTimeout, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var user entities.User
	var roleID uuid.UUID

	if tx, err := r.db.BeginTx(contextTimeout, pgx.TxOptions{AccessMode: pgx.ReadWrite}); err == nil {
		if err := tx.QueryRow(contextTimeout, "SELECT id FROM roles WHERE name = $1", "customer").Scan(&roleID); err != nil {
			return entities.User{}, errors.New("failed create new user")
		}

		if tx.QueryRow(contextTimeout, "INSERT INTO users (name, email, role_id) VALUES ($1, $2, $3) RETURNING *", name, email, roleID).Scan(&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID, &user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt) != nil {
			return entities.User{}, errors.New("failed create new user")
		}

		return user, tx.Commit(contextTimeout)
	}

	return user, errors.New("failed create new user")
}

func (r *Repository) UpdateUser(ctx context.Context, id uuid.UUID, name, email, image string) (entities.User, error) {
	var user entities.User
	if err := r.db.QueryRow(ctx, "UPDATE users SET name = $1, email = $2, image = $3, updated_at = $4 WHERE id = $5 RETURNING *", name, email, image, time.Now(), id.String()).Scan(&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID, &user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *Repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "UPDATE users SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL", time.Now(), id.String())

	return err
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (entities.User, error) {
	var user entities.User

	if err := r.db.QueryRow(ctx, "SELECT * FROM users WHERE email = $1", email).Scan(&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID, &user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *Repository) UpdateUserProfile(ctx context.Context, id uuid.UUID, name, preferredCurrency, image string) (entities.User, error) {
	var user entities.User
	if err := r.db.QueryRow(ctx,
		"UPDATE users SET name = $1, preferred_currency = $2, image = $3, updated_at = $4 WHERE id = $5 RETURNING *",
		name, preferredCurrency, image, time.Now(), id.String(),
	).Scan(&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID, &user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
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
