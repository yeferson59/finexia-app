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
	// ErrSnapshotNotFound means no daily snapshot exists as far back as the
	// caller asked. It is an ordinary state for an account in its first week,
	// not a failure, so callers comparing against a past value are expected to
	// handle it rather than propagate it.
	ErrSnapshotNotFound = httpx.AsNotFound(errors.New("portfolio snapshot not found"))
)
