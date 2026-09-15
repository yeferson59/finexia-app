package portfolio

import (
	"errors"
	"fmt"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/finance/compoundinterest"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/currency"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
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
	// Posting is how often what the version earns reaches the balance. A
	// monthly one computes every day just as a daily one does and holds the
	// days until the month closes; see creditsOn.
	Posting InterestPosting
	// MaxBalance is the most the account earns on, nil when it earns on all of
	// it. It belongs to the account, so its balances share it.
	MaxBalance    *decimal.Decimal
	EffectiveFrom time.Time
	EndedOn       *time.Time
}

// coversDay reports whether day falls between the version's first and last
// day. Whether a later version took over by then is the caller's to know.
func (v CashRateVersion) coversDay(day time.Time) bool {
	return !day.Before(v.EffectiveFrom) && (v.EndedOn == nil || !day.After(*v.EndedOn))
}

// creditsOn reports whether the interest of a day reaches the balance that
// day. A rate posted daily credits every day; a monthly one credits on the
// last day of the month, in one transaction for every day it holds.
func (v CashRateVersion) creditsOn(day time.Time) bool {
	if v.Posting != PostingMonthly {
		return true
	}

	return day.AddDate(0, 0, 1).Day() == 1
}

// earningBasis is the part of a balance that earns on a day. Without a cap that
// is all of it. With one, the cap belongs to the account rather than to any of
// its balances, so when they hold more than it together, each earns on its
// share of the cap, in proportion to what it holds.
//
// accountHeld is what every balance of the account held that day, this one
// included; a balance that is the whole account passes its own basis.
func (v CashRateVersion) earningBasis(basis, accountHeld decimal.Decimal) (decimal.Decimal, error) {
	if v.MaxBalance == nil || !accountHeld.GreaterThan(*v.MaxBalance) {
		return basis, nil
	}

	share, err := v.MaxBalance.Div(accountHeld)
	if err != nil {
		return decimal.Zero, err
	}

	return basis.Mul(share), nil
}

// ErrInvalidCashRecalculation rejects a recalculation that cannot be run as
// asked. It is wrapped with the rule that was broken.
var ErrInvalidCashRecalculation = httpx.AsBadRequest(errors.New("invalid cash interest recalculation"))

// CashAccrualFilter narrows a run to the balances of one account: one platform,
// in one currency, for one owner. The zero value is every balance there is,
// which is what the nightly job computes.
type CashAccrualFilter struct {
	UserID   uuid.UUID
	SourceID uuid.UUID
	Currency money.Currency
}

// RecalculateCashInterestInput asks for an account's interest to be computed
// again from a day.
//
// Days already computed are never revisited on their own: a day earns on what
// the balance held at its close, and a deposit recorded afterwards with a past
// date changes what that was. This is how the owner says so.
type RecalculateCashInterestInput struct {
	SourceID uuid.UUID
	Currency money.Currency
	From     time.Time
}

// Validate checks what the request states. today is the server's clock.
func (in RecalculateCashInterestInput) Validate(today time.Time) error {
	if in.SourceID == (uuid.UUID{}) {
		return fmt.Errorf("%w: sourceId is required", ErrInvalidCashRecalculation)
	}

	if !currency.IsSupported(in.Currency) {
		return fmt.Errorf("%w: currency must be one of: %s", ErrInvalidCashRecalculation, currency.List())
	}

	if in.From.IsZero() {
		return fmt.Errorf("%w: from is required", ErrInvalidCashRecalculation)
	}

	if cashRateDay(in.From).After(cashRateDay(today)) {
		return fmt.Errorf("%w: from cannot be in the future", ErrInvalidCashRecalculation)
	}

	return nil
}

// CashInterestCleared is what a recalculation threw away before computing the
// days again.
type CashInterestCleared struct {
	// From is the first day cleared: the day asked for, or the first day of a
	// credit that reached across it. A month's credit pays every day of its
	// month, so recomputing one of them recomputes them all.
	From     time.Time `json:"from"`
	Balances int       `json:"balances"`
	Days     int       `json:"days"`
}

// CashRecalculation is what a recalculation did.
type CashRecalculation struct {
	Cleared CashInterestCleared `json:"cleared"`
	// Through is the last day computed again, and Credited how many credits
	// that wrote.
	Through    time.Time `json:"through"`
	Credited   int       `json:"credited"`
	Recomputed int       `json:"recomputed"`
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

// CashDay is what a day of interest is computed from: where the balance and its
// ledger stood at the close of the day.
type CashDay struct {
	// Basis is what the balance held, and AccountHeld what every balance of its
	// account held together — the two are equal when the account has only this
	// balance. AccountHeld is only read by a rate with a cap.
	Basis       decimal.Decimal
	AccountHeld decimal.Decimal
	// Carry is what the rounding of the last credit left.
	Carry decimal.Decimal
	// Held is what earlier days earned and have not been credited: the days a
	// rate posted monthly holds until its month closes.
	Held decimal.Decimal
	// Credits is whether this day credits what it and the held days earned.
	// It is CashRateVersion.creditsOn for the day.
	Credits bool
}

// CashAccrual is one day of interest, computed.
type CashAccrual struct {
	// Basis is what the day earned on: what the balance held, or its share of
	// the cap when the account holds more than one.
	Basis       decimal.Decimal
	Gross       decimal.Decimal
	Withholding decimal.Decimal
	Net         decimal.Decimal
	// Credited is what goes into the balance: Net, plus what the days before it
	// held and the carry of the last credit, rounded to the currency's minor
	// unit. Carry is what that rounding left, for the next credit.
	//
	// A day that does not credit leaves Credited at zero and Carry as it found
	// it: its Net waits in the ledger, and the day that closes the month reads
	// it as Held.
	Credited decimal.Decimal
	Carry    decimal.Decimal
}

// accrueDay computes a day of interest at a version of the rate. A balance
// below zero earns nothing.
func accrueDay(day CashDay, version CashRateVersion, cur money.Currency) (CashAccrual, error) {
	basis := day.Basis
	if basis.IsNeg() {
		basis = decimal.Zero
	}

	basis, err := version.earningBasis(basis, day.AccountHeld)
	if err != nil {
		return CashAccrual{}, err
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

	accrual := CashAccrual{
		Basis:       basis,
		Gross:       gross,
		Withholding: withholding,
		Net:         net,
		Credited:    decimal.Zero,
		Carry:       day.Carry,
	}

	if !day.Credits {
		return accrual, nil
	}

	owed := net.Add(day.Held).Add(day.Carry)
	credited := owed.RoundHAZ(decimals)
	if credited.IsNeg() {
		credited = decimal.Zero
	}

	accrual.Credited = credited
	accrual.Carry = owed.Sub(credited)

	return accrual, nil
}

// creditHeld is what a balance that stopped earning before its month closed is
// owed: what its held days earned, plus the carry, rounded. The rest carries
// on, as it does on any other credit.
func creditHeld(held, carry decimal.Decimal, cur money.Currency) (credited, left decimal.Decimal, err error) {
	decimals, err := cur.GetCurrencyPrecisionCode()
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}

	owed := held.Add(carry)

	credited = owed.RoundHAZ(decimals)
	if credited.IsNeg() {
		credited = decimal.Zero
	}

	return credited, owed.Sub(credited), nil
}
