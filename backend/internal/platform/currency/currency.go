// Package currency holds the set of currencies this application can express
// money in.
//
// It is deliberately much narrower than ISO 4217, and the boundary is drawn by
// data rather than taste: a currency is usable only if some rate source reaches
// it, because every figure the app shows in a currency other than a portfolio's
// own is the product of a stored rate. Offering a currency with no rate behind
// it does not fail loudly — it shows unconverted amounts under the wrong
// symbol, which is worse than not offering it.
//
// The list below is exactly what the two keyless public sources publish against
// USD: the ECB reference feed (the majors) and the dolarapi TRM (COP). Those
// need no API key, so the conversion works for every account, not only for one
// that brought its own market-data key. Adding a currency here is safe once a
// source publishes a USD pair for it — nothing else has to change, since the
// sync derives the pairs it fetches from what users hold and prefer.
package currency

import (
	"slices"
	"strings"
)

// Supported is the application-wide currency universe: what an account may
// choose as its preferred currency, and what ?currency= accepts. Those two are
// one list on purpose — a preference the profile stores but the endpoints
// reject is a setting that silently does nothing.
//
// USD leads because it is the hub every stored pair goes through, and COP
// follows it as the app's home market; the rest are the ECB majors.
var Supported = []string{"USD", "COP", "EUR", "GBP", "CHF", "JPY", "CAD", "AUD", "CNY", "MXN", "BRL"}

// Normalize puts a user-supplied code in the shape the rest of the app stores
// and compares: trimmed and upper case. It does not validate — pair it with
// IsSupported.
func Normalize(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// IsSupported reports whether code is one the app can convert to. It expects an
// already-normalized code; callers reading user input should run Normalize
// first.
func IsSupported(code string) bool {
	return slices.Contains(Supported, code)
}

// List renders the set for an error message, so a rejected request tells the
// caller what would have been accepted instead of just saying no.
func List() string {
	return strings.Join(Supported, ", ")
}
