package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// ErrTwoFactorNotFound signals the user has no 2FA row at all (never started
// a setup, or it was removed).
var ErrTwoFactorNotFound = httpx.AsNotFound(errors.New("two-factor not found"))

func (r *PostgresRepository) GetTwoFactor(ctx context.Context, userID uuid.UUID) (TwoFactor, error) {
	var tf TwoFactor
	err := r.db.QueryRow(ctx,
		`SELECT user_id, secret, enabled, confirmed_at, created_at, updated_at
		 FROM user_two_factor WHERE user_id = $1`,
		userID.String(),
	).Scan(&tf.UserID, &tf.Secret, &tf.Enabled, &tf.ConfirmedAt, &tf.CreatedAt, &tf.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return TwoFactor{}, ErrTwoFactorNotFound
	}
	if err != nil {
		return TwoFactor{}, err
	}
	return tf, nil
}

// UpsertTwoFactorSecret stores a fresh (unconfirmed) secret. The ON CONFLICT
// guard refuses to overwrite an already-enabled enrollment, so a stray setup
// call can never silently rotate the secret of an active 2FA.
func (r *PostgresRepository) UpsertTwoFactorSecret(ctx context.Context, userID uuid.UUID, secret string) error {
	tag, err := r.db.Exec(ctx,
		`INSERT INTO user_two_factor (user_id, secret, enabled)
		 VALUES ($1, $2, FALSE)
		 ON CONFLICT (user_id) DO UPDATE
		   SET secret = EXCLUDED.secret,
		       enabled = FALSE,
		       confirmed_at = NULL,
		       updated_at = NOW()
		   WHERE user_two_factor.enabled = FALSE`,
		userID.String(), secret,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("two-factor already enabled")
	}
	return nil
}

// EnableTwoFactor flips a pending enrollment to enabled. It only matches
// enabled = FALSE rows so a double confirm cannot re-enable anything.
func (r *PostgresRepository) EnableTwoFactor(ctx context.Context, userID uuid.UUID) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE user_two_factor
		 SET enabled = TRUE, confirmed_at = NOW(), updated_at = NOW()
		 WHERE user_id = $1 AND enabled = FALSE`,
		userID.String(),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrTwoFactorNotFound
	}
	return nil
}

// DeleteTwoFactor removes the enrollment and every recovery code in one
// transaction, returning the account to the "2FA disabled" default.
func (r *PostgresRepository) DeleteTwoFactor(ctx context.Context, userID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx,
		"DELETE FROM user_two_factor_recovery_codes WHERE user_id = $1", userID.String(),
	); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		"DELETE FROM user_two_factor WHERE user_id = $1", userID.String(),
	); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ReplaceTwoFactorRecoveryCodes atomically swaps the user's recovery codes
// for a new batch of hashes.
func (r *PostgresRepository) ReplaceTwoFactorRecoveryCodes(ctx context.Context, userID uuid.UUID, codeHashes []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx,
		"DELETE FROM user_two_factor_recovery_codes WHERE user_id = $1", userID.String(),
	); err != nil {
		return err
	}
	for _, hash := range codeHashes {
		if _, err := tx.Exec(ctx,
			"INSERT INTO user_two_factor_recovery_codes (user_id, code_hash) VALUES ($1, $2)",
			userID.String(), hash,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ConsumeTwoFactorRecoveryCode marks a recovery code as used. The UPDATE only
// matches unused codes, so each code works exactly once.
func (r *PostgresRepository) ConsumeTwoFactorRecoveryCode(ctx context.Context, userID uuid.UUID, codeHash string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE user_two_factor_recovery_codes
		 SET used_at = NOW()
		 WHERE user_id = $1 AND code_hash = $2 AND used_at IS NULL`,
		userID.String(), codeHash,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httpx.AsBadRequest(errors.New("invalid recovery code"))
	}
	return nil
}

func (r *PostgresRepository) CountUnusedTwoFactorRecoveryCodes(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM user_two_factor_recovery_codes WHERE user_id = $1 AND used_at IS NULL",
		userID.String(),
	).Scan(&count)
	return count, err
}
