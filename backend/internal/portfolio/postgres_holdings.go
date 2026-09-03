package portfolio

import (
	"context"
	"errors"

	"uuid"

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

// GetRequiredCurrencyPairs returns every conversion the user's screens actually
// perform, which is what the BYO-key sync has to fetch with their key.
//
// There are three sources, and missing any of them leaves a conversion with no
// rate to use. Under the shared model that was invisible: the operator's key
// filled exchange_rates for everyone. Now nothing else fills them.
//
//  1. asset currency → portfolio base currency, for the market value of a
//     holding quoted in another currency (portfolio_summary).
//  2. cost currency → portfolio base currency, for the cost basis of a purchase
//     settled in another currency (portfolio_summary again — the two legs of
//     that view use different columns).
//  3. portfolio base currency → the user's preferred currency, for the display
//     conversion in GetPortfoliosSummaryInCurrency. Leaving this one out is what
//     made ?currency= fail outright once the shared table was emptied.
//
// Only one direction of each pair is asked for: GetConversionRate already
// inverts a stored rate and hops through USD.
func (r *PostgresRepository) GetRequiredCurrencyPairs(ctx context.Context, userID uuid.UUID) ([]CurrencyPair, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT from_currency, to_currency
		FROM (
			SELECT a.currency AS from_currency, p.base_currency AS to_currency
			FROM portfolio_entries pe
			JOIN portfolios p ON p.id = pe.portfolio_id
			JOIN assets a     ON a.id = pe.asset_id
			WHERE p.user_id = $1
		UNION
			SELECT pe.cost_currency, p.base_currency
			FROM portfolio_entries pe
			JOIN portfolios p ON p.id = pe.portfolio_id
			WHERE p.user_id = $1
		UNION
			SELECT p.base_currency, u.preferred_currency
			FROM portfolios p
			JOIN users u ON u.id = p.user_id
			WHERE p.user_id = $1
		) pairs
		WHERE from_currency IS NOT NULL
		  AND to_currency   IS NOT NULL
		  AND from_currency <> ''
		  AND to_currency   <> ''
		  AND from_currency <> to_currency
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
func (r *PostgresRepository) GetUserExchangeRateByPair(ctx context.Context, userID uuid.UUID, from, to money.Currency) (decimal.Decimal, error) {
	var rate decimal.Decimal

	err := r.db.QueryRow(ctx, `
		SELECT rate::text FROM user_exchange_rates
		WHERE user_id = $1 AND from_currency = $2 AND to_currency = $3
	`, userID, from, to).Scan(&rate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return decimal.Decimal{}, ErrExchangeRateNotFound
		}

		return decimal.Decimal{}, err
	}

	return rate, nil
}
