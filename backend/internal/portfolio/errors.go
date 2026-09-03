package portfolio

import (
	"errors"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// Domain sentinel errors for the "not found" family, tagged with their HTTP
// Kind so httpx.FromDomain maps them by type, never by the text of the
// message. Because each is tagged with AsNotFound, a handler that wraps one —
// e.g. fmt.Errorf("failed to load: %w", ...) — still resolves to 404.
//
// The messages are preserved verbatim from the persistence layer, so the
// response body and any errors.Is/message assertions keep working. Callers
// return these values directly (errors.Is-matchable) instead of building a
// fresh errors.New each time.
var (
	ErrPortfolioNotFound         = httpx.AsNotFound(errors.New("portfolio not found"))
	ErrPlatformNotFound          = httpx.AsNotFound(errors.New("platform not found"))
	ErrEntryNotFound             = httpx.AsNotFound(errors.New("portfolio entry not found"))
	ErrTransactionNotFound       = httpx.AsNotFound(errors.New("transaction not found"))
	ErrPortfolioOrSourceNotFound = httpx.AsNotFound(errors.New("portfolio or source not found"))
	ErrExchangeRateNotFound      = httpx.AsNotFound(errors.New("exchange rate not found"))
)

// ErrTransactionFXRate rejects a transaction whose currency and exchange rate
// contradict each other or the position they are being recorded on. It is a 400
// rather than a 422 for the same reason the rest of this module's input errors
// are: the request is malformed, not merely unprocessable, and the caller has
// to change what it sent. TransactionInput.Validate wraps it with which of the
// rules was broken, so the message the client sees names the two currencies
// involved instead of just refusing.
var ErrTransactionFXRate = httpx.AsBadRequest(errors.New("invalid transaction exchange rate"))
