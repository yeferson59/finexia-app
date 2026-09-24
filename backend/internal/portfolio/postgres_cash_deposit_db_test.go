package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// Fixed deposits, migration 000048. Same database contract as
// postgres_cash_db_test.go.
//
// The days are relative to the server's clock, because a deposit is validated
// against it: one cannot be opened tomorrow, and one whose term ran out is
// recorded as the movements it paid rather than as a deposit.

// depositDay is a calendar day `days` from cashToday, the first day not yet
// closed, at midnight UTC — negative for the past.
func depositDay(days int) time.Time {
	return cashToday().AddDate(0, 0, days)
}

// deposit opens a fixed deposit on the fixture's account, through the service,
// so the days it has already earned are computed with it.
func (f cashFixture) deposit(t *testing.T, name, amount, pct string, opened time.Time, matures *time.Time, posting InterestPosting) CashPocket {
	t.Helper()

	pocket, err := f.accrualService().OpenFixedDeposit(context.Background(), f.userID, NewFixedDepositInput{
		PortfolioID:    f.portfolioID,
		SourceID:       f.sourceID,
		Currency:       money.USD,
		Name:           name,
		Amount:         mustDecimal(t, amount),
		OpenedOn:       opened,
		MaturesOn:      matures,
		AnnualRatePct:  mustDecimal(t, pct),
		WithholdingPct: decimal.Zero,
		Posting:        posting,
	})
	if err != nil {
		t.Fatalf("OpenFixedDeposit(%q): %v", name, err)
	}

	return pocket
}

// netFlow is what the fixture's portfolio has taken in, by the convention
// transaction_cash_flow keeps: a withdrawal's fee stays in, so what the
// platform kept reads as a loss and not as money that left.
func (f cashFixture) netFlow(t *testing.T) string {
	t.Helper()

	var flow string
	if err := f.pool.QueryRow(context.Background(), `
		SELECT ROUND(COALESCE(SUM(
			transaction_cash_flow(t.type, t.quantity * t.fx_rate, t.price, t.fees)
		), 0), 8)::text
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		WHERE pe.portfolio_id = $1
	`, f.portfolioID).Scan(&flow); err != nil {
		t.Fatalf("read the net flow: %v", err)
	}

	return flow
}

// balanceIn is what one drawer of the fixture's account holds, "0" when it has
// no balance at all.
func (f cashFixture) balanceIn(t *testing.T, pocketID *uuid.UUID) string {
	t.Helper()

	balances, err := f.repo.GetCashBalancesByUserID(context.Background(), f.userID, money.XXX)
	if err != nil {
		t.Fatalf("GetCashBalancesByUserID: %v", err)
	}

	for _, b := range balances {
		switch {
		case pocketID == nil && b.PocketID == nil,
			pocketID != nil && b.PocketID != nil && *b.PocketID == *pocketID:
			return b.Balance
		}
	}

	return "0"
}

