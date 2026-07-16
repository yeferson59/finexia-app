package marketing

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository is the pgx-backed implementation of Repository.
type PostgresRepository struct {
	db *pgxpool.Pool
}

var _ Repository = (*PostgresRepository)(nil)

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) SaveWaitlistEmail(ctx context.Context, email string) error {
	_, err := r.db.Exec(ctx, "INSERT INTO waitlist (email) VALUES ($1)", email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	return nil
}

func (r *PostgresRepository) ListWaitlist(ctx context.Context, offset, limit uint) ([]Waitlist, uint, error) {
	var count uint
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM waitlist`).Scan(&count); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, email, status, invited_at, created_at
		 FROM waitlist
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	waitlist := make([]Waitlist, 0, limit)
	for rows.Next() {
		var w Waitlist
		if err := rows.Scan(&w.ID, &w.Email, &w.Status, &w.InvitedAt, &w.CreatedAt); err != nil {
			return nil, 0, err
		}
		waitlist = append(waitlist, w)
	}
	return waitlist, count, rows.Err()
}

// SetWaitlistInvited advances a waitlist row to "invited" and stamps the time.
// It is a no-op for emails that never joined the list, so admins can invite
// anyone without first seeding the waitlist.
func (r *PostgresRepository) SetWaitlistInvited(ctx context.Context, email string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE waitlist SET status = 'invited', invited_at = NOW()
		 WHERE email = $1 AND status = 'pending'`,
		email,
	)
	return err
}
