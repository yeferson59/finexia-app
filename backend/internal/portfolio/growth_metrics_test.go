package portfolio

import (
	"testing"
	"time"

	"github.com/yeferson59/gofinance/v2/money"
)

// day builds one point of the growth series. cost is the invested capital at
// that date and flow the money moved since the previous point: the pair is what
// separates a deposit from a gain.
func day(offset int, value, cost, flow string) GrowthPoint {
	return GrowthPoint{
		Date:          time.Date(2026, time.January, 1+offset, 0, 0, 0, 0, time.UTC),
		Currency:      money.USD,
		TotalValue:    value,
		TotalCostBase: cost,
		GainLoss:      "0",
		GainLossPct:   "0",
		NetFlow:       flow,
	}
}

// flat builds n daily points climbing by step, with no flows: a plain series to
// hang the risk thresholds off.
func flat(n int, start, step int) []GrowthPoint {
	points := make([]GrowthPoint, 0, n)
	for i := range n {
		value := start + step*i
		points = append(points, day(i, itoa(value), itoa(start), "0"))
	}

	return points
}

func itoa(n int) string {
	return decimalFromInt(n).String()
}

func TestSubperiodReturnsIgnoresContributions(t *testing.T) {
	// The value doubles, but only because the same amount was paid in.
	subperiods := SubperiodReturns([]GrowthPoint{
		day(0, "1000", "1000", "0"),
		day(1, "2000", "2000", "1000"),
	})

	if len(subperiods) != 1 {
		t.Fatalf("subperiods = %d, want 1", len(subperiods))
	}
	if got := subperiods[0].Rate.RoundBank(6).StringFixed(6); got != "0.000000" {
		t.Errorf("rate = %s, want a flat 0: a deposit is not a return", got)
	}
}

func TestSubperiodReturnsCreditsASaleAtMarketValue(t *testing.T) {
	// Shares bought for 600 are sold for 1000: the value drops by the proceeds
	// and so does the flow, so the day is flat. Netting by cost base instead
	// would have booked the 400 of realised gain as a loss.
	subperiods := SubperiodReturns([]GrowthPoint{
		day(0, "2000", "1600", "0"),
		day(1, "1000", "1000", "-1000"),
	})

	if got := subperiods[0].Rate.RoundBank(6).StringFixed(6); got != "0.000000" {
		t.Errorf("rate = %s, want 0: selling at a profit is not a loss", got)
	}
}

func TestSubperiodReturnsCountsADividendAsIncome(t *testing.T) {
	// A dividend leaves the tracked pool the moment it is paid: the value does
	// not move, the flow is negative, and the return is what the payment was.
	subperiods := SubperiodReturns([]GrowthPoint{
		day(0, "1000", "1000", "0"),
		day(1, "1000", "1000", "-50"),
	})

	// 50 / (1000 - 25) = 0.051282…
	if got := subperiods[0].Rate.RoundBank(6).StringFixed(6); got != "0.051282" {
		t.Errorf("rate = %s, want the dividend credited to the return", got)
	}
}

func TestSubperiodReturnsMeasuresTheMarketOnTopOfAFlow(t *testing.T) {
	// 1000 paid in halfway through and 30 of appreciation on top.
	subperiods := SubperiodReturns([]GrowthPoint{
		day(0, "1000", "1000", "0"),
		day(1, "2030", "2000", "1000"),
	})

	// Modified Dietz: 30 / (1000 + 1000/2).
	if got := subperiods[0].Rate.RoundBank(6).StringFixed(6); got != "0.020000" {
		t.Errorf("rate = %s, want 30/1500", got)
	}
}

func TestSubperiodReturnsSkipsAStretchWithoutCapital(t *testing.T) {
	if got := SubperiodReturns([]GrowthPoint{day(0, "0", "0", "0"), day(1, "0", "0", "0")}); len(got) != 0 {
		t.Errorf("subperiods = %d, want none: there was nothing at work", len(got))
	}
}

