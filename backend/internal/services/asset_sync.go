package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yeferson59/gofinance/money"

	"github.com/yeferson59/finexia-app/internal/alphavantage"
	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/logger"
)

func (s *Services) SyncAssetPrices(ctx context.Context) ([]entities.Asset, []error) {
	log := s.log.With(logger.Str("job", "asset_price_sync"))
	client := alphavantage.New(s.cfg.AlphaVantageAPIKey)

	allAssets, err := s.repos.GetAssets(ctx, 0, 1000)
	if err != nil {
		return nil, []error{fmt.Errorf("asset_price_sync: fetch assets: %w", err)}
	}

	results := make([]entities.Asset, 0, len(allAssets))
	var errs []error
	apiCallMade := false

	for _, asset := range allAssets {
		var priceStr string

		switch asset.AssetType {
		case entities.Stock, entities.ETF, entities.Bond:
			if apiCallMade {
				select {
				case <-ctx.Done():
					return results, errs
				case <-time.After(13 * time.Second):
				}
			}

			result, err := client.FetchGlobalQuote(ctx, asset.Ticker)
			apiCallMade = true
			if err != nil {
				log.Error("fetch global quote failed", logger.Err(err), logger.Str("ticker", asset.Ticker))
				errs = append(errs, err)
				continue
			}
			priceStr = result.Price

		case entities.Crypto:
			parts := strings.SplitN(asset.Ticker, "-", 2)
			if len(parts) != 2 {
				err := fmt.Errorf("asset_price_sync: cannot parse crypto ticker %q", asset.Ticker)
				log.Error("invalid crypto ticker format", logger.Err(err))
				errs = append(errs, err)
				continue
			}

			if apiCallMade {
				select {
				case <-ctx.Done():
					return results, errs
				case <-time.After(13 * time.Second):
				}
			}

			result, err := client.FetchExchangeRate(ctx, parts[0], parts[1])
			apiCallMade = true
			if err != nil {
				log.Error("fetch exchange rate failed", logger.Err(err), logger.Str("ticker", asset.Ticker))
				errs = append(errs, err)
				continue
			}
			priceStr = result.Rate

		default:
			log.Info("skipped — asset type not supported by Alpha Vantage", logger.Str("ticker", asset.Ticker), logger.Str("type", string(asset.AssetType)))
			continue
		}

		cur, err := money.CurrencyFromISOCode(asset.Currency)
		if err != nil {
			log.Error("unknown currency code", logger.Err(err), logger.Str("ticker", asset.Ticker), logger.Str("currency", asset.Currency))
			errs = append(errs, err)
			continue
		}

		price, err := money.NewMoneyFromString(priceStr, cur)
		if err != nil {
			log.Error("parse price failed", logger.Err(err), logger.Str("ticker", asset.Ticker), logger.Str("raw", priceStr))
			errs = append(errs, err)
			continue
		}

		updated, err := s.repos.UpdateAssetPrice(ctx, asset.ID, price)
		if err != nil {
			log.Error("update asset price failed", logger.Err(err), logger.Str("ticker", asset.Ticker))
			errs = append(errs, err)
			continue
		}

		log.Info("asset price updated", logger.Str("ticker", asset.Ticker), logger.Str("price", price.String()))
		results = append(results, updated)
	}

	return results, errs
}