// A deposit opened two weeks ago and registered today shows those two weeks the
// moment it is saved. Nothing is ever recorded in it by hand, so nothing can be
// counted twice, and it earns from the day the money went in rather than from
// the day the row was written.
func TestFixedDepositOpenedInThePastEarnsFromThatDay(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	opened := depositDay(-14)
	matures := depositDay(76)

	cdt := f.deposit(t, "CDT 90 días", "10000000", "10", opened, &matures, PostingDaily)

	if cdt.Kind != PocketFixed || cdt.MaturesOn == nil || !cdt.MaturesOn.Equal(matures) {
		t.Fatalf("deposit = %s maturing %v, want a fixed one maturing %s", cdt.Kind, cdt.MaturesOn, matures.Format(time.DateOnly))
	}

	// Fourteen days at 10 % E.A. on ten million: 36.624,23, less what the daily
	// rounding carries.
	earned, err := decimal.NewFromString(cdt.Balance)
	if err != nil {
		t.Fatalf("balance %q: %v", cdt.Balance, err)
	}
	near(t, "what the deposit is worth", earned, "10036624.23", "0.5")

	// One ledger row per day it was open, none for today, which is not over.
	var days int
	if err := f.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM cash_interest_accruals ac
		JOIN portfolio_entries pe ON pe.id = ac.entry_id
		WHERE pe.pocket_id = $1
	`, cdt.ID).Scan(&days); err != nil {
		t.Fatalf("count the ledger: %v", err)
	}
	if days != 14 {
		t.Errorf("days computed = %d, want 14", days)
	}

	// The money that went in is one deposit, dated the day it went in. The
	// interest is credited as cash_interest, which costs nothing and so is all
	// gain: what came in is the ten million and not a peso more.
	sameAmount(t, "what the portfolio took in", f.netFlow(t), "10000000")
}

// A deposit posted at maturity holds every day and pays them in one credit on
// the last day it earns — the day before it comes due.
func TestFixedDepositPostedAtMaturityCreditsOnce(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	// It comes due today, so the last day it earned was yesterday, which is the
	// last day the nightly accrual computes.
	matures := depositDay(0)
	cdt := f.deposit(t, "CDT al vencimiento", "10000000", "10", depositDay(-14), &matures, PostingAtMaturity)

	var entryID uuid.UUID
	if err := f.pool.QueryRow(ctx, `SELECT id FROM portfolio_entries WHERE pocket_id = $1`, cdt.ID).Scan(&entryID); err != nil {
		t.Fatalf("read the deposit's balance: %v", err)
	}

	if n := f.countInterest(t, entryID); n != 1 {
		t.Errorf("credits = %d, want one: a deposit posted at maturity pays once", n)
	}

	var pending int
	if err := f.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM cash_interest_accruals WHERE entry_id = $1 AND status <> 'posted'
	`, entryID).Scan(&pending); err != nil {
		t.Fatalf("count what is still waiting: %v", err)
	}
	if pending != 0 {
		t.Errorf("days still waiting = %d, want none once the term is up", pending)
	}
}

// On the day it comes due, the deposit's balance goes back to the main account.
// The two legs are dated that day and offset each other, so the portfolio's net
// flow — and its return — does not move.
func TestFixedDepositMatures(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	matures := depositDay(2)
	cdt := f.deposit(t, "CDT corto", "10000", "10", depositDay(-5), &matures, PostingDaily)

	before := f.netFlow(t)
	held := f.balanceIn(t, &cdt.ID)

	matured, err := f.repo.MatureCashPockets(ctx, matures)
	if err != nil {
		t.Fatalf("MatureCashPockets: %v", err)
	}
	if matured != 1 {
		t.Fatalf("matured = %d, want one", matured)
	}

	sameAmount(t, "what the deposit holds after", f.balanceIn(t, &cdt.ID), "0")
	sameAmount(t, "what the main account holds", f.balanceIn(t, nil), held)
	sameAmount(t, "the net flow", f.netFlow(t), before)

	closed, err := f.repo.GetCashPocketByID(ctx, f.userID, cdt.ID)
	if err != nil {
		t.Fatalf("GetCashPocketByID: %v", err)
	}
	if closed.ClosedOn == nil || !closed.ClosedOn.Equal(matures) {
		t.Errorf("closedOn = %v, want %s", closed.ClosedOn, matures.Format(time.DateOnly))
	}

	// Running the sweep again — the job catching up after a day down — moves
	// nothing a second time.
	again, err := f.repo.MatureCashPockets(ctx, matures.AddDate(0, 0, 3))
	if err != nil {
		t.Fatalf("MatureCashPockets again: %v", err)
	}
	if again != 0 {
		t.Errorf("matured again = %d, want none", again)
	}
	sameAmount(t, "what the main account holds after the second run", f.balanceIn(t, nil), held)
}

