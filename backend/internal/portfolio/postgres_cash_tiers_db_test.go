package portfolio

import (
	"context"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

// Tiers, migration 000046. Same database contract as postgres_cash_db_test.go.

// accrualRate is the rate the ledger kept for a balance's day.
func (f cashFixture) accrualRate(t *testing.T, entryID uuid.UUID, day time.Time) string {
	t.Helper()

	var rate string
	if err := f.pool.QueryRow(context.Background(), `
		SELECT annual_rate::text FROM cash_interest_accruals
		WHERE entry_id = $1 AND accrual_date = $2::date
	`, entryID, day.Format(time.DateOnly)).Scan(&rate); err != nil {
		t.Fatalf("read the day's rate: %v", err)
	}

	return rate
}

// wantTiers compares a rate's tiers with pairs of fromBalance and annualRatePct.
func wantTiers(t *testing.T, label string, got []CashRateTier, pairs ...string) {
	t.Helper()

	if len(got) != len(pairs)/2 {
		t.Fatalf("%s tiers = %+v, want %d of them", label, got, len(pairs)/2)
	}

	for i, tier := range got {
		if tier.FromBalance != pairs[2*i] || tier.AnnualRatePct != pairs[2*i+1] {
			t.Errorf("%s tier %d = %+v, want %s %% from %s", label, i, tier, pairs[2*i+1], pairs[2*i])
		}
	}
}

// 12 % up to 5 000 and 8 % above it: 5 000 × ((1.12)^(1/365) − 1) + 3 000 ×
// ((1.08)^(1/365) − 1) = 2.18531197 on 8 000, a day kept at the 10.483 % that
// comes to.
func TestCashAccrualEarnsInTiers(t *testing.T) {
	f := newCashFixture(t)

	deposit := f.mustMove(t, CashKindDeposit, "8000", cashDay)
	rate := f.rateWith(t, "12", cashDay, func(in *CashRateInput) {
		in.Tiers = tierSteps(t, "5000", "8")
	})

	f.accrue(t, deposit.EntryID, rate.ID, cashDay)

	row, _ := f.accrualOn(t, deposit.EntryID, cashDay)
	sameAmount(t, "basis", row.basis, "8000")
	sameAmount(t, "net", row.net, "2.18531197")
	sameAmount(t, "the rate of the day", f.accrualRate(t, deposit.EntryID, cashDay), "0.10483")
	sameAmount(t, "balance", f.balance(t), "8002.19")
}

// The tiers belong to the account: two portfolios holding 6 000 and 2 000 of it
// earn the account's day in proportion, 1.638984 and 0.546328, rather than each
// a day of its own that would put the smaller one wholly in the first step.
func TestCashAccrualSharesTiersBetweenBalances(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	other := uuid.New()
	f.exec(t, `INSERT INTO portfolios (id, user_id, name, type, risk_id, base_currency)
	           VALUES ($1, $2, 'reserve', 'cash', (SELECT id FROM risks LIMIT 1), 'USD')`, other, f.userID)

	first := f.mustMove(t, CashKindDeposit, "6000", cashDay)

	second, err := f.repo.CreateCashMovement(ctx, f.userID, other, f.sourceID, uuid.UUID{}, CashMovementInput{
		Kind: CashKindDeposit, Amount: mustDecimal(t, "2000"), Currency: money.USD, Date: cashDay,
	})
	if err != nil {
		t.Fatalf("CreateCashMovement on the second portfolio: %v", err)
	}

	rate := f.rateWith(t, "12", cashDay, func(in *CashRateInput) {
		in.Tiers = tierSteps(t, "5000", "8")
	})

	f.accrue(t, first.EntryID, rate.ID, cashDay)
	f.accrue(t, second.EntryID, rate.ID, cashDay)

	sameAmount(t, "the larger balance", f.balanceOf(t, first.EntryID).Balance, "6001.64")
	sameAmount(t, "the smaller balance", f.balanceOf(t, second.EntryID).Balance, "2000.55")

	// Both days are kept at the account's rate.
	sameAmount(t, "the rate of the larger", f.accrualRate(t, first.EntryID, cashDay), "0.10483")
	sameAmount(t, "the rate of the smaller", f.accrualRate(t, second.EntryID, cashDay), "0.10483")
}

// A version's tiers come back as they were stated, in order, and a correction
// states them whole: the steps it leaves out are gone.
func TestCashRateTiersAreStatedWhole(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	rate := f.rateWith(t, "12", cashDay, func(in *CashRateInput) {
		in.Tiers = tierSteps(t, "5000", "8", "20000.5", "0")
	})
	wantTiers(t, "created", rate.Tiers, "5000", "8", "20000.5", "0")

	rates, err := f.repo.GetCashRatesByUserID(ctx, f.userID)
	if err != nil || len(rates) != 1 {
		t.Fatalf("GetCashRatesByUserID = %+v, %v", rates, err)
	}
	wantTiers(t, "listed", rates[0].Tiers, "5000", "8", "20000.5", "0")

	values := CashRateInput{
		AnnualRatePct:  mustDecimal(t, "12"),
		WithholdingPct: mustDecimal(t, "0"),
		Posting:        PostingDaily,
		Tiers:          tierSteps(t, "7000", "7.25"),
	}

	corrected, err := f.repo.UpdateCashRate(ctx, f.userID, rate.ID, values)
	if err != nil {
		t.Fatalf("UpdateCashRate: %v", err)
	}
	wantTiers(t, "corrected", corrected.Tiers, "7000", "7.25")

	values.Tiers = nil

	cleared, err := f.repo.UpdateCashRate(ctx, f.userID, rate.ID, values)
	if err != nil {
		t.Fatalf("UpdateCashRate without tiers: %v", err)
	}
	if cleared.Tiers == nil || len(cleared.Tiers) != 0 {
		t.Errorf("tiers after a correction without them = %#v, want an empty list", cleared.Tiers)
	}
}
