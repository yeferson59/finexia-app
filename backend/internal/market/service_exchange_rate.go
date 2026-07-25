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

type CurrencyPair struct{ From, To string }

var defaultPairs = []CurrencyPair{
	{"EUR", "USD"},
	{"GBP", "USD"},
	{"USD", "COP"},
}

func (s *Service) SyncExchangeRates(ctx context.Context) ([]ExchangeRate, []error) {
	log := s.log.With(logger.Str("job", "exchange_rate"))

	size := len(defaultPairs)
	results := make([]ExchangeRate, 0, size)
	errs := make([]error, 0, 3*size)

	for i, pair := range defaultPairs {
		result, err := s.provider.FetchExchangeRate(ctx, pair.From, pair.To)
		if err != nil {
			log.Error(ctx, "fetch failed", logger.Str("from", pair.From), logger.Str("to", pair.To), logger.Err(err))

			errs = append(errs, err)

			continue
		}

		rate, err := decimal.NewFromString(result.Rate)
		if err != nil {
			log.Error(ctx, "parse rate failed", logger.Str("from", pair.From), logger.Str("to", pair.To), logger.Err(err), logger.Str("raw", result.Rate))

			errs = append(errs, err)

			continue
		}

		er, err := s.repo.UpsertExchangeRate(ctx, pair.From, pair.To, rate, result.FetchedAt)
		if err != nil {
			log.Error(ctx, "upsert failed", logger.Str("from", pair.From), logger.Str("to", pair.To), logger.Err(err))

			errs = append(errs, err)

			continue
		}

		log.Info(ctx, "rate upserted", logger.Str("from", pair.From), logger.Str("to", pair.To), logger.Str("rate", er.Rate.String()))

		results = append(results, er)

		if i < len(defaultPairs)-1 {
			select {
			case <-ctx.Done():
				return results, errs
			case <-time.After(13 * time.Second):
			}
		}
	}

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
	log := s.log.With(logger.Str("job", "exchange_rate_ingle"), logger.Str("id", id.String()))

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