// Cancelling early stops the rate the day before, pays what it earned and moves
// the rest back. The penalty rides on the withdrawal as its fee, so the money
// that arrives is short by it while the net flow is not: it reads as a loss,
// not as money the owner took out.
func TestCancellingAFixedDepositCountsThePenaltyAsALoss(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	matures := depositDay(60)
	cdt := f.deposit(t, "CDT cancelado", "10000", "10", depositDay(-10), &matures, PostingDaily)

	before := f.netFlow(t)
	held, err := decimal.NewFromString(f.balanceIn(t, &cdt.ID))
	if err != nil {
		t.Fatalf("balance: %v", err)
	}

	closed, err := f.accrualService().CloseFixedDeposit(ctx, f.userID, cdt.ID, CloseFixedDepositInput{
		ClosesOn: depositDay(0),
		Penalty:  mustDecimal(t, "100"),
	})
	if err != nil {
		t.Fatalf("CloseFixedDeposit: %v", err)
	}

	if closed.ClosedOn == nil {
		t.Error("the deposit is not closed")
	}
	sameAmount(t, "what the deposit holds after", f.balanceIn(t, &cdt.ID), "0")
	sameAmount(t, "what the main account holds", f.balanceIn(t, nil), held.Sub(mustDecimal(t, "100")).String())
	sameAmount(t, "the net flow", f.netFlow(t), before)

	// Its rate stopped the day before it closed, so tonight's run computes
	// nothing more for it.
	var endedOn *time.Time
	if err := f.pool.QueryRow(ctx, `SELECT ended_on FROM cash_yield_rates WHERE pocket_id = $1`, cdt.ID).Scan(&endedOn); err != nil {
		t.Fatalf("read the rate: %v", err)
	}
	if endedOn == nil || !endedOn.Equal(depositDay(-1)) {
		t.Errorf("endedOn = %v, want %s", endedOn, depositDay(-1).Format(time.DateOnly))
	}

	// And it cannot be cancelled twice.
	if _, err := f.accrualService().CloseFixedDeposit(ctx, f.userID, cdt.ID, CloseFixedDepositInput{ClosesOn: depositDay(0)}); !errors.Is(err, ErrCashPocketClosed) {
		t.Errorf("cancelling it again = %v, want ErrCashPocketClosed", err)
	}
}

