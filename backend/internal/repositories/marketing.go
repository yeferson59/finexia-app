package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) SaveWaitlistEmail(ctx context.Context, email string) error {
	_, err := r.db.Exec(ctx, "INSERT INTO waitlist (email) VALUES ($1)", email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	return nil
}
