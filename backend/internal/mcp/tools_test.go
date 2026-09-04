package mcp

import (
	"testing"
	"time"

	"github.com/yeferson59/gofinance/v2/money"
)

func TestParseCurrency(t *testing.T) {
	for _, tc := range []struct {
		name    string
		in      string
		want    money.Currency
		wantErr bool
	}{
		// Omitted is not an error: money.XXX is how every service below spells
		// "the account's preferred currency".
		{name: "omitted", in: "", want: money.XXX},
		{name: "lowercase", in: "cop", want: money.COP},
		{name: "padded", in: " usd ", want: money.USD},
		// A real ISO code the app has no rate source for is refused for the
		// same reason it is refused on the REST routes: an unconvertible
		// currency shows unconverted amounts under the wrong symbol.
		{name: "unsupported ISO code", in: "SEK", wantErr: true},
		{name: "not a currency", in: "ZZZ", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseCurrency(tc.in)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("parseCurrency(%q) = %s, want an error", tc.in, got)
				}

				return
			}

			if err != nil {
				t.Fatalf("parseCurrency(%q): %v", tc.in, err)
			}

			if got != tc.want {
				t.Errorf("parseCurrency(%q) = %s, want %s", tc.in, got, tc.want)
			}
		})
	}
}

func TestClampLimit(t *testing.T) {
	for _, tc := range []struct {
		name  string
		limit int
		want  int
	}{
		{name: "omitted", limit: 0, want: 20},
		{name: "negative", limit: -5, want: 20},
		{name: "in range", limit: 50, want: 50},
		{name: "over the ceiling", limit: 1000, want: 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := clampLimit(tc.limit, 20, 200); got != tc.want {
				t.Errorf("clampLimit(%d) = %d, want %d", tc.limit, got, tc.want)
			}
		})
	}
}

func TestTimeText(t *testing.T) {
	if got := timeText(time.Time{}); got != "" {
		t.Errorf("the zero time rendered as %q, want empty — year 1 reads as a date", got)
	}

	want := "2026-03-01T12:00:00Z"
	if got := timeText(time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)); got != want {
		t.Errorf("timeText = %q, want %q", got, want)
	}
}
