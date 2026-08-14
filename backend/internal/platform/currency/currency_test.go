package currency

import (
	"slices"
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	cases := []struct{ in, want string }{
		{"  eur ", "EUR"},
		{"usd", "USD"},
		{"COP", "COP"},
		{"", ""},
	}

	for _, c := range cases {
		if got := Normalize(c.in); got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsSupported(t *testing.T) {
	// ARS is the case that matters: a real ISO code with no source publishing a
	// USD pair for it. The set is about what can be converted, not what exists.
	for _, code := range []string{"ARS", "ZZZ", "eur", " EUR", ""} {
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
	if !slices.Contains(Supported, "USD") {
		t.Fatal("USD missing: every pair is stored against it and cross rates hop through it")
	}
}

func TestSupportedIsNormalized(t *testing.T) {
	for _, code := range Supported {
		if code != Normalize(code) || len(code) != 3 {
			t.Errorf("%q is not a normalized three-letter code", code)
		}
	}

	if len(slices.Compact(slices.Sorted(slices.Values(Supported)))) != len(Supported) {
		t.Errorf("Supported has duplicates: %s", List())
	}
}

func TestList(t *testing.T) {
	got := List()

	for _, code := range Supported {
		if !strings.Contains(got, code) {
			t.Errorf("List() = %q, missing %s", got, code)
		}
	}
}