func TestSubperiodReturnsNeedsTwoPoints(t *testing.T) {
	if got := SubperiodReturns([]GrowthPoint{day(0, "1000", "1000", "0")}); got != nil {
		t.Errorf("subperiods = %v, want nil", got)
	}
}

func TestBuildGrowthMetricsChainsTheTotalReturn(t *testing.T) {
	m := BuildGrowthMetrics([]GrowthPoint{
		day(0, "1000", "1000", "0"),
		day(1, "1100", "1000", "0"),
		day(2, "1210", "1000", "0"),
	})

	// +10% chained with +10% is +21%, not +20%.
	if got := m.TotalReturn.Mul(oneHundred).RoundBank(2).StringFixed(2); got != "21.00" {
		t.Errorf("total return = %s, want 21.00", got)
	}
	if m.Points != 3 || m.SpanDays != 2 {
		t.Errorf("points = %d, span = %d days", m.Points, m.SpanDays)
	}
}

func TestBuildGrowthMetricsDrawdownRunsOnTheReturnIndex(t *testing.T) {
	m := BuildGrowthMetrics([]GrowthPoint{
		day(0, "1000", "1000", "0"),
		day(1, "1200", "1000", "0"),
		day(2, "900", "1000", "0"),
	})

	if got := m.MaxDrawdown.Mul(oneHundred).RoundBank(2).StringFixed(2); got != "-25.00" {
		t.Errorf("max drawdown = %s, want -25.00", got)
	}
}

func TestBuildGrowthMetricsDoesNotCallAWithdrawalADrawdown(t *testing.T) {
	m := BuildGrowthMetrics([]GrowthPoint{
		day(0, "2000", "2000", "0"),
		day(1, "1000", "1000", "-1000"),
		day(2, "1000", "1000", "0"),
	})

	if got := m.MaxDrawdown.Mul(oneHundred).RoundBank(2).StringFixed(2); got != "0.00" {
		t.Errorf("max drawdown = %s, want 0.00: the owner took the money out", got)
	}
}

func TestBuildGrowthMetricsWithholdsRiskUntilTheHistoryEarnsIt(t *testing.T) {
	m := BuildGrowthMetrics(flat(5, 1000, 10))

	if m.HasVolatility || m.HasSharpe {
		t.Error("volatility/sharpe reported on four subperiods")
	}
	if m.HasAnnualized {
		t.Error("annualized reported on four days of history")
	}
}

// wobbly is n daily points climbing by three with every other one stepping up,
// so the series has something to vary by.
func wobbly(n int) []GrowthPoint {
	points := make([]GrowthPoint, 0, n)
	for i := range n {
		value := 1000 + i*3
		if i%2 == 1 {
			value += 12
		}
		points = append(points, day(i, itoa(value), "1000", "0"))
	}

	return points
}

func TestBuildGrowthMetricsPublishesEveryYearlyFigureAtOnce(t *testing.T) {
	m := BuildGrowthMetrics(wobbly(120))

	if !m.HasVolatility || m.Volatility.Sign() <= 0 {
		t.Errorf("volatility = %s (has = %v), want a positive figure", m.Volatility.String(), m.HasVolatility)
	}
	if !m.VolatilityAnnualized {
		t.Error("volatility left unannualized past the quarter")
	}
	if !m.HasSharpe {
		t.Error("sharpe not reported with a hundred and twenty points")
	}
	if !m.HasAnnualized {
		t.Error("annualized not reported with a hundred and twenty days")
	}
}

