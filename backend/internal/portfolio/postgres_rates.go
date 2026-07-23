package portfolio

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// GetExchangeRateByPair looks up the stored rate for one direction of a
// currency pair. Rates are synced one-directional (e.g. only EUR->USD is
// stored, never USD->EUR); callers needing the reverse direction should
// invert the returned rate themselves. Read-only: writing/syncing exchange
// rates is owned by the market module.
func (r *PostgresRepository) GetExchangeRateByPair(ctx context.Context, from, to string) (money.Decimal, error) {
	var rateStr string
	err := r.db.QueryRow(ctx, `
		SELECT rate::text
		FROM exchange_rates WHERE from_currency = $1 AND to_currency = $2
	`, from, to).Scan(&rateStr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return money.Decimal{}, ErrExchangeRateNotFound
		}
		return money.Decimal{}, err
	}

	return decimal.MustFromString(rateStr), nil
}
