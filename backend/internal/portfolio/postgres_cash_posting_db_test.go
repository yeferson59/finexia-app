package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

// Interest posted monthly and a cap on what earns — migration 000045. Same
// database contract as postgres_cash_db_test.go.

// rateWith records a USD rate on the fixture's platform, with the posting and
// the tiers the case is about. A cap is a tier at 0 % (000046).
func (f cashFixture) rateWith(t *testing.T, pct string, from time.Time, edit func(*CashRateInput)) CashRate {
	t.Helper()

	in := CashRateInput{
		AnnualRatePct:  mustDecimal(t, pct),
		WithholdingPct: mustDecimal(t, "0"),
		Posting:        PostingDaily,
	}
	edit(&in)

	rate, err := f.repo.CreateCashRate(context.Background(), f.userID, NewCashRateInput{
		SourceID:      f.sourceID,
		Currency:      money.USD,
		EffectiveFrom: from,
		CashRateInput: in,
	})
	if err != nil {
		t.Fatalf("CreateCashRate(%s from %s): %v", pct, from.Format(time.DateOnly), err)
	}

	return rate
}

func monthly(in *CashRateInput) { in.Posting = PostingMonthly }

// accrueThrough computes every day from cashDay through the given one.
func (f cashFixture) accrueThrough(t *testing.T, entryID, rateID uuid.UUID, last time.Time) {
	t.Helper()

	for day := cashDay; !day.After(last); day = day.AddDate(0, 0, 1) {
		f.accrue(t, entryID, rateID, day)
	}
}

// balanceOf is what one of several balances holds.
func (f cashFixture) balanceOf(t *testing.T, entryID uuid.UUID) CashBalance {
	t.Helper()

	balances, err := f.repo.GetCashBalancesByUserID(context.Background(), f.userID, money.XXX)
	if err != nil {
		t.Fatalf("GetCashBalancesByUserID: %v", err)
	}

	for _, b := range balances {
		if b.EntryID == entryID {
			return b
		}
	}

	t.Fatalf("balances = %+v, want one on %v", balances, entryID)

	return CashBalance{}
}

