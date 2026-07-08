package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/money"

	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/logger"
)

const assetSyncCacheKey = "finexia:sync:asset_prices"
const assetSyncTTL = 24 * time.Hour

func (s *Services) WasAssetPriceSyncedRecently() bool {
	v, err := s.storage.Get(assetSyncCacheKey)
	return err == nil && len(v) > 0
}

type defaultAsset struct {
	Ticker    string
	Name      string
	AssetType entities.AssetType
	Exchange  string
	Currency  string
}

var defaultAssets = []defaultAsset{
	{"AAPL", "Apple Inc.", entities.Stock, "NASDAQ", "USD"},
	{"MSFT", "Microsoft Corporation", entities.Stock, "NASDAQ", "USD"},
	{"SPY", "SPDR S&P 500 ETF Trust", entities.ETF, "NYSEARCA", "USD"},
	{"BTC-USD", "Bitcoin", entities.Crypto, "Coinbase", "USD"},
	{"ETH-USD", "Ethereum", entities.Crypto, "Coinbase", "USD"},
	{"BND", "Vanguard Total Bond Market ETF", entities.Bond, "NASDAQ", "USD"},
}

// fetchAndUpdatePrice fetches the current market price for asset and persists it.
// Returns (updated asset, skipped, error). skipped=true means the asset type
// is not supported by the provider — not an error, just not actionable.
func (s *Services) fetchAndUpdatePrice(ctx context.Context, asset entities.Asset, log logger.Logger) (entities.Asset, bool, error) {
	var priceStr string

	switch asset.AssetType {
	case entities.Stock, entities.ETF, entities.Bond:
		result, err := s.priceProvider.FetchQuote(ctx, asset.Ticker)
		if err != nil {
			return entities.Asset{}, false, fmt.Errorf("fetch quote %q: %w", asset.Ticker, err)
		}
		priceStr = result.Price

	case entities.Crypto:
		parts := strings.SplitN(asset.Ticker, "-", 2)
		if len(parts) != 2 {
			return entities.Asset{}, false, fmt.Errorf("cannot parse crypto ticker %q", asset.Ticker)
		}
		result, err := s.priceProvider.FetchExchangeRate(ctx, parts[0], parts[1])
		if err != nil {
			return entities.Asset{}, false, fmt.Errorf("fetch exchange rate %q: %w", asset.Ticker, err)
		}
		priceStr = result.Rate

	default:
		log.Info(ctx, "skipped — asset type not supported by Alpha Vantage", logger.Str("ticker", asset.Ticker), logger.Str("type", string(asset.AssetType)))
		return entities.Asset{}, true, nil
	}

	cur, err := money.CurrencyFromISOCode(asset.Currency)
	if err != nil {
		return entities.Asset{}, false, fmt.Errorf("unknown currency %q for %q: %w", asset.Currency, asset.Ticker, err)
	}
	price, err := money.NewMoneyFromString(priceStr, cur)
	if err != nil {
		return entities.Asset{}, false, fmt.Errorf("parse price %q for %q: %w", priceStr, asset.Ticker, err)
	}
	updated, err := s.repos.UpdateAssetPrice(ctx, asset.ID, price)
	if err != nil {
		return entities.Asset{}, false, fmt.Errorf("persist price for %q: %w", asset.Ticker, err)
	}
	log.Info(ctx, "asset price updated", logger.Str("ticker", asset.Ticker), logger.Str("price", price.String()))
	return updated, false, nil
}

func (s *Services) SyncAssetPrices(ctx context.Context) ([]entities.Asset, []error) {
	log := s.log.With(logger.Str("job", "asset_price_sync"))

	var errs []error

	for _, da := range defaultAssets {
		if _, err := s.repos.UpsertAsset(ctx, da.Ticker, da.Name, da.AssetType, da.Exchange, da.Currency); err != nil {
			log.Error(ctx, "upsert default asset failed", logger.Err(err), logger.Str("ticker", da.Ticker))
			errs = append(errs, err)
		}
	}

	allAssets, err := s.repos.GetAssets(ctx, 0, 1000)
	if err != nil {
		return nil, append(errs, fmt.Errorf("asset_price_sync: fetch assets: %w", err))
	}

	results := make([]entities.Asset, 0, len(allAssets))
	apiCallMade := false

	for _, asset := range allAssets {
		if apiCallMade {
			select {
			case <-ctx.Done():
				return results, errs
			case <-time.After(13 * time.Second):
			}
		}

		updated, skipped, err := s.fetchAndUpdatePrice(ctx, asset, log)
		if err != nil {
			log.Error(ctx, "sync asset failed", logger.Err(err), logger.Str("ticker", asset.Ticker))
			errs = append(errs, err)
			apiCallMade = true
			continue
		}
		if skipped {
			continue
		}

		apiCallMade = true
		results = append(results, updated)
	}

	_ = s.storage.Set(assetSyncCacheKey, []byte(time.Now().UTC().Format(time.RFC3339)), assetSyncTTL)
	return results, errs
}

// SyncAssetByID fetches and updates the price for a single asset by ID.
func (s *Services) SyncAssetByID(ctx context.Context, assetID uuid.UUID) (entities.Asset, error) {
	log := s.log.With(logger.Str("job", "asset_price_sync_single"), logger.Str("assetID", assetID.String()))

	asset, err := s.repos.GetAssetByID(ctx, assetID)
	if err != nil {
		return entities.Asset{}, err
	}

	updated, skipped, err := s.fetchAndUpdatePrice(ctx, asset, log)
	if err != nil {
		return entities.Asset{}, err
	}
	if skipped {
		return entities.Asset{}, fmt.Errorf("asset type %q is not supported by the price provider", asset.AssetType)
	}
	return updated, nil
}
