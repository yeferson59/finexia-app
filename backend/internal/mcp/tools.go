package mcp

import (
	"fmt"
	"strings"
	"time"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/currency"
)

// parseCurrency resolves an ISO 4217 code coming from a tool argument.
//
// The empty string is not an error: it is what every service below reads as
// "the account's preferred currency", and money.XXX is how that is spelled. A
// code outside the supported set is rejected with the list that would have been
// accepted, because the caller here is a model and an error it can act on is
// worth more than a 400 it can only report.
func parseCurrency(code string) (money.Currency, error) {
	if code == "" {
		return money.XXX, nil
	}

	cur, err := money.GetCurrencyFromISOCode(strings.ToUpper(strings.TrimSpace(code)))
	if err != nil || !currency.IsSupported(cur) {
		return money.XXX, fmt.Errorf("unsupported currency %q: must be one of %s", code, currency.List())
	}

	return cur, nil
}

// clampLimit bounds a caller-supplied page size. Zero means the caller omitted
// it and gets the default; anything past the ceiling is capped rather than
// refused, since an over-large limit is a guess, not a mistake worth failing on.
func clampLimit(limit, fallback, ceiling int) int {
	switch {
	case limit <= 0:
		return fallback
	case limit > ceiling:
		return ceiling
	default:
		return limit
	}
}

// timeText renders a timestamp for the wire. The zero time becomes the empty
// string rather than year 1: every time field on these DTOs is optional, and
// "0001-01-01T00:00:00Z" reads as a date to a model, which is worse than absent.
func timeText(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format(time.RFC3339)
}