// A month posted monthly holds every day and credits them on the last one, in
// one transaction. It earns the same as a month posted daily: a held day
// compounds on the days held before it, whether or not they are credited yet.
func TestCashAccrualPostsAMonthOnItsLastDay(t *testing.T) {
	f := newCashFixture(t)

	deposit := f.mustMove(t, CashKindDeposit, "10000", cashDay)
	rate := f.rateWith(t, "9", cashDay, monthly)

	lastOfMonth := time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)
	f.accrueThrough(t, deposit.EntryID, rate.ID, lastOfMonth.AddDate(0, 0, -1))

	// Twenty-nine days computed, nothing credited.
	if n := f.countInterest(t, deposit.EntryID); n != 0 {
		t.Errorf("credits = %d, want none before the month closes", n)
	}
	sameAmount(t, "balance", f.balance(t), "10000")

	// 10 000 × ((1.09)^(29/365) − 1) = 68,704893…, because a held day compounds
	// into the next one just as a credited one would.
	held := f.balanceOf(t, deposit.EntryID)
	aboutAmount(t, "pending interest", held.PendingInterest, "68.7049")

	row, ok := f.accrualOn(t, deposit.EntryID, cashDay.AddDate(0, 0, 1))
	if !ok || row.status != "pending" || row.credited {
		t.Errorf("the second day = %+v (found %v), want pending", row, ok)
	}
	// The second day earned on the first day's interest, held and not credited.
	sameAmount(t, "second basis", row.basis, "10002.36131152")

	if !f.accrue(t, deposit.EntryID, rate.ID, lastOfMonth) {
		t.Fatal("the last day of the month credited nothing")
	}

	// 10 000 × ((1.09)^(30/365) − 1) = 71.0827…, as a month posted daily earns.
	if n := f.countInterest(t, deposit.EntryID); n != 1 {
		t.Errorf("credits = %d, want one for the whole month", n)
	}
	sameAmount(t, "balance", f.balance(t), "10071.08")

	after := f.balanceOf(t, deposit.EntryID)
	sameAmount(t, "pending interest", after.PendingInterest, "0")
	sameAmount(t, "interest earned", after.InterestEarned, "71.08")

	// The credit pays every day of the month, so each one names it.
	var unpaid int
	if err := f.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM cash_interest_accruals
		WHERE entry_id = $1 AND (status <> 'posted' OR transaction_id IS NULL)
	`, deposit.EntryID).Scan(&unpaid); err != nil {
		t.Fatalf("count unpaid: %v", err)
	}
	if unpaid != 0 {
		t.Errorf("%d days of the month were left unpaid", unpaid)
	}
}

// A rate paused partway through a month leaves days that no month end will ever
// close. The job credits them on the last day they were earned.
func TestCashAccrualCreditsWhatAPausedRateHolds(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	deposit := f.mustMove(t, CashKindDeposit, "10000", cashDay)
	rate := f.rateWith(t, "9", cashDay, monthly)

	lastEarned := cashDay.AddDate(0, 0, 4)
	f.accrueThrough(t, deposit.EntryID, rate.ID, lastEarned)

	// While the rate runs on, the days wait for the end of the month.
	if held, err := f.repo.GetHeldCashInterest(ctx, lastEarned.AddDate(0, 0, 3), CashAccrualFilter{}); err != nil || len(held) != 0 {
		t.Fatalf("held = %v, %v; want none while the rate runs on", held, err)
	}

	if _, err := f.repo.EndCashRate(ctx, f.userID, rate.ID, lastEarned.AddDate(0, 0, 1)); err != nil {
		t.Fatalf("EndCashRate: %v", err)
	}

	if _, errs := f.accrualService().AccrueCashInterest(ctx, lastEarned.AddDate(0, 0, 3)); len(errs) > 0 {
		t.Fatalf("AccrueCashInterest: %v", errs)
	}

	// 10 000 × ((1.09)^(5/365) − 1) = 11.8112…
	if n := f.countInterest(t, deposit.EntryID); n != 1 {
		t.Errorf("credits = %d, want one for the days it held", n)
	}
	sameAmount(t, "balance", f.balance(t), "10011.81")

	after := f.balanceOf(t, deposit.EntryID)
	sameAmount(t, "pending interest", after.PendingInterest, "0")

	var dated time.Time
	if err := f.pool.QueryRow(ctx, `
		SELECT transaction_date FROM transactions WHERE entry_id = $1 AND type = 'cash_interest'
	`, deposit.EntryID).Scan(&dated); err != nil {
		t.Fatalf("read the credit: %v", err)
	}
	if !dated.Equal(lastEarned) {
		t.Errorf("credit dated %s, want the last day it earned, %s", dated.Format(time.DateOnly), lastEarned.Format(time.DateOnly))
	}

	// Nothing is held any more, so a second run credits nothing.
	if _, errs := f.accrualService().AccrueCashInterest(ctx, lastEarned.AddDate(0, 0, 4)); len(errs) > 0 {
		t.Fatalf("AccrueCashInterest again: %v", errs)
	}
	if n := f.countInterest(t, deposit.EntryID); n != 1 {
		t.Errorf("credits = %d after a second run, want 1", n)
	}
}

// A cap belongs to the account, so two portfolios holding the same account's
// money earn on their share of it.
func TestCashAccrualSharesTheCapBetweenBalances(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	other := uuid.New()
	f.exec(t, `INSERT INTO portfolios (id, user_id, name, type, risk_id, base_currency)
	           VALUES ($1, $2, 'reserve', 'cash', (SELECT id FROM risks LIMIT 1), 'USD')`, other, f.userID)

	first := f.mustMove(t, CashKindDeposit, "7500", cashDay)

	second, err := f.repo.CreateCashMovement(ctx, f.userID, other, f.sourceID, uuid.UUID{}, CashMovementInput{
		Kind: CashKindDeposit, Amount: mustDecimal(t, "2500"), Currency: money.USD, Date: cashDay,
	})
	if err != nil {
		t.Fatalf("CreateCashMovement on the second portfolio: %v", err)
	}

	rate := f.rateWith(t, "9", cashDay, func(in *CashRateInput) {
		in.Tiers = tierSteps(t, "5000", "0")
	})

	f.accrue(t, first.EntryID, rate.ID, cashDay)
	f.accrue(t, second.EntryID, rate.ID, cashDay)

	// The account holds 10 000 against a cap of 5 000, so each balance earns
	// what half of it would, 3 750 and 1 250, never more than the cap. The
	// ledger keeps what each held, and the rate that came to: 4.4033 %.
	firstRow, _ := f.accrualOn(t, first.EntryID, cashDay)
	secondRow, _ := f.accrualOn(t, second.EntryID, cashDay)
	sameAmount(t, "the larger basis", firstRow.basis, "7500")
	sameAmount(t, "the smaller basis", secondRow.basis, "2500")
	sameAmount(t, "the rate of the day", f.accrualRate(t, first.EntryID, cashDay), "0.044033")

	// 3 750 × 0.000236131 = 0.8855, 1 250 × 0.000236131 = 0.2952.
	sameAmount(t, "the larger balance", f.balanceOf(t, first.EntryID).Balance, "7500.89")
	sameAmount(t, "the smaller balance", f.balanceOf(t, second.EntryID).Balance, "2500.3")
}

// Under the cap nothing changes: the account earns on all of it.
func TestCashAccrualUnderTheCapEarnsOnEverything(t *testing.T) {
	f := newCashFixture(t)

	deposit := f.mustMove(t, CashKindDeposit, "10000", cashDay)
	rate := f.rateWith(t, "9", cashDay, func(in *CashRateInput) {
		in.Tiers = tierSteps(t, "25000", "0")
	})

	f.accrue(t, deposit.EntryID, rate.ID, cashDay)

	sameAmount(t, "balance", f.balance(t), "10002.36")
}

// creditedSum is every interest the ledger credited to a balance.
func (f cashFixture) creditedSum(t *testing.T, entryID uuid.UUID) string {
	t.Helper()

	var total string
	if err := f.pool.QueryRow(context.Background(), `
		SELECT COALESCE(SUM(t.quantity), 0)::text FROM transactions t
		WHERE t.entry_id = $1 AND t.type = 'cash_interest'
	`, entryID).Scan(&total); err != nil {
		t.Fatalf("sum credits: %v", err)
	}

	return total
}

// A deposit recorded with a past date after those days were computed leaves
// every day after it earning on less than the balance held. Recalculating from
// that day throws the days away and computes them again on what the balance
// holds now.
func TestRecalculateCashInterestRedoesTheDays(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	today := time.Now().UTC().Truncate(24 * time.Hour)
	start := today.AddDate(0, 0, -3)
	yesterday := today.AddDate(0, 0, -1)

	deposit := f.mustMove(t, CashKindDeposit, "1000", start)
	f.openedOn(t, deposit.EntryID, start)
	f.rateWith(t, "9", start, func(*CashRateInput) {})

	if _, errs := f.accrualService().AccrueCashInterest(ctx, yesterday); len(errs) > 0 {
		t.Fatalf("AccrueCashInterest: %v", errs)
	}
	if n := f.countInterest(t, deposit.EntryID); n != 3 {
		t.Fatalf("credits = %d, want one per day", n)
	}

	first, _ := f.accrualOn(t, deposit.EntryID, start)
	sameAmount(t, "the first basis", first.basis, "1000")

	// The 9 000 that was there all along, recorded now.
	if _, err := f.move(t, CashKindDeposit, "9000", "", start); err != nil {
		t.Fatalf("the backdated deposit: %v", err)
	}

	done, err := f.accrualService().RecalculateCashInterest(ctx, f.userID, RecalculateCashInterestInput{
		SourceID: f.sourceID,
		Currency: money.USD,
		From:     start,
	})
	if err != nil {
		t.Fatalf("RecalculateCashInterest: %v", err)
	}

	if done.Cleared.Days != 3 || done.Cleared.Balances != 1 || !done.Cleared.From.Equal(start) {
		t.Errorf("cleared = %+v, want the three days of one balance from %s", done.Cleared, start.Format(time.DateOnly))
	}
	if done.Recomputed != 3 || done.Credited != 3 {
		t.Errorf("recomputed %d days into %d credits, want 3 and 3", done.Recomputed, done.Credited)
	}

	if n := f.countInterest(t, deposit.EntryID); n != 3 {
		t.Errorf("credits = %d after recalculating, want one per day", n)
	}

	redone, _ := f.accrualOn(t, deposit.EntryID, start)
	sameAmount(t, "the first basis, redone", redone.basis, "10000")

	// The balance is the deposits plus exactly what the ledger credited.
	credited := f.creditedSum(t, deposit.EntryID)
	sameAmount(t, "balance", f.balance(t), mustDecimal(t, "10000").Add(mustDecimal(t, credited)).String())

	// Three days on 10 000 at 9 % E.A. is about 7.08, not the 0.71 the first run
	// credited on 1 000.
	aboutAmount(t, "credited", credited, "7.09")

	// Nothing is owed any more: a run right after finds no day to compute.
	if _, errs := f.accrualService().AccrueCashInterest(ctx, yesterday); len(errs) > 0 {
		t.Fatalf("AccrueCashInterest after recalculating: %v", errs)
	}
	if n := f.countInterest(t, deposit.EntryID); n != 3 {
		t.Errorf("credits = %d after a run, want 3", n)
	}
}

// A month's credit pays every day of its month, so half of one cannot be
// undone: clearing a day inside it clears the month from its first day.
func TestRecalculateCashInterestClearsAWholeMonthlyCredit(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	deposit := f.mustMove(t, CashKindDeposit, "10000", cashDay)
	rate := f.rateWith(t, "9", cashDay, monthly)

	lastOfMonth := time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)
	f.accrueThrough(t, deposit.EntryID, rate.ID, lastOfMonth)

	if n := f.countInterest(t, deposit.EntryID); n != 1 {
		t.Fatalf("credits = %d, want one for the month", n)
	}

	cleared, err := f.repo.ClearCashInterest(ctx, CashAccrualFilter{
		UserID: f.userID, SourceID: f.sourceID, Currency: money.USD,
	}, lastOfMonth.AddDate(0, 0, -5))
	if err != nil {
		t.Fatalf("ClearCashInterest: %v", err)
	}

	if !cleared.From.Equal(cashDay) || cleared.Days != 30 {
		t.Errorf("cleared = %+v, want the thirty days of the month from %s", cleared, cashDay.Format(time.DateOnly))
	}
	if n := f.countInterest(t, deposit.EntryID); n != 0 {
		t.Errorf("credits = %d, want the month's credit gone with its days", n)
	}
	sameAmount(t, "balance", f.balance(t), "10000")
}

// Everything above lives behind a rate. Without one, an account behaves exactly
// as it did before the interest ledger existed: the job walks past it and
// writes nothing.
func TestCashWithoutARateIsUntouched(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	deposit := f.mustMove(t, CashKindDeposit, "10000", cashDay)
	f.openedOn(t, deposit.EntryID, cashDay)

	// Interest recorded by hand stays the owner's, not the ledger's.
	f.mustMove(t, CashKindInterest, "25", cashDay.AddDate(0, 0, 1))

	before := f.balanceOf(t, deposit.EntryID)

	if _, errs := f.accrualService().AccrueCashInterest(ctx, cashDay.AddDate(0, 0, 30)); len(errs) > 0 {
		t.Fatalf("AccrueCashInterest: %v", errs)
	}

	after := f.balanceOf(t, deposit.EntryID)

	sameAmount(t, "balance", after.Balance, before.Balance)
	sameAmount(t, "balance", after.Balance, "10025")
	sameAmount(t, "interest earned", after.InterestEarned, "25")
	sameAmount(t, "pending interest", after.PendingInterest, "0")
	if after.LastAccrualDate != nil {
		t.Errorf("last accrual = %v, want none without a rate", after.LastAccrualDate)
	}

	var ledger int
	if err := f.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM cash_interest_accruals WHERE entry_id = $1`, deposit.EntryID,
	).Scan(&ledger); err != nil {
		t.Fatalf("count the ledger: %v", err)
	}
	if ledger != 0 {
		t.Errorf("ledger rows = %d, want none without a rate", ledger)
	}

	// The one the owner wrote does not read as automatic.
	movements, err := f.repo.GetCashMovementsPaginated(ctx, f.userID, 10, 0)
	if err != nil {
		t.Fatalf("GetCashMovementsPaginated: %v", err)
	}
	for _, m := range movements {
		if m.Automatic {
			t.Errorf("movement %+v reads as automatic", m)
		}
	}
}

