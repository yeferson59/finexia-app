package portfolio

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// SupportedDisplayCurrencies lists the currencies a user can pick to view
// their portfolio totals in. Kept intentionally small for now; extending it is
// enough on its own — the sync derives the pairs it fetches from what each user
// actually holds and prefers (GetRequiredCurrencyPairs), not from a fixed list.
var SupportedDisplayCurrencies = []string{"USD", "COP"}

func IsSupportedDisplayCurrency(currency string) bool {
	return slices.Contains(SupportedDisplayCurrencies, currency)
}

// ErrExchangeRateUnavailable means no stored rate (direct, inverse, or via a
// USD hop) connects the requested currency pair. Tagged as NotFound so it maps
// to 404 by type, whatever a caller wraps around it.
var ErrExchangeRateUnavailable = httpx.AsNotFound(errors.New("exchange rate not found for currency pair"))

// GetConversionRate returns the multiplier that turns an amount in `from`
// into an amount in `to` (amountInFrom * rate = amountInTo). It tries a
// direct pair, then its inverse, then a two-hop conversion through USD,
// since rates are only stored one-directional.
//
// It takes a userID because rates are BYO-key data: the rate this user's own
// key fetched is consulted first, and only then the shared table, which now
// holds admin-entered rows alone. Serving another user's fetched rate would be
// the same redistribution problem as serving their prices.
func (s *Service) GetConversionRate(ctx context.Context, userID uuid.UUID, from, to string) (money.Decimal, error) {
	from = strings.ToUpper(strings.TrimSpace(from))
	to = strings.ToUpper(strings.TrimSpace(to))

	if from == to {
		return decimal.One, nil
	}

	if rate, err := s.pairRate(ctx, userID, from, to); err == nil {
		return rate, nil
	}

	fromToUSD, err := s.pairRate(ctx, userID, from, "USD")
	if err != nil {
		return money.Decimal{}, ErrExchangeRateUnavailable
	}
	usdToTarget, err := s.pairRate(ctx, userID, "USD", to)
	if err != nil {
		return money.Decimal{}, ErrExchangeRateUnavailable
	}

	return fromToUSD.Mul(usdToTarget), nil
}

// usableRate reports whether a stored rate is one money.Convert would accept.
// gofinance treats a zero or negative rate as invalid (ErrInvalidExchangeRate),
// so screening it here lets a corrupt row fall through to the other directions
// below and, failing those, surface as the same 404 as a missing rate.
func usableRate(rate money.Decimal) bool {
	return rate.IsPos()
}

// pairRate resolves a single pair directly, falling back to inverting the
// opposite direction if that's what was fetched.
//
// Each direction is looked up in the user's own cache first, then in the shared
// table, so a user with their own rate never falls back to a stale shared one.
func (s *Service) pairRate(ctx context.Context, userID uuid.UUID, from, to string) (money.Decimal, error) {
	if from == to {
		return decimal.One, nil
	}

	if rate, err := s.storedRate(ctx, userID, from, to); err == nil && usableRate(rate) {
		return rate, nil
	}

	rate, err := s.storedRate(ctx, userID, to, from)
	if err != nil || !usableRate(rate) {
		return money.Decimal{}, ErrExchangeRateUnavailable
	}

	return decimal.One.Div(rate)
}

// storedRate reads one direction, preferring the user's own data.
func (s *Service) storedRate(ctx context.Context, userID uuid.UUID, from, to string) (money.Decimal, error) {
	if rate, err := s.repo.GetUserExchangeRateByPair(ctx, userID, from, to); err == nil {
		return rate, nil
	}

	return s.repo.GetExchangeRateByPair(ctx, from, to)
}
