package support

import (
	"errors"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// ErrDisabled means the Bold keys are not configured: the page stays up and
// says contributions are unavailable. The handler answers it with 503, which
// httpx has no Kind for.
var ErrDisabled = errors.New("contributions are not available: the Bold keys are not configured")

// ErrContributionNotFound covers an unknown order id and a malformed one alike:
// telling them apart would only help someone guessing ids.
var ErrContributionNotFound = httpx.AsNotFound(errors.New("contribution not found"))

// ErrAmountOutOfRange is an amount outside [MinAmount, MaxAmount].
var ErrAmountOutOfRange = httpx.AsBadRequest(errors.New("amount out of range"))

// ErrInvalidSignature is a webhook whose HMAC does not match. The handler
// answers 401 so a forged or misconfigured delivery shows up as such in Bold's
// panel and in the logs.
var ErrInvalidSignature = errors.New("invalid webhook signature")

// ErrMalformedEvent is a correctly signed body this module cannot read.
var ErrMalformedEvent = httpx.AsBadRequest(errors.New("malformed webhook event"))