// A month that closes while the job is catching up: the last day of the month
// credits it, and the days of the next month start a new one.
func TestCashAccrualClosesAMonthMidCatchUp(t *testing.T) {
	f := newCashFixture(t)

	from := time.Date(2026, time.September, 25, 0, 0, 0, 0, time.UTC)
	deposit := f.mustMove(t, CashKindDeposit, "10000", from)
	rate := f.rateWith(t, "9", from, monthly)

	for day := from; !day.After(from.AddDate(0, 0, 10)); day = day.AddDate(0, 0, 1) {
		f.accrue(t, deposit.EntryID, rate.ID, day)
	}

	// One credit, on 30 September, for the six days of that month.
	if n := f.countInterest(t, deposit.EntryID); n != 1 {
		t.Errorf("credits = %d, want one for September", n)
	}

	closed, ok := f.accrualOn(t, deposit.EntryID, time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC))
	if !ok || closed.status != "posted" || !closed.credited {
		t.Errorf("30 September = %+v (found %v), want the month's credit", closed, ok)
	}

	// October has started over: its days wait for their own month end.
	october, ok := f.accrualOn(t, deposit.EntryID, time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC))
	if !ok || october.status != "pending" || october.credited {
		t.Errorf("1 October = %+v (found %v), want pending", october, ok)
	}

	held := f.balanceOf(t, deposit.EntryID)
	// 10 000 × ((1.09)^(6/365) − 1) = 14.176235…, credited as 14.18.
	sameAmount(t, "balance", held.Balance, "10014.18")
	// Five days of October on 10 014.17, still held.
	if parsed := mustDecimal(t, held.PendingInterest); !parsed.IsPos() {
		t.Errorf("pending interest = %s, want the five days of October", held.PendingInterest)
	}
}

