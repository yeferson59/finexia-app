package repositories

import (
	"context"
	"time"

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
