package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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

func (s *Services) GetExchangeRates(ctx context.Context, offset, limit uint) ([]entities.ExchangeRate, error) {
	return s.repos.GetExchangeRates(ctx, offset, limit)
}

func (s *Services) CreateExchangeRate(ctx context.Context, from, to string, rate money.Decimal) (entities.ExchangeRate, error) {
	return s.repos.UpsertExchangeRate(ctx, from, to, rate, time.Now())
}

func (s *Services) UpdateExchangeRate(ctx context.Context, id uuid.UUID, rate money.Decimal) (entities.ExchangeRate, error) {
	return s.repos.UpdateExchangeRateByID(ctx, id, rate)
}

// SyncExchangeRateByID fetches and updates the rate for a single currency pair by ID.
func (s *Services) SyncExchangeRateByID(ctx context.Context, id uuid.UUID) (entities.ExchangeRate, error) {
	log := s.log.With(logger.Str("job", "exchange_rate_sync_single"), logger.Str("id", id.String()))

	existing, err := s.repos.GetExchangeRateByID(ctx, id)
	if err != nil {
		return entities.ExchangeRate{}, err
	}

	result, err := s.priceProvider.FetchExchangeRate(ctx, existing.FromCurrency, existing.ToCurrency)
	if err != nil {
		return entities.ExchangeRate{}, fmt.Errorf("fetch exchange rate %q/%q: %w", existing.FromCurrency, existing.ToCurrency, err)
	}

	rate, err := money.NewFromString(result.Rate)
	if err != nil {
		return entities.ExchangeRate{}, fmt.Errorf("parse rate %q for %q/%q: %w", result.Rate, existing.FromCurrency, existing.ToCurrency, err)
	}

	updated, err := s.repos.UpsertExchangeRate(ctx, existing.FromCurrency, existing.ToCurrency, rate, result.FetchedAt)
	if err != nil {
		return entities.ExchangeRate{}, err
	}

	log.Info("rate synced", logger.Str("pair", existing.FromCurrency+"/"+existing.ToCurrency), logger.Str("rate", updated.Rate.String()))
	return updated, nil
}
