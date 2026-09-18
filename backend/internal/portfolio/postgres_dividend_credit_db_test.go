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

// A dividend paid into cash (000049, 000050): a credit on the platform's
// balance that follows its dividend. Same database contract as
// postgres_cash_db_test.go.

type dividendFixture struct {
	cashFixture
	holdingID uuid.UUID
	ticker    string
}

// newDividendFixture is newCashFixture plus a share on the same platform: ten
// units bought at 100 on opened, with a catalog price of 100 so the holding is
// worth what it cost and every gain below is the dividend's.
func newDividendFixture(t *testing.T, opened time.Time) dividendFixture {
	t.Helper()

	f := dividendFixture{
		cashFixture: newCashFixture(t),
		holdingID:   uuid.New(),
		ticker:      "DIV" + uuid.New().String()[:6],
	}
	assetID := uuid.New()

	// A second teardown for the share. It runs before newCashFixture's, drops
	// the entries that point at the asset, and then the asset.
	dropFixture(t, f.pool, f.userID)(assetID)

	f.exec(t, `INSERT INTO assets (id, ticker, name, asset_type, exchange, currency, current_price)
	           VALUES ($1, $2, 'dividend probe', 'stock', NULL, 'USD', 100)`, assetID, f.ticker)
	f.exec(t, `INSERT INTO portfolio_entries (id, portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
	           VALUES ($1, $2, $3, $4, 0, 100, 'USD', $5)`,
		f.holdingID, f.portfolioID, assetID, f.sourceID, opened)

	if _, err := f.repo.CreateTransaction(context.Background(), f.userID, f.holdingID, TransactionInput{
		Type:            Buy,
		Quantity:        mustDecimal(t, "10"),
		Price:           mustUSD(t, "100"),
		Currency:        money.USD,
		TransactionDate: opened,
	}); err != nil {
		t.Fatalf("opening buy: %v", err)
	}

	return f
}

func (f dividendFixture) dividend(t *testing.T, amount string, credit bool, date time.Time) (Transaction, error) {
	t.Helper()

	return f.repo.CreateTransaction(context.Background(), f.userID, f.holdingID, dividendInput(t, amount, credit, date))
}

func (f dividendFixture) mustDividend(t *testing.T, amount string, credit bool, date time.Time) Transaction {
	t.Helper()

	txn, err := f.dividend(t, amount, credit, date)
	if err != nil {
		t.Fatalf("dividend %s (credit %v): %v", amount, credit, err)
	}

	return txn
}

func dividendInput(t *testing.T, amount string, credit bool, date time.Time) TransactionInput {
	t.Helper()

	return TransactionInput{
		Type:            Dividend,
		Quantity:        decimal.One,
		Price:           mustUSD(t, amount),
		Currency:        money.USD,
		TransactionDate: date,
		CreditCash:      credit,
	}
}

// credits lists every cash movement that pays a dividend.
func (f dividendFixture) credits(t *testing.T) []CashMovement {
	t.Helper()

	movements, err := f.repo.GetCashMovementsPaginated(context.Background(), f.userID, 100, 0)
	if err != nil {
		t.Fatalf("GetCashMovementsPaginated: %v", err)
	}

	credits := make([]CashMovement, 0)
	for _, m := range movements {
		if m.Type == CashDividend {
			credits = append(credits, m)
		}
	}

	return credits
}

// mustCredit is the one credit the fixture should have.
func (f dividendFixture) mustCredit(t *testing.T) CashMovement {
	t.Helper()

	credits := f.credits(t)
	if len(credits) != 1 {
		t.Fatalf("credits = %+v, want exactly one", credits)
	}

	return credits[0]
}

// balanceIn is what the account holds in cur, zero when there is no such
// balance.
func (f dividendFixture) balanceIn(t *testing.T, cur money.Currency) string {
	t.Helper()

	balances, err := f.repo.GetCashBalancesByUserID(context.Background(), f.userID, money.XXX)
	if err != nil {
		t.Fatalf("GetCashBalancesByUserID: %v", err)
	}

	total := decimal.Zero
	for _, b := range balances {
		if b.Currency == cur {
			total = total.Add(mustDecimal(t, b.Balance))
		}
	}

	return total.String()
}