// What a balance holds is one account's business. A recalculation for one owner
// never reaches another's held days.
func TestHeldCashInterestStaysInsideTheFilter(t *testing.T) {
	mine, theirs := newCashFixture(t), newCashFixture(t)
	ctx := context.Background()

	stop := cashDay.AddDate(0, 0, 4)

	for _, f := range []cashFixture{mine, theirs} {
		deposit := f.mustMove(t, CashKindDeposit, "10000", cashDay)
		rate := f.rateWith(t, "9", cashDay, monthly)
		f.accrueThrough(t, deposit.EntryID, rate.ID, stop)

		if _, err := f.repo.EndCashRate(ctx, f.userID, rate.ID, stop.AddDate(0, 0, 1)); err != nil {
			t.Fatalf("EndCashRate: %v", err)
		}
	}

	held, err := mine.repo.GetHeldCashInterest(ctx, stop.AddDate(0, 0, 3), CashAccrualFilter{UserID: mine.userID})
	if err != nil {
		t.Fatalf("GetHeldCashInterest: %v", err)
	}
	if len(held) != 1 {
		t.Fatalf("held = %v, want only the one balance of the owner asked for", held)
	}

	balances, err := mine.repo.GetCashBalancesByUserID(ctx, mine.userID, money.XXX)
	if err != nil || len(balances) != 1 {
		t.Fatalf("GetCashBalancesByUserID = %+v, %v", balances, err)
	}
	if held[0] != balances[0].EntryID {
		t.Errorf("held = %v, want %v", held[0], balances[0].EntryID)
	}

	// And a recalculation of one account leaves the other's days waiting.
	if _, err := mine.accrualService().RecalculateCashInterest(ctx, mine.userID, RecalculateCashInterestInput{
		SourceID: mine.sourceID, Currency: money.USD, From: cashDay,
	}); err != nil {
		t.Fatalf("RecalculateCashInterest: %v", err)
	}

	other := theirs.balanceOf(t, theirs.mustMove(t, CashKindDeposit, "0.00000001", stop).EntryID)
	if !mustDecimal(t, other.PendingInterest).IsPos() {
		t.Errorf("the other owner's pending interest = %s, want it untouched", other.PendingInterest)
	}
}

