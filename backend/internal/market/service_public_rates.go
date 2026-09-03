package market

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/yeferson59/gofinance/v2/decimal"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

// The application has no market-data key of its own, so until now a user
// without one saw their holdings at cost and could not switch the dashboard to
// pesos at all: nothing filled USD→COP. A public feed closes that gap without
// reopening the one 000018 shut. What may not be shared is data a user's key
// paid for; the TRM is Colombia's official rate, published to anyone who asks
// and fetched here with no credential, so one refresh serves every user and
// belongs in the shared table portfolio already reads.
//
// The user's own rate still wins where there is one — portfolio.storedRate
// consults user_exchange_rates first — so this is a floor, not a replacement.

// maxSharedRates bounds the unpaginated read below. It is a guard against a
// table that grew unexpectedly, not a page size.
const maxSharedRates = 100

// GetLatestExchangeRates returns the shared rates as they stand.
//
// Unpaginated, unlike the admin listing: the table holds one row per currency
// pair the app converts between, which is a handful, and the caller is a
// dashboard asking "what is a dollar worth today", not a browser of records.
func (s *service) GetLatestExchangeRates(ctx context.Context) ([]ExchangeRate, error) {
	return s.repo.GetExchangeRates(ctx, 0, maxSharedRates)
}

// ErrPublicRatesUnavailable means no public feed is wired, so there is nothing
// to refresh. Untagged: a deployment composed without the source is a
// misconfiguration, not a bad request.
var ErrPublicRatesUnavailable = errors.New("market: no public exchange rate source configured")

// RefreshPublicRates re-reads the public feed and stores what it published.
//
// One bad pair does not sink the refresh: every rate is validated and written
// on its own, and the failures are joined and returned after the loop, so a
// feed that adds a currency gofinance does not know still updates the ones it
// does. A pair that fails keeps its previous value, which is the right failure
// mode for a rate — stale beats absent.
func (s *service) RefreshPublicRates(ctx context.Context) ([]ExchangeRate, error) {
	if s.publicRates == nil {
		return nil, ErrPublicRatesUnavailable
	}

	// A partial fetch comes back as rates *and* an error: the source is several
	// feeds read as one (marketdata.PublicRateSources), and one of them being
	// down is no reason to drop the pairs the others published. Only an empty
	// result is a failed refresh.
	published, fetchErr := s.publicRates.FetchRates(ctx)
	if len(published) == 0 {
		return nil, fetchErr
	}

	stored := make([]ExchangeRate, 0, len(published))
	errs := []error{fetchErr}

	for _, pr := range published {
		rate, err := s.storePublicRate(ctx, pr)
		if err != nil {
			errs = append(errs, err)

			continue
		}

		stored = append(stored, rate)
	}

	return stored, errors.Join(errs...)
}

// storePublicRate validates one published rate against the same rules the admin
// endpoints apply — the ISO 4217 table money.CurrencyFromISOCode reads, and
// money.Convert's positive-rate rule — before it reaches the table. A feed is
// not more trusted than an operator: whatever gets in here is what every
// portfolio valuation then converts with.
func (s *service) storePublicRate(ctx context.Context, pr marketdata.PublicRate) (ExchangeRate, error) {
	rate, err := decimal.NewFromString(pr.Rate)
	if err != nil {
		return ExchangeRate{}, fmt.Errorf("public rate %s/%s: parse %q: %w", pr.From, pr.To, pr.Rate, err)
	}

	if err := validRate(rate); err != nil {
		return ExchangeRate{}, fmt.Errorf("public rate %s/%s: %w", pr.From, pr.To, err)
	}

	// AsOf dates the rate, so a feed that stopped updating stores the day it
	// last did rather than today. Only a source that published no timestamp
	// falls back to now.
	rateDate := pr.AsOf
	if rateDate.IsZero() {
		rateDate = time.Now()
	}

	return s.repo.UpsertPublicExchangeRate(ctx, pr.From, pr.To, rate, rateDate, pr.Source)
}
