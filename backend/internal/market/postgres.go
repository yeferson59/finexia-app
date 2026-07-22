package market

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return new(PostgresRepository{db})
}

func (r *PostgresRepository) UpsertExchangeRate(ctx context.Context, from, to string, rate money.Decimal, rateDate time.Time) (ExchangeRate, error) {
	var er ExchangeRate
	var rateStr string

	err := r.db.QueryRow(ctx, `
		INSERT INTO exchange_rates (from_currency, to_currency, rate, rate_date, fetched_at)
		VALUES ($1, $2, $3::numeric, $4::date, NOW())
		ON CONFLICT (from_currency, to_currency)
		DO UPDATE SET rate = EXCLUDED.rate, rate_date = EXCLUDED.rate_date, fetched_at = NOW()
		RETURNING id, from_currency, to_currency, rate::text, rate_date, fetched_at
	`, from, to, rate.String(), rateDate).Scan(
		&er.ID,
		&er.FromCurrency,
		&er.ToCurrency,
		&rateStr,
		&er.RateDate,
		&er.CreatedAt, // fetched_at mapped to CreatedAt; table has no separate created_at/updated_at
	)
	if err != nil {
		return ExchangeRate{}, err
	}

	er.Rate = decimal.MustFromString(rateStr)
	return er, nil
}

func (r *PostgresRepository) GetExchangeRates(ctx context.Context, offset, limit uint) ([]ExchangeRate, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, from_currency, to_currency, rate::text, rate_date, fetched_at
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

		if err := rows.Scan(&er.ID, &er.FromCurrency, &er.ToCurrency, &rateStr, &er.RateDate, &er.CreatedAt); err != nil {
			return nil, err
		}

		er.Rate = decimal.MustFromString(rateStr)
		rates = append(rates, er)
	}

	return rates, nil
}

// GetExchangeRateByPair looks up the stored rate for one direction of a
// currency pair. Rates are synced one-directional (e.g. only EUR->USD is
// stored, never USD->EUR); callers needing the reverse direction should
// invert the returned rate themselves.
func (r *PostgresRepository) GetExchangeRateByPair(ctx context.Context, from, to string) (ExchangeRate, error) {
	var er ExchangeRate
	var rateStr string

	err := r.db.QueryRow(ctx, `
		SELECT id, from_currency, to_currency, rate::text, rate_date, fetched_at
		FROM exchange_rates WHERE from_currency = $1 AND to_currency = $2
	`, from, to).Scan(&er.ID, &er.FromCurrency, &er.ToCurrency, &rateStr, &er.RateDate, &er.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ExchangeRate{}, errors.New("exchange rate not found")
		}
		return ExchangeRate{}, err
	}

	er.Rate = decimal.MustFromString(rateStr)
	return er, nil
}

func (r *PostgresRepository) GetExchangeRateByID(ctx context.Context, id uuid.UUID) (ExchangeRate, error) {
	var er ExchangeRate
	var rateStr string

	err := r.db.QueryRow(ctx, `
		SELECT id, from_currency, to_currency, rate::text, rate_date, fetched_at
		FROM exchange_rates WHERE id = $1
	`, id).Scan(&er.ID, &er.FromCurrency, &er.ToCurrency, &rateStr, &er.RateDate, &er.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ExchangeRate{}, errors.New("exchange rate not found")
		}
		return ExchangeRate{}, err
	}

	er.Rate = decimal.MustFromString(rateStr)
	return er, nil
}

func (r *PostgresRepository) UpdateExchangeRateByID(ctx context.Context, id uuid.UUID, rate money.Decimal) (ExchangeRate, error) {
	var er ExchangeRate
	var rateStr string

	err := r.db.QueryRow(ctx, `
		UPDATE exchange_rates
		SET rate = $2::numeric, fetched_at = NOW()
		WHERE id = $1
		RETURNING id, from_currency, to_currency, rate::text, rate_date, fetched_at
	`, id, rate.String()).Scan(&er.ID, &er.FromCurrency, &er.ToCurrency, &rateStr, &er.RateDate, &er.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ExchangeRate{}, errors.New("exchange rate not found")
		}
		return ExchangeRate{}, err
	}

	er.Rate = decimal.MustFromString(rateStr)
	return er, nil
}
