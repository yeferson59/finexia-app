package portfolio

import (
	"testing"

	"github.com/yeferson59/gofinance/v2/money"
)

// The queries that let a caller omit the display currency all read the
// parameter through the same COALESCE + NULLIF pattern: only an empty string
// falls back to the account's own currency.
//
// money.XXX is how "omitted" arrives from the handler, and passing it straight
// to pgx sent the literal "XXX" — the ISO code for "no currency" — which the
// NULLIF keeps. Every amount then came back at a rate of 1 labelled XXX, and
// the reports page printed the generic currency sign (¤) next to the numbers.
func TestCurrencyParam(t *testing.T) {
	if got := currencyParam(money.XXX); got != "" {
		t.Errorf("currencyParam(XXX) = %q, want \"\" so the query falls back to the account's currency", got)
	}

	for _, c := range []money.Currency{money.USD, money.COP, money.EUR} {
		if got := currencyParam(c); got != c.String() {
			t.Errorf("currencyParam(%v) = %q, want %q", c, got, c.String())
		}
	}
}
