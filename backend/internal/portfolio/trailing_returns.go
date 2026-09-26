package portfolio

import (
	"time"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/finance/returns"
)

// Trailing returns: what the account earned over the last day, week, month…
//
// Each period is measured on the growth series, from an anchor point in the
// past to the series' last point — today, read live. Like every other figure
// over that series (see growth_metrics.go), a deposit is not a gain: the money
// figure nets out the flows of the window and the percentage chains the
// flow-adjusted subperiod returns. The percentage is therefore the same one the
// chart's percent view and the reports page draw for the same window, which is
// the point — the dashboard must not quote a third number for the same month.

// TrailingPeriod names one look-back window. The values travel as they are in
// the JSON and match the ones ?period= already accepts where they overlap.
type TrailingPeriod string

const (
	TrailingDay     TrailingPeriod = "1D"
	TrailingWeek    TrailingPeriod = "1W"
	TrailingMonth   TrailingPeriod = "1M"
	TrailingQuarter TrailingPeriod = "3M"
	TrailingYTD     TrailingPeriod = "YTD"
	TrailingYear    TrailingPeriod = "1Y"
	TrailingAll     TrailingPeriod = "ALL"
)

// TrailingPeriods is every window, shortest first: the order the response keeps
// so a client can lay them out without sorting.
var TrailingPeriods = []TrailingPeriod{
	TrailingDay, TrailingWeek, TrailingMonth, TrailingQuarter, TrailingYTD, TrailingYear, TrailingAll,
}

// TrailingReturn is one window's result. Nothing but Period and HistoryStart is
// meaningful unless Available: a history shorter than the window has no anchor,
// and publishing the whole history under a "1 year" label would be claiming a
// year nobody watched.
type TrailingReturn struct {
	Period    TrailingPeriod
	Available bool
	// HistoryStart is the series' first date, so a client can say when the
	// window will fill in. Zero when there is no history at all.
	HistoryStart time.Time
	// From is the anchor actually used: the last point on or before the
	// window's target date, which is earlier than the target when the snapshot
	// job missed days. To is the series' last point.
	From, To   time.Time
	StartValue decimal.Decimal
	EndValue   decimal.Decimal
	// NetFlow is the money moved in or out across (From, To].
	NetFlow decimal.Decimal
	// Gain is what the window earned in money: the change in value less the
	// money that came in or went out.
	Gain decimal.Decimal
	// Rate is the time-weighted return of the window, as a fraction. HasRate is
	// false when no stretch of the window had capital at work — an account that
	// was empty — because a 0% there would be a measurement nobody made.
	Rate    decimal.Decimal
	HasRate bool
}

// BuildTrailingReturns measures every period of TrailingPeriods on the series,
// which must be in date order, as the repository returns it. It always yields
// one entry per period.
func BuildTrailingReturns(points []GrowthPoint) []TrailingReturn {
	out := make([]TrailingReturn, 0, len(TrailingPeriods))

	var historyStart time.Time
	if len(points) > 0 {
		historyStart = points[0].Date
	}

	for _, period := range TrailingPeriods {
		entry := TrailingReturn{Period: period, HistoryStart: historyStart}

		if anchor, ok := trailingAnchor(points, period); ok {
			fillTrailingReturn(&entry, points[anchor:])
		}

		out = append(out, entry)
	}

	return out
}

// trailingAnchor is the index of the point the period is measured from, or
// false when the history does not reach back that far. The anchor is always
// before the last point: a window needs at least one stretch.
func trailingAnchor(points []GrowthPoint, period TrailingPeriod) (int, bool) {
	if len(points) < 2 {
		return 0, false
	}

	if period == TrailingAll {
		return 0, true
	}

	target := trailingTarget(points[len(points)-1].Date, period)

	anchor := -1
	for i, p := range points[:len(points)-1] {
		if p.Date.After(target) {
			break
		}
		anchor = i
	}

	return anchor, anchor >= 0
}

// trailingTarget is the date a period looks back to from the series' last day.
//
// The year to date is measured from the close of the previous year, the way a
// statement reports it: on 1 January it is one day long, not zero.
func trailingTarget(last time.Time, period TrailingPeriod) time.Time {
	switch period {
	case TrailingDay:
		return last.AddDate(0, 0, -1)
	case TrailingWeek:
		return last.AddDate(0, 0, -7)
	case TrailingMonth:
		return monthsBefore(last, 1)
	case TrailingQuarter:
		return monthsBefore(last, 3)
	case TrailingYTD:
		return time.Date(last.Year()-1, time.December, 31, 0, 0, 0, 0, last.Location())
	case TrailingYear:
		return monthsBefore(last, 12)
	default:
		return last
	}
}

// monthsBefore steps back n calendar months and clamps to the end of the month
// it lands in. time.AddDate normalizes instead, so a month before 31 March would
// be 3 March, and the "1 month" window would be four weeks long.
func monthsBefore(t time.Time, n int) time.Time {
	firstOfMonth := time.Date(t.Year(), t.Month()-time.Month(n), 1, 0, 0, 0, 0, t.Location())
	lastDay := firstOfMonth.AddDate(0, 1, -1).Day()

	return firstOfMonth.AddDate(0, 0, min(t.Day(), lastDay)-1)
}

// fillTrailingReturn measures the window that opens on window[0] and closes on
// its last point.
func fillTrailingReturn(entry *TrailingReturn, window []GrowthPoint) {
	first, last := window[0], window[len(window)-1]

	netFlow := decimal.Zero
	for _, p := range window[1:] {
		netFlow = netFlow.Add(growthDecimal(p.NetFlow))
	}

	entry.Available = true
	entry.From, entry.To = first.Date, last.Date
	entry.StartValue = growthDecimal(first.TotalValue)
	entry.EndValue = growthDecimal(last.TotalValue)
	entry.NetFlow = netFlow
	entry.Gain = entry.EndValue.Sub(entry.StartValue).Sub(netFlow)

	subperiods := SubperiodReturns(window)
	if len(subperiods) == 0 {
		return
	}

	if rate, err := returns.ChainReturns(rates(subperiods)); err == nil {
		entry.Rate, entry.HasRate = rate, true
	}
}
