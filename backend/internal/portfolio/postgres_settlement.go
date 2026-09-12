package portfolio

import (
	"context"
	"errors"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/finexia-app/internal/platform/database"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// ChangeEntrySettlement restates a position the caller owns in another cost
// currency, with the rate each of its transactions converted at.
//
// It is one database transaction for the reason UpdateTransaction is one: the
// rates can only be judged against the whole history, which has to be read, and
// locked, before anything is written. The position is locked first, so a
// transaction added to it meanwhile waits and is judged against the currency it
// will actually be recorded in.
//
// Only the currency and the rates change. trg_recalculate_avg_cost reprices the
// position from the rewritten rates, and no rewrite moves a quantity, so the
// growth series keeps every flow where it was and retires nothing (000038): the
// cost was wrong, the holding never was.
func (r *PostgresRepository) ChangeEntrySettlement(ctx context.Context, userID, entryID uuid.UUID, costCurrency money.Currency, rates map[uuid.UUID]decimal.Decimal) (int, error) {
	var changed int

	if err := database.WithinTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		var owned uuid.UUID
		if err := tx.QueryRow(ctx, `
		SELECT pe.id
		FROM portfolio_entries pe
		JOIN portfolios p ON p.id = pe.portfolio_id
		WHERE pe.id = $1 AND p.user_id = $2
		FOR UPDATE OF pe
	`, entryID, userID).Scan(&owned); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrEntryNotFound
			}

			return err
		}

		txns, err := lockEntryTransactions(ctx, tx, entryID)
		if err != nil {
			return err
		}

		plan, err := PlanSettlement(txns, costCurrency, rates)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
		UPDATE portfolio_entries SET cost_currency = $2::char(3) WHERE id = $1
	`, entryID, costCurrency); err != nil {
			return err
		}

		for _, settled := range plan {
			if _, err := tx.Exec(ctx, `
			UPDATE transactions SET
				fx_rate       = $2::numeric,
				fees_currency = $3::char(3),
				updated_at    = NOW()
			WHERE id = $1
		`, settled.TransactionID, settled.FXRate.String(), settled.FeesCurrency); err != nil {
				return err
			}
		}

		changed = len(plan)

		return nil
	}); err != nil {
		return 0, err
	}

	return changed, nil
}

// lockEntryTransactions reads a position's transactions in the order they
// happened, locked for the rewrite that follows.
func lockEntryTransactions(ctx context.Context, tx pgx.Tx, entryID uuid.UUID) ([]Transaction, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, entry_id, type, quantity, price, currency, fx_rate, fees, fees_currency,
		       transaction_date, COALESCE(notes, '')
		FROM transactions
		WHERE entry_id = $1
		ORDER BY transaction_date, created_at
		FOR UPDATE
	`, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txns := make([]Transaction, 0)
	for rows.Next() {
		var txn Transaction
		if err := rows.Scan(
			&txn.ID,
			&txn.EntryID,
			&txn.Type,
			&txn.Quantity,
			&txn.Price,
			&txn.Currency,
			&txn.FXRate,
			&txn.Fees,
			&txn.FeesCurrency,
			&txn.TransactionDate,
			&txn.Notes,
		); err != nil {
			return nil, err
		}

		txn.Price.SetCurrency(txn.Currency)
		txn.Fees.SetCurrency(txn.FeesCurrency)
		txns = append(txns, txn)
	}

	return txns, rows.Err()
}
