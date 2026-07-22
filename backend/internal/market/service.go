package market

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/portfolio"
)

type defaultAsset struct {
	Ticker    string
	Name      string
	AssetType portfolio.AssetType
	Exchange  string
	Currency  string
}

const assetSyncCacheKey = "finexia:sync:asset_prices"
const assetSyncTTL = 24 * time.Hour

var defaultAssets = []defaultAsset{
	{"AAPL", "Apple Inc.", portfolio.Stock, "NASDAQ", "USD"},
	{"MSFT", "Microsoft Corporation", portfolio.Stock, "NASDAQ", "USD"},
	{"SPY", "SPDR S&P 500 ETF Trust", portfolio.ETF, "NYSEARCA", "USD"},
	{"BTC-USD", "Bitcoin", portfolio.Crypto, "Coinbase", "USD"},
	{"ETH-USD", "Ethereum", portfolio.Crypto, "Coinbase", "USD"},
	{"BND", "Vanguard Total Bond Market ETF", portfolio.Bond, "NASDAQ", "USD"},
}

type portfolioService interface {
	UpdateAssetPrice(ctx context.Context, assetID uuid.UUID, price money.Money) (portfolio.Asset, error)
	CreateAsset(ctx context.Context, ticker, name string, assetType portfolio.AssetType, exchange, currency string) (portfolio.Asset, error)
	GetAssets(ctx context.Context, offset, limit uint) ([]portfolio.Asset, error)
	GetAssetByID(ctx context.Context, assetID uuid.UUID) (portfolio.Asset, error)
	ImportAssetsFromFile(ctx context.Context, data []byte, filename, sheet string) (portfolio.ImportResultResponseDTO, error)
}

type Service struct {
	repo     Repository
	port     portfolioService
	storage  fiber.Storage
	provider marketdata.Provider
	log      logger.Logger
}

func NewService(repo Repository, port portfolioService, storage fiber.Storage, provider marketdata.Provider, log logger.Logger) *Service {
	return new(Service{
		repo:     repo,
		port:     port,
		storage:  storage,
		provider: provider,
		log:      log,
	})
}

func (s *Service) WasAssetPriceSyncedRecently() bool {
	v, err := s.storage.Get(assetSyncCacheKey)

	return err == nil && len(v) > 0
}

// fetchAndUpdatePrice fetches the current market price for asset and persists it.
// Returns (updated asset, skipped, error). skipped=true means the asset type
// is not supported by the provider — not an error, just not actionable.
func (s *Service) fetchAndUpdatePrice(ctx context.Context, asset portfolio.Asset, log logger.Logger) (portfolio.Asset, bool, error) {
	var priceStr string

	switch asset.AssetType {
	case portfolio.Stock, portfolio.ETF, portfolio.Bond:
		result, err := s.provider.FetchQuote(ctx, asset.Ticker)
		if err != nil {
			return portfolio.Asset{}, false, fmt.Errorf("fetch quote %q: %w", asset.Ticker, err)
		}
		priceStr = result.Price

	case portfolio.Crypto:
		parts := strings.SplitN(asset.Ticker, "-", 2)
		if len(parts) != 2 {
			return portfolio.Asset{}, false, fmt.Errorf("cannot parse crypto ticker %q", asset.Ticker)
		}
		result, err := s.provider.FetchExchangeRate(ctx, parts[0], parts[1])
		if err != nil {
			return portfolio.Asset{}, false, fmt.Errorf("fetch exchange rate %q: %w", asset.Ticker, err)
		}
		priceStr = result.Rate

	default:
		log.Info(ctx, "skipped — asset type not supported by Alpha Vantage", logger.Str("ticker", asset.Ticker), logger.Str("type", string(asset.AssetType)))
		return portfolio.Asset{}, true, nil
	}

	cur, err := money.CurrencyFromISOCode(asset.Currency)
	if err != nil {
		return portfolio.Asset{}, false, fmt.Errorf("unknown currency %q for %q: %w", asset.Currency, asset.Ticker, err)
	}
	price, err := money.NewMoneyFromString(priceStr, cur)
	if err != nil {
		return portfolio.Asset{}, false, fmt.Errorf("parse price %q for %q: %w", priceStr, asset.Ticker, err)
	}
	updated, err := s.port.UpdateAssetPrice(ctx, asset.ID, price)
	if err != nil {
		return portfolio.Asset{}, false, fmt.Errorf("persist price for %q: %w", asset.Ticker, err)
	}
	log.Info(ctx, "asset price updated", logger.Str("ticker", asset.Ticker), logger.Str("price", price.String()))
	return updated, false, nil
}

func (s *Service) SyncAssetPrices(ctx context.Context) ([]portfolio.Asset, []error) {
	log := s.log.With(logger.Str("job", "asset_price_sync"))

	var errs []error

	for _, da := range defaultAssets {
		if _, err := s.port.CreateAsset(ctx, da.Ticker, da.Name, da.AssetType, da.Exchange, da.Currency); err != nil {
			log.Error(ctx, "upsert default asset failed", logger.Err(err), logger.Str("ticker", da.Ticker))
			errs = append(errs, err)
		}
	}

	allAssets, err := s.port.GetAssets(ctx, 0, 1000)
	if err != nil {
		return nil, append(errs, fmt.Errorf("asset_price_sync: fetch assets: %w", err))
	}

	results := make([]portfolio.Asset, 0, len(allAssets))
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
func (s *Service) SyncAssetByID(ctx context.Context, assetID uuid.UUID) (portfolio.Asset, error) {
	log := s.log.With(logger.Str("job", "asset_price_sync_single"), logger.Str("assetID", assetID.String()))

	asset, err := s.port.GetAssetByID(ctx, assetID)
	if err != nil {
		return portfolio.Asset{}, err
	}

	updated, skipped, err := s.fetchAndUpdatePrice(ctx, asset, log)
	if err != nil {
		return portfolio.Asset{}, err
	}
	if skipped {
		return portfolio.Asset{}, fmt.Errorf("asset type %q is not supported by the price provider", asset.AssetType)
	}
	return updated, nil
}
