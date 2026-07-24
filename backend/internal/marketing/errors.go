package marketing

import (
	"errors"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// ErrWaitlistEmailExists is returned when an email is already on the waitlist
// (the unique constraint fires). It is tagged with its HTTP Kind so
// httpx.FromDomain maps it to 409 by type; before the typed-error migration the
// raw Postgres "duplicate key…" message reached the status mapper by substring
// (docs/TECH_DEBT.md #1). The repository translates the unique violation into
// this sentinel so callers never depend on the driver's error text.
var ErrWaitlistEmailExists = httpx.AsConflict(errors.New("email already on the waitlist"))
