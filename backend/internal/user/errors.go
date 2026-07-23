package user

import (
	"errors"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// ErrUserNotFound is returned when an operation targets a user that does not
// exist (or was soft-deleted). It is tagged with its HTTP Kind so
// httpx.FromDomain maps it to 404 by type instead of by the "not found"
// substring of the message (docs/TECH_DEBT.md #1), staying correct even if a
// caller wraps it with a message containing "failed"/"invalid". The message is
// preserved verbatim so response bodies and message assertions keep working.
var ErrUserNotFound = httpx.AsNotFound(errors.New("not found user"))
