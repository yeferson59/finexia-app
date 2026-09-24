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
			Posting: portfolio.PostingMonthly, Tiers: []portfolio.CashRateTier{{FromBalance: limit, AnnualRatePct: "0"}},
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

	if rows[0].AnnualRatePct != "9.25" || rows[0].Posting != "monthly" || len(rows[0].Tiers) != 1 || rows[0].Tiers[0].FromBalance != limit {
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

// A fund row carries its returns as the manager publishes them, leaves a period
// the history does not cover without a figure, and keeps the gain of a fund
// priced at cost out of the answer.
func TestFundRow(t *testing.T) {
	from := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	days, pct, ea := 29, "0.6929", "9.08"
	unrealized := "85541.10"

	fund := portfolio.Fund{
		Name: "FIC Renta Fija", Currency: money.COP, Tracking: portfolio.FundUnits,
		Value: "12431220", Cost: "12345678.90", UnitValue: new("12431.22"), ValuedOn: new(from.AddDate(0, 0, 29)),
		Positions: []portfolio.FundPosition{{SourceName: "Fiduciaria"}, {SourceName: "Fiduciaria"}},
	}
	perf := portfolio.FundPerformance{
		Invested: "12345678.90", Withdrawn: "0", RealizedGain: "0", UnrealizedGain: &unrealized,
		Periods: []portfolio.FundPeriod{
			{Key: "30d", To: from.AddDate(0, 0, 29)},
			{Key: "inception", From: &from, To: from.AddDate(0, 0, 29), Days: &days, Pct: &pct, EAPct: &ea},
		},
	}

	row := fundRow(fund, perf)

	if len(row.Platforms) != 1 || row.Platforms[0] != "Fiduciaria" {
		t.Errorf("platforms = %v, want the one platform once", row.Platforms)
	}

	if row.UnitValue != "12431.22" || row.ValuedOn != "2026-09-30T00:00:00Z" || row.UnrealizedGain != unrealized {
		t.Errorf("row = %+v", row)
	}

	if len(row.Returns) != 2 || row.Returns[0].Pct != "" || row.Returns[0].From != "" {
		t.Errorf("an uncovered period has a figure: %+v", row.Returns)
	}

	if r := row.Returns[1]; r.Pct != pct || r.EAPct != ea || r.Days != days || r.From != "2026-09-01T00:00:00Z" {
		t.Errorf("inception = %+v", r)
	}

	if row.PublishedBy != "" {
		t.Errorf("a fund the owner marks says it is published by %q", row.PublishedBy)
	}

	fund.PublicFund = &portfolio.PublicFundLink{EntityName: "Fiduciaria Bancolombia", FundName: "FIC Renta Fija", Participation: 501}
	if row := fundRow(fund, perf); row.PublishedBy != "Fiduciaria Bancolombia — FIC Renta Fija (participación 501)" {
		t.Errorf("publishedBy = %q", row.PublishedBy)
	}

	fund.PricedAtCost = true
	perf.UnrealizedGain = nil

	if row := fundRow(fund, perf); row.UnrealizedGain != "" || !row.PricedAtCost {
		t.Errorf("a fund priced at cost = %+v", row)
	}
}
