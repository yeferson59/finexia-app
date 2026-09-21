package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

// Purchases paid out of the platform's cash (000053, 000054): the debit that
// mirrors a sale's credit. Same database contract as postgres_cash_db_test.go,
// and the same fixture as postgres_cash_link_db_test.go — a share worth what it
// cost, so every figure below comes from the money moving.

// buyInput is a purchase of quantity units at price, funded from cash or not.
// Out of the main account: fromPocket is the version that names a drawer.
func buyInput(t *testing.T, quantity, price string, fromCash bool, date time.Time) TransactionInput {
	t.Helper()

	in := tradeInput(t, Buy, quantity, price, false, date)
	in.PayFromCash = fromCash

	return in
}

// fromPocket is buyInput paid out of one drawer of the account.
func fromPocket(in TransactionInput, pocketID uuid.UUID) TransactionInput {
	in.PayFromCash = true
	in.CashPocketID = pocketID

	return in
}

// balanceOfPocket is what one drawer of the fixture's account holds.
func (f creditFixture) balanceOfPocket(t *testing.T, pocketID *uuid.UUID) string {
	t.Helper()

	balances, err := f.repo.GetCashBalancesByUserID(context.Background(), f.userID, money.XXX)
	if err != nil {
		t.Fatalf("GetCashBalancesByUserID: %v", err)
	}

	for _, b := range balances {
		if samePocket(b.PocketID, pocketID) {
			return b.Balance
		}
	}

	return "0"
}

// debits lists every cash movement that holds a purchase.
func (f creditFixture) debits(t *testing.T) []CashMovement {
	t.Helper()

	debits := make([]CashMovement, 0)
	for _, m := range f.credits(t) {
		if m.Type == CashPurchase {
			debits = append(debits, m)
		}
	}

	return debits
}

// fund opens the balance the purchases below are paid from.
func (f creditFixture) fund(t *testing.T, amount string) {
	t.Helper()

	f.mustMove(t, CashKindDeposit, amount, cashDay)
}

func TestPurchasePaidFromCashComesOutOfTheBalance(t *testing.T) {
	f := newCreditFixture(t, cashDay)
	f.fund(t, "1000")

	txn := f.mustTrade(t, buyInput(t, "5", "100", true, cashDay.AddDate(0, 0, 1)))
	if !txn.CashPaid {
		t.Error("CashPaid = false, want true")
	}

	sameAmount(t, "balance", f.balanceIn(t, money.USD), "500")

	debits := f.debits(t)
	if len(debits) != 1 {
		t.Fatalf("debits = %+v, want exactly one", debits)
	}

	debit := debits[0]
	if debit.Kind != CashKindPurchase || debit.Editable || debit.OriginTicker != f.ticker {
		t.Errorf("debit = %+v, want a read-only purchase of %s", debit, f.ticker)
	}
	if debit.Ticker != "CASH-USD" || debit.SourceID != f.sourceID || debit.PortfolioID != f.portfolioID {
		t.Errorf("debit landed on %s of %v/%v, want CASH-USD of the holding's platform", debit.Ticker, debit.PortfolioID, debit.SourceID)
	}
	if !debit.Date.Equal(cashDay.AddDate(0, 0, 1)) {
		t.Errorf("debit date = %v, want the purchase's", debit.Date)
	}
	sameAmount(t, "debit amount", debit.Amount, "500")

	// Nothing arrived and nothing left: the money was already in the portfolio
	// and turned into shares. Ten units from the fixture plus five, at 100, and
	// the 500 still in cash.
	s := f.summary(t)
	aboutAmount(t, "portfolio value", s.TotalMarketValue, "2000")
	aboutAmount(t, "portfolio gain", s.TotalGainLoss, "0")

	listed, err := f.repo.GetTransactionsByEntryID(context.Background(), f.userID, f.holdingID)
	if err != nil {
		t.Fatalf("GetTransactionsByEntryID: %v", err)
	}
	for _, l := range listed {
		if l.ID == txn.ID && (!l.CashPaid || l.CashCredited) {
			t.Errorf("the purchase reads back as paid=%v credited=%v", l.CashPaid, l.CashCredited)
		}
	}
}

