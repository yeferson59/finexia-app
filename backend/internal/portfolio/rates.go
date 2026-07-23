package portfolio

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// SupportedDisplayCurrencies lists the currencies a user can pick to view
// their portfolio totals in. Kept intentionally small for now; extend this
// list (and, if needed, the legacy sync's defaultPairs so rates stay fresh)
// to support more.
var SupportedDisplayCurrencies = []string{"USD", "COP"}

func IsSupportedDisplayCurrency(currency string) bool {
	return slices.Contains(SupportedDisplayCurrencies, currency)
}

// ErrExchangeRateUnavailable means no stored rate (direct, inverse, or via a
// USD hop) connects the requested currency pair. Tagged as NotFound so it maps
// to 404 by type rather than by the "not found" substring (docs/TECH_DEBT.md #1).
var ErrExchangeRateUnavailable = httpx.AsNotFound(errors.New("exchange rate not found for currency pair"))

// GetConversionRate returns the multiplier that turns an amount in `from`
// into an amount in `to` (amountInFrom * rate = amountInTo). It tries a
// direct pair, then its inverse, then a two-hop conversion through USD
// (every synced pair involves USD, see the legacy sync's defaultPairs),
// since rates are only stored one-directional.
func (s *Service) GetConversionRate(ctx context.Context, from, to string) (money.Decimal, error) {
	from = strings.ToUpper(strings.TrimSpace(from))
	to = strings.ToUpper(strings.TrimSpace(to))

	if from == to {
		return decimal.One, nil
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
func (s *Service) pairRate(ctx context.Context, from, to string) (money.Decimal, error) {
	if from == to {
		return decimal.One, nil
	}

	if rate, err := s.repo.GetExchangeRateByPair(ctx, from, to); err == nil {
		return rate, nil
	}

	rate, err := s.repo.GetExchangeRateByPair(ctx, to, from)
	if err != nil {
		return money.Decimal{}, ErrExchangeRateUnavailable
	}
	if rate.IsZero() {
		return money.Decimal{}, ErrExchangeRateUnavailable
	}

	return decimal.One.Div(rate)
}
