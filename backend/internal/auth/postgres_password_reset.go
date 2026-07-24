package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// ErrPasswordResetNotFound signals that no reset row matched the lookup, or
// that the row was no longer redeemable at the moment it was consumed.
var ErrPasswordResetNotFound = httpx.AsNotFound(errors.New("password reset not found"))

// CreatePasswordReset invalidates any still-usable reset for the user (so at
// most one link stays redeemable, mirroring how invitations replace a live
// token instead of piling up) and inserts a fresh one.
func (r *PostgresRepository) CreatePasswordReset(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (PasswordReset, error) {
	if _, err := r.db.Exec(ctx,
		`UPDATE password_resets SET used_at = NOW() WHERE user_id = $1 AND used_at IS NULL`,
		userID,
	); err != nil {
		return PasswordReset{}, err
	}

	var pr PasswordReset
	if err := r.db.QueryRow(ctx,
		`INSERT INTO password_resets (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, token_hash, expires_at, used_at, created_at`,
		userID, tokenHash, expiresAt,
	).Scan(&pr.ID, &pr.UserID, &pr.TokenHash, &pr.ExpiresAt, &pr.UsedAt, &pr.CreatedAt); err != nil {
		return PasswordReset{}, err
	}
	return pr, nil
}

func (r *PostgresRepository) GetPasswordResetByHash(ctx context.Context, tokenHash string) (PasswordReset, error) {
	var pr PasswordReset
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, used_at, created_at
		 FROM password_resets WHERE token_hash = $1`,
		tokenHash,
	).Scan(&pr.ID, &pr.UserID, &pr.TokenHash, &pr.ExpiresAt, &pr.UsedAt, &pr.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PasswordReset{}, ErrPasswordResetNotFound
		}
		return PasswordReset{}, err
	}
	return pr, nil
}

// ConsumePasswordReset atomically marks the reset used and updates the local
// account's password in one transaction, so a token can never be replayed
// even under concurrent confirm requests.
func (r *PostgresRepository) ConsumePasswordReset(ctx context.Context, resetID, userID uuid.UUID, hashedPassword string) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctxTimeout, pgx.TxOptions{AccessMode: pgx.ReadWrite})
	if err != nil {
		return httpx.AsBadRequest(errors.New("failed to reset password"))
	}
	defer func() { _ = tx.Rollback(ctxTimeout) }()

	tag, err := tx.Exec(ctxTimeout,
		`UPDATE password_resets SET used_at = NOW() WHERE id = $1 AND used_at IS NULL`,
		resetID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrPasswordResetNotFound
	}

	if _, err := tx.Exec(ctxTimeout,
		`UPDATE accounts SET password = $1, updated_at = NOW() WHERE user_id = $2 AND provider_id = 'local'`,
		hashedPassword, userID,
	); err != nil {
		return err
	}

	return tx.Commit(ctxTimeout)
}
