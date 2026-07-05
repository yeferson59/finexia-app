package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/yeferson59/finexia-app/internal/entities"
)

// ErrVerificationNotFound signals that no verification row matched the
// lookup, or that the row was gone by the time it was consumed.
var ErrVerificationNotFound = errors.New("verification not found")

// CreateEmailVerification invalidates any pending verification for the email
// (so at most one link stays redeemable) and inserts a fresh one.
func (r *Repository) CreateEmailVerification(ctx context.Context, email, tokenHash string, expiresAt time.Time) (entities.Verification, error) {
	if _, err := r.db.Exec(ctx, `DELETE FROM verifications WHERE identifier = $1`, email); err != nil {
		return entities.Verification{}, err
	}

	var v entities.Verification
	if err := r.db.QueryRow(ctx,
		`INSERT INTO verifications (identifier, value, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, identifier, value, expires_at, created_at, updated_at`,
		email, tokenHash, expiresAt,
	).Scan(&v.ID, &v.Identifier, &v.Value, &v.ExpiresAt, &v.CreatedAt, &v.UpdatedAt); err != nil {
		return entities.Verification{}, err
	}
	return v, nil
}

func (r *Repository) GetEmailVerificationByHash(ctx context.Context, tokenHash string) (entities.Verification, error) {
	var v entities.Verification
	err := r.db.QueryRow(ctx,
		`SELECT id, identifier, value, expires_at, created_at, updated_at
		 FROM verifications WHERE value = $1`,
		tokenHash,
	).Scan(&v.ID, &v.Identifier, &v.Value, &v.ExpiresAt, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Verification{}, ErrVerificationNotFound
		}
		return entities.Verification{}, err
	}
	return v, nil
}

// ConsumeEmailVerification marks the account's email verified and deletes the
// verification row in one transaction, scoped by id so a token can never be
// redeemed twice even under concurrent confirm requests.
func (r *Repository) ConsumeEmailVerification(ctx context.Context, id uuid.UUID, email string) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctxTimeout, pgx.TxOptions{AccessMode: pgx.ReadWrite})
	if err != nil {
		return errors.New("failed to verify email")
	}
	defer func() { _ = tx.Rollback(ctxTimeout) }()

	tag, err := tx.Exec(ctxTimeout, `DELETE FROM verifications WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrVerificationNotFound
	}

	if _, err := tx.Exec(ctxTimeout,
		`UPDATE users SET email_verified = true, updated_at = NOW() WHERE email = $1 AND deleted_at IS NULL`,
		email,
	); err != nil {
		return err
	}

	return tx.Commit(ctxTimeout)
}
