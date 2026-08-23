package market

import (
	"context"
	"errors"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/gofinance/v2/decimal"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return new(PostgresRepository{db})
}

func (r *PostgresRepository) UpsertExchangeRate(ctx context.Context, from, to string, rate decimal.Decimal, rateDate time.Time) (ExchangeRate, error) {
	var er ExchangeRate
	var rateStr string

	err := r.db.QueryRow(ctx, `
		INSERT INTO exchange_rates (from_currency, to_currency, rate, rate_date, source, fetched_at)
		VALUES ($1, $2, $3::numeric, $4::date, $5, NOW())
		ON CONFLICT (from_currency, to_currency)
		DO UPDATE SET rate = EXCLUDED.rate, rate_date = EXCLUDED.rate_date, source = EXCLUDED.source, fetched_at = NOW()
		RETURNING id, from_currency, to_currency, rate::text, rate_date, source, fetched_at
	`, from, to, rate.String(), rateDate, ManualRateSource).Scan(
		&er.ID,
		&er.FromCurrency,
		&er.ToCurrency,
		&rateStr,
		&er.RateDate,
		&er.Source,
		&er.CreatedAt, // fetched_at mapped to CreatedAt; table has no separate created_at/updated_at
	)
	if err != nil {
		return ExchangeRate{}, err
	}

	er.Rate = decimal.MustFromString(rateStr)
	return er, nil
}

// UpsertPublicExchangeRate writes a rate a keyless feed published, stamping the
// row with the feed that produced it.
//
// It overwrites whatever the pair held, including a rate an operator typed.
// That is the intended precedence and the alternative is worse: a DO UPDATE
// that skipped manual rows would pin the pair to a hand-entered number
// permanently, and there is no endpoint to delete one and hand the pair back to
// the feed. An operator correcting a pair the feed covers is correcting it
// until the next refresh, and the source column is what makes that legible.
func (r *PostgresRepository) UpsertPublicExchangeRate(ctx context.Context, from, to string, rate decimal.Decimal, rateDate time.Time, source ProviderID) (ExchangeRate, error) {
	var er ExchangeRate
	var rateStr string

	err := r.db.QueryRow(ctx, `
		INSERT INTO exchange_rates (from_currency, to_currency, rate, rate_date, source, fetched_at)
		VALUES ($1, $2, $3::numeric, $4::date, $5, NOW())
		ON CONFLICT (from_currency, to_currency)
		DO UPDATE SET rate = EXCLUDED.rate, rate_date = EXCLUDED.rate_date, source = EXCLUDED.source, fetched_at = NOW()
		RETURNING id, from_currency, to_currency, rate::text, rate_date, source, fetched_at
	`, from, to, rate.String(), rateDate, string(source)).Scan(
		&er.ID,
		&er.FromCurrency,
		&er.ToCurrency,
		&rateStr,
		&er.RateDate,
		&er.Source,
		&er.CreatedAt,
	)
	if err != nil {
		return ExchangeRate{}, err
	}

	er.Rate = decimal.MustFromString(rateStr)
	return er, nil
}

func (r *PostgresRepository) GetExchangeRates(ctx context.Context, offset, limit uint) ([]ExchangeRate, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, from_currency, to_currency, rate::text, rate_date, source, fetched_at
		FROM exchange_rates
		ORDER BY from_currency, to_currency, rate_date DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return []ExchangeRate{}, err
	}
	defer rows.Close()

	rates := make([]ExchangeRate, 0)
	for rows.Next() {
		var er ExchangeRate
		var rateStr string

		if err := rows.Scan(&er.ID, &er.FromCurrency, &er.ToCurrency, &rateStr, &er.RateDate, &er.Source, &er.CreatedAt); err != nil {
			return nil, err
		}

		er.Rate = decimal.MustFromString(rateStr)
		rates = append(rates, er)
	}

	return rates, nil
}

// Reading a single shared rate used to live here, by pair and by id. Both went
// unused once portfolio grew its own reads: under BYO-key a valuation asks for
// the user's rate first (portfolio.GetUserExchangeRateByPair) and only then the
// shared one (portfolio.GetExchangeRateByPair). This module writes the shared
// table on behalf of an admin; it no longer reads a single row out of it.

func (r *PostgresRepository) UpdateExchangeRateByID(ctx context.Context, id uuid.UUID, rate decimal.Decimal) (ExchangeRate, error) {
	var er ExchangeRate
	var rateStr string

	// The edit also claims the row for the operator. Leaving source alone would
	// label a hand-typed number with the feed that no longer produced it, which
	// is the one thing the column exists to prevent.
	err := r.db.QueryRow(ctx, `
		UPDATE exchange_rates
		SET rate = $2::numeric, source = $3, fetched_at = NOW()
		WHERE id = $1
		RETURNING id, from_currency, to_currency, rate::text, rate_date, source, fetched_at
	`, id, rate.String(), ManualRateSource).Scan(&er.ID, &er.FromCurrency, &er.ToCurrency, &rateStr, &er.RateDate, &er.Source, &er.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ExchangeRate{}, ErrExchangeRateNotFound
		}
		return ExchangeRate{}, err
	}

	er.Rate = decimal.MustFromString(rateStr)
	return er, nil
}