func TestDividendCreditLandsInThePlatformsCash(t *testing.T) {
	f := newDividendFixture(t, cashDay)

	txn := f.mustDividend(t, "25", true, cashDay.AddDate(0, 0, 1))
	if !txn.CashCredited {
		t.Error("CashCredited = false, want true")
	}

	sameAmount(t, "balance", f.balanceIn(t, money.USD), "25")

	credit := f.mustCredit(t)
	if credit.Kind != CashKindDividend || credit.Editable || credit.DividendTicker != f.ticker {
		t.Errorf("credit = %+v, want a read-only dividend of %s", credit, f.ticker)
	}
	if credit.Ticker != "CASH-USD" || credit.SourceID != f.sourceID || credit.PortfolioID != f.portfolioID {
		t.Errorf("credit landed on %s of %v/%v, want CASH-USD of the holding's platform", credit.Ticker, credit.PortfolioID, credit.SourceID)
	}
	if !credit.Date.Equal(cashDay.AddDate(0, 0, 1)) {
		t.Errorf("credit date = %v, want the dividend's", credit.Date)
	}

	// The credit cost nothing: the portfolio's gain is the dividend, and the
	// share, worth what it cost, adds none.
	s := f.summary(t)
	aboutAmount(t, "portfolio value", s.TotalMarketValue, "1025")
	aboutAmount(t, "portfolio gain", s.TotalGainLoss, "25")

	listed, err := f.repo.GetTransactionsByEntryID(context.Background(), f.userID, f.holdingID)
	if err != nil {
		t.Fatalf("GetTransactionsByEntryID: %v", err)
	}
	for _, l := range listed {
		if l.ID == txn.ID && !l.CashCredited {
			t.Error("the dividend reads back as not credited")
		}
	}

	// One payment, one line in the activity: the dividend, not its credit.
	recent, err := f.repo.GetRecentTransactionsByUserID(context.Background(), f.userID, 50)
	if err != nil {
		t.Fatalf("GetRecentTransactionsByUserID: %v", err)
	}
	for _, r := range recent {
		if r.Type == CashDividend {
			t.Errorf("recent activity lists the credit %v beside its dividend", r.ID)
		}
	}
}

func TestDividendNotCreditedLeavesTheCashAlone(t *testing.T) {
	f := newDividendFixture(t, cashDay)

	txn := f.mustDividend(t, "25", false, cashDay)
	if txn.CashCredited {
		t.Error("CashCredited = true, want false")
	}

	if got := f.credits(t); len(got) != 0 {
		t.Errorf("credits = %+v, want none", got)
	}
	sameAmount(t, "balance", f.balanceIn(t, money.USD), "0")
}

// A dividend settled in another currency is credited in the account's, at the
// rate it converted at and less the commission billed to the account.
func TestDividendCreditIsWhatTheAccountReceived(t *testing.T) {
	f := newDividendFixture(t, cashDay)

	eur, err := money.NewMoneyFromString("10", money.EUR)
	if err != nil {
		t.Fatalf("NewMoneyFromString: %v", err)
	}

	if _, err := f.repo.CreateTransaction(context.Background(), f.userID, f.holdingID, TransactionInput{
		Type:            Dividend,
		Quantity:        decimal.One,
		Price:           eur,
		Currency:        money.EUR,
		FXRate:          mustDecimal(t, "1.1"),
		Fees:            mustUSD(t, "0.5"),
		FeesCurrency:    money.USD,
		TransactionDate: cashDay,
		CreditCash:      true,
	}); err != nil {
		t.Fatalf("dividend in EUR: %v", err)
	}

	credit := f.mustCredit(t)
	sameAmount(t, "credit", credit.Amount, "10.5")
	if credit.Currency != money.USD {
		t.Errorf("credit currency = %s, want the account's USD", credit.Currency)
	}
}

// The reason cash_dividend carries a flow: the dividend's flow out of the
// holding and the credit's into the balance cancel, and the value the balance
// gained is return once.
func TestDividendCreditIsReturnOnce(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	f := newDividendFixture(t, today)

	f.mustDividend(t, "25", true, today)

	series, err := f.repo.GetPortfolioGrowthByPortfolioID(context.Background(), f.userID, f.portfolioID, false, time.Time{}, today)
	if err != nil {
		t.Fatalf("GetPortfolioGrowthByPortfolioID: %v", err)
	}
	if len(series) == 0 {
		t.Fatal("empty series")
	}

	last := series[len(series)-1]
	sameAmount(t, "value", last.TotalValue, "1025")
	sameAmount(t, "net flow", last.NetFlow, "1000")
}

