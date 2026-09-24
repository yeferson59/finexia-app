package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// The interest ledger — migration 000044 and CashAccrualStore — and the loop the
// job runs over it. Same database contract as postgres_cash_db_test.go.

// accrualService is the service with only what AccrueCashInterest reads. It
// computes every balance in the database that has a rate; the checks below read
// only the fixture's.
func (f cashFixture) accrualService() *service {
	return newService(f.repo, testConfig(), nil, nil, nil, logger.Noop())
}

func (f cashFixture) usdRate(t *testing.T, pct string, from time.Time) CashRate {
	t.Helper()

	rate, err := f.rate(t, money.USD, pct, from)
	if err != nil {
		t.Fatalf("CreateCashRate(%s from %s): %v", pct, from.Format(time.DateOnly), err)
	}

	return rate
}

// openedOn backdates the day a balance was opened. The fixtures open it now, and
// a balance earns nothing before it was opened or its first movement, whichever
// is earlier.
func (f cashFixture) openedOn(t *testing.T, entryID uuid.UUID, day time.Time) {
	t.Helper()

	f.exec(t, `UPDATE portfolio_entries SET created_at = ($2::date)::timestamp AT TIME ZONE 'UTC' WHERE id = $1`,
		entryID, day.Format(time.DateOnly))
}

func (f cashFixture) accrue(t *testing.T, entryID, rateID uuid.UUID, day time.Time) bool {
	t.Helper()

	credited, err := f.repo.AccrueCashInterestDay(context.Background(), entryID, rateID, day)
	if err != nil {
		t.Fatalf("AccrueCashInterestDay(%s): %v", day.Format(time.DateOnly), err)
	}

	return credited
}

type accrualRow struct {
	basis, net, carry, status string
	credited                  bool
}

func (f cashFixture) accrualOn(t *testing.T, entryID uuid.UUID, day time.Time) (accrualRow, bool) {
	t.Helper()

	var row accrualRow
	err := f.pool.QueryRow(context.Background(), `
		SELECT balance_basis::text, net_amount::text, rounding_carry::text, status, transaction_id IS NOT NULL
		FROM cash_interest_accruals
		WHERE entry_id = $1 AND accrual_date = $2::date
	`, entryID, day.Format(time.DateOnly)).Scan(&row.basis, &row.net, &row.carry, &row.status, &row.credited)
	if errors.Is(err, pgx.ErrNoRows) {
		return row, false
	}
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}

	return row, true
}

