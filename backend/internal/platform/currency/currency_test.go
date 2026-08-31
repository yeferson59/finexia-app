package currency

import (
	"slices"
	"strings"
	"testing"

	"github.com/yeferson59/gofinance/v2/money"
)

func TestIsSupported(t *testing.T) {
	// ARS is the case that matters: a real ISO code with no source publishing a
	// USD pair for it. The set is about what can be converted, not what exists.
	for _, code := range []money.Currency{money.ARS, money.Currency(255), money.EUR} {
		if IsSupported(code) {
			t.Errorf("IsSupported(%q) = true, want false", code)
		}
	}

	for _, code := range Supported {
		if !IsSupported(code) {
			t.Errorf("IsSupported(%q) = false for a listed currency", code)
		}
	}
}

// Every stored rate is a USD pair, and a conversion between two other
// currencies is resolved by hopping through it. Dropping USD from the set would
// not fail here, it would silently leave every cross pair unconvertible.
func TestSupportedContainsTheHubCurrency(t *testing.T) {
	if !slices.Contains(Supported, money.USD) {
		t.Fatal("USD missing: every pair is stored against it and cross rates hop through it")
	}
}

func TestList(t *testing.T) {
	got := List()

	for _, code := range Supported {
		if !strings.Contains(got, code.String()) {
			t.Errorf("List() = %q, missing %s", got, code)
		}
	}
}
