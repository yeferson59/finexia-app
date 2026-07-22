package market

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"
)

type Repository interface {
	UpsertExchangeRate(ctx context.Context, from, to string, rate money.Decimal, rateDate time.Time) (ExchangeRate, error)
	GetExchangeRates(ctx context.Context, offset, limit uint) ([]ExchangeRate, error)
	GetExchangeRateByPair(ctx context.Context, from, to string) (ExchangeRate, error)
	GetExchangeRateByID(ctx context.Context, id uuid.UUID) (ExchangeRate, error)
	UpdateExchangeRateByID(ctx context.Context, id uuid.UUID, rate money.Decimal) (ExchangeRate, error)
}
