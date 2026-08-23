package portfolio

import (
	"strconv"
	"time"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/finance/returns"
)

// Risk and rentability metrics over a growth series.
//
// The series carries the market value of the account on each date and the money
// its owner moved in or out between one date and the next. Only the first is a
// return: adding a position raises the value without anyone having earned
// anything. Every figure here is therefore computed on flow-adjusted subperiod
// returns, never on the raw change in value.
//
// The arithmetic itself comes from gofinance's returns package — the same one
// the growth summary already uses for ROI — so the formulas quoted in the
// spreadsheet are the library's, not a second implementation of them.
//
// The formulas match the ones the reports page runs on the same series
// (frontend/src/lib/features/reports/returns.ts). They have to: the page and
// the spreadsheet describe the same portfolio, and a user who downloads the
// file to check a number expects to find the one they were shown.

// daysPerYear is the length of a year used to annualize. The series is
// calendar-daily, not trading-daily, so 365.25 is the right divisor and the 252
// of trading-day conventions would overstate every annualized figure.
var (
	daysPerYear = decimal.MustFromString("365.25")
	two         = decimal.MustFromString("2")
)

// Below these the risk figures are noise: ten subperiods and three weeks is the
// least that says anything about how much a portfolio swings.
const (
	minRiskReturns = 10
	minRiskDays    = 21
	// Annualizing a fortnight compounds its noise into an absurd yearly rate.
	minAnnualizedDays = 90
)

// SubperiodReturn is one stretch of the series, from one snapshot to the next,
// with the owner's own deposits and withdrawals already netted out.
type SubperiodReturn struct {
	// Date closes the stretch.
	Date time.Time
	// Rate is the fractional return of the stretch: 0.012 is +1.2%.
	Rate decimal.Decimal
	// Days the stretch spans.
	Days int
}

// MonthReturn is the chained return of one calendar month, keyed "2006-01".
type MonthReturn struct {
	Month string
	Rate  decimal.Decimal
}

// GrowthMetrics is what the risk report publishes. The decimal fields are
// fractions; the presentation layer turns them into percentages.
//
// Available reports which of the three risk figures the history sustains, so a
// reader can tell "not enough data yet" from "zero".
type GrowthMetrics struct {
	Currency  string
	FirstDate time.Time
	LastDate  time.Time
	SpanDays  int
	Points    int
	Subperiod []SubperiodReturn

	TotalReturn decimal.Decimal
	MaxDrawdown decimal.Decimal

	Annualized      decimal.Decimal
	HasAnnualized   bool
	Volatility      decimal.Decimal
	HasVolatility   bool
	Sharpe          decimal.Decimal
	HasSharpe       bool
	Best, Worst     MonthReturn
	HasMonthReturns bool

	CurrentValue  decimal.Decimal
	InvestedCost  decimal.Decimal
	GainLoss      decimal.Decimal
	NetInvestment decimal.Decimal
}

// growthDecimal reads one amount of the series, treating an unparsable or empty
// value as zero the way the summary already does.
func growthDecimal(raw string) decimal.Decimal {
	if raw == "" {
		return decimal.Zero
	}

	d, err := decimal.NewFromString(raw)
	if err != nil {
		return decimal.Zero
	}

	return d
}

// SubperiodReturns is the return of every stretch of the series by modified
// Dietz:
//
//	r = (End − Begin − Flow) / (Begin + Flow/2)
//
// The flow leaves the numerator so a deposit is not mistaken for a gain, and
// half of it stays in the denominator because the money that arrived during the
// stretch only worked part of it — which is also what keeps the first day of an
// account, opening at zero, from dividing by it.
//
// A stretch whose invested base is not positive is skipped: there was no
// capital at work to earn anything on.
func SubperiodReturns(points []GrowthPoint) []SubperiodReturn {
	if len(points) < 2 {
		return nil
	}

	subperiods := make([]SubperiodReturn, 0, len(points)-1)

	for i := 1; i < len(points); i++ {
		prev, curr := points[i-1], points[i]

		begin := growthDecimal(prev.TotalValue)
		end := growthDecimal(curr.TotalValue)
		flow := growthDecimal(curr.NetFlow)

		half, err := flow.Div(two)
		if err != nil {
			continue
		}

		base := begin.Add(half)
		if base.Sign() <= 0 {
			continue
		}

		rate, err := end.Sub(begin).Sub(flow).Div(base)
		if err != nil {
			continue
		}

		days := int(curr.Date.Sub(prev.Date).Hours() / 24)
		if days <= 0 {
			continue
		}

		subperiods = append(subperiods, SubperiodReturn{Date: curr.Date, Rate: rate, Days: days})
	}

	return subperiods
}

// rates pulls the bare fractions out, which is the shape gofinance works in.
func rates(subperiods []SubperiodReturn) []decimal.Decimal {
	out := make([]decimal.Decimal, len(subperiods))
	for i, s := range subperiods {
		out[i] = s.Rate
	}

	return out
}

