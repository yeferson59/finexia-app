package portfolio

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// GetHeldAssetIDs returns the distinct assets the user holds across all their
// portfolios, most recently traded first so a run cut short by an exhausted
// quota still refreshes what the user touched last.
func (r *PostgresRepository) GetHeldAssetIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `
		SELECT pe.asset_id
		FROM portfolio_entries pe
		JOIN portfolios p ON p.id = pe.portfolio_id
		WHERE p.user_id = $1
		GROUP BY pe.asset_id
		ORDER BY MAX(pe.updated_at) DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// GetRequiredCurrencyPairs returns the conversions the user's portfolios need:
// every distinct (asset currency → portfolio base currency) combination where
// the two differ.
func (r *PostgresRepository) GetRequiredCurrencyPairs(ctx context.Context, userID uuid.UUID) ([]CurrencyPair, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT a.currency, p.base_currency
		FROM portfolio_entries pe
		JOIN portfolios p ON p.id = pe.portfolio_id
		JOIN assets a     ON a.id = pe.asset_id
		WHERE p.user_id = $1 AND a.currency <> p.base_currency
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pairs := make([]CurrencyPair, 0)
	for rows.Next() {
		var pair CurrencyPair
		if err := rows.Scan(&pair.From, &pair.To); err != nil {
			return nil, err
		}
		pairs = append(pairs, pair)
	}

	return pairs, rows.Err()
}

// GetUserExchangeRateByPair reads a rate from the user's own cache — the rates
// their key fetched. Falls back to nothing: the shared exchange_rates table is
// read separately by GetExchangeRateByPair, and only holds admin-entered rows.
func (r *PostgresRepository) GetUserExchangeRateByPair(ctx context.Context, userID uuid.UUID, from, to string) (money.Decimal, error) {
	var rateStr string

	err := r.db.QueryRow(ctx, `
		SELECT rate::text FROM user_exchange_rates
		WHERE user_id = $1 AND from_currency = $2 AND to_currency = $3
	`, userID, from, to).Scan(&rateStr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return money.Decimal{}, ErrExchangeRateNotFound
		}

		return money.Decimal{}, err
	}

	return decimal.NewFromString(rateStr)
}
