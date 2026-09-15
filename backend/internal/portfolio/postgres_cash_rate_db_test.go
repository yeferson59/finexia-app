package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

// The versions of an account's rate, as migration 000043 and CashRateStore keep
// them. Same database contract as postgres_cash_db_test.go. The dates are fixed:
// the store takes them as given, and the grace around today is the service's.

func (f cashFixture) rate(t *testing.T, cur money.Currency, pct string, from time.Time) (CashRate, error) {
	t.Helper()

	return f.repo.CreateCashRate(context.Background(), f.userID, NewCashRateInput{
		SourceID:      f.sourceID,
		Currency:      cur,
		EffectiveFrom: from,
		CashRateInput: CashRateInput{
			AnnualRatePct:  mustDecimal(t, pct),
			WithholdingPct: mustDecimal(t, "0"),
			Posting:        PostingDaily,
		},
	})
}

func (f cashFixture) mustRate(t *testing.T, pct string, from time.Time) CashRate {
	t.Helper()

	rate, err := f.rate(t, money.COP, pct, from)
	if err != nil {
		t.Fatalf("CreateCashRate(%s from %s): %v", pct, from.Format(time.DateOnly), err)
	}

	return rate
}

func (f cashFixture) rates(t *testing.T) []CashRate {
	t.Helper()

	rates, err := f.repo.GetCashRatesByUserID(context.Background(), f.userID)
	if err != nil {
		t.Fatalf("GetCashRatesByUserID: %v", err)
	}

	return rates
}

func wantEndedOn(t *testing.T, label string, rate CashRate, want *time.Time) {
	t.Helper()

	switch {
	case want == nil && rate.EndedOn != nil:
		t.Errorf("%s ended on %s, want no end", label, rate.EndedOn.Format(time.DateOnly))
	case want != nil && (rate.EndedOn == nil || !rate.EndedOn.Equal(*want)):
		t.Errorf("%s ended on %v, want %s", label, rate.EndedOn, want.Format(time.DateOnly))
	}
}

func TestCashRateNewVersionEndsTheOneBefore(t *testing.T) {
	f := newCashFixture(t)

	first := f.mustRate(t, "9.25", cashDay)
	if first.AnnualRatePct != "9.25" || first.WithholdingPct != "0" || first.Posting != PostingDaily {
		t.Errorf("first = %+v, want 9.25 %% with nothing withheld, posted daily", first)
	}
	if first.SourceName != "bank" || first.Currency != money.COP || !first.Latest || !first.EffectiveFrom.Equal(cashDay) {
		t.Errorf("first = %+v, want the latest COP rate on bank from %s", first, cashDay.Format(time.DateOnly))
	}
	wantEndedOn(t, "first", first, nil)

	change := cashDay.AddDate(0, 0, 19)
	second := f.mustRate(t, "8.75", change)

	rates := f.rates(t)
	if len(rates) != 2 || rates[0].ID != second.ID || rates[1].ID != first.ID {
		t.Fatalf("rates = %+v, want the second version before the first", rates)
	}

	dayBefore := change.AddDate(0, 0, -1)
	wantEndedOn(t, "first", rates[1], &dayBefore)
	wantEndedOn(t, "second", rates[0], nil)
	if rates[1].Latest || !rates[0].Latest {
		t.Errorf("latest = %v, %v; want only the second", rates[1].Latest, rates[0].Latest)
	}
}

// A version cannot be slotted in on or before the latest one's first day: some
// day would have two rates, or what the latest said would change.
func TestCashRateCannotStartBeforeTheLatest(t *testing.T) {
	f := newCashFixture(t)
	f.mustRate(t, "9", cashDay.AddDate(0, 0, 9))

	for _, from := range []time.Time{cashDay.AddDate(0, 0, 9), cashDay.AddDate(0, 0, 4)} {
		if _, err := f.rate(t, money.COP, "8", from); !errors.Is(err, ErrCashRateOverlaps) {
			t.Errorf("a version from %s = %v, want ErrCashRateOverlaps", from.Format(time.DateOnly), err)
		}
	}

	// The same day is free in another currency: it is another account.
	if _, err := f.rate(t, money.USD, "4.5", cashDay.AddDate(0, 0, 9)); err != nil {
		t.Errorf("a USD rate on the same platform and day: %v", err)
	}
}

