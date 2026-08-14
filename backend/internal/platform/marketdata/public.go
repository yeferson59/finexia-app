package marketdata

import (
	"context"
	"errors"
	"time"
)

// DolarAPI names the keyless public feed at dolarapi.com.
//
// It is a ProviderName so it can be persisted in the same source columns every
// other provider uses, but it is deliberately absent from IsValid and from
// SupportedProviders: those two gate the BYO-key surface, and a user must not
// be able to store a credential for a provider that takes none. Nothing here
// ever reaches Factory.For.
const DolarAPI ProviderName = "dolarapi"

// ECB names the keyless feed of euro reference rates at ecb.europa.eu. It is a
// ProviderName on the same terms as DolarAPI above: persistable as a source,
// never offered as a credential.
const ECB ProviderName = "ecb"

// PublicRate is one exchange rate from a source that needs no API key.
//
// It is not an ExchangeRateResult, and the difference is the whole point.
// ExchangeRateResult answers "what is this pair worth", asked with one user's
// key and answerable only for that user (see 000018). A PublicRate is data the
// source publishes to anyone who asks, so one fetch serves every user and the
// result belongs in the shared exchange_rates table.
type PublicRate struct {
	// From and To are ISO 4217 codes: one unit of From is worth Rate of To.
	From string
	To   string
	// Rate is a decimal string, never a float, so the value the source
	// published survives to the numeric column unrounded.
	Rate string
	// Source is the feed that published it, recorded on the stored row.
	Source ProviderName
	// AsOf is when the source last updated the value, not when we read it. The
	// two differ by up to a full refresh interval, and it is the source's
	// timestamp that dates the rate.
	AsOf time.Time
}

// PublicRateSource publishes exchange rates that cost no credential.
//
// Implemented by marketdata/dolarapi; declared here so the market module can
// depend on the capability without importing the concrete client, the same
// shape Factory has for the BYO-key providers.
type PublicRateSource interface {
	FetchRates(ctx context.Context) ([]PublicRate, error)
}

// PublicRateSources reads several keyless feeds as one.
//
// No feed covers every pair the app converts — the ECB does not quote the
// Colombian peso and dolarapi quotes nothing else — so the sources are additive
// rather than a fallback chain: each is asked, and everything published is
// returned together.
//
// A failing feed does not silence the others. Its error is joined and returned
// alongside whatever the rest published, which is what lets one source being
// down leave the pairs it does not own untouched. Later rates win where two
// feeds publish the same pair, so ordering the slice orders the preference.
type PublicRateSources []PublicRateSource

var _ PublicRateSource = PublicRateSources(nil)

func (s PublicRateSources) FetchRates(ctx context.Context) ([]PublicRate, error) {
	var (
		rates []PublicRate
		errs  []error
	)

	for _, source := range s {
		published, err := source.FetchRates(ctx)
		if err != nil {
			errs = append(errs, err)

			continue
		}

		rates = append(rates, published...)
	}

	// Every source failing is a failed refresh; some of them failing is a
	// partial one, and RefreshPublicRates stores what came back either way.
	if len(rates) == 0 && len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return rates, errors.Join(errs...)
}