// The commission goes out with the money: the account paid the price and the
// fee, which is exactly what the purchase cost the growth series.
func TestPurchaseDebitCarriesTheCommission(t *testing.T) {
	f := newCreditFixture(t, cashDay)
	f.fund(t, "1000")

	in := buyInput(t, "1", "100", true, cashDay)
	in.Fees = mustUSD(t, "5")
	f.mustTrade(t, in)

	sameAmount(t, "balance", f.balanceIn(t, money.USD), "895")
}

// A purchase in another currency comes out of the balance the position settles
// in, at the rate it converted at.
func TestPurchaseDebitIsWhatTheAccountPaid(t *testing.T) {
	f := newCreditFixture(t, cashDay)
	f.fund(t, "1000")

	f.mustTrade(t, TransactionInput{
		Type:            Buy,
		Quantity:        mustDecimal(t, "2"),
		Price:           mustEUR(t, "100"),
		Currency:        money.EUR,
		FXRate:          mustDecimal(t, "1.1"),
		TransactionDate: cashDay,
		PayFromCash:     true,
	})

	// 2 × 100 EUR at 1.1 = 220 USD, the currency the position costs in.
	sameAmount(t, "balance", f.balanceIn(t, money.USD), "780")
}

// The refusal the whole feature turns on: a balance that does not hold enough
// stops the purchase, and stops it whole — no transaction, no debit, no
// position moved.
func TestPurchaseFromCashIsRefusedWhenTheBalanceIsShort(t *testing.T) {
	f := newCreditFixture(t, cashDay)
	f.fund(t, "100")

	if _, err := f.trade(t, buyInput(t, "5", "100", true, cashDay)); !errors.Is(err, ErrInsufficientCash) {
		t.Fatalf("error = %v, want ErrInsufficientCash", err)
	}

	sameAmount(t, "balance", f.balanceIn(t, money.USD), "100")
	if got := f.debits(t); len(got) != 0 {
		t.Errorf("debits = %+v, want none", got)
	}

	listed, err := f.repo.GetTransactionsByEntryID(context.Background(), f.userID, f.holdingID)
	if err != nil {
		t.Fatalf("GetTransactionsByEntryID: %v", err)
	}
	// The fixture's own opening purchase, and nothing else.
	if len(listed) != 1 {
		t.Errorf("transactions = %d, want only the fixture's opening buy", len(listed))
	}
}

// What the owner asked for: a purchase recorded before any of this was offered
// can be paid from cash afterwards, and a purchase paid by mistake can stop
// being, with the money going back where it was.
func TestPurchaseCanBePaidFromCashAfterItWasRecorded(t *testing.T) {
	f := newCreditFixture(t, cashDay)
	f.fund(t, "1000")

	txn := f.mustTrade(t, buyInput(t, "5", "100", false, cashDay))
	sameAmount(t, "balance after recording it", f.balanceIn(t, money.USD), "1000")

	paid, err := f.repo.UpdateTransaction(context.Background(), f.userID, txn.ID, buyInput(t, "5", "100", true, cashDay))
	if err != nil {
		t.Fatalf("UpdateTransaction: %v", err)
	}
	if !paid.CashPaid {
		t.Error("CashPaid = false after the edit, want true")
	}
	sameAmount(t, "balance once it is paid from cash", f.balanceIn(t, money.USD), "500")

	// And an edit of the same purchase that changes what it cost moves the
	// debit with it.
	if _, err := f.repo.UpdateTransaction(context.Background(), f.userID, txn.ID, buyInput(t, "4", "100", true, cashDay)); err != nil {
		t.Fatalf("UpdateTransaction (resized): %v", err)
	}
	sameAmount(t, "balance after the purchase shrank", f.balanceIn(t, money.USD), "600")

	back, err := f.repo.UpdateTransaction(context.Background(), f.userID, txn.ID, buyInput(t, "4", "100", false, cashDay))
	if err != nil {
		t.Fatalf("UpdateTransaction (unpaid): %v", err)
	}
	if back.CashPaid {
		t.Error("CashPaid = true after taking it back, want false")
	}
	sameAmount(t, "balance once it is not paid from cash", f.balanceIn(t, money.USD), "1000")
	if got := f.debits(t); len(got) != 0 {
		t.Errorf("debits = %+v, want none", got)
	}
}

