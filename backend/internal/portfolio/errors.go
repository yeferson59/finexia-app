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

// ErrPlatformHasPositions refuses to delete a platform that positions still
// point at.
//
// The refusal is not a policy invented here, it is the only honest answer the
// schema can give. portfolio_entries.source_id is NOT NULL and its foreign key
// says ON DELETE SET NULL (migration 000003), two rules that contradict each
// other the moment a platform holding anything is deleted: Postgres tries to
// null the column, the NOT NULL rejects it, and what reached the client was a
// 500 with a constraint name in the log and nothing in the response. Deleting
// the positions along with the platform is not a fix either — they are the
// owner's trade history, and a stray click on a delete button is not consent to
// erase it.
//
// So the answer is 409: the request is refused, nothing is destroyed, and the
// caller is told what to remove first. It counts *every* entry, including
// positions sold down to nothing, because those still reference the row and
// would still break the delete — which is why the count here can exceed the
// open positions PlatformStats.Investments reports.
var ErrPlatformHasPositions = httpx.AsConflict(errors.New("platform still has positions"))

// ErrTransactionFXRate rejects a transaction whose currency and exchange rate
// contradict each other or the position they are being recorded on. It is a 400
// rather than a 422 for the same reason the rest of this module's input errors
// are: the request is malformed, not merely unprocessable, and the caller has
// to change what it sent. TransactionInput.Validate wraps it with which of the
// rules was broken, so the message the client sees names the two currencies
// involved instead of just refusing.
var ErrTransactionFXRate = httpx.AsBadRequest(errors.New("invalid transaction exchange rate"))

// ErrTransactionFeesCurrency rejects a commission billed in a currency the
// transaction has no way to convert. A row carries exactly one rate, between
// the trade's currency and the position's, so a fee in a third currency could
// only be reached by inventing a second one.
var ErrTransactionFeesCurrency = httpx.AsBadRequest(errors.New("invalid transaction fees currency"))