// Nothing is written in a deposit by hand: not a movement, not a move, not a
// version of its rate, and not from the positions screen either.
func TestFixedDepositRefusesWritesByHand(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	matures := depositDay(60)
	cdt := f.deposit(t, "CDT cerrado", "10000", "10", depositDay(-3), &matures, PostingDaily)
	f.mustMove(t, CashKindDeposit, "5000", cashDay)

	var (
		entryID uuid.UUID
		rateID  uuid.UUID
		openID  uuid.UUID
	)
	if err := f.pool.QueryRow(ctx, `SELECT id FROM portfolio_entries WHERE pocket_id = $1`, cdt.ID).Scan(&entryID); err != nil {
		t.Fatalf("read the deposit's balance: %v", err)
	}
	if err := f.pool.QueryRow(ctx, `SELECT id FROM cash_yield_rates WHERE pocket_id = $1`, cdt.ID).Scan(&rateID); err != nil {
		t.Fatalf("read the deposit's rate: %v", err)
	}
	if err := f.pool.QueryRow(ctx, `
		SELECT id FROM transactions WHERE entry_id = $1 AND type = 'transfer_in'
	`, entryID).Scan(&openID); err != nil {
		t.Fatalf("read the opening deposit: %v", err)
	}

	usd := CashMovementInput{Kind: CashKindDeposit, Amount: mustDecimal(t, "100"), Currency: money.USD, Date: cashDay}

	refusals := map[string]func() error{
		"a deposit by hand": func() error {
			_, err := f.repo.CreateCashMovement(ctx, f.userID, f.portfolioID, f.sourceID, cdt.ID, usd)
			return err
		},
		"moving money into it": func() error {
			_, err := f.repo.MoveCash(ctx, f.userID, f.portfolioID, f.sourceID, CashMoveInput{
				Currency: money.USD, To: cdt.ID, Amount: mustDecimal(t, "100"), Date: cashDay,
			})
			return err
		},
		"moving money out of it": func() error {
			_, err := f.repo.MoveCash(ctx, f.userID, f.portfolioID, f.sourceID, CashMoveInput{
				Currency: money.USD, From: cdt.ID, Amount: mustDecimal(t, "100"), Date: cashDay,
			})
			return err
		},
		"a new version of its rate": func() error {
			_, err := f.repo.CreateCashRate(ctx, f.userID, NewCashRateInput{
				SourceID: f.sourceID, Currency: money.USD, PocketID: cdt.ID, EffectiveFrom: depositDay(0),
				AnnualRatePct: mustDecimal(t, "12"), Posting: PostingDaily,
			})
			return err
		},
		"correcting its rate": func() error {
			_, err := f.repo.UpdateCashRate(ctx, f.userID, rateID, CashRateInput{
				AnnualRatePct: mustDecimal(t, "12"), Posting: PostingDaily,
			})
			return err
		},
		"pausing its rate": func() error {
			_, err := f.repo.EndCashRate(ctx, f.userID, rateID, depositDay(0))
			return err
		},
		"deleting its rate": func() error {
			return f.repo.DeleteCashRate(ctx, f.userID, rateID)
		},
		"a transaction from the positions screen": func() error {
			_, err := f.repo.CreateTransaction(ctx, f.userID, entryID, TransactionInput{
				Type:            TransferIn,
				Quantity:        mustDecimal(t, "100"),
				Price:           money.NewFromDecimal(decimal.One, money.USD),
				Currency:        money.USD,
				FXRate:          decimal.One,
				Fees:            money.NewFromDecimal(decimal.Zero, money.USD),
				FeesCurrency:    money.USD,
				TransactionDate: cashDay,
			})
			return err
		},
		"rewriting the opening deposit": func() error {
			_, err := f.repo.UpdateCashMovement(ctx, f.userID, openID, usd)
			return err
		},
		"deleting the opening deposit": func() error {
			return f.repo.DeleteCashMovement(ctx, f.userID, openID)
		},
		"deleting it as a transaction": func() error {
			return f.repo.DeleteTransaction(ctx, f.userID, openID)
		},
	}

	for name, write := range refusals {
		t.Run(name, func(t *testing.T) {
			if err := write(); !errors.Is(err, ErrCashPocketFixed) {
				t.Errorf("%s = %v, want ErrCashPocketFixed", name, err)
			}
		})
	}

	// The main account of the same platform is untouched by any of it.
	sameAmount(t, "the main account", f.balanceIn(t, nil), "5000")
}

// Deleting a deposit takes the whole thing: the money, the interest, the rate
// and the pocket. It is what "I recorded this wrong" means, and the reason a
// flexible pocket with movements is refused instead — that one has a history
// that happened.
func TestDeletingAFixedDepositTakesItWhole(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	matures := depositDay(60)
	cdt := f.deposit(t, "CDT mal anotado", "10000", "10", depositDay(-5), &matures, PostingDaily)

	if err := f.repo.DeleteCashPocket(ctx, f.userID, cdt.ID); err != nil {
		t.Fatalf("DeleteCashPocket: %v", err)
	}

	for _, q := range []struct {
		what string
		sql  string
	}{
		{"the pocket", `SELECT COUNT(*) FROM cash_pockets WHERE id = $1`},
		{"its balance", `SELECT COUNT(*) FROM portfolio_entries WHERE pocket_id = $1`},
		{"its rate", `SELECT COUNT(*) FROM cash_yield_rates WHERE pocket_id = $1`},
	} {
		var n int
		if err := f.pool.QueryRow(ctx, q.sql, cdt.ID).Scan(&n); err != nil {
			t.Fatalf("count %s: %v", q.what, err)
		}
		if n != 0 {
			t.Errorf("%s is still there (%d rows)", q.what, n)
		}
	}
}
