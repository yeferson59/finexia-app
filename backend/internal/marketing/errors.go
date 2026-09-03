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

// ErrWaitlistNotFound is returned when an admin deletes a waitlist row that is
// no longer there — another admin got to it first, or the id is stale. Tagged
// so httpx.FromDomain answers 404 instead of the 500 a bare error would give.
var ErrWaitlistNotFound = httpx.AsNotFound(errors.New("waitlist entry not found"))
