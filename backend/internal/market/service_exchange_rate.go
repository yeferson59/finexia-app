package market

import (
	"context"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

type CurrencyPair struct{ From, To money.Currency }

// The exchange_rates table below is the shared, admin-maintained one. Rates
// fetched with a user's own key do not come here — they go to
// user_exchange_rates via SyncRatesForUser, because provider terms do not allow
// serving one user's data to another.

func (s *service) GetExchangeRates(ctx context.Context, offset, limit uint) ([]ExchangeRate, error) {
	return s.repo.GetExchangeRates(ctx, offset, limit)
}

func (s *service) CreateExchangeRate(ctx context.Context, from, to money.Currency, rate decimal.Decimal) (ExchangeRate, error) {
	if err := validRate(rate); err != nil {
		return ExchangeRate{}, err
	}

	return s.repo.UpsertExchangeRate(ctx, from, to, rate, time.Now())
}

func (s *service) UpdateExchangeRate(ctx context.Context, id uuid.UUID, rate decimal.Decimal) (ExchangeRate, error) {
	if err := validRate(rate); err != nil {
		return ExchangeRate{}, err
	}

	return s.repo.UpdateExchangeRateByID(ctx, id, rate)
}

// validRate rejects a rate money.Convert would refuse. gofinance returns
// ErrInvalidExchangeRate for anything not strictly positive, and the spreadsheet
// importer already screens rows on the same rule; storing one through the admin
// endpoint only moved the failure to the portfolios that later convert with it.
func validRate(rate decimal.Decimal) error {
	if !rate.IsPos() {
		return errExchangeRateInvalid
	}

	return nil
}
