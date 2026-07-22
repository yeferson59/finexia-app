package market

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"
)

type Repository interface {
	// Exchange rates
	UpsertExchangeRate(ctx context.Context, from, to string, rate money.Decimal, rateDate time.Time) (ExchangeRate, error)
	GetExchangeRates(ctx context.Context, offset, limit uint) ([]ExchangeRate, error)
	GetExchangeRateByPair(ctx context.Context, from, to string) (ExchangeRate, error)
	GetExchangeRateByID(ctx context.Context, id uuid.UUID) (ExchangeRate, error)
	UpdateExchangeRateByID(ctx context.Context, id uuid.UUID, rate money.Decimal) (ExchangeRate, error)

	// Assets (catalog owned by this module; portfolio reads them via AssetReader)
	GetAssetByID(ctx context.Context, assetID uuid.UUID) (Asset, error)
	GetAssets(ctx context.Context, offset, limit uint) ([]Asset, error)
	SearchAssets(ctx context.Context, search string, offset, limit uint) ([]Asset, error)
	UpsertAsset(ctx context.Context, ticker, name string, assetType AssetType, exchange, currency string) (Asset, error)
	UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (Asset, error)
}
