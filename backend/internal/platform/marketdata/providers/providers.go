// Package providers builds a market-data provider from a user's own API keys.
//
// It lives beside the clients rather than inside marketdata because the clients
// import marketdata for the Provider interface; a factory in marketdata itself
// would close that import cycle.
package providers

import (
	"net/http"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata/alphavantage"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata/finnhub"
)

var _ marketdata.Factory = (*Factory)(nil)

// Factory turns a set of credentials into a provider chain. It holds no keys
// of its own — that is the whole point of the BYO-key model — only the shared
// HTTP client the per-call clients borrow.
type Factory struct {
	httpClient *http.Client
}

// New builds the factory. A nil httpClient uses marketdata.DefaultHTTPClient.
func New(httpClient *http.Client) *Factory {
	if httpClient == nil {
		httpClient = marketdata.DefaultHTTPClient
	}

	return new(Factory{httpClient: httpClient})
}

// For assembles the chain for one caller, in marketdata.SupportedProviders
// order so the cheapest quota is spent first. Credentials for unknown providers
// are skipped rather than failing the whole run: a provider retired from the
// code should not lock a user out of the keys that still work.
//
// It returns marketdata.ErrNoCredentials when nothing usable was supplied, so
// callers can tell "this user has not set up a key" apart from "the fetch
// failed".
func (f *Factory) For(creds []marketdata.Credential) (marketdata.Provider, error) {
	byProvider := make(map[marketdata.ProviderName]string, len(creds))

	for _, cred := range creds {
		if cred.APIKey != "" && cred.Provider.IsValid() {
			byProvider[cred.Provider] = cred.APIKey
		}
	}

	chain := make([]marketdata.Provider, 0, len(byProvider))

	for _, name := range marketdata.SupportedProviders {
		apiKey, ok := byProvider[name]
		if !ok {
			continue
		}

		switch name {
		case marketdata.Finnhub:
			chain = append(chain, finnhub.New(apiKey, f.httpClient))
		case marketdata.AlphaVantage:
			chain = append(chain, alphavantage.New(apiKey, f.httpClient))
		}
	}

	if len(chain) == 0 {
		return nil, marketdata.ErrNoCredentials
	}

	return marketdata.NewFallback(chain...), nil
}