// A cap and a monthly posting at once: every day earns on the share of the cap,
// and the month still closes in one credit.
func TestCashAccrualCapsAMonthlyRate(t *testing.T) {
	f := newCashFixture(t)

	deposit := f.mustMove(t, CashKindDeposit, "10000", cashDay)
	rate := f.rateWith(t, "9", cashDay, func(in *CashRateInput) {
		monthly(in)
		in.Tiers = tierSteps(t, "4000", "0")
	})

	f.accrueThrough(t, deposit.EntryID, rate.ID, cashDay.AddDate(0, 0, 2))

	// Every day earns on the cap alone. The ledger keeps what the balance held:
	// the deposit on the first day, and on the second the day before as well,
	// which leaves the account over the cap either way.
	first, _ := f.accrualOn(t, deposit.EntryID, cashDay)
	sameAmount(t, "the first basis", first.basis, "10000")

	second, _ := f.accrualOn(t, deposit.EntryID, cashDay.AddDate(0, 0, 1))
	sameAmount(t, "the second basis", second.basis, mustDecimal(t, "10000").Add(mustDecimal(t, first.net)).String())
	sameAmount(t, "the second day", second.net, first.net)

	if n := f.countInterest(t, deposit.EntryID); n != 0 {
		t.Errorf("credits = %d, want none before the month closes", n)
	}

	lastOfMonth := time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)
	f.accrueThrough(t, deposit.EntryID, rate.ID, lastOfMonth)

	// The cap pins the basis: the account is over it every day, so there is no
	// compounding to have. 30 × 4 000 × ((1.09)^(1/365) − 1) = 28.335738…
	sameAmount(t, "balance", f.balance(t), "10028.34")
}