func (f cashFixture) countInterest(t *testing.T, entryID uuid.UUID) int {
	t.Helper()

	var n int
	if err := f.pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM transactions WHERE entry_id = $1 AND type = 'cash_interest'`, entryID,
	).Scan(&n); err != nil {
		t.Fatalf("count interest: %v", err)
	}

	return n
}

func TestCashAccrualCreditsADayOfInterest(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	deposit := f.mustMove(t, CashKindDeposit, "10000", cashDay)
	rate := f.usdRate(t, "9", cashDay)

	if !f.accrue(t, deposit.EntryID, rate.ID, cashDay) {
		t.Fatal("nothing was credited")
	}

	// 10 000 × ((1.09)^(1/365) − 1) = 2.3613…: 2.36 credited, the rest carried.
	sameAmount(t, "balance", f.balance(t), "10002.36")

	row, ok := f.accrualOn(t, deposit.EntryID, cashDay)
	if !ok {
		t.Fatal("the day has no ledger row")
	}
	sameAmount(t, "basis", row.basis, "10000")
	aboutAmount(t, "net", row.net, "2.3613")
	aboutAmount(t, "carry", row.carry, "0.0013")
	if row.status != "posted" || !row.credited {
		t.Errorf("row = %+v, want posted with its transaction", row)
	}

	movements, err := f.repo.GetCashMovementsPaginated(ctx, f.userID, 10, 0)
	if err != nil {
		t.Fatalf("GetCashMovementsPaginated: %v", err)
	}
	if len(movements) != 2 {
		t.Fatalf("movements = %+v, want the deposit and the credit", movements)
	}

	credit := movements[0]
	if credit.Kind != CashKindInterest || !credit.Automatic || !credit.Date.Equal(cashDay) {
		t.Errorf("credit = %+v, want automatic interest on %s", credit, cashDay.Format(time.DateOnly))
	}
	sameAmount(t, "credit", credit.Amount, "2.36")
	if movements[1].Automatic {
		t.Error("the deposit reads as automatic")
	}

	balances, err := f.repo.GetCashBalancesByUserID(ctx, f.userID, money.XXX)
	if err != nil {
		t.Fatalf("GetCashBalancesByUserID: %v", err)
	}
	sameAmount(t, "interest earned", balances[0].InterestEarned, "2.36")
	if last := balances[0].LastAccrualDate; last == nil || !last.Equal(cashDay) {
		t.Errorf("last accrual = %v, want %s", last, cashDay.Format(time.DateOnly))
	}
}

func TestCashAccrualComputesADayOnce(t *testing.T) {
	f := newCashFixture(t)

	deposit := f.mustMove(t, CashKindDeposit, "10000", cashDay)
	rate := f.usdRate(t, "9", cashDay)

	f.accrue(t, deposit.EntryID, rate.ID, cashDay)
	if f.accrue(t, deposit.EntryID, rate.ID, cashDay) {
		t.Error("the same day was credited twice")
	}

	if n := f.countInterest(t, deposit.EntryID); n != 1 {
		t.Errorf("credits = %d, want 1", n)
	}
	sameAmount(t, "balance", f.balance(t), "10002.36")
}

// The second day earns on the first day's interest, and on what its rounding
// carried.
func TestCashAccrualCompoundsAndCarries(t *testing.T) {
	f := newCashFixture(t)

	deposit := f.mustMove(t, CashKindDeposit, "10000", cashDay)
	rate := f.usdRate(t, "9", cashDay)

	f.accrue(t, deposit.EntryID, rate.ID, cashDay)
	f.accrue(t, deposit.EntryID, rate.ID, cashDay.AddDate(0, 0, 1))

	second, ok := f.accrualOn(t, deposit.EntryID, cashDay.AddDate(0, 0, 1))
	if !ok {
		t.Fatal("the second day has no ledger row")
	}
	sameAmount(t, "second basis", second.basis, "10002.36")
	// 2.3618688 earned, plus 0.0013115 carried, is 2.3631803: 2.36 and 0.0031803.
	aboutAmount(t, "second carry", second.carry, "0.0032")

	sameAmount(t, "balance", f.balance(t), "10004.72")
}

func TestCashAccrualWithholds(t *testing.T) {
	f := newCashFixture(t)

	deposit := f.mustMove(t, CashKindDeposit, "10000", cashDay)
	rate, err := f.repo.CreateCashRate(context.Background(), f.userID, NewCashRateInput{
		SourceID:       f.sourceID,
		Currency:       money.USD,
		EffectiveFrom:  cashDay,
		AnnualRatePct:  mustDecimal(t, "9"),
		WithholdingPct: mustDecimal(t, "7"),
		Posting:        PostingDaily,
	})
	if err != nil {
		t.Fatalf("CreateCashRate: %v", err)
	}

	f.accrue(t, deposit.EntryID, rate.ID, cashDay)

	// 2.3613115 × 0.93 = 2.1960197.
	sameAmount(t, "balance", f.balance(t), "10002.2")
}

func TestCashAccrualSkipsTheDaysWithoutARate(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	deposit := f.mustMove(t, CashKindDeposit, "1000", cashDay)
	f.openedOn(t, deposit.EntryID, cashDay.AddDate(0, 0, -1))

	first := f.usdRate(t, "9", cashDay)
	if _, err := f.repo.EndCashRate(ctx, f.userID, first.ID, cashDay.AddDate(0, 0, 2)); err != nil {
		t.Fatalf("EndCashRate: %v", err)
	}
	f.usdRate(t, "12", cashDay.AddDate(0, 0, 4))

	if _, errs := f.accrualService().AccrueCashInterest(ctx, cashDay.AddDate(0, 0, 5)); len(errs) > 0 {
		t.Fatalf("AccrueCashInterest: %v", errs)
	}

	rows, err := f.pool.Query(ctx, `
		SELECT accrual_date, trim_scale(annual_rate * 100)::text
		FROM cash_interest_accruals WHERE entry_id = $1 ORDER BY accrual_date
	`, deposit.EntryID)
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	defer rows.Close()

	type day struct {
		date time.Time
		pct  string
	}

	var got []day
	for rows.Next() {
		var d day
		if err := rows.Scan(&d.date, &d.pct); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, d)
	}

	want := []day{
		{cashDay, "9"},
		{cashDay.AddDate(0, 0, 1), "9"},
		{cashDay.AddDate(0, 0, 4), "12"},
		{cashDay.AddDate(0, 0, 5), "12"},
	}
	if len(got) != len(want) {
		t.Fatalf("ledger = %+v, want %+v", got, want)
	}
	for i := range want {
		if !got[i].date.Equal(want[i].date) || got[i].pct != want[i].pct {
			t.Errorf("day %d = %s at %s %%, want %s at %s %%", i, got[i].date.Format(time.DateOnly), got[i].pct, want[i].date.Format(time.DateOnly), want[i].pct)
		}
	}
}

// Ten days caught up the morning after a snapshot are dated on the snapshot's
// day. Dated on the days they were earned, the growth series would read them as
// loaded history — money walking in — and not as the return they are.
func TestCashAccrualCaughtUpStillCountsAsReturn(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	today := time.Now().UTC().Truncate(24 * time.Hour)
	start := today.AddDate(0, 0, -10)
	yesterday := today.AddDate(0, 0, -1)

	deposit := f.mustMove(t, CashKindDeposit, "1000", start)
	f.openedOn(t, deposit.EntryID, start)
	f.usdRate(t, "9", start)

	f.exec(t, `INSERT INTO portfolio_snapshots (portfolio_id, snapshot_date, total_value, currency, created_at)
	           VALUES ($1, $2::date, 1000, 'USD', NOW())`, f.portfolioID, yesterday.Format(time.DateOnly))

	if _, errs := f.accrualService().AccrueCashInterest(ctx, yesterday); len(errs) > 0 {
		t.Fatalf("AccrueCashInterest: %v", errs)
	}

	if n := f.countInterest(t, deposit.EntryID); n != 10 {
		t.Errorf("credits = %d, want one per day", n)
	}

	var misdated int
	if err := f.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM transactions
		WHERE entry_id = $1 AND type = 'cash_interest' AND transaction_date <> $2::date
	`, deposit.EntryID, yesterday.Format(time.DateOnly)).Scan(&misdated); err != nil {
		t.Fatalf("count misdated: %v", err)
	}
	if misdated != 0 {
		t.Errorf("%d credits are dated before the snapshot", misdated)
	}

	// Each caught-up day still earned on the credits of the days before it.
	var basis, heldBefore string
	if err := f.pool.QueryRow(ctx, `
		SELECT ac.balance_basis::text,
		       (1000 + COALESCE((
		         SELECT SUM(t.quantity)
		         FROM cash_interest_accruals prior
		         JOIN transactions t ON t.id = prior.transaction_id
		         WHERE prior.entry_id = $1 AND prior.accrual_date < $2::date
		       ), 0))::text
		FROM cash_interest_accruals ac
		WHERE ac.entry_id = $1 AND ac.accrual_date = $2::date
	`, deposit.EntryID, yesterday.Format(time.DateOnly)).Scan(&basis, &heldBefore); err != nil {
		t.Fatalf("read basis: %v", err)
	}
	sameAmount(t, "basis of the last day", basis, heldBefore)

	series, err := f.repo.GetPortfolioGrowthByPortfolioID(ctx, f.userID, f.portfolioID, false, time.Time{}, today)
	if err != nil {
		t.Fatalf("GetPortfolioGrowthByPortfolioID: %v", err)
	}
	if len(series) == 0 {
		t.Fatal("empty series")
	}

	last := series[len(series)-1]
	sameAmount(t, "net flow", last.NetFlow, "0")
	sameAmount(t, "value", last.TotalValue, f.balance(t))
}