// A dividend loaded with history is no return of the day it was typed in: the
// cash it left walks in with the balance.
func TestDividendCreditLoadedWithHistoryWalksInWithTheBalance(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	monthAgo := today.AddDate(0, -1, 0)
	f := newDividendFixture(t, monthAgo)

	f.mustDividend(t, "25", true, monthAgo)

	series, err := f.repo.GetPortfolioGrowthByPortfolioID(context.Background(), f.userID, f.portfolioID, false, time.Time{}, today)
	if err != nil {
		t.Fatalf("GetPortfolioGrowthByPortfolioID: %v", err)
	}
	if len(series) == 0 {
		t.Fatal("empty series")
	}

	last := series[len(series)-1]
	sameAmount(t, "value", last.TotalValue, "1025")
	sameAmount(t, "net flow", last.NetFlow, "1025")
}

func TestDividendEditRewritesItsCredit(t *testing.T) {
	f := newDividendFixture(t, cashDay)
	ctx := context.Background()

	txn := f.mustDividend(t, "25", true, cashDay)
	first := f.mustCredit(t)

	edited, err := f.repo.UpdateTransaction(ctx, f.userID, txn.ID, dividendInput(t, "40", true, cashDay.AddDate(0, 0, 2)))
	if err != nil {
		t.Fatalf("resize: %v", err)
	}
	if !edited.CashCredited {
		t.Error("CashCredited after resize = false, want true")
	}

	resized := f.mustCredit(t)
	if resized.ID != first.ID {
		t.Errorf("credit %v was replaced by %v, want it rewritten in place", first.ID, resized.ID)
	}
	sameAmount(t, "credit", resized.Amount, "40")
	if !resized.Date.Equal(cashDay.AddDate(0, 0, 2)) {
		t.Errorf("credit date = %v, want the dividend's new date", resized.Date)
	}

	if _, err := f.repo.UpdateTransaction(ctx, f.userID, txn.ID, dividendInput(t, "40", false, cashDay)); err != nil {
		t.Fatalf("uncredit: %v", err)
	}
	if got := f.credits(t); len(got) != 0 {
		t.Errorf("credits after uncrediting = %+v, want none", got)
	}
	sameAmount(t, "balance after uncrediting", f.balanceIn(t, money.USD), "0")

	if _, err := f.repo.UpdateTransaction(ctx, f.userID, txn.ID, dividendInput(t, "40", true, cashDay)); err != nil {
		t.Fatalf("credit again: %v", err)
	}
	sameAmount(t, "balance credited again", f.balanceIn(t, money.USD), "40")

	// A dividend turned into something else keeps no credit.
	fee := dividendInput(t, "0", false, cashDay)
	fee.Type = Fee
	fee.Fees = mustUSD(t, "1")
	if _, err := f.repo.UpdateTransaction(ctx, f.userID, txn.ID, fee); err != nil {
		t.Fatalf("turn into a fee: %v", err)
	}
	if got := f.credits(t); len(got) != 0 {
		t.Errorf("credits of a fee = %+v, want none", got)
	}
}

func TestDividendCreditGoesWithItsDividend(t *testing.T) {
	f := newDividendFixture(t, cashDay)
	ctx := context.Background()

	f.mustMove(t, CashKindDeposit, "100", cashDay)
	txn := f.mustDividend(t, "25", true, cashDay)

	if err := f.repo.DeleteTransaction(ctx, f.userID, txn.ID); err != nil {
		t.Fatalf("DeleteTransaction: %v", err)
	}

	if got := f.credits(t); len(got) != 0 {
		t.Errorf("credits = %+v, want none", got)
	}
	sameAmount(t, "balance", f.balanceIn(t, money.USD), "100")

	// Deleting the position takes the credits of all its dividends.
	f.mustDividend(t, "5", true, cashDay)
	f.mustDividend(t, "7", true, cashDay)
	sameAmount(t, "balance with two credits", f.balanceIn(t, money.USD), "112")

	if _, err := f.repo.DeletePortfolioEntry(ctx, f.userID, f.holdingID); err != nil {
		t.Fatalf("DeletePortfolioEntry: %v", err)
	}
	sameAmount(t, "balance without the position", f.balanceIn(t, money.USD), "100")
}

