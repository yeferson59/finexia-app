package portfolio

import (
	"math"
	"sort"
	"strconv"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// The return of a fund, the way its manager publishes it (D10 of
// docs/PLAN_FONDOS_INVERSION.md): how the unit value moved over a period, and
// that movement annualized.
//
//	period = VU_t / VU_s − 1
//	E.A.   = (VU_t / VU_s)^(365 / days) − 1
//
// t is the latest point of the fund's series and s the latest point on or
// before t − n days. A period the series does not reach back to is left empty
// rather than extrapolated, and the annualized figure is only given over 28
// days or more: annualizing a week turns a quiet week into a headline.
//
// Next to it, the money: what went in and out, and the gain realized by the
// withdrawals and still held by the units. Those two answer different
// questions — the percentage is how the fund did, the money is what the owner
// made with their own timing — and both are shown.

// fundMinAnnualizedDays is the shortest period whose return is annualized.
const fundMinAnnualizedDays = 28

// fundPeriodDays are the fixed-length periods, by key.
var fundPeriodDays = []struct {
	key  string
	days int
}{
	{"30d", 30}, {"90d", 90}, {"180d", 180}, {"365d", 365},
}

var oneHundredPct = decimal.MustFromString("100")

// FundPoint is the unit value of a fund on a day.
type FundPoint struct {
	Date      time.Time `json:"date"`
	UnitValue string    `json:"unitValue"`
}

// FundPeriod is the return of a fund over one period. From, Days, Pct and
// EAPct are nil when the series does not reach back that far; EAPct is also nil
// for a period shorter than fundMinAnnualizedDays.
type FundPeriod struct {
	Key   string     `json:"key"`
	From  *time.Time `json:"from"`
	To    time.Time  `json:"to"`
	Days  *int       `json:"days"`
	Pct   *string    `json:"pct"`
	EAPct *string    `json:"eaPct"`
}

// FundPerformance is everything the fund's detail shows about how it did.
type FundPerformance struct {
	AssetID      uuid.UUID      `json:"assetId"`
	Tracking     FundTracking   `json:"tracking"`
	Currency     money.Currency `json:"currency"`
	ValuedOn     *time.Time     `json:"valuedOn"`
	UnitValue    *string        `json:"unitValue"`
	Units        string         `json:"units"`
	Value        string         `json:"value"`
	Cost         string         `json:"cost"`
	PricedAtCost bool           `json:"pricedAtCost"`
	// Invested is what every contribution put in, Withdrawn what every
	// withdrawal paid out before its fees, and Fees those fees.
	Invested  string `json:"invested"`
	Withdrawn string `json:"withdrawn"`
	Fees      string `json:"fees"`
	// RealizedGain is what the withdrawals made over what their units cost, net
	// of fees; UnrealizedGain what the units still held are worth over their
	// cost, nil while the fund is valued at cost.
	RealizedGain   string       `json:"realizedGain"`
	UnrealizedGain *string      `json:"unrealizedGain"`
	Periods        []FundPeriod `json:"periods"`
	Series         []FundPoint  `json:"series"`
}

// fundSeries is the fund's unit value by day: its marks, plus the points its
// movements prove.
//
// Followed by units, a purchase or a sale is made at the unit value of its day,
// so its price is a point too — which is what gives "since inception" a start
// before anyone typed a mark. Followed by balance, the one point a movement
// proves is the first contribution, which bought at the opening unit value; the
// others traded at the last balance before them, which is already a point. A
// mark on the same day wins: it is what the statement said.
func fundSeries(tracking FundTracking, marks []FundMark, movements []FundMovement) ([]FundPoint, error) {
	byDay := make(map[time.Time]decimal.Decimal)

	switch tracking {
	case FundUnits:
		for _, m := range movements {
			v, err := decimal.NewFromString(m.UnitValue)
			if err != nil {
				return nil, err
			}

			if v.IsPos() {
				byDay[cashRateDay(m.Date)] = v
			}
		}
	case FundBalance:
		var first *FundMovement

		for i := range movements {
			m := &movements[i]
			if m.Kind != FundContribution {
				continue
			}

			if first == nil || m.Date.Before(first.Date) || (m.Date.Equal(first.Date) && m.CreatedAt.Before(first.CreatedAt)) {
				first = m
			}
		}

		if first != nil {
			byDay[cashRateDay(first.Date)] = fundOpeningUnitValue
		}
	}

	for _, m := range marks {
		v, err := decimal.NewFromString(m.UnitValue)
		if err != nil {
			return nil, err
		}

		byDay[cashRateDay(m.Date)] = v
	}

	series := make([]FundPoint, 0, len(byDay))
	for day, v := range byDay {
		series = append(series, FundPoint{Date: day, UnitValue: v.String()})
	}

	sort.Slice(series, func(i, j int) bool { return series[i].Date.Before(series[j].Date) })

	return series, nil
}

// fundPeriods measures the series over each period, ending at its last point.
func fundPeriods(series []FundPoint) ([]FundPeriod, error) {
	if len(series) == 0 {
		return []FundPeriod{}, nil
	}

	last := series[len(series)-1]
	to := last.Date

	end, err := decimal.NewFromString(last.UnitValue)
	if err != nil {
		return nil, err
	}

	// The latest point on or before a day, or nil.
	atOrBefore := func(day time.Time) *FundPoint {
		i := sort.Search(len(series), func(i int) bool { return series[i].Date.After(day) })
		if i == 0 {
			return nil
		}

		return &series[i-1]
	}

	measure := func(key string, start *FundPoint) (FundPeriod, error) {
		p := FundPeriod{Key: key, To: to}
		if start == nil || !start.Date.Before(to) {
			return p, nil
		}

		begin, err := decimal.NewFromString(start.UnitValue)
		if err != nil || !begin.IsPos() {
			return p, err
		}

		ratio, err := end.Div(begin)
		if err != nil {
			return p, err
		}

		days := int(to.Sub(start.Date).Hours() / 24)
		pct := ratio.Sub(decimal.One).Mul(oneHundredPct).RoundHAZ(4).String()

		from := start.Date
		p.From, p.Days, p.Pct = &from, &days, &pct

		if days >= fundMinAnnualizedDays {
			// Annualized in float: the exponent is fractional, and the figure is
			// shown at two decimals.
			ea := (math.Pow(ratio.InexactFloat64(), 365/float64(days)) - 1) * 100
			if !math.IsInf(ea, 0) && !math.IsNaN(ea) {
				if d, err := decimal.NewFromString(strconv.FormatFloat(ea, 'f', 6, 64)); err == nil {
					s := d.RoundHAZ(2).String()
					p.EAPct = &s
				}
			}
		}

		return p, nil
	}

	periods := make([]FundPeriod, 0, len(fundPeriodDays)+2)

	for _, pd := range fundPeriodDays {
		p, err := measure(pd.key, atOrBefore(to.AddDate(0, 0, -pd.days)))
		if err != nil {
			return nil, err
		}

		periods = append(periods, p)
	}

	// The year so far starts at the close of the year before.
	yearEnd := time.Date(to.Year()-1, time.December, 31, 0, 0, 0, 0, time.UTC)

	ytd, err := measure("ytd", atOrBefore(yearEnd))
	if err != nil {
		return nil, err
	}

	inception, err := measure("inception", &series[0])
	if err != nil {
		return nil, err
	}

	return append(periods, ytd, inception), nil
}

// fundMoney is the money side of a fund's performance, from its movements.
type fundMoney struct {
	invested, withdrawn, fees, realized decimal.Decimal
}

// fundMoneyFlows adds up the fund's movements and the gain its withdrawals
// realized.
//
// The cost of a sold unit is the position's average cost the way the database
// keeps it (recalculate_avg_cost): every purchase's cost over every purchase's
// units, fees left out and not moved by sales. Measured the same way, the
// realized gain and the unrealized one — value over the cost the position
// shows — add up to what came out plus what is left over what went in.
func fundMoneyFlows(movements []FundMovement) (fundMoney, error) {
	type position struct{ boughtUnits, boughtCost decimal.Decimal }

	var (
		out       fundMoney
		positions = make(map[uuid.UUID]*position)
		sales     []FundMovement
	)

	for _, m := range movements {
		amount, err := decimal.NewFromString(m.Amount)
		if err != nil {
			return out, err
		}

		p := positions[m.EntryID]
		if p == nil {
			p = &position{}
			positions[m.EntryID] = p
		}

		if m.Kind == FundContribution {
			units, err := decimal.NewFromString(m.Units)
			if err != nil {
				return out, err
			}

			out.invested = out.invested.Add(amount)
			p.boughtUnits = p.boughtUnits.Add(units)
			p.boughtCost = p.boughtCost.Add(amount)

			continue
		}

		sales = append(sales, m)
	}

	for _, m := range sales {
		amount, err := decimal.NewFromString(m.Amount)
		if err != nil {
			return out, err
		}

		units, err := decimal.NewFromString(m.Units)
		if err != nil {
			return out, err
		}

		fees, err := decimal.NewFromString(m.Fees)
		if err != nil {
			return out, err
		}

		out.withdrawn = out.withdrawn.Add(amount)
		out.fees = out.fees.Add(fees)

		cost := decimal.Decimal{}
		if p := positions[m.EntryID]; p != nil && p.boughtUnits.IsPos() {
			avg, err := p.boughtCost.Div(p.boughtUnits)
			if err != nil {
				return out, err
			}

			cost = units.Mul(avg)
		}

		out.realized = out.realized.Add(amount.Sub(fees).Sub(cost))
	}

	out.realized = out.realized.RoundHAZ(2)

	return out, nil
}

// buildFundPerformance puts a fund, its marks and its movements together.
func buildFundPerformance(fund Fund, marks []FundMark, movements []FundMovement) (FundPerformance, error) {
	series, err := fundSeries(fund.Tracking, marks, movements)
	if err != nil {
		return FundPerformance{}, err
	}

	periods, err := fundPeriods(series)
	if err != nil {
		return FundPerformance{}, err
	}

	flows, err := fundMoneyFlows(movements)
	if err != nil {
		return FundPerformance{}, err
	}

	perf := FundPerformance{
		AssetID:      fund.AssetID,
		Tracking:     fund.Tracking,
		Currency:     fund.Currency,
		ValuedOn:     fund.ValuedOn,
		UnitValue:    fund.UnitValue,
		Units:        fund.Units,
		Value:        fund.Value,
		Cost:         fund.Cost,
		PricedAtCost: fund.PricedAtCost,
		Invested:     flows.invested.String(),
		Withdrawn:    flows.withdrawn.String(),
		Fees:         flows.fees.String(),
		RealizedGain: flows.realized.String(),
		Periods:      periods,
		Series:       series,
	}

	if !fund.PricedAtCost {
		value, err := decimal.NewFromString(fund.Value)
		if err != nil {
			return FundPerformance{}, err
		}

		cost, err := decimal.NewFromString(fund.Cost)
		if err != nil {
			return FundPerformance{}, err
		}

		gain := value.Sub(cost).RoundHAZ(2).String()
		perf.UnrealizedGain = &gain
	}

	return perf, nil
}