func TestDeletingAPurchaseGivesTheMoneyBack(t *testing.T) {
	f := newCreditFixture(t, cashDay)
	f.fund(t, "1000")

	txn := f.mustTrade(t, buyInput(t, "5", "100", true, cashDay))
	sameAmount(t, "balance", f.balanceIn(t, money.USD), "500")

	if err := f.repo.DeleteTransaction(context.Background(), f.userID, txn.ID); err != nil {
		t.Fatalf("DeleteTransaction: %v", err)
	}

	sameAmount(t, "balance once the purchase is gone", f.balanceIn(t, money.USD), "1000")
	if got := f.debits(t); len(got) != 0 {
		t.Errorf("debits = %+v, want none", got)
	}
}

// The debit belongs to its purchase, like every other cash row that does: the
// cash screens can neither edit it nor delete it.
func TestPurchaseDebitCannotBeTouchedFromTheCashScreens(t *testing.T) {
	f := newCreditFixture(t, cashDay)
	f.fund(t, "1000")
	f.mustTrade(t, buyInput(t, "5", "100", true, cashDay))

	debits := f.debits(t)
	if len(debits) != 1 {
		t.Fatalf("debits = %+v, want exactly one", debits)
	}

	ctx := context.Background()
	in := CashMovementInput{Kind: CashKindWithdrawal, Amount: mustDecimal(t, "10"), Currency: money.USD, Date: cashDay}

	if _, err := f.repo.UpdateCashMovement(ctx, f.userID, debits[0].ID, in); !errors.Is(err, ErrCashCreditLinked) {
		t.Errorf("UpdateCashMovement error = %v, want ErrCashCreditLinked", err)
	}
	if err := f.repo.DeleteCashMovement(ctx, f.userID, debits[0].ID); !errors.Is(err, ErrCashCreditLinked) {
		t.Errorf("DeleteCashMovement error = %v, want ErrCashCreditLinked", err)
	}
}

// The other door into a purchase: opening a position pays for it the same way,
// and a balance that cannot cover it opens no position at all.
func TestOpeningAPositionPaidFromCash(t *testing.T) {
	f := newCreditFixture(t, cashDay)
	f.fund(t, "1000")

	ctx := context.Background()
	in := buyInput(t, "3", "100", true, cashDay)

	entry, err := f.repo.CreatePortfolioEntry(ctx, f.userID, f.portfolioID, f.assetID, f.sourceID, money.USD, in)
	if err != nil {
		t.Fatalf("CreatePortfolioEntry: %v", err)
	}
	if entry.ID != f.holdingID {
		t.Fatalf("entry = %v, want the fixture's position %v", entry.ID, f.holdingID)
	}

	sameAmount(t, "balance", f.balanceIn(t, money.USD), "700")

	short := buyInput(t, "20", "100", true, cashDay)
	if _, err := f.repo.CreatePortfolioEntry(ctx, f.userID, f.portfolioID, f.assetID, f.sourceID, money.USD, short); !errors.Is(err, ErrInsufficientCash) {
		t.Fatalf("error = %v, want ErrInsufficientCash", err)
	}

	sameAmount(t, "balance after the refusal", f.balanceIn(t, money.USD), "700")
}

// A purchase paid from a balance the position does not settle in has no cash it
// could have come from, and says so rather than guessing at a rate.
func TestPurchaseFromCashNeedsABalanceInThePositionsCurrency(t *testing.T) {
	f := newCreditFixture(t, cashDay)

	// The holding costs in USD; a purchase on a cash position is the other
	// refusal, and the one this fixture can reach.
	cash := f.mustMove(t, CashKindDeposit, "1000", cashDay)

	in := buyInput(t, "1", "1", true, cashDay)
	if _, err := f.repo.CreateTransaction(context.Background(), f.userID, cash.EntryID, in); !errors.Is(err, ErrNotPayableFromCash) {
		t.Fatalf("error = %v, want ErrNotPayableFromCash", err)
	}
}