// A dividend whose money was spent cannot take it back out of the balance: the
// same refusal a withdrawal gets.
func TestDividendCreditCannotBeTakenBackOnceSpent(t *testing.T) {
	f := newDividendFixture(t, cashDay)
	ctx := context.Background()

	txn := f.mustDividend(t, "25", true, cashDay)
	f.mustMove(t, CashKindWithdrawal, "20", cashDay.AddDate(0, 0, 1))

	if err := f.repo.DeleteTransaction(ctx, f.userID, txn.ID); !errors.Is(err, ErrInsufficientCash) {
		t.Errorf("deleting a spent dividend = %v, want ErrInsufficientCash", err)
	}
	if _, err := f.repo.UpdateTransaction(ctx, f.userID, txn.ID, dividendInput(t, "10", true, cashDay)); !errors.Is(err, ErrInsufficientCash) {
		t.Errorf("shrinking a spent dividend = %v, want ErrInsufficientCash", err)
	}
	if _, err := f.repo.UpdateTransaction(ctx, f.userID, txn.ID, dividendInput(t, "25", false, cashDay)); !errors.Is(err, ErrInsufficientCash) {
		t.Errorf("uncrediting a spent dividend = %v, want ErrInsufficientCash", err)
	}
	if _, err := f.repo.DeletePortfolioEntry(ctx, f.userID, f.holdingID); !errors.Is(err, ErrInsufficientCash) {
		t.Errorf("deleting the position of a spent dividend = %v, want ErrInsufficientCash", err)
	}

	// Nothing moved on the way.
	sameAmount(t, "balance", f.balanceIn(t, money.USD), "5")
	f.mustCredit(t)

	// Once the balance holds the money again, the dividend can go.
	f.mustMove(t, CashKindDeposit, "20", cashDay.AddDate(0, 0, 2))
	if err := f.repo.DeleteTransaction(ctx, f.userID, txn.ID); err != nil {
		t.Fatalf("DeleteTransaction after the deposit: %v", err)
	}
	sameAmount(t, "balance after the deletion", f.balanceIn(t, money.USD), "0")
}

func TestDividendCreditIsOnlyWrittenThroughItsDividend(t *testing.T) {
	f := newDividendFixture(t, cashDay)
	ctx := context.Background()

	f.mustDividend(t, "25", true, cashDay)
	credit := f.mustCredit(t)

	if err := f.repo.DeleteCashMovement(ctx, f.userID, credit.ID); !errors.Is(err, ErrDividendCreditLinked) {
		t.Errorf("DeleteCashMovement(credit) = %v, want ErrDividendCreditLinked", err)
	}
	if _, err := f.repo.UpdateCashMovement(ctx, f.userID, credit.ID, CashMovementInput{
		Kind: CashKindDeposit, Amount: mustDecimal(t, "25"), Date: cashDay,
	}); !errors.Is(err, ErrDividendCreditLinked) {
		t.Errorf("UpdateCashMovement(credit) = %v, want ErrDividendCreditLinked", err)
	}
	if err := f.repo.DeleteTransaction(ctx, f.userID, credit.ID); !errors.Is(err, ErrDividendCreditLinked) {
		t.Errorf("DeleteTransaction(credit) = %v, want ErrDividendCreditLinked", err)
	}

	deposit := dividendInput(t, "1", false, cashDay)
	deposit.Type = TransferIn
	if _, err := f.repo.UpdateTransaction(ctx, f.userID, credit.ID, deposit); !errors.Is(err, ErrDividendCreditLinked) {
		t.Errorf("UpdateTransaction(credit) = %v, want ErrDividendCreditLinked", err)
	}

	byHand := dividendInput(t, "1", false, cashDay)
	byHand.Type = CashDividend
	if _, err := f.repo.CreateTransaction(ctx, f.userID, credit.EntryID, byHand); !errors.Is(err, ErrDividendCreditLinked) {
		t.Errorf("cash_dividend by hand = %v, want ErrDividendCreditLinked", err)
	}

	sameAmount(t, "balance", f.balanceIn(t, money.USD), "25")
}

