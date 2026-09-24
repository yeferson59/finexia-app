package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

// A rate recorded after the fact, and a change of rate moved to another day:
// CreateCashRate and RescheduleCashRate with a first day in the past, and the
// service computing those days at once. Same database contract as
// postgres_cash_db_test.go.
//
// The service computes through the last day closed by the real clock
// (lastClosedCashDay), so these days are counted back from the day after it
// rather than from cashDay.

// cashToday is the first day not yet closed: today in UTC once the nightly run's
// hour has come, and yesterday before it. "cashToday - 1" is always the last day
// the service computes, whatever the hour the suite runs at.
func cashToday() time.Time {
	return lastClosedCashDay(time.Now()).AddDate(0, 0, 1)
}

func (f cashFixture) usdRateInput(t *testing.T, pct string, from time.Time, recompute bool) NewCashRateInput {
	t.Helper()

	return NewCashRateInput{
		SourceID:      f.sourceID,
		Currency:      money.USD,
		EffectiveFrom: from,
		Recompute:     recompute,
		CashRateInput: CashRateInput{
			AnnualRatePct:  mustDecimal(t, pct),
			WithholdingPct: mustDecimal(t, "0"),
			Posting:        PostingDaily,
		},
	}
}

// daysAt counts the balance's days computed at a version.
func (f cashFixture) daysAt(t *testing.T, entryID, rateID uuid.UUID) int {
	t.Helper()

	var n int
	if err := f.pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM cash_interest_accruals WHERE entry_id = $1 AND rate_id = $2`, entryID, rateID,
	).Scan(&n); err != nil {
		t.Fatalf("count days: %v", err)
	}

	return n
}

// rateByID is the version as the list reads it now.
func (f cashFixture) rateByID(t *testing.T, rateID uuid.UUID) CashRate {
	t.Helper()

	for _, rate := range f.rates(t) {
		if rate.ID == rateID {
			return rate
		}
	}

	t.Fatalf("rate %v is gone", rateID)

	return CashRate{}
}

// A deposit recorded today with a date ten days back, and a rate recorded today
// from that same day: the balance earns all ten days, not only from today.
func TestCashRateFromThePastEarnsFromTheDeposit(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()
	start := cashToday().AddDate(0, 0, -10)

	deposit := f.mustMove(t, CashKindDeposit, "10000", start)

	rate, err := f.accrualService().CreateCashRate(ctx, f.userID, f.usdRateInput(t, "9", start, false))
	if err != nil {
		t.Fatalf("CreateCashRate: %v", err)
	}

	if n := f.daysAt(t, deposit.EntryID, rate.ID); n != 10 {
		t.Errorf("days computed = %d, want the ten from the deposit through yesterday", n)
	}

	row, ok := f.accrualOn(t, deposit.EntryID, start)
	if !ok {
		t.Fatal("the deposit's own day was not computed")
	}
	sameAmount(t, "first day's basis", row.basis, "10000")
}

// A change of rate recorded late: the days already computed at the old rate
// are kept unless the owner asks for them to be computed again, and then they
// are, at the new one.
func TestCashRateFromThePastRecomputesTheComputedDays(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()
	svc := f.accrualService()
	start := cashToday().AddDate(0, 0, -10)
	change := cashToday().AddDate(0, 0, -4)

	deposit := f.mustMove(t, CashKindDeposit, "10000", start)
	old := f.usdRate(t, "5", start)

	if _, errs := svc.AccrueCashInterest(ctx, cashToday().AddDate(0, 0, -1)); len(errs) > 0 {
		t.Fatalf("AccrueCashInterest: %v", errs)
	}
	if n := f.daysAt(t, deposit.EntryID, old.ID); n != 10 {
		t.Fatalf("days at the old rate = %d, want 10", n)
	}

	if _, err := svc.CreateCashRate(ctx, f.userID, f.usdRateInput(t, "9", change, false)); !errors.Is(err, ErrCashRateInUse) {
		t.Fatalf("a version inside the computed days without recompute = %v, want ErrCashRateInUse", err)
	}

	current, err := svc.CreateCashRate(ctx, f.userID, f.usdRateInput(t, "9", change, true))
	if err != nil {
		t.Fatalf("CreateCashRate with recompute: %v", err)
	}

	if n := f.daysAt(t, deposit.EntryID, old.ID); n != 6 {
		t.Errorf("days kept at the old rate = %d, want the six before the change", n)
	}
	if n := f.daysAt(t, deposit.EntryID, current.ID); n != 4 {
		t.Errorf("days at the new rate = %d, want the four from the change", n)
	}
	if n := f.countInterest(t, deposit.EntryID); n != 10 {
		t.Errorf("credits = %d, want one a day, none twice", n)
	}

	lastOld := change.AddDate(0, 0, -1)
	wantEndedOn(t, "old version", f.rateByID(t, old.ID), &lastOld)
}

// A change of rate moved to another day, forward and back, while no day has
// been computed at it.
func TestRescheduleCashRate(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()
	svc := f.accrualService()
	start := cashToday().AddDate(0, 0, -10)

	deposit := f.mustMove(t, CashKindDeposit, "10000", start)
	base := f.usdRate(t, "5", start)

	if _, errs := svc.AccrueCashInterest(ctx, cashToday().AddDate(0, 0, -1)); len(errs) > 0 {
		t.Fatalf("AccrueCashInterest: %v", errs)
	}

	announced := cashToday().AddDate(0, 0, 10)
	change, err := svc.CreateCashRate(ctx, f.userID, f.usdRateInput(t, "9", announced, false))
	if err != nil {
		t.Fatalf("CreateCashRate: %v", err)
	}

	t.Run("later", func(t *testing.T) {
		later := cashToday().AddDate(0, 0, 20)

		moved, err := svc.RescheduleCashRate(ctx, f.userID, change.ID, RescheduleCashRateInput{EffectiveFrom: later})
		if err != nil {
			t.Fatalf("RescheduleCashRate: %v", err)
		}
		if !moved.EffectiveFrom.Equal(later) {
			t.Errorf("effectiveFrom = %v, want %v", moved.EffectiveFrom, later)
		}

		lastBase := later.AddDate(0, 0, -1)
		wantEndedOn(t, "the version before", f.rateByID(t, base.ID), &lastBase)
	})

	t.Run("not before the version before", func(t *testing.T) {
		_, err := svc.RescheduleCashRate(ctx, f.userID, change.ID, RescheduleCashRateInput{EffectiveFrom: start})
		checkCashRateError(t, err, "must be after")
	})

	t.Run("back into computed days", func(t *testing.T) {
		back := cashToday().AddDate(0, 0, -3)

		if _, err := svc.RescheduleCashRate(ctx, f.userID, change.ID, RescheduleCashRateInput{EffectiveFrom: back}); !errors.Is(err, ErrCashRateInUse) {
			t.Fatalf("without recompute = %v, want ErrCashRateInUse", err)
		}

		if _, err := svc.RescheduleCashRate(ctx, f.userID, change.ID, RescheduleCashRateInput{EffectiveFrom: back, Recompute: true}); err != nil {
			t.Fatalf("RescheduleCashRate with recompute: %v", err)
		}

		if n := f.daysAt(t, deposit.EntryID, base.ID); n != 7 {
			t.Errorf("days kept at the version before = %d, want 7", n)
		}
		if n := f.daysAt(t, deposit.EntryID, change.ID); n != 3 {
			t.Errorf("days at the moved version = %d, want 3", n)
		}

		lastBase := back.AddDate(0, 0, -1)
		wantEndedOn(t, "the version before", f.rateByID(t, base.ID), &lastBase)
	})

	t.Run("not once it earned interest", func(t *testing.T) {
		_, err := svc.RescheduleCashRate(ctx, f.userID, change.ID, RescheduleCashRateInput{EffectiveFrom: cashToday().AddDate(0, 0, 5)})
		if !errors.Is(err, ErrCashRateInUse) {
			t.Errorf("err = %v, want ErrCashRateInUse", err)
		}
	})
}

// A paused version keeps its pause: it cannot be moved past the day it stops.
func TestRescheduleCashRateKeepsThePause(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()
	svc := f.accrualService()
	from := cashToday().AddDate(0, 0, 10)

	rate, err := svc.CreateCashRate(ctx, f.userID, f.usdRateInput(t, "9", from, false))
	if err != nil {
		t.Fatalf("CreateCashRate: %v", err)
	}
	if _, err := svc.EndCashRate(ctx, f.userID, rate.ID, from.AddDate(0, 0, 5)); err != nil {
		t.Fatalf("EndCashRate: %v", err)
	}

	_, err = svc.RescheduleCashRate(ctx, f.userID, rate.ID, RescheduleCashRateInput{EffectiveFrom: from.AddDate(0, 0, 6)})
	checkCashRateError(t, err, "cannot start later")

	if _, err := svc.RescheduleCashRate(ctx, f.userID, rate.ID, RescheduleCashRateInput{EffectiveFrom: from.AddDate(0, 0, 2)}); err != nil {
		t.Errorf("moving it inside its days: %v", err)
	}
}

// A rate recorded from the wrong day after its days were computed: the money was
// there, and earning, days before. Moving it back with recompute computes the
// days it gains; moving it forward throws away the ones it gives up.
func TestRescheduleCashRateMovesAComputedVersion(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()
	svc := f.accrualService()
	opened := cashToday().AddDate(0, 0, -10)
	recorded := cashToday().AddDate(0, 0, -5)

	deposit := f.mustMove(t, CashKindDeposit, "10000", opened)

	rate, err := svc.CreateCashRate(ctx, f.userID, f.usdRateInput(t, "9", recorded, false))
	if err != nil {
		t.Fatalf("CreateCashRate: %v", err)
	}
	if n := f.daysAt(t, deposit.EntryID, rate.ID); n != 5 {
		t.Fatalf("days at the rate = %d, want the five from its first day", n)
	}

	if _, err := svc.RescheduleCashRate(ctx, f.userID, rate.ID, RescheduleCashRateInput{EffectiveFrom: opened}); !errors.Is(err, ErrCashRateInUse) {
		t.Fatalf("without recompute = %v, want ErrCashRateInUse", err)
	}

	moved, err := svc.RescheduleCashRate(ctx, f.userID, rate.ID, RescheduleCashRateInput{EffectiveFrom: opened, Recompute: true})
	if err != nil {
		t.Fatalf("moving it back: %v", err)
	}
	if !moved.EffectiveFrom.Equal(opened) {
		t.Errorf("effectiveFrom = %v, want %v", moved.EffectiveFrom, opened)
	}
	if n := f.daysAt(t, deposit.EntryID, rate.ID); n != 10 {
		t.Errorf("days after moving it back = %d, want the ten since the deposit", n)
	}

	later := cashToday().AddDate(0, 0, -3)
	if _, err := svc.RescheduleCashRate(ctx, f.userID, rate.ID, RescheduleCashRateInput{EffectiveFrom: later, Recompute: true}); err != nil {
		t.Fatalf("moving it forward: %v", err)
	}
	if n := f.daysAt(t, deposit.EntryID, rate.ID); n != 3 {
		t.Errorf("days after moving it forward = %d, want the three from its new first day", n)
	}
	if n := f.countInterest(t, deposit.EntryID); n != 3 {
		t.Errorf("credits = %d, want one per day left", n)
	}
}

// A recalculation redoes the days the account had computed and stops there: the
// day after is the nightly job's, even when it is already over by the server's
// clock. What it reports sets the days as they were against the days as they
// are now, apart from any day computed for the first time.
func TestRecalculateCashInterestStopsAtTheComputedDays(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()
	svc := f.accrualService()
	start := cashToday().AddDate(0, 0, -5)
	computed := cashToday().AddDate(0, 0, -2)

	deposit := f.mustMove(t, CashKindDeposit, "1000", start)
	f.openedOn(t, deposit.EntryID, start)
	f.usdRate(t, "9", start)

	if _, errs := svc.AccrueCashInterest(ctx, computed); len(errs) > 0 {
		t.Fatalf("AccrueCashInterest: %v", errs)
	}

	if _, err := f.move(t, CashKindDeposit, "9000", "", start); err != nil {
		t.Fatalf("the backdated deposit: %v", err)
	}

	done, err := svc.RecalculateCashInterest(ctx, f.userID, RecalculateCashInterestInput{
		SourceID: f.sourceID,
		Currency: money.USD,
		From:     start,
	})
	if err != nil {
		t.Fatalf("RecalculateCashInterest: %v", err)
	}

	if !done.Through.Equal(computed) {
		t.Errorf("through = %s, want the last day computed before, %s", done.Through.Format(time.DateOnly), computed.Format(time.DateOnly))
	}
	if n := f.countInterest(t, deposit.EntryID); n != 4 {
		t.Errorf("credits = %d, want the four days that were computed, not yesterday", n)
	}

	if done.Before.Days != 4 || done.After.Days != 4 || done.New.Days != 0 {
		t.Fatalf("before %d, after %d, new %d days; want 4, 4 and 0", done.Before.Days, done.After.Days, done.New.Days)
	}
	if done.Before.From == nil || !done.Before.From.Equal(start) || done.Before.Through == nil || !done.Before.Through.Equal(computed) {
		t.Errorf("before runs %v to %v, want %s to %s", done.Before.From, done.Before.Through, start.Format(time.DateOnly), computed.Format(time.DateOnly))
	}
	sameAmount(t, "new", done.New.Net, "0")

	// Four days on 1 000 against four days on 10 000.
	before, after := mustDecimal(t, done.Before.Net), mustDecimal(t, done.After.Net)
	ratio, err := after.Div(before)
	if err != nil {
		t.Fatalf("ratio: %v", err)
	}
	aboutAmount(t, "after / before", ratio.String(), "10")
}
