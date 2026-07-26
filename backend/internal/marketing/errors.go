package marketing

import (
	"errors"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// ErrWaitlistEmailExists is returned when an email is already on the waitlist
// (the unique constraint fires). It is tagged with its HTTP Kind so
// httpx.FromDomain maps it to 409 by type. The repository translates the
// unique violation into this sentinel so neither the status nor the response
// body ever depends on the driver's "duplicate key…" text.
var ErrWaitlistEmailExists = httpx.AsConflict(errors.New("email already on the waitlist"))
