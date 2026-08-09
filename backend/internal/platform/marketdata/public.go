package marketdata

import (
	"context"
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
