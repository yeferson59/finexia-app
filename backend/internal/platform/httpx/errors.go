package httpx

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// Kind classifies a domain error into the HTTP status FromDomain must emit, so
// the status stops depending on substrings of the error message
// (docs/TECH_DEBT.md #1). A domain tags its errors with one of these values;
// FromDomain resolves the kind through errors.As and only falls back to the
// frozen message-substring mapping for errors that haven't been tagged yet.
type Kind int

const (
	// KindInternal is the zero value and maps to 500. It is also the result
	// when neither a tag nor a substring matches.
	KindInternal Kind = iota
	// KindBadRequest maps to 400.
	KindBadRequest
	// KindNotFound maps to 404.
	KindNotFound
	// KindConflict maps to 409.
	KindConflict
	// KindTooManyRequests maps to 429.
	KindTooManyRequests
)

// httpStatus is the HTTP status a Kind maps to.
func (k Kind) httpStatus() int {
	switch k {
	case KindBadRequest:
		return fiber.StatusBadRequest
	case KindNotFound:
		return fiber.StatusNotFound
	case KindConflict:
		return fiber.StatusConflict
	case KindTooManyRequests:
		return fiber.StatusTooManyRequests
	default:
		return fiber.StatusInternalServerError
	}
}

// statusError couples an error with an explicit Kind. It is unexported on
// purpose: domains build one through the tagging helpers below and keep
// matching the wrapped sentinel with errors.Is/As, never asserting this type
// directly.
type statusError struct {
	kind Kind
	err  error
}

func (e *statusError) Error() string { return e.err.Error() }

// Unwrap keeps the tag transparent to errors.Is/As, so a sentinel wrapped by
// Tagged (directly or through a further fmt.Errorf("...: %w", …)) stays
// matchable by the domain and its tests.
func (e *statusError) Unwrap() error { return e.err }

// Tagged wraps err so FromDomain maps it to kind's HTTP status regardless of
// the error's message. It returns nil for a nil err, so it is safe to write
// `return Tagged(kind, mayBeNil)` at a call site.
func Tagged(kind Kind, err error) error {
	if err == nil {
		return nil
	}

	return &statusError{kind: kind, err: err}
}

// AsBadRequest, AsNotFound, AsConflict and AsTooManyRequests read better at a
// call site than Tagged(KindX, err) and are the intended way to tag domain
// errors. (There is no AsInternal: an untagged error already maps to 500.)
func AsBadRequest(err error) error      { return Tagged(KindBadRequest, err) }
func AsNotFound(err error) error        { return Tagged(KindNotFound, err) }
func AsConflict(err error) error        { return Tagged(KindConflict, err) }
func AsTooManyRequests(err error) error { return Tagged(KindTooManyRequests, err) }

// domainStatus resolves the HTTP status for a domain error: a tagged Kind wins;
// otherwise the frozen substring mapping (docs/API.md §1.2) applies so that
// modules not yet migrated to typed errors keep their exact behavior.
func domainStatus(err error) int {
	var se *statusError
	if errors.As(err, &se) {
		return se.kind.httpStatus()
	}

	return legacyStatus(err.Error())
}

// legacyStatus is the original message-substring mapping, preserved verbatim as
// the fallback for untagged errors. The matching order is part of the frozen
// contract: "failed"/"invalid" win over "not found".
func legacyStatus(msg string) int {
	switch {
	case strings.Contains(msg, "too many"):
		return fiber.StatusTooManyRequests
	case strings.Contains(msg, "failed") || strings.Contains(msg, "invalid"):
		return fiber.StatusBadRequest
	case strings.Contains(msg, "not found"):
		return fiber.StatusNotFound
	case strings.Contains(msg, "already exist") || strings.Contains(msg, "already found") || strings.Contains(msg, "duplicate"):
		return fiber.StatusConflict
	default:
		return fiber.StatusInternalServerError
	}
}