func TestCashRateOnlyTheLatestVersionChanges(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	first := f.mustRate(t, "9.25", cashDay)
	second := f.mustRate(t, "8.75", cashDay.AddDate(0, 0, 19))

	values := CashRateInput{AnnualRatePct: mustDecimal(t, "8.5"), WithholdingPct: mustDecimal(t, "7"), Posting: PostingDaily}

	if _, err := f.repo.UpdateCashRate(ctx, f.userID, first.ID, values); !errors.Is(err, ErrCashRateNotLatest) {
		t.Errorf("correcting the first version = %v, want ErrCashRateNotLatest", err)
	}
	if _, err := f.repo.EndCashRate(ctx, f.userID, first.ID, cashDay.AddDate(0, 0, 5)); !errors.Is(err, ErrCashRateNotLatest) {
		t.Errorf("ending the first version = %v, want ErrCashRateNotLatest", err)
	}
	if err := f.repo.DeleteCashRate(ctx, f.userID, first.ID); !errors.Is(err, ErrCashRateNotLatest) {
		t.Errorf("deleting the first version = %v, want ErrCashRateNotLatest", err)
	}

	corrected, err := f.repo.UpdateCashRate(ctx, f.userID, second.ID, values)
	if err != nil {
		t.Fatalf("UpdateCashRate: %v", err)
	}
	if corrected.AnnualRatePct != "8.5" || corrected.WithholdingPct != "7" || !corrected.EffectiveFrom.Equal(second.EffectiveFrom) {
		t.Errorf("corrected = %+v, want 8.5 %% with 7 %% withheld, from the same day", corrected)
	}
}

func TestCashRatePauseAndResume(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	rate := f.mustRate(t, "9", cashDay)

	// A rate that has not started on the day it would stop is deleted, not paused.
	if _, err := f.repo.EndCashRate(ctx, f.userID, rate.ID, cashDay); !errors.Is(err, ErrInvalidCashRate) {
		t.Errorf("pausing from its first day = %v, want ErrInvalidCashRate", err)
	}

	paused, err := f.repo.EndCashRate(ctx, f.userID, rate.ID, cashDay.AddDate(0, 0, 14))
	if err != nil {
		t.Fatalf("EndCashRate: %v", err)
	}
	lastDay := cashDay.AddDate(0, 0, 13)
	wantEndedOn(t, "paused", paused, &lastDay)

	// Resuming later is a new version, and leaves the pause where it was.
	f.mustRate(t, "8.5", cashDay.AddDate(0, 0, 19))

	rates := f.rates(t)
	if len(rates) != 2 {
		t.Fatalf("rates = %+v, want two versions", rates)
	}
	wantEndedOn(t, "paused", rates[1], &lastDay)
	wantEndedOn(t, "resumed", rates[0], nil)
}

// Deleting a change of rate puts the account back on the rate it replaced.
func TestCashRateDeletingAChangeRestoresTheRateBefore(t *testing.T) {
	f := newCashFixture(t)

	first := f.mustRate(t, "9.25", cashDay)
	second := f.mustRate(t, "12", cashDay.AddDate(0, 0, 19))

	if err := f.repo.DeleteCashRate(context.Background(), f.userID, second.ID); err != nil {
		t.Fatalf("DeleteCashRate: %v", err)
	}

	rates := f.rates(t)
	if len(rates) != 1 || rates[0].ID != first.ID || !rates[0].Latest {
		t.Fatalf("rates = %+v, want the first version back as the latest", rates)
	}
	wantEndedOn(t, "first", rates[0], nil)
}

func TestCashRateBelongsToItsPlatformOwner(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()
	stranger := uuid.New()

	rate := f.mustRate(t, "9", cashDay)
	values := CashRateInput{AnnualRatePct: mustDecimal(t, "20"), WithholdingPct: mustDecimal(t, "0"), Posting: PostingDaily}

	if _, err := f.repo.CreateCashRate(ctx, stranger, NewCashRateInput{
		SourceID: f.sourceID, Currency: money.USD, EffectiveFrom: cashDay, CashRateInput: values,
	}); !errors.Is(err, ErrPlatformNotFound) {
		t.Errorf("a rate on someone else's platform = %v, want ErrPlatformNotFound", err)
	}
	if _, err := f.repo.UpdateCashRate(ctx, stranger, rate.ID, values); !errors.Is(err, ErrCashRateNotFound) {
		t.Errorf("correcting someone else's rate = %v, want ErrCashRateNotFound", err)
	}
	if _, err := f.repo.EndCashRate(ctx, stranger, rate.ID, cashDay.AddDate(0, 0, 5)); !errors.Is(err, ErrCashRateNotFound) {
		t.Errorf("ending someone else's rate = %v, want ErrCashRateNotFound", err)
	}
	if err := f.repo.DeleteCashRate(ctx, stranger, rate.ID); !errors.Is(err, ErrCashRateNotFound) {
		t.Errorf("deleting someone else's rate = %v, want ErrCashRateNotFound", err)
	}

	others, err := f.repo.GetCashRatesByUserID(ctx, stranger)
	if err != nil || len(others) != 0 {
		t.Errorf("a stranger's rates = %+v, %v; want none", others, err)
	}
}

func TestCashRateOnAnInactivePlatformIsRefused(t *testing.T) {
	f := newCashFixture(t)
	f.exec(t, `UPDATE investment_sources SET is_active = FALSE WHERE id = $1`, f.sourceID)

	if _, err := f.rate(t, money.COP, "9", cashDay); !errors.Is(err, ErrInvalidCashRate) {
		t.Errorf("a rate on an inactive platform = %v, want ErrInvalidCashRate", err)
	}
}
