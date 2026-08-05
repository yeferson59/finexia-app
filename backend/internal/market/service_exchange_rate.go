package market

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"
)

type CurrencyPair struct{ From, To string }

// The exchange_rates table below is the shared, admin-maintained one. Rates
// fetched with a user's own key do not come here — they go to
// user_exchange_rates via SyncRatesForUser, because provider terms do not allow
// serving one user's data to another.

func (s *Service) GetExchangeRates(ctx context.Context, offset, limit uint) ([]ExchangeRate, error) {
	return s.repo.GetExchangeRates(ctx, offset, limit)
}

func (s *Service) CreateExchangeRate(ctx context.Context, from, to string, rate money.Decimal) (ExchangeRate, error) {
	fromCode, ok := NormalizeCurrencyCode(from)
	if !ok {
		return ExchangeRate{}, errExchangeRateCurrencyInvalid
	}
	toCode, ok := NormalizeCurrencyCode(to)
	if !ok {
		return ExchangeRate{}, errExchangeRateCurrencyInvalid
	}
	if err := validRate(rate); err != nil {
		return ExchangeRate{}, err
	}

	return s.repo.UpsertExchangeRate(ctx, fromCode, toCode, rate, time.Now())
}

func (s *Service) UpdateExchangeRate(ctx context.Context, id uuid.UUID, rate money.Decimal) (ExchangeRate, error) {
	if err := validRate(rate); err != nil {
		return ExchangeRate{}, err
	}

	return s.repo.UpdateExchangeRateByID(ctx, id, rate)
}

// validRate rejects a rate money.Convert would refuse. gofinance returns
// ErrInvalidExchangeRate for anything not strictly positive, and the spreadsheet
// importer already screens rows on the same rule; storing one through the admin
// endpoint only moved the failure to the portfolios that later convert with it.
func validRate(rate money.Decimal) error {
	if !rate.IsPos() {
		return errExchangeRateInvalid
	}

	return nil
}
