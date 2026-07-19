package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// SupportedDisplayCurrencies lists the currencies a user can pick to view
// their portfolio totals in. Kept intentionally small for now; extend this
// list (and, if needed, defaultPairs so rates stay fresh) to support more.
var SupportedDisplayCurrencies = []string{"USD", "COP"}

func IsSupportedDisplayCurrency(currency string) bool {
	for _, c := range SupportedDisplayCurrencies {
		if c == currency {
			return true
		}
	}
	return false
}

// ErrExchangeRateUnavailable means no stored rate (direct, inverse, or via a
// USD hop) connects the requested currency pair.
var ErrExchangeRateUnavailable = errors.New("exchange rate not found for currency pair")

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
			log.Error(ctx, "fetch failed", logger.Err(err), logger.Str("pair", pair.From+"/"+pair.To))
			errs = append(errs, err)
			continue
		}

		rate, err := money.NewFromString(result.Rate)
		if err != nil {
			log.Error(ctx, "parse rate failed", logger.Err(err), logger.Str("pair", pair.From+"/"+pair.To), logger.Str("raw", result.Rate))
			errs = append(errs, err)
			continue
		}

		er, err := s.repos.UpsertExchangeRate(ctx, pair.From, pair.To, rate, result.FetchedAt)
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

	log.Info(ctx, "rate synced", logger.Str("pair", existing.FromCurrency+"/"+existing.ToCurrency), logger.Str("rate", updated.Rate.String()))
	return updated, nil
}

// GetConversionRate returns the multiplier that turns an amount in `from`
// into an amount in `to` (amountInFrom * rate = amountInTo). It tries a
// direct pair, then its inverse, then a two-hop conversion through USD
// (every synced pair involves USD, see defaultPairs), since rates are only
// stored one-directional.
func (s *Services) GetConversionRate(ctx context.Context, from, to string) (money.Decimal, error) {
	from = strings.ToUpper(strings.TrimSpace(from))
	to = strings.ToUpper(strings.TrimSpace(to))

	if from == to {
		return money.One, nil
	}

	if rate, err := s.pairRate(ctx, from, to); err == nil {
		return rate, nil
	}

	fromToUSD, err := s.pairRate(ctx, from, "USD")
	if err != nil {
		return money.Decimal{}, ErrExchangeRateUnavailable
	}
	usdToTarget, err := s.pairRate(ctx, "USD", to)
	if err != nil {
		return money.Decimal{}, ErrExchangeRateUnavailable
	}

	return fromToUSD.Mul(usdToTarget), nil
}

// pairRate resolves a single pair directly, falling back to inverting the
// opposite direction if that's what was synced.
func (s *Services) pairRate(ctx context.Context, from, to string) (money.Decimal, error) {
	if from == to {
		return money.One, nil
	}

	if er, err := s.repos.GetExchangeRateByPair(ctx, from, to); err == nil {
		return er.Rate, nil
	}

	er, err := s.repos.GetExchangeRateByPair(ctx, to, from)
	if err != nil {
		return money.Decimal{}, ErrExchangeRateUnavailable
	}
	if er.Rate.IsZero() {
		return money.Decimal{}, ErrExchangeRateUnavailable
	}

	return money.One.Div(er.Rate)
}
