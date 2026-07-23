package portfolio

import (
	"errors"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// Domain sentinel errors for the "not found" family, tagged with their HTTP
// Kind so httpx.FromDomain maps them by type instead of by substring of the
// message (docs/TECH_DEBT.md #1). Because each is tagged with AsNotFound, a
// handler that wraps one — e.g. fmt.Errorf("failed to load: %w", ...) — still
// resolves to 404, whereas the old substring mapping would have seen "failed"
// and wrongly returned 400.
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