// An owner who deletes a credit meant it: the day stays computed.
func TestCashAccrualDoesNotCreditADeletedInterestAgain(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	deposit := f.mustMove(t, CashKindDeposit, "10000", cashDay)
	rate := f.usdRate(t, "9", cashDay)
	f.accrue(t, deposit.EntryID, rate.ID, cashDay)

	movements, err := f.repo.GetCashMovementsPaginated(ctx, f.userID, 10, 0)
	if err != nil || len(movements) == 0 || !movements[0].Automatic {
		t.Fatalf("movements = %+v, %v; want the credit first", movements, err)
	}
	if err := f.repo.DeleteCashMovement(ctx, f.userID, movements[0].ID); err != nil {
		t.Fatalf("DeleteCashMovement: %v", err)
	}

	if f.accrue(t, deposit.EntryID, rate.ID, cashDay) {
		t.Error("the deleted day was credited again")
	}

	f.openedOn(t, deposit.EntryID, cashDay)
	if _, errs := f.accrualService().AccrueCashInterest(ctx, cashDay); len(errs) > 0 {
		t.Fatalf("AccrueCashInterest: %v", errs)
	}

	sameAmount(t, "balance", f.balance(t), "10000")
	if n := f.countInterest(t, deposit.EntryID); n != 0 {
		t.Errorf("credits = %d, want none", n)
	}

	if row, ok := f.accrualOn(t, deposit.EntryID, cashDay); !ok || row.credited {
		t.Errorf("row = %+v (found %v), want the day kept without its transaction", row, ok)
	}
}