// periodsPerYear is how many stretches of this series fit in a year.
//
// It comes from the median spacing rather than the mean: the series is daily,
// but a snapshot job that missed a run leaves a gap of several days that would
// drag the mean and, through it, every annualized figure.
func periodsPerYear(subperiods []SubperiodReturn) decimal.Decimal {
	gaps := make([]int, len(subperiods))
	for i, s := range subperiods {
		gaps[i] = s.Days
	}

	slicesSortInts(gaps)

	median := gaps[len(gaps)/2]
	if median <= 0 {
		median = 1
	}

	perYear, err := daysPerYear.Div(decimalFromInt(median))
	if err != nil {
		return decimal.One
	}

	return perYear
}

// decimalFromInt is the counted-things bridge into decimal: the package parses
// from text and there is no integer constructor.
func decimalFromInt(n int) decimal.Decimal {
	return decimal.MustFromString(strconv.Itoa(n))
}

// slicesSortInts is an insertion sort: the slice is one entry per snapshot and
// pulling in a generic sort for it is not worth the import.
func slicesSortInts(values []int) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}

// maxDrawdown is the worst fall from a peak of the compounded return index, as
// a non-positive fraction. It runs on the index and not on the value on
// purpose: a withdrawal drops the value without anyone having lost anything.
func maxDrawdown(subperiods []SubperiodReturn) decimal.Decimal {
	index, peak, worst := decimal.One, decimal.One, decimal.Zero

	for _, s := range subperiods {
		index = index.Mul(decimal.One.Add(s.Rate))
		if index.Cmp(peak) > 0 {
			peak = index
		}

		if peak.Sign() <= 0 {
			continue
		}

		fall, err := index.Div(peak)
		if err != nil {
			continue
		}

		if drop := fall.Sub(decimal.One); drop.Cmp(worst) < 0 {
			worst = drop
		}
	}

	return worst
}

// MonthlyReturns chains every stretch that closes in the same calendar month.
// The first month of a series covers only from the day the history starts.
func MonthlyReturns(subperiods []SubperiodReturn) []MonthReturn {
	if len(subperiods) == 0 {
		return nil
	}

	order := make([]string, 0, 12)
	chained := make(map[string]decimal.Decimal, 12)

	for _, s := range subperiods {
		month := s.Date.Format("2006-01")
		if _, seen := chained[month]; !seen {
			order = append(order, month)
			chained[month] = decimal.One
		}

		chained[month] = chained[month].Mul(decimal.One.Add(s.Rate))
	}

	months := make([]MonthReturn, 0, len(order))
	for _, month := range order {
		months = append(months, MonthReturn{Month: month, Rate: chained[month].Sub(decimal.One)})
	}

	return months
}

// BuildGrowthMetrics derives every published figure from the series. An empty
// or one-point series yields a zero value whose Points says so; the caller
// reports "not enough history" rather than a column of zeros pretending to be
// measurements.
func BuildGrowthMetrics(points []GrowthPoint) GrowthMetrics {
	if len(points) == 0 {
		return GrowthMetrics{}
	}

	first, last := points[0], points[len(points)-1]

	metrics := GrowthMetrics{
		Currency:      last.Currency,
		FirstDate:     first.Date,
		LastDate:      last.Date,
		SpanDays:      int(last.Date.Sub(first.Date).Hours() / 24),
		Points:        len(points),
		Subperiod:     SubperiodReturns(points),
		CurrentValue:  growthDecimal(last.TotalValue),
		InvestedCost:  growthDecimal(last.TotalCostBase),
		GainLoss:      growthDecimal(last.GainLoss),
		NetInvestment: growthDecimal(last.TotalCostBase).Sub(growthDecimal(first.TotalCostBase)),
	}

	if len(metrics.Subperiod) == 0 {
		return metrics
	}

	series := rates(metrics.Subperiod)

	if total, err := returns.ChainReturns(series); err == nil {
		metrics.TotalReturn = total
	}

	metrics.MaxDrawdown = maxDrawdown(metrics.Subperiod)

	if months := MonthlyReturns(metrics.Subperiod); len(months) > 0 {
		metrics.HasMonthReturns = true
		metrics.Best, metrics.Worst = months[0], months[0]

		for _, month := range months[1:] {
			if month.Rate.Cmp(metrics.Best.Rate) > 0 {
				metrics.Best = month
			}

			if month.Rate.Cmp(metrics.Worst.Rate) < 0 {
				metrics.Worst = month
			}
		}
	}

	if metrics.SpanDays >= minAnnualizedDays {
		years, err := decimalFromInt(metrics.SpanDays).Div(daysPerYear)
		if err == nil {
			if annualized, err := returns.Annualized(metrics.TotalReturn, years); err == nil {
				metrics.Annualized, metrics.HasAnnualized = annualized, true
			}
		}
	}

	if len(series) < minRiskReturns || metrics.SpanDays < minRiskDays {
		return metrics
	}

	perYear := periodsPerYear(metrics.Subperiod)

	if volatility, err := returns.Volatility(series); err == nil {
		if annualized, err := returns.AnnualizedVolatility(volatility, perYear); err == nil {
			metrics.Volatility, metrics.HasVolatility = annualized, true
		}
	}

	// Risk-free rate zero: the app knows nothing about the user's currency of
	// reference or their alternative, and inventing one would be worse than
	// saying which convention the number follows.
	if sharpe, err := returns.AnnualizedSharpeRatio(series, decimal.Zero, perYear); err == nil {
		metrics.Sharpe, metrics.HasSharpe = sharpe, true
	}

	return metrics
}
