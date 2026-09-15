package mcp

import (
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/portfolio"
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

// A cash balance is matched to the rate of its account — a platform and a
// currency — and only to a version in effect today.
func TestCashAccountRows(t *testing.T) {
	today := time.Date(2026, time.September, 15, 0, 0, 0, 0, time.UTC)
	nu, davivienda := uuid.New(), uuid.New()
	yesterday := today.AddDate(0, 0, -1)
	limit := "25000000"

	balances := []portfolio.CashBalance{
		{SourceID: nu, SourceName: "Nu", Balance: "10000", Currency: money.COP, PendingInterest: "0"},
		// Same platform, another currency: the COP rate is not this one's.
		{SourceID: nu, SourceName: "Nu", Balance: "500", Currency: money.USD, PendingInterest: "0"},
		// A platform whose rate was paused yesterday.
		{SourceID: davivienda, SourceName: "Davivienda", Balance: "2000", Currency: money.COP, PendingInterest: "1.5"},
	}

	rates := []portfolio.CashRate{
		{
			SourceID: nu, Currency: money.COP, AnnualRatePct: "9.25", WithholdingPct: "0",
			Posting: portfolio.PostingMonthly, MaxBalance: &limit,
			EffectiveFrom: today.AddDate(0, -1, 0), Latest: true,
		},
		// Superseded: a version a later one follows is never in effect.
		{
			SourceID: davivienda, Currency: money.COP, AnnualRatePct: "12",
			EffectiveFrom: today.AddDate(0, -2, 0), Latest: false,
		},
		{
			SourceID: davivienda, Currency: money.COP, AnnualRatePct: "11",
			EffectiveFrom: today.AddDate(0, -1, 0), EndedOn: &yesterday, Latest: true,
		},
	}

	rows := cashAccountRows(balances, rates, today)
	if len(rows) != 3 {
		t.Fatalf("rows = %+v, want one per balance", rows)
	}

	if rows[0].AnnualRatePct != "9.25" || rows[0].Posting != "monthly" || rows[0].MaxBalance != limit {
		t.Errorf("the account with a rate = %+v", rows[0])
	}
	if rows[1].AnnualRatePct != "" {
		t.Errorf("the dollar balance took the peso rate: %+v", rows[1])
	}
	if rows[2].AnnualRatePct != "" {
		t.Errorf("a rate paused yesterday still reads as in effect: %+v", rows[2])
	}
	if rows[2].PendingInterest != "1.5" {
		t.Errorf("pending interest = %q, want 1.5", rows[2].PendingInterest)
	}
}
