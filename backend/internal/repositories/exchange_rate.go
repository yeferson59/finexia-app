package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/money"

	"github.com/yeferson59/finexia-app/internal/entities"
)

func (r *Repository) UpsertExchangeRate(
	ctx context.Context,
	from, to string,
	rate money.Decimal,
	rateDate time.Time,
) (entities.ExchangeRate, error) {
	var er entities.ExchangeRate
	var rateStr string

	err := r.db.QueryRow(ctx, `
		INSERT INTO exchange_rates (from_currency, to_currency, rate, rate_date, fetched_at)
		VALUES ($1, $2, $3::numeric, $4::date, NOW())
		ON CONFLICT (from_currency, to_currency, rate_date)
		DO UPDATE SET rate = EXCLUDED.rate, fetched_at = NOW()
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
		return entities.ExchangeRate{}, err
	}

	er.Rate = money.MustFromString(rateStr)
	return er, nil
}

func (r *Repository) GetExchangeRates(ctx context.Context, offset, limit uint) ([]entities.ExchangeRate, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, from_currency, to_currency, rate::text, rate_date, fetched_at
		FROM exchange_rates
		ORDER BY from_currency, to_currency, rate_date DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return []entities.ExchangeRate{}, err
	}
	defer rows.Close()

	rates := make([]entities.ExchangeRate, 0)
	for rows.Next() {
		var er entities.ExchangeRate
		var rateStr string

		if err := rows.Scan(&er.ID, &er.FromCurrency, &er.ToCurrency, &rateStr, &er.RateDate, &er.CreatedAt); err != nil {
			return nil, err
		}

		er.Rate = money.MustFromString(rateStr)
		rates = append(rates, er)
	}

	return rates, nil
}

func (r *Repository) GetExchangeRateByID(ctx context.Context, id uuid.UUID) (entities.ExchangeRate, error) {
	var er entities.ExchangeRate
	var rateStr string

	err := r.db.QueryRow(ctx, `
		SELECT id, from_currency, to_currency, rate::text, rate_date, fetched_at
		FROM exchange_rates WHERE id = $1
	`, id).Scan(&er.ID, &er.FromCurrency, &er.ToCurrency, &rateStr, &er.RateDate, &er.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.ExchangeRate{}, errors.New("exchange rate not found")
		}
		return entities.ExchangeRate{}, err
	}

	er.Rate = money.MustFromString(rateStr)
	return er, nil
}

func (r *Repository) UpdateExchangeRateByID(ctx context.Context, id uuid.UUID, rate money.Decimal) (entities.ExchangeRate, error) {
	var er entities.ExchangeRate
	var rateStr string

	err := r.db.QueryRow(ctx, `
		UPDATE exchange_rates
		SET rate = $2::numeric, fetched_at = NOW()
		WHERE id = $1
		RETURNING id, from_currency, to_currency, rate::text, rate_date, fetched_at
	`, id, rate.String()).Scan(&er.ID, &er.FromCurrency, &er.ToCurrency, &rateStr, &er.RateDate, &er.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.ExchangeRate{}, errors.New("exchange rate not found")
		}
		return entities.ExchangeRate{}, err
	}

	er.Rate = money.MustFromString(rateStr)
	return er, nil
}
