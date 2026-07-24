package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// invitationCols is the column list shared by the invitation SELECT queries.
const invitationCols = `id, email, name, role, token_hash, invited_by,
	expires_at, accepted_at, revoked_at, created_at, updated_at`

// ErrInvitationNotFound signals that no invitation matched the lookup, letting
// the service map it to a 404 without leaking whether a token ever existed.
var ErrInvitationNotFound = httpx.AsNotFound(errors.New("invitation not found"))

func scanInvitation(row interface {
	Scan(...any) error
}, inv *Invitation) error {
	return row.Scan(
		&inv.ID, &inv.Email, &inv.Name, &inv.Role, &inv.TokenHash, &inv.InvitedBy,
		&inv.ExpiresAt, &inv.AcceptedAt, &inv.RevokedAt, &inv.CreatedAt, &inv.UpdatedAt,
	)
}

// CreateInvitation inserts a fresh invitation, or — when a live (neither
// accepted nor revoked) one already exists for the email — replaces its token
// and expiry. The partial unique index guarantees only one redeemable link per
// email, so resending or re-inviting can never leave two valid tokens behind.
func (r *PostgresRepository) CreateInvitation(ctx context.Context, email, name, role, tokenHash string, invitedBy *uuid.UUID, expiresAt time.Time) (Invitation, error) {
	var inv Invitation
	row := r.db.QueryRow(ctx,
		`INSERT INTO invitations (email, name, role, token_hash, invited_by, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (email) WHERE accepted_at IS NULL AND revoked_at IS NULL
		 DO UPDATE SET name = EXCLUDED.name, role = EXCLUDED.role,
		               token_hash = EXCLUDED.token_hash, invited_by = EXCLUDED.invited_by,
		               expires_at = EXCLUDED.expires_at, updated_at = NOW()
		 RETURNING `+invitationCols,
		email, name, role, tokenHash, invitedBy, expiresAt,
	)
	if err := scanInvitation(row, &inv); err != nil {
		return Invitation{}, err
	}
	return inv, nil
}

func (r *PostgresRepository) GetInvitationByHash(ctx context.Context, tokenHash string) (Invitation, error) {
	var inv Invitation
	row := r.db.QueryRow(ctx, `SELECT `+invitationCols+` FROM invitations WHERE token_hash = $1`, tokenHash)
	if err := scanInvitation(row, &inv); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Invitation{}, ErrInvitationNotFound
		}
		return Invitation{}, err
	}
	return inv, nil
}

func (r *PostgresRepository) GetInvitationByID(ctx context.Context, id uuid.UUID) (Invitation, error) {
	var inv Invitation
	row := r.db.QueryRow(ctx, `SELECT `+invitationCols+` FROM invitations WHERE id = $1`, id)
	if err := scanInvitation(row, &inv); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Invitation{}, ErrInvitationNotFound
		}
		return Invitation{}, err
	}
	return inv, nil
}

// ListInvitations returns the still-open invitations (neither accepted nor
// revoked) newest first, so the dashboard shows exactly the ones an admin can
// still act on. Expired-but-unrevoked rows stay in the list so they can be
// resent or revoked explicitly.
func (r *PostgresRepository) ListInvitations(ctx context.Context, offset, limit uint) ([]Invitation, uint, error) {
	var count uint
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM invitations WHERE accepted_at IS NULL AND revoked_at IS NULL`,
	).Scan(&count); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT `+invitationCols+`
		 FROM invitations
		 WHERE accepted_at IS NULL AND revoked_at IS NULL
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	invitations := make([]Invitation, 0, limit)
	for rows.Next() {
		var inv Invitation
		if err := scanInvitation(rows, &inv); err != nil {
			return nil, 0, err
		}
		invitations = append(invitations, inv)
	}
	return invitations, count, rows.Err()
}

// RevokeInvitation marks a still-pending invitation as revoked. Revoking an
// already-accepted or already-revoked invitation is a no-op that reports
// ErrInvitationNotFound, so the caller never double-revokes.
func (r *PostgresRepository) RevokeInvitation(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE invitations SET revoked_at = NOW(), updated_at = NOW()
		 WHERE id = $1 AND accepted_at IS NULL AND revoked_at IS NULL`,
		id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrInvitationNotFound
	}
	return nil
}

// AcceptInvitation atomically turns an invitation into a real account: it
// creates the user (email already verified — clicking the link proves control
// of the inbox) and their local credentials, marks the invitation consumed, and
// advances any matching waitlist row to "registered". The invitation is only
// consumed if still pending, so two concurrent accepts cannot both succeed.
func (r *PostgresRepository) AcceptInvitation(ctx context.Context, invitationID uuid.UUID, name, email, role, passwordHash string) (identity.User, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctxTimeout, pgx.TxOptions{AccessMode: pgx.ReadWrite})
	if err != nil {
		return identity.User{}, httpx.AsBadRequest(errors.New("failed to accept invitation"))
	}
	defer func() { _ = tx.Rollback(ctxTimeout) }()

	// Consume the invitation first: if it is no longer pending (expired rows are
	// filtered by the service, but a race could have revoked or accepted it),
	// RowsAffected is 0 and we abort before creating any account.
	tag, err := tx.Exec(ctxTimeout,
		`UPDATE invitations SET accepted_at = NOW(), updated_at = NOW()
		 WHERE id = $1 AND accepted_at IS NULL AND revoked_at IS NULL`,
		invitationID,
	)
	if err != nil {
		return identity.User{}, err
	}
	if tag.RowsAffected() == 0 {
		return identity.User{}, ErrInvitationNotFound
	}

	var roleID uuid.UUID
	if err := tx.QueryRow(ctxTimeout, "SELECT id FROM roles WHERE name = $1", role).Scan(&roleID); err != nil {
		return identity.User{}, httpx.AsBadRequest(errors.New("invalid role"))
	}

	var user identity.User
	if err := tx.QueryRow(ctxTimeout,
		`INSERT INTO users (name, email, email_verified, role_id)
		 VALUES ($1, $2, TRUE, $3)
		 RETURNING id, name, email, email_verified, image, role_id, preferred_currency, created_at, updated_at, deleted_at, banned_at`,
		name, email, roleID,
	).Scan(
		&user.ID, &user.Name, &user.Email, &user.EmailVerified, &user.Image, &user.RoleID,
		&user.PreferredCurrency, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.BannedAt,
	); err != nil {
		return identity.User{}, httpx.AsConflict(errors.New("user already exists"))
	}

	if _, err := tx.Exec(ctxTimeout,
		`INSERT INTO accounts (user_id, account_id, provider_id, password)
		 VALUES ($1, $2, $3, $4)`,
		user.ID, "credentials", "local", passwordHash,
	); err != nil {
		return identity.User{}, err
	}

	// Best-effort within the transaction: advancing the funnel must not fail the
	// accept if the person was invited directly without ever joining the list.
	if _, err := tx.Exec(ctxTimeout,
		`UPDATE waitlist SET status = 'registered' WHERE email = $1`, email,
	); err != nil {
		return identity.User{}, err
	}

	user.Role.Name = role
	return user, tx.Commit(ctxTimeout)
}
