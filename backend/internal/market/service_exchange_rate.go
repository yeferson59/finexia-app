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
	return s.repo.UpsertExchangeRate(ctx, from, to, rate, time.Now())
}

func (s *Service) UpdateExchangeRate(ctx context.Context, id uuid.UUID, rate money.Decimal) (ExchangeRate, error) {
	return s.repo.UpdateExchangeRateByID(ctx, id, rate)
}
