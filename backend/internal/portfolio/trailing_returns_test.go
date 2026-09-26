package portfolio

import (
	"testing"
	"time"

	"github.com/yeferson59/gofinance/v2/money"
)

// on builds one point of the series on a calendar date.
func on(year int, month time.Month, dayOfMonth int, value, flow string) GrowthPoint {
	return GrowthPoint{
		Date:       time.Date(year, month, dayOfMonth, 0, 0, 0, 0, time.UTC),
		Currency:   money.USD,
		TotalValue: value,
		NetFlow:    flow,
	}
}

func trailingFor(t *testing.T, all []TrailingReturn, period TrailingPeriod) TrailingReturn {
	t.Helper()

	for _, r := range all {
		if r.Period == period {
			return r
		}
	}

	t.Fatalf("period %s missing from %v", period, all)

	return TrailingReturn{}
}

func TestBuildTrailingReturnsAlwaysListsEveryPeriodInOrder(t *testing.T) {
	for _, points := range [][]GrowthPoint{nil, {day(0, "1000", "1000", "0")}, flat(40, 1000, 5)} {
		got := BuildTrailingReturns(points)
		if len(got) != len(TrailingPeriods) {
			t.Fatalf("len = %d, want %d", len(got), len(TrailingPeriods))
		}

		for i, period := range TrailingPeriods {
			if got[i].Period != period {
				t.Errorf("got[%d] = %s, want %s", i, got[i].Period, period)
			}
		}
	}
}

func TestBuildTrailingReturnsWithoutHistory(t *testing.T) {
	for _, r := range BuildTrailingReturns(nil) {
		if r.Available || !r.HistoryStart.IsZero() {
			t.Errorf("%s = %+v, want unavailable with no history start", r.Period, r)
		}
	}

	// One point is a value, not a return: there is no stretch to measure.
	single := day(0, "1000", "1000", "0")
	for _, r := range BuildTrailingReturns([]GrowthPoint{single}) {
		if r.Available || !r.HistoryStart.Equal(single.Date) {
			t.Errorf("%s = %+v, want unavailable, history from %v", r.Period, r, single.Date)
		}
	}
}

func TestBuildTrailingReturnsDoesNotCountADepositAsAGain(t *testing.T) {
	got := BuildTrailingReturns([]GrowthPoint{
		on(2026, time.September, 1, "1000", "0"),
		on(2026, time.September, 24, "1000", "0"),
		// 500 paid in and 30 earned on top.
		on(2026, time.September, 25, "1530", "500"),
	})

	d := trailingFor(t, got, TrailingDay)
	if !d.Available || !d.HasRate {
		t.Fatalf("1D = %+v, want available with a rate", d)
	}
	if g := d.Gain.StringFixed(2); g != "30.00" {
		t.Errorf("gain = %s, want 30.00: the 500 was a deposit", g)
	}
	if f := d.NetFlow.StringFixed(2); f != "500.00" {
		t.Errorf("netFlow = %s, want 500.00", f)
	}
	// Modified Dietz: 30 / (1000 + 500/2).
	if r := d.Rate.RoundBank(6).StringFixed(6); r != "0.024000" {
		t.Errorf("rate = %s, want 30/1250", r)
	}
}

func TestBuildTrailingReturnsDoesNotCountAWithdrawalAsALoss(t *testing.T) {
	got := BuildTrailingReturns([]GrowthPoint{
		on(2026, time.September, 18, "2000", "0"),
		on(2026, time.September, 25, "1000", "-1000"),
	})

	w := trailingFor(t, got, TrailingWeek)
	if !w.Available {
		t.Fatalf("1W = %+v, want available", w)
	}
	if !w.Gain.IsZero() || !w.Rate.IsZero() {
		t.Errorf("1W gain/rate = %s/%s, want 0/0: taking money out is not losing it", w.Gain, w.Rate)
	}
}

func TestBuildTrailingReturnsWithholdsAWindowTheHistoryDoesNotReach(t *testing.T) {
	points := []GrowthPoint{
		on(2026, time.August, 1, "1000", "0"),
		on(2026, time.September, 24, "1100", "0"),
		on(2026, time.September, 25, "1100", "0"),
	}
	got := BuildTrailingReturns(points)

	for _, period := range []TrailingPeriod{TrailingQuarter, TrailingYTD, TrailingYear} {
		r := trailingFor(t, got, period)
		if r.Available {
			t.Errorf("%s = %+v, want unavailable: the history starts on 1 August", period, r)
		}
		if !r.HistoryStart.Equal(points[0].Date) {
			t.Errorf("%s history start = %v, want %v", period, r.HistoryStart, points[0].Date)
		}
	}

	if m := trailingFor(t, got, TrailingMonth); !m.Available || !m.From.Equal(points[0].Date) {
		t.Errorf("1M = %+v, want it anchored on 1 August", m)
	}
	if a := trailingFor(t, got, TrailingAll); !a.Available || a.Gain.StringFixed(2) != "100.00" {
		t.Errorf("ALL = %+v, want +100.00 since the first point", a)
	}
}

