package market

import (
	"errors"
	"fmt"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

// Domain sentinel errors for the "not found" family, tagged with their HTTP
// Kind so httpx.FromDomain maps them by type rather than by the text of the
// message — the 404 survives a caller wrapping the error with a message that
// happens to contain "failed"/"invalid".
//
// Messages are preserved verbatim from the persistence layer so response
// bodies and any errors.Is/message assertions keep working.
var (
	ErrAssetNotFound        = httpx.AsNotFound(errors.New("asset not found"))
	ErrExchangeRateNotFound = httpx.AsNotFound(errors.New("exchange rate not found"))
	ErrCredentialNotFound   = httpx.AsNotFound(errors.New("market data credential not found"))
)

// BYO-key errors. A user with no key is not a failure of the application, so
// ErrNoCredentials is a 400 the UI turns into "add your key in settings"
// rather than a 500.
var (
	// Wraps the marketdata sentinel so callers can match either: this is the
	// same condition, carrying an HTTP kind. Without the wrap the two would be
	// distinct errors for one state.
	ErrNoCredentials   = httpx.AsBadRequest(fmt.Errorf("no market data credential configured: %w", marketdata.ErrNoCredentials))
	ErrInvalidProvider = httpx.AsBadRequest(errors.New("unknown market data provider"))
	ErrInvalidAPIKey   = httpx.AsBadRequest(errors.New("the provider rejected this API key"))
	// ErrProviderUnavailable means the provider could not be reached or answered
	// something we cannot classify — a timeout, a 5xx, a malformed body.
	//
	// It exists to keep that case apart from ErrInvalidAPIKey. Collapsing the two
	// told the user their key was rejected during an outage, and, worse, let the
	// verify endpoint persist status='invalid', which the sync queries then skip
	// for good: one bad afternoon at the provider would silently retire a working
	// key. Untagged, so it maps to 500 like every other upstream failure.
	ErrProviderUnavailable = errors.New("the market data provider could not be reached")
	ErrKeyEncryptionOff    = errors.New("market: credential encryption is not configured")
)

// isDomainCredentialError reports whether err is one this package authored, and
// is therefore safe to show the user. Anything else — a provider's own text, a
// transport failure — is replaced by a generic message before it reaches a
// response body.
func isDomainCredentialError(err error) bool {
	for _, domain := range []error{ErrNoCredentials, ErrInvalidProvider, ErrInvalidAPIKey, ErrProviderUnavailable, ErrCredentialNotFound} {
		if errors.Is(err, domain) {
			return true
		}
	}

	return false
}
