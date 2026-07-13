package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// TxBeginner is the subset of *pgxpool.Pool needed to open a transaction,
// kept as an interface so WithinTx can be tested without a live database.
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// WithinTx runs fn inside a single transaction: it commits when fn returns
// nil and rolls back when fn returns an error or panics. Intended for
// services that need multi-statement atomicity (e.g. bulk imports).
func WithinTx(ctx context.Context, db TxBeginner, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	// Rollback after a successful Commit is a harmless no-op; this defer is
	// what guarantees the rollback on early returns and panics.
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(ctx, tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
