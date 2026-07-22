package market

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

const rateSyncCacheKey = "finexia:sync:exchange_rates"
const rateSyncTTL = 24 * time.Hour

func (s *Service) WasExchangeRateSyncedRecently() bool {
	v, err := s.storage.Get(rateSyncCacheKey)
	return err == nil && len(v) > 0
}

type CurrencyPair struct{ From, To string }

var defaultPairs = []CurrencyPair{
	{"EUR", "USD"},
	{"GBP", "USD"},
	{"USD", "COP"},
}

func (s *Service) SyncExchangeRates(ctx context.Context) ([]ExchangeRate, []error) {
	log := s.log.With(logger.Str("job", "exchange_rate_sync"))

	results := make([]ExchangeRate, 0, len(defaultPairs))
	var errs []error

	for i, pair := range defaultPairs {
		result, err := s.provider.FetchExchangeRate(ctx, pair.From, pair.To)
		if err != nil {
			log.Error(ctx, "fetch failed", logger.Err(err), logger.Str("pair", pair.From+"/"+pair.To))
			errs = append(errs, err)
			continue
		}

		rate, err := decimal.NewFromString(result.Rate)
		if err != nil {
			log.Error(ctx, "parse rate failed", logger.Err(err), logger.Str("pair", pair.From+"/"+pair.To), logger.Str("raw", result.Rate))
			errs = append(errs, err)
			continue
		}

		er, err := s.repo.UpsertExchangeRate(ctx, pair.From, pair.To, rate, result.FetchedAt)
		if err != nil {
			log.Error(ctx, "upsert failed", logger.Err(err), logger.Str("pair", pair.From+"/"+pair.To))
			errs = append(errs, err)
			continue
		}

		log.Info(ctx, "rate upserted", logger.Str("pair", pair.From+"/"+pair.To), logger.Str("rate", er.Rate.String()))
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

func (s *Service) GetExchangeRates(ctx context.Context, offset, limit uint) ([]ExchangeRate, error) {
	return s.repo.GetExchangeRates(ctx, offset, limit)
}

func (s *Service) CreateExchangeRate(ctx context.Context, from, to string, rate money.Decimal) (ExchangeRate, error) {
	return s.repo.UpsertExchangeRate(ctx, from, to, rate, time.Now())
}

func (s *Service) UpdateExchangeRate(ctx context.Context, id uuid.UUID, rate money.Decimal) (ExchangeRate, error) {
	return s.repo.UpdateExchangeRateByID(ctx, id, rate)
}

// SyncExchangeRateByID fetches and updates the rate for a single currency pair by ID.
func (s *Service) SyncExchangeRateByID(ctx context.Context, id uuid.UUID) (ExchangeRate, error) {
	log := s.log.With(logger.Str("job", "exchange_rate_sync_single"), logger.Str("id", id.String()))

	existing, err := s.repo.GetExchangeRateByID(ctx, id)
	if err != nil {
		return ExchangeRate{}, err
	}

	result, err := s.provider.FetchExchangeRate(ctx, existing.FromCurrency, existing.ToCurrency)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("fetch exchange rate %q/%q: %w", existing.FromCurrency, existing.ToCurrency, err)
	}

	rate, err := decimal.NewFromString(result.Rate)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("parse rate %q for %q/%q: %w", result.Rate, existing.FromCurrency, existing.ToCurrency, err)
	}

	updated, err := s.repo.UpsertExchangeRate(ctx, existing.FromCurrency, existing.ToCurrency, rate, result.FetchedAt)
	if err != nil {
		return ExchangeRate{}, err
	}

	log.Info(ctx, "rate synced", logger.Str("pair", existing.FromCurrency+"/"+existing.ToCurrency), logger.Str("rate", updated.Rate.String()))
	return updated, nil
}

// GetConversionRate and the display-currency helpers migrated to the
// portfolio module (Fase 6): converting summary totals is a portfolio
// concern and it was their only consumer.