func TestBuildGrowthMetricsWithholdsEveryYearlyFigureUnderTheQuarter(t *testing.T) {
	// Thirty daily points: subperiods to spare, history nowhere near a quarter.
	// The Sharpe is the annualized return scaled by another rule, so publishing
	// it while withholding the first said the report both had and did not have
	// the number.
	m := BuildGrowthMetrics(wobbly(30))

	if m.HasSharpe {
		t.Error("sharpe reported on twenty-nine days of history")
	}
	if m.HasAnnualized {
		t.Error("annualized reported on twenty-nine days of history")
	}
	// The dispersion itself is measurable and comes out, only unannualized:
	// √(periods per year) is the premature part, not the standard deviation.
	if !m.HasVolatility || m.Volatility.Sign() <= 0 {
		t.Errorf("volatility = %s (has = %v), want the per-subperiod figure",
			m.Volatility.String(), m.HasVolatility)
	}
	if m.VolatilityAnnualized {
		t.Error("volatility annualized on twenty-nine days of history")
	}
	// And it is the raw deviation, not the annual one under another name: this
	// series moves by tenths of a point from one day to the next.
	if m.Volatility.Mul(oneHundred).Cmp(decimalFromInt(5)) >= 0 {
		t.Errorf("volatility = %s, too high to be the per-subperiod figure", m.Volatility.String())
	}
	// What is not annual still comes out: the period and its worst fall.
	if m.TotalReturn.Sign() <= 0 {
		t.Errorf("total return = %s, want the period figure regardless", m.TotalReturn.String())
	}
}

func TestBuildGrowthMetricsPicksTheExtremesAmongWholeMonthsOnly(t *testing.T) {
	// June opens on the 28th and gains 20 % in two days; August, whole, gains
	// 5 %. Two days do not compete with thirty-one.
	at := func(y int, mo time.Month, d int, value string) GrowthPoint {
		return GrowthPoint{
			Date: time.Date(y, mo, d, 0, 0, 0, 0, time.UTC), Currency: money.USD,
			TotalValue: value, TotalCostBase: "1000", GainLoss: "0", GainLossPct: "0", NetFlow: "0",
		}
	}

	m := BuildGrowthMetrics([]GrowthPoint{
		at(2026, time.June, 28, "1000"),
		at(2026, time.June, 30, "1200"),
		at(2026, time.July, 31, "1140"),
		at(2026, time.August, 31, "1197"),
	})

	if m.Best.Month != "2026-08" || m.Worst.Month != "2026-07" {
		t.Errorf("best/worst = %s/%s, want 2026-08/2026-07: June is a two-day month",
			m.Best.Month, m.Worst.Month)
	}
	if m.MonthsPartialOnly {
		t.Error("MonthsPartialOnly set with two whole months on the series")
	}
}

func TestBuildGrowthMetricsFallsBackToAPartialMonthAndSaysSo(t *testing.T) {
	// Ten days inside one month: nothing whole to compare with, so the figure
	// comes out flagged instead of pretending to be a month.
	m := BuildGrowthMetrics([]GrowthPoint{
		day(9, "1000", "1000", "0"),
		day(19, "1100", "1000", "0"),
	})

	if !m.MonthsPartialOnly {
		t.Error("MonthsPartialOnly unset with no whole month on the series")
	}
	if !m.Best.Partial || !m.Worst.Partial {
		t.Errorf("best/worst partial = %v/%v, want both flagged", m.Best.Partial, m.Worst.Partial)
	}
}

func TestMonthlyReturnsFlagsTheMonthStillRunning(t *testing.T) {
	// The series stops on the 12th: August is twelve days old, not thirty-one.
	at := func(mo time.Month, d int, value string) GrowthPoint {
		return GrowthPoint{
			Date: time.Date(2026, mo, d, 0, 0, 0, 0, time.UTC), Currency: money.USD,
			TotalValue: value, TotalCostBase: "1000", GainLoss: "0", GainLossPct: "0", NetFlow: "0",
		}
	}

	months := MonthlyReturns(SubperiodReturns([]GrowthPoint{
		at(time.June, 30, "1000"),
		at(time.July, 31, "1100"),
		at(time.August, 12, "1150"),
	}))

	if len(months) != 2 {
		t.Fatalf("months = %d, want July and August", len(months))
	}
	if months[0].Partial {
		t.Error("July flagged partial: the series covers it end to end")
	}
	if !months[1].Partial {
		t.Error("August not flagged partial: the series stops on the 12th")
	}
}

