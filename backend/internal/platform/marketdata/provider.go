package marketdata

import (
	"context"
	"time"
)

type QuoteResult struct {
	Price string
	// Source names the provider that actually answered. With a fallback chain
	// built from several of the user's keys, it is the only way to record which
	// key paid for a given price.
	Source    ProviderName
	FetchedAt time.Time
}

type ExchangeRateResult struct {
	Rate      string
	Source    ProviderName
	FetchedAt time.Time
}

type Provider interface {
	FetchQuote(ctx context.Context, symbol string) (QuoteResult, error)
	FetchExchangeRate(ctx context.Context, from, to string) (ExchangeRateResult, error)
}

// Factory builds a Provider from the credentials of whoever is asking. Under
// the BYO-key model there is no process-wide provider to inject: the chain is
// assembled per sync run from that user's own keys, so this is what the market
// module depends on instead of a Provider.
//
// Implemented by marketdata/providers; declared here so the consumer does not
// have to import the concrete package.
type Factory interface {
	For(creds []Credential) (Provider, error)
}