// A balance emptied halfway through its month keeps what it earned: interest
// computed is interest owed, and the month's credit still arrives.
func TestCashAccrualEmptiedMidMonthStillCredits(t *testing.T) {
	f := newCashFixture(t)

	deposit := f.mustMove(t, CashKindDeposit, "10000", cashDay)
	rate := f.rateWith(t, "9", cashDay, monthly)

	f.accrueThrough(t, deposit.EntryID, rate.ID, cashDay.AddDate(0, 0, 9))

	// The whole balance goes; the days it earned are not part of it yet.
	f.mustMove(t, CashKindWithdrawal, "10000", cashDay.AddDate(0, 0, 10))
	sameAmount(t, "balance after the withdrawal", f.balance(t), "0")

	lastOfMonth := time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)
	for day := cashDay.AddDate(0, 0, 10); !day.After(lastOfMonth); day = day.AddDate(0, 0, 1) {
		f.accrue(t, deposit.EntryID, rate.ID, day)
	}

	// Ten days on 10 000 compound to 23.638222…, and the days after the
	// withdrawal earn only on what those ten held.
	if n := f.countInterest(t, deposit.EntryID); n != 1 {
		t.Errorf("credits = %d, want the month's", n)
	}

	left := mustDecimal(t, f.balance(t))
	if !left.IsPos() || left.GreaterThan(mustDecimal(t, "24")) {
		t.Errorf("balance = %s, want a little over 23.64 and nothing like the deposit", left)
	}

	// All of it is gain: the deposit left with its cost, the interest cost
	// nothing.
	summary := f.summary(t)
	sameAmount(t, "gain", summary.TotalGainLoss, left.String())
}

