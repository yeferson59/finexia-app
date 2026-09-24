package portfolio

import (
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

// planMovements and planMarks are the example of §4 of
// docs/PLAN_FONDOS_INVERSION.md as the repository would read them back.
func planMovements() ([]FundMovement, uuid.UUID) {
	entry := uuid.New()

	return []FundMovement{
		{EntryID: entry, Kind: FundWithdrawal, Date: planDay(time.September, 20), Amount: "2000000", Fees: "0", Units: "19801.87618916", UnitValue: "101.0005305"},
		{EntryID: entry, Kind: FundContribution, Date: planDay(time.August, 15), Amount: "5000000", Fees: "0", Units: "49603.17460317", UnitValue: "100.8"},
		{EntryID: entry, Kind: FundContribution, Date: planDay(time.July, 1), Amount: "10000000", Fees: "0", Units: "100000", UnitValue: "100"},
	}, entry
}

func planMarks() []FundMark {
	return []FundMark{
		{Date: planDay(time.September, 30), UnitValue: "100.53828551"},
		{Date: planDay(time.August, 31), UnitValue: "101.0005305"},
		{Date: planDay(time.July, 31), UnitValue: "100.8"},
	}
}

func periodByKey(t *testing.T, periods []FundPeriod, key string) FundPeriod {
	t.Helper()

	for _, p := range periods {
		if p.Key == key {
			return p
		}
	}

	t.Fatalf("no period %q", key)

	return FundPeriod{}
}

func wantPeriod(t *testing.T, p FundPeriod, from time.Time, days int, pct, ea string) {
	t.Helper()

	if p.From == nil || !p.From.Equal(from) || p.Days == nil || *p.Days != days || p.Pct == nil {
		t.Fatalf("%s = from %v, days %v, pct %v; want from %s over %d days", p.Key, p.From, p.Days, p.Pct, from.Format(time.DateOnly), days)
	}

	sameAmount(t, p.Key+" pct", *p.Pct, pct)

	if ea == "" {
		if p.EAPct != nil {
			t.Errorf("%s E.A. = %s, want none", p.Key, *p.EAPct)
		}

		return
	}

	if p.EAPct == nil {
		t.Fatalf("%s has no E.A., want %s", p.Key, ea)
	}

	sameAmount(t, p.Key+" E.A.", *p.EAPct, ea)
}

func wantEmptyPeriod(t *testing.T, p FundPeriod) {
	t.Helper()

	if p.From != nil || p.Days != nil || p.Pct != nil || p.EAPct != nil {
		t.Errorf("%s = %+v, want empty: the series does not reach back that far", p.Key, p)
	}
}

// The example of the plan: September lost 0,4577 % (−5,43 % E.A.), and the
// fund made 0,5383 % since it opened.
func TestFundPerformancePlanExample(t *testing.T) {
	movements, _ := planMovements()

	series, err := fundSeries(FundBalance, planMarks(), movements)
	if err != nil {
		t.Fatalf("fundSeries: %v", err)
	}

	// The first contribution opens the series at 100; the others traded at a
	// balance already in it.
	if len(series) != 4 || !series[0].Date.Equal(planDay(time.July, 1)) || series[0].UnitValue != "100" {
		t.Fatalf("series = %+v", series)
	}

	periods, err := fundPeriods(series)
	if err != nil {
		t.Fatalf("fundPeriods: %v", err)
	}

	wantPeriod(t, periodByKey(t, periods, "30d"), planDay(time.August, 31), 30, "-0.4577", "-5.43")
	// Ninety days back is the 2nd of July; the latest point on or before it is
	// the opening, 91 days before the end.
	wantPeriod(t, periodByKey(t, periods, "90d"), planDay(time.July, 1), 91, "0.5383", "2.18")
	wantEmptyPeriod(t, periodByKey(t, periods, "180d"))
	wantEmptyPeriod(t, periodByKey(t, periods, "365d"))
	wantEmptyPeriod(t, periodByKey(t, periods, "ytd"))
	wantPeriod(t, periodByKey(t, periods, "inception"), planDay(time.July, 1), 91, "0.5383", "2.18")
}

// The money of the example: 50.000 made, 14.559,89 of it by the withdrawal and
// 35.440,11 still in the units.
func TestFundPerformanceMoney(t *testing.T) {
	movements, entry := planMovements()

	fund := Fund{
		AssetID:   uuid.New(),
		Tracking:  FundBalance,
		Currency:  money.COP,
		Units:     "129801.29841401",
		Value:     "13049999.99951645",
		Cost:      "13014559.89410990", // the units at the average the database keeps, 100.26525199
		UnitValue: new("100.53828551"),
		Positions: []FundPosition{{EntryID: entry}},
	}

	perf, err := buildFundPerformance(fund, planMarks(), movements)
	if err != nil {
		t.Fatalf("buildFundPerformance: %v", err)
	}

	sameAmount(t, "invested", perf.Invested, "15000000")
	sameAmount(t, "withdrawn", perf.Withdrawn, "2000000")
	sameAmount(t, "realized", perf.RealizedGain, "14559.89")

	if perf.UnrealizedGain == nil {
		t.Fatal("no unrealized gain on a fund with a mark")
	}

	sameAmount(t, "unrealized", *perf.UnrealizedGain, "35440.11")

	// Valued at cost there is nothing unrealized to report.
	fund.PricedAtCost = true

	perf, err = buildFundPerformance(fund, nil, movements)
	if err != nil || perf.UnrealizedGain != nil {
		t.Fatalf("at cost: unrealized %v, %v; want none", perf.UnrealizedGain, err)
	}
}

// Fees on a withdrawal are a loss, and the cost of the units sold is the
// average of every purchase of their position, as the database keeps it.
func TestFundMoneyFlowsFeesAndPositions(t *testing.T) {
	a, b := uuid.New(), uuid.New()

	flows, err := fundMoneyFlows([]FundMovement{
		{EntryID: a, Kind: FundContribution, Amount: "1000", Units: "10", Fees: "0"},
		{EntryID: a, Kind: FundContribution, Amount: "1200", Units: "10", Fees: "0"},
		{EntryID: b, Kind: FundContribution, Amount: "500", Units: "10", Fees: "0"},
		// Ten of a's units, at an average of 110, sold for 1300 less 20.
		{EntryID: a, Kind: FundWithdrawal, Amount: "1300", Units: "10", Fees: "20"},
	})
	if err != nil {
		t.Fatalf("fundMoneyFlows: %v", err)
	}

	sameAmount(t, "invested", flows.invested.String(), "2700")
	sameAmount(t, "fees", flows.fees.String(), "20")
	sameAmount(t, "realized", flows.realized.String(), "180")
}

// Followed by units, every purchase and sale is a point: its price is the
// unit value of its day. A mark on the same day wins.
func TestFundSeriesByUnits(t *testing.T) {
	series, err := fundSeries(FundUnits,
		[]FundMark{{Date: planDay(time.September, 1), UnitValue: "12346"}, {Date: planDay(time.September, 30), UnitValue: "12431.22"}},
		[]FundMovement{
			{Kind: FundContribution, Date: planDay(time.September, 1), UnitValue: "12345.678901"},
			{Kind: FundContribution, Date: planDay(time.September, 10), UnitValue: "12380"},
		})
	if err != nil {
		t.Fatalf("fundSeries: %v", err)
	}

	if len(series) != 3 || series[0].UnitValue != "12346" || series[1].UnitValue != "12380" {
		t.Fatalf("series = %+v", series)
	}

	periods, err := fundPeriods(series)
	if err != nil {
		t.Fatalf("fundPeriods: %v", err)
	}

	// Twenty days is too short to annualize.
	short := periodByKey(t, periods, "inception")
	if short.Days == nil || *short.Days != 29 || short.EAPct == nil {
		t.Fatalf("inception = %+v, want 29 days, annualized", short)
	}

	if periods, _ := fundPeriods(series[1:]); periodByKey(t, periods, "inception").EAPct != nil {
		t.Error("a 20-day period was annualized")
	}
}

// An empty series has nothing to measure, and a single point measures nothing.
func TestFundPeriodsWithoutHistory(t *testing.T) {
	if periods, err := fundPeriods(nil); err != nil || len(periods) != 0 {
		t.Fatalf("fundPeriods(nil) = %v, %v", periods, err)
	}

	periods, err := fundPeriods([]FundPoint{{Date: planDay(time.July, 1), UnitValue: "100"}})
	if err != nil {
		t.Fatalf("fundPeriods: %v", err)
	}

	for _, p := range periods {
		wantEmptyPeriod(t, p)
	}
}