// Where the money actually is. Most of what an owner keeps sits in a pocket —
// the bank's «cajita», the broker's sub-balance — so a purchase has to be able
// to come out of one, and to stay in the one it came out of.
func TestPurchaseIsPaidFromThePocketItNames(t *testing.T) {
	f := newCreditFixture(t, cashDay)
	f.fund(t, "1000")

	pocket := f.pocket(t, "acciones")
	f.moveInto(t, pocket.ID, CashKindDeposit, "600", cashDay)

	// Each deposit arrives from outside, so the two balances are independent.
	sameAmount(t, "main account", f.balanceOfPocket(t, nil), "1000")
	sameAmount(t, "pocket", f.balanceOfPocket(t, &pocket.ID), "600")

	txn := f.mustTrade(t, fromPocket(buyInput(t, "5", "100", true, cashDay), pocket.ID))
	if !txn.CashPaid {
		t.Error("CashPaid = false, want true")
	}

	// It comes out of the drawer that paid, and the main account never hears
	// about it.
	sameAmount(t, "main account after the purchase", f.balanceOfPocket(t, nil), "1000")
	sameAmount(t, "pocket after the purchase", f.balanceOfPocket(t, &pocket.ID), "100")

	// Y una escritura que no tiene nada que ver no lo saca de ahí.
	f.mustDividend(t, "10", false, cashDay.AddDate(0, 0, 1))
	sameAmount(t, "pocket after an unrelated write", f.balanceOfPocket(t, &pocket.ID), "100")
	sameAmount(t, "main account after an unrelated write", f.balanceOfPocket(t, nil), "1000")
}

// The drawer is part of the answer: naming another moves the money, and gives
// the first one back what had come out of it.
func TestEditingAPurchaseMovesItToTheNewPocket(t *testing.T) {
	f := newCreditFixture(t, cashDay)

	one := f.pocket(t, "uno")
	two := f.pocket(t, "dos")
	f.fund(t, "1000")
	f.moveInto(t, one.ID, CashKindDeposit, "500", cashDay)
	f.moveInto(t, two.ID, CashKindDeposit, "500", cashDay)

	txn := f.mustTrade(t, fromPocket(buyInput(t, "3", "100", true, cashDay), one.ID))
	sameAmount(t, "the pocket that paid", f.balanceOfPocket(t, &one.ID), "200")

	moved := fromPocket(buyInput(t, "3", "100", true, cashDay), two.ID)
	if _, err := f.repo.UpdateTransaction(context.Background(), f.userID, txn.ID, moved); err != nil {
		t.Fatalf("UpdateTransaction: %v", err)
	}

	sameAmount(t, "the pocket that paid first", f.balanceOfPocket(t, &one.ID), "500")
	sameAmount(t, "the pocket that pays now", f.balanceOfPocket(t, &two.ID), "200")
}

// A fixed deposit is closed until it matures: no money leaves it to buy
// anything, just as none leaves it in a withdrawal.
func TestPurchaseCannotBePaidFromAFixedDeposit(t *testing.T) {
	f := newCreditFixture(t, cashDay)
	f.fund(t, "1000")

	matures := cashDay.AddDate(0, 6, 0)
	deposit := f.deposit(t, "cdt 180", "500", "10", cashDay, &matures, PostingAtMaturity)

	in := fromPocket(buyInput(t, "1", "100", true, cashDay), deposit.ID)
	if _, err := f.trade(t, in); !errors.Is(err, ErrCashPocketFixed) {
		t.Fatalf("error = %v, want ErrCashPocketFixed", err)
	}

	sameAmount(t, "the deposit", f.balanceOfPocket(t, &deposit.ID), "500")
}

// A pocket that does not exist, or belongs to another account, pays for
// nothing: the whole purchase is refused before anything is written.
func TestPurchaseFromAnUnknownPocketIsRefused(t *testing.T) {
	f := newCreditFixture(t, cashDay)
	f.fund(t, "1000")

	in := fromPocket(buyInput(t, "1", "100", true, cashDay), uuid.New())
	if _, err := f.trade(t, in); !errors.Is(err, ErrCashPocketNotFound) {
		t.Fatalf("error = %v, want ErrCashPocketNotFound", err)
	}

	sameAmount(t, "balance", f.balanceIn(t, money.USD), "1000")
}

// A drawer without the answer that qualifies it is a request that disagrees
// with itself: it is neither honoured nor ignored.
func TestAPocketWithoutPayingFromCashIsRefused(t *testing.T) {
	f := newCreditFixture(t, cashDay)
	f.fund(t, "1000")

	in := buyInput(t, "1", "100", false, cashDay)
	in.CashPocketID = uuid.New()

	if _, err := f.trade(t, in); !errors.Is(err, ErrNotPayableFromCash) {
		t.Fatalf("error = %v, want ErrNotPayableFromCash", err)
	}
}