func TestBuildGrowthMetricsAnnualizesFromNinetyDays(t *testing.T) {
	m := BuildGrowthMetrics([]GrowthPoint{
		day(0, "1000", "1000", "0"),
		day(89, "1000", "1000", "0"),
		day(120, "1200", "1000", "0"),
	})

	if !m.HasAnnualized {
		t.Fatal("annualized withheld over 120 days of history")
	}
	if m.Annualized.Cmp(m.TotalReturn) <= 0 {
		t.Errorf("annualized %s should exceed the %s earned in under a year",
			m.Annualized.String(), m.TotalReturn.String())
	}
}

func TestMonthlyReturnsChainWithinEachMonth(t *testing.T) {
	months := MonthlyReturns(SubperiodReturns([]GrowthPoint{
		{Date: time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC), TotalValue: "1000", NetFlow: "0"},
		{Date: time.Date(2026, time.February, 14, 0, 0, 0, 0, time.UTC), TotalValue: "1100", NetFlow: "0"},
		{Date: time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC), TotalValue: "1210", NetFlow: "0"},
	}))

	if len(months) != 1 || months[0].Month != "2026-02" {
		t.Fatalf("months = %v, want only February", months)
	}
	if got := months[0].Rate.Mul(oneHundred).RoundBank(2).StringFixed(2); got != "21.00" {
		t.Errorf("February = %s, want 21.00", got)
	}
}

func TestBuildGrowthMetricsPicksTheBestAndWorstMonth(t *testing.T) {
	m := BuildGrowthMetrics([]GrowthPoint{
		{Date: time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC), TotalValue: "1000", NetFlow: "0"},
		{Date: time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC), TotalValue: "1100", NetFlow: "0"},
		{Date: time.Date(2026, time.March, 31, 0, 0, 0, 0, time.UTC), TotalValue: "990", NetFlow: "0"},
	})

	if !m.HasMonthReturns || m.Best.Month != "2026-02" || m.Worst.Month != "2026-03" {
		t.Errorf("best = %v, worst = %v", m.Best, m.Worst)
	}
}

func TestBuildGrowthMetricsWithoutHistory(t *testing.T) {
	if got := BuildGrowthMetrics(nil); got.Points != 0 {
		t.Errorf("points = %d, want 0", got.Points)
	}

	single := BuildGrowthMetrics([]GrowthPoint{day(0, "1000", "1000", "0")})
	if single.Points != 1 || len(single.Subperiod) != 0 {
		t.Errorf("a single point yielded %d subperiods", len(single.Subperiod))
	}
	if single.HasVolatility || single.HasAnnualized || single.HasMonthReturns {
		t.Error("a single point reported metrics it cannot support")
	}
}

func TestBuildGrowthMetricsCarriesTheLastPointsAmounts(t *testing.T) {
	m := BuildGrowthMetrics([]GrowthPoint{
		{Date: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), Currency: money.COP, TotalValue: "1000", TotalCostBase: "1000", GainLoss: "0", NetFlow: "0"},
		{Date: time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC), Currency: money.COP, TotalValue: "1500", TotalCostBase: "1200", GainLoss: "300", NetFlow: "200"},
	})

	if m.Currency != money.COP {
		t.Errorf("currency = %q, want the series' own", m.Currency)
	}
	if got := m.CurrentValue.StringFixed(0); got != "1500" {
		t.Errorf("current value = %s", got)
	}
	if got := m.GainLoss.StringFixed(0); got != "300" {
		t.Errorf("gain/loss = %s", got)
	}
	// The flow of the measured span, not the change in cost base: they differ
	// as soon as there is a fee or a dividend in it.
	if got := m.NetFlow.StringFixed(0); got != "200" {
		t.Errorf("net flow = %s, want 200", got)
	}
}

func TestBuildGrowthMetricsNetFlowSkipsTheOpeningPoint(t *testing.T) {
	// The first point carries whatever fell on or before its own date, which is
	// outside the span the report measures.
	m := BuildGrowthMetrics([]GrowthPoint{
		day(0, "1000", "1000", "1000"),
		day(1, "1500", "1400", "400"),
		day(2, "1600", "1400", "-100"),
	})

	if got := m.NetFlow.StringFixed(0); got != "300" {
		t.Errorf("net flow = %s, want 400 - 100 with the opening 1000 left out", got)
	}
}
