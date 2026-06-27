package services

import (
	"context"
	"time"

	"github.com/yeferson59/gofinance/money"

	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/logger"
)

const rateSyncCacheKey = "finexia:sync:exchange_rates"
const rateSyncTTL = 24 * time.Hour

func (s *Services) WasExchangeRateSyncedRecently() bool {
	v, err := s.storage.Get(rateSyncCacheKey)
	return err == nil && len(v) > 0
}

type CurrencyPair struct{ From, To string }

var defaultPairs = []CurrencyPair{
	{"EUR", "USD"},
	{"GBP", "USD"},
	{"USD", "COP"},
}

func (s *Services) SyncExchangeRates(ctx context.Context) ([]entities.ExchangeRate, []error) {
	log := s.log.With(logger.Str("job", "exchange_rate_sync"))

	results := make([]entities.ExchangeRate, 0, len(defaultPairs))
	var errs []error

	for i, pair := range defaultPairs {
		result, err := s.priceProvider.FetchExchangeRate(ctx, pair.From, pair.To)
		if err != nil {
			log.Error("fetch failed", logger.Err(err), logger.Str("pair", pair.From+"/"+pair.To))
			errs = append(errs, err)
			continue
		}

		rate, err := money.NewFromString(result.Rate)
		if err != nil {
			log.Error("parse rate failed", logger.Err(err), logger.Str("pair", pair.From+"/"+pair.To), logger.Str("raw", result.Rate))
			errs = append(errs, err)
			continue
		}

		er, err := s.repos.UpsertExchangeRate(ctx, pair.From, pair.To, rate, result.FetchedAt)
		if err != nil {
			log.Error("upsert failed", logger.Err(err), logger.Str("pair", pair.From+"/"+pair.To))
			errs = append(errs, err)
			continue
		}

		log.Info("rate upserted", logger.Str("pair", pair.From+"/"+pair.To), logger.Str("rate", er.Rate.String()))
		results = append(results, er)

		// Alpha Vantage free tier allows 5 req/min; sleep between pairs to avoid hitting the limit
		if i < len(defaultPairs)-1 {
			select {
			case <-ctx.Done():
				return results, errs
			case <-time.After(13 * time.Second):
			}
		}
	}

	_ = s.storage.Set(rateSyncCacheKey, []byte(time.Now().UTC().Format(time.RFC3339)), rateSyncTTL)
	return results, errs
}
