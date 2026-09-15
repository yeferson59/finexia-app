package portfolio

import (
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/finance/compoundinterest"
	"github.com/yeferson59/gofinance/v2/money"
)

// The interest a cash balance earns, computed a day at a time from the rate of
// its account (cash_rate.go) and credited as cash_interest. The ledger is
// migration 000044, and the job that walks it, cash_interest_job.go.
//
// A day earns on what the balance held at its close, at the daily equivalent of
// the effective annual rate. Credited daily and earned on the next day, that
// compounds to exactly the annual rate over a year.

// CashRateVersion is a version of a rate as the ledger reads it: the annual and
// withholding rates as fractions, not percentages.
type CashRateVersion struct {
	ID              uuid.UUID
	AnnualRate      decimal.Decimal
	WithholdingRate decimal.Decimal
	EffectiveFrom   time.Time
	EndedOn         *time.Time
}

// coversDay reports whether day falls between the version's first and last
// day. Whether a later version took over by then is the caller's to know.
func (v CashRateVersion) coversDay(day time.Time) bool {
	return !day.Before(v.EffectiveFrom) && (v.EndedOn == nil || !day.After(*v.EndedOn))
}

// CashAccrualTarget is a cash balance whose account has a rate, and where its
// ledger stands.
type CashAccrualTarget struct {
	EntryID uuid.UUID
	// OpenedOn is the day the balance was opened. It earns from then, not from
	// its account's first day with a rate: whatever it was opened with was
	// loaded with the interest it had already earned.
	OpenedOn time.Time
	// LastAccrued is the last day already computed, nil if none.
	LastAccrued *time.Time
	// Versions is every version of the account's rate that started by the day
	// being computed through, oldest first.
	Versions []CashRateVersion
}

// CashAccrualDay is one day to compute, and the version to compute it at.
type CashAccrualDay struct {
	Day    time.Time
	RateID uuid.UUID
}

// PendingDays lists, in order, the days not yet computed through `through` that
// have a rate: from the day after the last one computed, or the first day the
// balance could earn. Days a rate was paused, or before any started, are left
// out.
func (t CashAccrualTarget) PendingDays(through time.Time) []CashAccrualDay {
	if len(t.Versions) == 0 {
		return nil
	}

	from := cashRateDay(t.OpenedOn)
	if first := t.Versions[0].EffectiveFrom; from.Before(first) {
		from = first
	}
	if t.LastAccrued != nil {
		if next := t.LastAccrued.AddDate(0, 0, 1); from.Before(next) {
			from = next
		}
	}

	var days []CashAccrualDay
	v := 0
	for day := from; !day.After(cashRateDay(through)); day = day.AddDate(0, 0, 1) {
		// The version of a day is the latest one that started by then.
		for v+1 < len(t.Versions) && !t.Versions[v+1].EffectiveFrom.After(day) {
			v++
		}

		if version := t.Versions[v]; version.coversDay(day) {
			days = append(days, CashAccrualDay{Day: day, RateID: version.ID})
		}
	}

	return days
}

// dailyRate is the daily equivalent of an effective annual rate, on 365 days:
// (1 + EA)^(1/365) − 1.
func dailyRate(annual decimal.Decimal) (decimal.Decimal, error) {
	return compoundinterest.NewRateConversion().
		RateDecimal(annual).
		EffectiveAnnual().
		Annually().
		ToPeriodicAt(compoundinterest.Daily)
}

// CashAccrual is one day of interest, computed.
type CashAccrual struct {
	Basis       decimal.Decimal
	Gross       decimal.Decimal
	Withholding decimal.Decimal
	Net         decimal.Decimal
	// Credited is what goes into the balance: Net plus the carry of the day
	// before, rounded to the currency's minor unit. Carry is what that rounding
	// left, for the next day.
	Credited decimal.Decimal
	Carry    decimal.Decimal
}

// accrueDay computes a day of interest on basis, at a version of the rate, with
// the carry the day before left. A balance below zero earns nothing.
func accrueDay(basis decimal.Decimal, version CashRateVersion, carry decimal.Decimal, cur money.Currency) (CashAccrual, error) {
	if basis.IsNeg() {
		basis = decimal.Zero
	}

	daily, err := dailyRate(version.AnnualRate)
	if err != nil {
		return CashAccrual{}, err
	}

	decimals, err := cur.GetCurrencyPrecisionCode()
	if err != nil {
		return CashAccrual{}, err
	}

	gross := basis.Mul(daily)
	withholding := gross.Mul(version.WithholdingRate)
	net := gross.Sub(withholding)

	owed := net.Add(carry)
	credited := owed.RoundHAZ(decimals)
	if credited.IsNeg() {
		credited = decimal.Zero
	}

	return CashAccrual{
		Basis:       basis,
		Gross:       gross,
		Withholding: withholding,
		Net:         net,
		Credited:    credited,
		Carry:       owed.Sub(credited),
	}, nil
}