// A change of posting is a new version of the rate, so a month can be half
// daily and half monthly. Whichever way it turns, the days a balance holds are
// paid by the first day that credits, and the carry crosses the boundary.
func TestCashAccrualSwitchesPostingMidMonth(t *testing.T) {
	lastOfMonth := time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)

	t.Run("daily then monthly", func(t *testing.T) {
		f := newCashFixture(t)

		deposit := f.mustMove(t, CashKindDeposit, "10000", cashDay)
		daily := f.rateWith(t, "9", cashDay, func(*CashRateInput) {})
		f.accrueThrough(t, deposit.EntryID, daily.ID, cashDay.AddDate(0, 0, 8))

		if n := f.countInterest(t, deposit.EntryID); n != 9 {
			t.Fatalf("credits = %d, want one per daily day", n)
		}

		switched := cashDay.AddDate(0, 0, 9)
		next := f.rateWith(t, "9", switched, monthly)

		for day := switched; !day.After(lastOfMonth); day = day.AddDate(0, 0, 1) {
			f.accrue(t, deposit.EntryID, next.ID, day)
		}

		// The nine daily credits, plus one for the rest of the month.
		if n := f.countInterest(t, deposit.EntryID); n != 10 {
			t.Errorf("credits = %d, want the nine daily ones and the month's", n)
		}
		if held := f.balanceOf(t, deposit.EntryID); mustDecimal(t, held.PendingInterest).IsPos() {
			t.Errorf("pending interest = %s, want nothing held once the month closed", held.PendingInterest)
		}

		// A month at 9 % E.A. is the same money either way: 10 000 ×
		// ((1.09)^(30/365) − 1) = 71.0824…, within the rounding of ten credits.
		aboutAmount(t, "balance", f.balance(t), "10071.08")
	})

	t.Run("monthly then daily", func(t *testing.T) {
		f := newCashFixture(t)

		deposit := f.mustMove(t, CashKindDeposit, "10000", cashDay)
		held := f.rateWith(t, "9", cashDay, monthly)
		f.accrueThrough(t, deposit.EntryID, held.ID, cashDay.AddDate(0, 0, 8))

		if n := f.countInterest(t, deposit.EntryID); n != 0 {
			t.Fatalf("credits = %d, want none while the month is held", n)
		}

		switched := cashDay.AddDate(0, 0, 9)
		daily := f.rateWith(t, "9", switched, func(*CashRateInput) {})

		// The first day that credits pays the nine it was holding too.
		if !f.accrue(t, deposit.EntryID, daily.ID, switched) {
			t.Fatal("the first daily day credited nothing")
		}

		if n := f.countInterest(t, deposit.EntryID); n != 1 {
			t.Errorf("credits = %d, want the held days and that day in one", n)
		}
		if after := f.balanceOf(t, deposit.EntryID); mustDecimal(t, after.PendingInterest).IsPos() {
			t.Errorf("pending interest = %s, want the held days paid", after.PendingInterest)
		}

		// Ten days of interest arrived at once: 10 000 × ((1.09)^(10/365) − 1).
		aboutAmount(t, "balance", f.balance(t), "10023.64")

		for day := switched.AddDate(0, 0, 1); !day.After(lastOfMonth); day = day.AddDate(0, 0, 1) {
			f.accrue(t, deposit.EntryID, daily.ID, day)
		}

		aboutAmount(t, "balance", f.balance(t), "10071.08")
	})
}

// A recalculation names an account. One that is not the owner's answers the way
// every other write to a rate does, instead of reporting that it did nothing.
func TestRecalculateCashInterestRefusesSomeoneElsesPlatform(t *testing.T) {
	mine, theirs := newCashFixture(t), newCashFixture(t)
	ctx := context.Background()

	deposit := theirs.mustMove(t, CashKindDeposit, "10000", cashDay)
	rate := theirs.rateWith(t, "9", cashDay, func(*CashRateInput) {})
	theirs.accrue(t, deposit.EntryID, rate.ID, cashDay)

	_, err := mine.accrualService().RecalculateCashInterest(ctx, mine.userID, RecalculateCashInterestInput{
		SourceID: theirs.sourceID,
		Currency: money.USD,
		From:     cashDay,
	})
	if !errors.Is(err, ErrPlatformNotFound) {
		t.Errorf("recalculating someone else's account = %v, want ErrPlatformNotFound", err)
	}

	// And it left their day exactly as it was.
	if n := theirs.countInterest(t, deposit.EntryID); n != 1 {
		t.Errorf("their credits = %d, want the one they had", n)
	}

	// An inactive platform of the owner's own is still theirs to fix.
	own := mine.mustMove(t, CashKindDeposit, "1000", cashDay)
	ownRate := mine.rateWith(t, "9", cashDay, func(*CashRateInput) {})
	mine.accrue(t, own.EntryID, ownRate.ID, cashDay)
	mine.exec(t, `UPDATE investment_sources SET is_active = false WHERE id = $1`, mine.sourceID)

	if _, err := mine.accrualService().RecalculateCashInterest(ctx, mine.userID, RecalculateCashInterestInput{
		SourceID: mine.sourceID,
		Currency: money.USD,
		From:     cashDay,
	}); err != nil {
		t.Errorf("recalculating an inactive platform of their own: %v", err)
	}
}
