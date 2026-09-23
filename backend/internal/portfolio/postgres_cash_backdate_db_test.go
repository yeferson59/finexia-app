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
// The service computes through yesterday by the real clock, so these days are
// counted back from today rather than from cashDay.

func cashToday() time.Time {
	return snapshotDay(time.Now())
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
