package marketdata

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// redacted is what a credential renders as anywhere it might be printed. The
// point is that no logger, error formatter or JSON encoder can be the thing
// that leaks a user's key.
const redacted = "[REDACTED]"

// ProviderName identifies a market-data provider. The values are persisted (as
// market_credentials.provider) and form part of the AAD binding a sealed key to
// its row, so they must stay stable.
type ProviderName string

const (
	AlphaVantage ProviderName = "alphavantage"
	Finnhub      ProviderName = "finnhub"
)

func (p ProviderName) IsValid() bool {
	switch p {
	case AlphaVantage, Finnhub:
		return true
	default:
		return false
	}
}

// SupportedProviders is the order a user's keys are tried in. Finnhub leads
// because its free tier allows 60 calls/minute against Alpha Vantage's 5, so it
// drains a personal quota far more slowly.
var SupportedProviders = []ProviderName{Finnhub, AlphaVantage}

// Credential is a user's own API key, in plaintext. It exists only for the
// duration of the calls that need it: the service opens it from the sealed
// store, builds a provider, and drops it.
//
// The three formatting methods below are not decoration. A Credential reaches
// error paths and log lines that were written without it in mind, and the
// default rendering of a struct would print the key.
type Credential struct {
	Provider ProviderName
	APIKey   string
}

// String keeps the key out of explicit stringification.
func (c Credential) String() string { return string(c.Provider) + ":" + redacted }

// Format makes every fmt verb render the redacted form, including %#v.
//
// String alone is not enough: %#v asks for the Go-syntax representation and
// bypasses Stringer entirely, printing every field — which is exactly the
// accident this type exists to prevent. Implementing Formatter is the only way
// to intercept it.
func (c Credential) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v':
		if f.Flag('#') {
			_, _ = fmt.Fprintf(f, "marketdata.Credential{Provider:%q, APIKey:%q}", string(c.Provider), redacted)

			return
		}

		fallthrough
	default:
		_, _ = io.WriteString(f, c.String())
	}
}

// MarshalJSON keeps the key out of any response body a Credential is
// accidentally embedded in.
func (c Credential) MarshalJSON() ([]byte, error) {
	return []byte(`{"provider":"` + string(c.Provider) + `","apiKey":"` + redacted + `"}`), nil
}

// LogValue keeps the key out of structured logs.
func (c Credential) LogValue() slog.Value {
	return slog.GroupValue(slog.String("provider", string(c.Provider)), slog.String("apiKey", redacted))
}

// scrub removes an API key from arbitrary text. Providers take the key as a URL
// query parameter, and Go's transport errors quote the full URL, so an error
// returned verbatim would carry the key into logs and response bodies. Every
// client funnels its errors through this.
func scrub(text, apiKey string) string {
	if apiKey == "" {
		return text
	}

	return strings.ReplaceAll(text, apiKey, redacted)
}

// DefaultHTTPClient is shared by every provider client. Clients are now built
// per sync run rather than once at startup, so each one minting its own client
// would leak connections and defeat keep-alive.
var DefaultHTTPClient = new(http.Client{Timeout: 10 * time.Second})