func TestDividendCreditRefusedWhereItCannotLand(t *testing.T) {
	f := newDividendFixture(t, cashDay)
	ctx := context.Background()

	fee := dividendInput(t, "0", true, cashDay)
	fee.Type = Fee
	fee.Fees = mustUSD(t, "1")
	if _, err := f.repo.CreateTransaction(ctx, f.userID, f.holdingID, fee); !errors.Is(err, ErrDividendNotCreditable) {
		t.Errorf("crediting a fee = %v, want ErrDividendNotCreditable", err)
	}

	eaten := dividendInput(t, "1", true, cashDay)
	eaten.Fees = mustUSD(t, "1")
	if _, err := f.repo.CreateTransaction(ctx, f.userID, f.holdingID, eaten); !errors.Is(err, ErrDividendNotCreditable) {
		t.Errorf("crediting a dividend its fees ate = %v, want ErrDividendNotCreditable", err)
	}

	deposit := f.mustMove(t, CashKindDeposit, "10", cashDay)
	if _, err := f.repo.CreateTransaction(ctx, f.userID, deposit.EntryID, dividendInput(t, "1", true, cashDay)); !errors.Is(err, ErrDividendNotCreditable) {
		t.Errorf("crediting a dividend recorded on cash = %v, want ErrDividendNotCreditable", err)
	}

	// The refused writes left nothing behind: not the dividend, not a credit.
	listed, err := f.repo.GetTransactionsByEntryID(ctx, f.userID, f.holdingID)
	if err != nil {
		t.Fatalf("GetTransactionsByEntryID: %v", err)
	}
	if len(listed) != 1 {
		t.Errorf("holding transactions = %d, want only the opening buy", len(listed))
	}
	sameAmount(t, "balance", f.balanceIn(t, money.USD), "10")
}

// Restating the position in the currency its account really settled in moves
// the credit to the balance kept in that currency, at the new rate.
func TestDividendCreditMovesWithTheSettlementCurrency(t *testing.T) {
	f := newDividendFixture(t, cashDay)
	ctx := context.Background()

	txn := f.mustDividend(t, "25", true, cashDay)

	txns, err := f.repo.GetTransactionsByEntryID(ctx, f.userID, f.holdingID)
	if err != nil {
		t.Fatalf("GetTransactionsByEntryID: %v", err)
	}

	rates := make(map[uuid.UUID]decimal.Decimal, len(txns))
	for _, l := range txns {
		rates[l.ID] = mustDecimal(t, "4000")
	}

	if _, err := f.repo.ChangeEntrySettlement(ctx, f.userID, f.holdingID, money.COP, rates); err != nil {
		t.Fatalf("ChangeEntrySettlement: %v", err)
	}

	sameAmount(t, "USD balance", f.balanceIn(t, money.USD), "0")
	sameAmount(t, "COP balance", f.balanceIn(t, money.COP), "100000")

	credit := f.mustCredit(t)
	if credit.Currency != money.COP {
		t.Errorf("credit currency = %s, want COP", credit.Currency)
	}

	var dividendID uuid.UUID
	if err := f.pool.QueryRow(ctx, `SELECT dividend_id FROM transactions WHERE id = $1`, credit.ID).Scan(&dividendID); err != nil {
		t.Fatalf("read link: %v", err)
	}
	if dividendID != txn.ID {
		t.Errorf("credit pays %v, want the dividend %v", dividendID, txn.ID)
	}
}

// A credit a snapshot already holds is kept when its dividend goes, like any
// row that moved a quantity (000038): the day it left must not read as a loss.
func TestDividendCreditASnapshotSawIsRetiredWhenItsDividendIsDeleted(t *testing.T) {
	f := newDividendFixture(t, cashDay)
	ctx := context.Background()

	txn := f.mustDividend(t, "25", true, cashDay)

	f.exec(t, `INSERT INTO portfolio_snapshots (portfolio_id, snapshot_date, total_value, currency, created_at)
	           VALUES ($1, CURRENT_DATE - 1, 1025, 'USD', NOW() + INTERVAL '1 minute')`, f.portfolioID)

	if err := f.repo.DeleteTransaction(ctx, f.userID, txn.ID); err != nil {
		t.Fatalf("DeleteTransaction: %v", err)
	}

	var retired int
	if err := f.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM retired_transactions WHERE portfolio_id = $1 AND type = 'cash_dividend'
	`, f.portfolioID).Scan(&retired); err != nil {
		t.Fatalf("count retired: %v", err)
	}
	if retired != 1 {
		t.Errorf("retired cash_dividend versions = %d, want 1", retired)
	}
}
