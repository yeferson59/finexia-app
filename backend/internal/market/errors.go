package market

import (
	"errors"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// Domain sentinel errors for the "not found" family, tagged with their HTTP
// Kind so httpx.FromDomain maps them by type instead of by substring of the
// message (docs/TECH_DEBT.md #1). Tagging keeps the same 404 the substring
// mapping produced, and stays correct if a caller wraps the error with a
// message that happens to contain "failed"/"invalid".
//
// Messages are preserved verbatim from the persistence layer so response
// bodies and any errors.Is/message assertions keep working.
var (
	ErrAssetNotFound        = httpx.AsNotFound(errors.New("asset not found"))
	ErrExchangeRateNotFound = httpx.AsNotFound(errors.New("exchange rate not found"))
)