func TestBuildTrailingReturnsReportsTheAnchorActuallyUsed(t *testing.T) {
	// The job missed the 22nd to the 24th: the day is measured from the 21st,
	// and says so.
	got := BuildTrailingReturns([]GrowthPoint{
		on(2026, time.September, 21, "1000", "0"),
		on(2026, time.September, 25, "1010", "0"),
	})

	d := trailingFor(t, got, TrailingDay)
	want := time.Date(2026, time.September, 21, 0, 0, 0, 0, time.UTC)
	if !d.Available || !d.From.Equal(want) {
		t.Errorf("1D from = %v, want %v", d.From, want)
	}
}

func TestBuildTrailingReturnsMeasuresTheYearFromLastYearsClose(t *testing.T) {
	points := []GrowthPoint{
		on(2025, time.December, 30, "900", "0"),
		on(2025, time.December, 31, "1000", "0"),
		on(2026, time.January, 1, "1020", "0"),
	}

	y := trailingFor(t, BuildTrailingReturns(points), TrailingYTD)
	if !y.Available || !y.From.Equal(points[1].Date) {
		t.Fatalf("YTD = %+v, want it anchored on 31 December", y)
	}
	if g := y.Gain.StringFixed(2); g != "20.00" {
		t.Errorf("YTD gain on 1 January = %s, want the one day's 20.00", g)
	}

	// Opened this year: no close of the last one to measure from.
	opened := trailingFor(t, BuildTrailingReturns(points[2:]), TrailingYTD)
	if opened.Available {
		t.Errorf("YTD = %+v, want unavailable for an account opened this year", opened)
	}
}

func TestBuildTrailingReturnsWithholdsTheRateOfAnEmptyAccount(t *testing.T) {
	got := BuildTrailingReturns([]GrowthPoint{
		on(2026, time.September, 24, "0", "0"),
		on(2026, time.September, 25, "0", "0"),
	})

	d := trailingFor(t, got, TrailingDay)
	if !d.Available || d.HasRate {
		t.Errorf("1D = %+v, want a window with no rate: nothing was at work", d)
	}
}

func TestBuildTrailingReturnsAllMatchesTheReportsTotalReturn(t *testing.T) {
	points := wobbly(120)

	all := trailingFor(t, BuildTrailingReturns(points), TrailingAll)
	metrics := BuildGrowthMetrics(points)

	if !all.HasRate || all.Rate.Cmp(metrics.TotalReturn) != 0 {
		t.Errorf("ALL rate = %s, want the reports' total return %s", all.Rate, metrics.TotalReturn)
	}
}

func TestBuildTrailingReturnsGainIsTheSumOfEachDays(t *testing.T) {
	points := []GrowthPoint{
		on(2026, time.September, 20, "1000", "0"),
		on(2026, time.September, 21, "1050", "0"),   // +50
		on(2026, time.September, 22, "1600", "600"), // −50
		on(2026, time.September, 23, "1580", "-40"), // +20
		on(2026, time.September, 25, "1610", "0"),   // +30
	}

	all := trailingFor(t, BuildTrailingReturns(points), TrailingAll)
	if g := all.Gain.StringFixed(2); g != "50.00" {
		t.Errorf("ALL gain = %s, want 50.00", g)
	}
}

func TestMonthsBeforeClampsToTheEndOfTheMonth(t *testing.T) {
	cases := []struct {
		from time.Time
		n    int
		want time.Time
	}{
		{time.Date(2026, time.March, 31, 0, 0, 0, 0, time.UTC), 1, time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC)},
		{time.Date(2028, time.March, 31, 0, 0, 0, 0, time.UTC), 1, time.Date(2028, time.February, 29, 0, 0, 0, 0, time.UTC)},
		{time.Date(2026, time.September, 25, 0, 0, 0, 0, time.UTC), 3, time.Date(2026, time.June, 25, 0, 0, 0, 0, time.UTC)},
		{time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC), 3, time.Date(2025, time.November, 15, 0, 0, 0, 0, time.UTC)},
		{time.Date(2028, time.February, 29, 0, 0, 0, 0, time.UTC), 12, time.Date(2027, time.February, 28, 0, 0, 0, 0, time.UTC)},
	}

	for _, tc := range cases {
		if got := monthsBefore(tc.from, tc.n); !got.Equal(tc.want) {
			t.Errorf("monthsBefore(%s, %d) = %s, want %s", tc.from.Format(time.DateOnly), tc.n, got.Format(time.DateOnly), tc.want.Format(time.DateOnly))
		}
	}
}