// Once interest was computed at a rate, the days it covered keep it.
func TestCashRateThatEarnedInterestKeepsItsPast(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	deposit := f.mustMove(t, CashKindDeposit, "10000", cashDay)
	rate := f.usdRate(t, "9", cashDay)
	f.accrue(t, deposit.EntryID, rate.ID, cashDay)
	f.accrue(t, deposit.EntryID, rate.ID, cashDay.AddDate(0, 0, 1))

	values := CashRateInput{AnnualRatePct: mustDecimal(t, "20"), WithholdingPct: mustDecimal(t, "0"), Posting: PostingDaily}

	if _, err := f.repo.UpdateCashRate(ctx, f.userID, rate.ID, values); !errors.Is(err, ErrCashRateInUse) {
		t.Errorf("correcting it = %v, want ErrCashRateInUse", err)
	}
	if err := f.repo.DeleteCashRate(ctx, f.userID, rate.ID); !errors.Is(err, ErrCashRateInUse) {
		t.Errorf("deleting it = %v, want ErrCashRateInUse", err)
	}
	if _, err := f.repo.EndCashRate(ctx, f.userID, rate.ID, cashDay.AddDate(0, 0, 1)); !errors.Is(err, ErrCashRateInUse) {
		t.Errorf("stopping it on a computed day = %v, want ErrCashRateInUse", err)
	}
	if _, err := f.rate(t, money.USD, "8", cashDay.AddDate(0, 0, 1)); !errors.Is(err, ErrCashRateInUse) {
		t.Errorf("a new version from a computed day = %v, want ErrCashRateInUse", err)
	}

	// From the first day not yet computed, both are open.
	if _, err := f.repo.EndCashRate(ctx, f.userID, rate.ID, cashDay.AddDate(0, 0, 2)); err != nil {
		t.Errorf("stopping it after the computed days: %v", err)
	}
	if _, err := f.rate(t, money.USD, "8", cashDay.AddDate(0, 0, 2)); err != nil {
		t.Errorf("a new version after the computed days: %v", err)
	}

	rates := f.rates(t)
	if len(rates) != 2 || rates[1].AccruedThrough == nil || !rates[1].AccruedThrough.Equal(cashDay.AddDate(0, 0, 1)) {
		t.Errorf("rates = %+v, want the first computed through %s", rates, cashDay.AddDate(0, 0, 1).Format(time.DateOnly))
	}
}
