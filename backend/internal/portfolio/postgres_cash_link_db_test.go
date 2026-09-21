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

// Dividends and sales paid into cash (000049 to 000052), and purchases paid out
// of it (000053, 000054): a row on the platform's balance that follows its
// transaction. Same database contract as postgres_cash_db_test.go. The
// purchases have their own file, postgres_cash_purchase_db_test.go.

type creditFixture struct {
	cashFixture
	holdingID uuid.UUID
	assetID   uuid.UUID
	ticker    string
}

// newCreditFixture is newCashFixture plus a share on the same platform: ten
// units bought at 100 on opened, with a catalog price of 100 so the holding is
// worth what it cost and every gain below comes from the credits.
func newCreditFixture(t *testing.T, opened time.Time) creditFixture {
	t.Helper()

	f := creditFixture{
		cashFixture: newCashFixture(t),
		holdingID:   uuid.New(),
		assetID:     uuid.New(),
		ticker:      "CRD" + uuid.New().String()[:6],
	}
	assetID := f.assetID

	// A second teardown for the share. It runs before newCashFixture's, drops
	// the entries that point at the asset, and then the asset.
	dropFixture(t, f.pool, f.userID)(assetID)

	f.exec(t, `INSERT INTO assets (id, ticker, name, asset_type, exchange, currency, current_price)
	           VALUES ($1, $2, 'credit probe', 'stock', NULL, 'USD', 100)`, assetID, f.ticker)
	f.exec(t, `INSERT INTO portfolio_entries (id, portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
	           VALUES ($1, $2, $3, $4, 0, 100, 'USD', $5)`,
		f.holdingID, f.portfolioID, assetID, f.sourceID, opened)

	f.mustTrade(t, tradeInput(t, Buy, "10", "100", false, opened))

	return f
}

func tradeInput(t *testing.T, txnType TransactionType, quantity, price string, credit bool, date time.Time) TransactionInput {
	t.Helper()

	return TransactionInput{
		Type:            txnType,
		Quantity:        mustDecimal(t, quantity),
		Price:           mustUSD(t, price),
		Currency:        money.USD,
		TransactionDate: date,
		CreditCash:      credit,
	}
}

func dividendInput(t *testing.T, amount string, credit bool, date time.Time) TransactionInput {
	t.Helper()

	return tradeInput(t, Dividend, "1", amount, credit, date)
}

func (f creditFixture) trade(t *testing.T, in TransactionInput) (Transaction, error) {
	t.Helper()

	return f.repo.CreateTransaction(context.Background(), f.userID, f.holdingID, in)
}

func (f creditFixture) mustTrade(t *testing.T, in TransactionInput) Transaction {
	t.Helper()

	txn, err := f.trade(t, in)
	if err != nil {
		t.Fatalf("%s %s at %s (credit %v): %v", in.Type, in.Quantity, in.Price, in.CreditCash, err)
	}

	return txn
}

func (f creditFixture) mustDividend(t *testing.T, amount string, credit bool, date time.Time) Transaction {
	t.Helper()

	return f.mustTrade(t, dividendInput(t, amount, credit, date))
}

// credits lists every cash movement that holds a dividend or a sale.
func (f creditFixture) credits(t *testing.T) []CashMovement {
	t.Helper()

	movements, err := f.repo.GetCashMovementsPaginated(context.Background(), f.userID, 100, 0)
	if err != nil {
		t.Fatalf("GetCashMovementsPaginated: %v", err)
	}

	credits := make([]CashMovement, 0)
	for _, m := range movements {
		if m.Type.isCashLinked() {
			credits = append(credits, m)
		}
	}

	return credits
}

// mustCredit is the one credit the fixture should have.
func (f creditFixture) mustCredit(t *testing.T) CashMovement {
	t.Helper()

	credits := f.credits(t)
	if len(credits) != 1 {
		t.Fatalf("credits = %+v, want exactly one", credits)
	}

	return credits[0]
}

// balanceIn is what the account holds in cur, zero when there is no such
// balance.
func (f creditFixture) balanceIn(t *testing.T, cur money.Currency) string {
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

// lastPoint is the live point the growth series closes on today.
func (f creditFixture) lastPoint(t *testing.T, today time.Time) GrowthPoint {
	t.Helper()

	series, err := f.repo.GetPortfolioGrowthByPortfolioID(context.Background(), f.userID, f.portfolioID, false, time.Time{}, today)
	if err != nil {
		t.Fatalf("GetPortfolioGrowthByPortfolioID: %v", err)
	}
	if len(series) == 0 {
		t.Fatal("empty series")
	}

	return series[len(series)-1]
}

func TestDividendCreditLandsInThePlatformsCash(t *testing.T) {
	f := newCreditFixture(t, cashDay)

	txn := f.mustDividend(t, "25", true, cashDay.AddDate(0, 0, 1))
	if !txn.CashCredited {
		t.Error("CashCredited = false, want true")
	}

	sameAmount(t, "balance", f.balanceIn(t, money.USD), "25")

	credit := f.mustCredit(t)
	if credit.Kind != CashKindDividend || credit.Editable || credit.OriginTicker != f.ticker {
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
		if r.Type.isCashLinked() {
			t.Errorf("recent activity lists the credit %v beside its dividend", r.ID)
		}
	}
}

func TestDividendNotCreditedLeavesTheCashAlone(t *testing.T) {
	f := newCreditFixture(t, cashDay)

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
	f := newCreditFixture(t, cashDay)

	f.mustTrade(t, TransactionInput{
		Type:            Dividend,
		Quantity:        decimal.One,
		Price:           mustEUR(t, "10"),
		Currency:        money.EUR,
		FXRate:          mustDecimal(t, "1.1"),
		Fees:            mustUSD(t, "0.5"),
		FeesCurrency:    money.USD,
		TransactionDate: cashDay,
		CreditCash:      true,
	})

	credit := f.mustCredit(t)
	sameAmount(t, "credit", credit.Amount, "10.5")
	if credit.Currency != money.USD {
		t.Errorf("credit currency = %s, want the account's USD", credit.Currency)
	}
}

// The reason a credit carries a flow: the dividend's flow out of the holding
// and the credit's into the balance cancel, and the value the balance gained is
// return once.
func TestDividendCreditIsReturnOnce(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	f := newCreditFixture(t, today)

	f.mustDividend(t, "25", true, today)

	last := f.lastPoint(t, today)
	sameAmount(t, "value", last.TotalValue, "1025")
	sameAmount(t, "net flow", last.NetFlow, "1000")
}

// A dividend loaded with history is no return of the day it was typed in: the
// cash it left walks in with the balance.
func TestDividendCreditLoadedWithHistoryWalksInWithTheBalance(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	monthAgo := today.AddDate(0, -1, 0)
	f := newCreditFixture(t, monthAgo)

	f.mustDividend(t, "25", true, monthAgo)

	last := f.lastPoint(t, today)
	sameAmount(t, "value", last.TotalValue, "1025")
	sameAmount(t, "net flow", last.NetFlow, "1025")
}

func TestDividendEditRewritesItsCredit(t *testing.T) {
	f := newCreditFixture(t, cashDay)
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

	// The emptied balance is still the account's (000052's price fix), so the
	// credit lands back on it rather than being refused.
	if _, err := f.repo.UpdateTransaction(ctx, f.userID, txn.ID, dividendInput(t, "40", true, cashDay)); err != nil {
		t.Fatalf("credit again: %v", err)
	}
	sameAmount(t, "balance credited again", f.balanceIn(t, money.USD), "40")
	if again := f.mustCredit(t); again.EntryID != first.EntryID {
		t.Errorf("credit again landed on %v, want the balance %v", again.EntryID, first.EntryID)
	}

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

func TestCashCreditGoesWithItsTransaction(t *testing.T) {
	f := newCreditFixture(t, cashDay)
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

	// Deleting the position takes the credits of all its dividends and sales.
	f.mustDividend(t, "5", true, cashDay)
	f.mustTrade(t, tradeInput(t, Sell, "1", "7", true, cashDay))
	sameAmount(t, "balance with two credits", f.balanceIn(t, money.USD), "112")

	if _, err := f.repo.DeletePortfolioEntry(ctx, f.userID, f.holdingID); err != nil {
		t.Fatalf("DeletePortfolioEntry: %v", err)
	}
	sameAmount(t, "balance without the position", f.balanceIn(t, money.USD), "100")
}

// Money the balance already spent cannot be taken back out of it: the same
// refusal a withdrawal gets.
func TestCashCreditCannotBeTakenBackOnceSpent(t *testing.T) {
	f := newCreditFixture(t, cashDay)
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

func TestCashCreditIsOnlyWrittenThroughItsTransaction(t *testing.T) {
	f := newCreditFixture(t, cashDay)
	ctx := context.Background()

	f.mustDividend(t, "25", true, cashDay)
	f.mustTrade(t, tradeInput(t, Sell, "1", "100", true, cashDay))

	credits := f.credits(t)
	if len(credits) != 2 {
		t.Fatalf("credits = %+v, want the dividend's and the sale's", credits)
	}

	for _, credit := range credits {
		if err := f.repo.DeleteCashMovement(ctx, f.userID, credit.ID); !errors.Is(err, ErrCashCreditLinked) {
			t.Errorf("DeleteCashMovement(%s) = %v, want ErrCashCreditLinked", credit.Kind, err)
		}
		if _, err := f.repo.UpdateCashMovement(ctx, f.userID, credit.ID, CashMovementInput{
			Kind: CashKindDeposit, Amount: mustDecimal(t, "25"), Date: cashDay,
		}); !errors.Is(err, ErrCashCreditLinked) {
			t.Errorf("UpdateCashMovement(%s) = %v, want ErrCashCreditLinked", credit.Kind, err)
		}
		if err := f.repo.DeleteTransaction(ctx, f.userID, credit.ID); !errors.Is(err, ErrCashCreditLinked) {
			t.Errorf("DeleteTransaction(%s) = %v, want ErrCashCreditLinked", credit.Kind, err)
		}
		if _, err := f.repo.UpdateTransaction(ctx, f.userID, credit.ID, tradeInput(t, TransferIn, "1", "1", false, cashDay)); !errors.Is(err, ErrCashCreditLinked) {
			t.Errorf("UpdateTransaction(%s) = %v, want ErrCashCreditLinked", credit.Kind, err)
		}
	}

	for _, byHand := range []TransactionType{CashDividend, CashSale} {
		in := tradeInput(t, byHand, "1", "1", false, cashDay)
		if _, err := f.repo.CreateTransaction(ctx, f.userID, credits[0].EntryID, in); !errors.Is(err, ErrCashCreditLinked) {
			t.Errorf("%s by hand = %v, want ErrCashCreditLinked", byHand, err)
		}
	}

	sameAmount(t, "balance", f.balanceIn(t, money.USD), "125")
}

func TestCashCreditRefusedWhereItCannotLand(t *testing.T) {
	f := newCreditFixture(t, cashDay)
	ctx := context.Background()

	fee := dividendInput(t, "0", true, cashDay)
	fee.Type = Fee
	fee.Fees = mustUSD(t, "1")
	if _, err := f.trade(t, fee); !errors.Is(err, ErrNotCreditable) {
		t.Errorf("crediting a fee = %v, want ErrNotCreditable", err)
	}

	eaten := dividendInput(t, "1", true, cashDay)
	eaten.Fees = mustUSD(t, "1")
	if _, err := f.trade(t, eaten); !errors.Is(err, ErrNotCreditable) {
		t.Errorf("crediting a dividend its fees ate = %v, want ErrNotCreditable", err)
	}

	deposit := f.mustMove(t, CashKindDeposit, "10", cashDay)
	for _, onCash := range []TransactionInput{
		dividendInput(t, "1", true, cashDay),
		tradeInput(t, Sell, "1", "1", true, cashDay),
	} {
		if _, err := f.repo.CreateTransaction(ctx, f.userID, deposit.EntryID, onCash); !errors.Is(err, ErrNotCreditable) {
			t.Errorf("crediting a %s recorded on cash = %v, want ErrNotCreditable", onCash.Type, err)
		}
	}

	// The refused writes left nothing behind: not the transaction, not a credit.
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
func TestCashCreditMovesWithTheSettlementCurrency(t *testing.T) {
	f := newCreditFixture(t, cashDay)
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

	var creditedFrom uuid.UUID
	if err := f.pool.QueryRow(ctx, `SELECT credited_from FROM transactions WHERE id = $1`, credit.ID).Scan(&creditedFrom); err != nil {
		t.Fatalf("read link: %v", err)
	}
	if creditedFrom != txn.ID {
		t.Errorf("credit holds %v, want the dividend %v", creditedFrom, txn.ID)
	}
}

// A credit a snapshot already holds is kept when its transaction goes, like any
// row that moved a quantity (000038): the day it left must not read as a loss.
func TestCashCreditASnapshotSawIsRetiredWhenItsTransactionIsDeleted(t *testing.T) {
	f := newCreditFixture(t, cashDay)
	ctx := context.Background()

	dividend := f.mustDividend(t, "25", true, cashDay)
	sale := f.mustTrade(t, tradeInput(t, Sell, "1", "100", true, cashDay))

	f.exec(t, `INSERT INTO portfolio_snapshots (portfolio_id, snapshot_date, total_value, currency, created_at)
	           VALUES ($1, CURRENT_DATE - 1, 1025, 'USD', NOW() + INTERVAL '1 minute')`, f.portfolioID)

	for _, txn := range []Transaction{dividend, sale} {
		if err := f.repo.DeleteTransaction(ctx, f.userID, txn.ID); err != nil {
			t.Fatalf("DeleteTransaction(%s): %v", txn.Type, err)
		}
	}

	for _, credit := range []TransactionType{CashDividend, CashSale} {
		var retired int
		if err := f.pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM retired_transactions WHERE portfolio_id = $1 AND type = $2::transaction_type
		`, f.portfolioID, credit).Scan(&retired); err != nil {
			t.Fatalf("count retired: %v", err)
		}
		if retired != 1 {
			t.Errorf("retired %s versions = %d, want 1", credit, retired)
		}
	}
}

// A sale paid into cash: the proceeds land in the balance, and they cost what
// the sold shares cost, so the gain the sale realised stays gain and the
// capital stays what was paid.
func TestSaleCreditLandsInThePlatformsCashAtTheSharesCost(t *testing.T) {
	f := newCreditFixture(t, cashDay)

	sell := tradeInput(t, Sell, "4", "150", true, cashDay.AddDate(0, 0, 1))
	sell.Fees = mustUSD(t, "2")
	txn := f.mustTrade(t, sell)
	if !txn.CashCredited {
		t.Error("CashCredited = false, want true")
	}

	// What the account received: 4 × 150, less the commission.
	sameAmount(t, "balance", f.balanceIn(t, money.USD), "598")

	credit := f.mustCredit(t)
	if credit.Kind != CashKindSale || credit.Type != CashSale || credit.Editable || credit.OriginTicker != f.ticker {
		t.Errorf("credit = %+v, want a read-only sale of %s", credit, f.ticker)
	}

	// Six shares left at 100 and 598 in cash; the capital is the 1000 paid for
	// the ten shares, and the gain is the 198 the sale realised.
	s := f.summary(t)
	aboutAmount(t, "portfolio value", s.TotalMarketValue, "1198")
	aboutAmount(t, "portfolio cost", s.TotalCostBase, "1000")
	aboutAmount(t, "portfolio gain", s.TotalGainLoss, "198")

	recent, err := f.repo.GetRecentTransactionsByUserID(context.Background(), f.userID, 50)
	if err != nil {
		t.Fatalf("GetRecentTransactionsByUserID: %v", err)
	}
	for _, r := range recent {
		if r.Type.isCashLinked() {
			t.Errorf("recent activity lists the credit %v beside its sale", r.ID)
		}
	}
}

// The sale's flow out of the holding and its credit's into the balance cancel:
// what is left is the value the portfolio gained, here the 200 the shares sold
// for over the price the series held them at.
func TestSaleCreditIsReturnOnce(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	f := newCreditFixture(t, today)

	f.mustTrade(t, tradeInput(t, Sell, "4", "150", true, today))

	last := f.lastPoint(t, today)
	sameAmount(t, "value", last.TotalValue, "1200")
	sameAmount(t, "net flow", last.NetFlow, "1000")
}

// The cost the proceeds carry is the position's average cost, which a later
// purchase moves; the credit follows it, so the capital stays the sum of what
// every purchase cost.
func TestSaleCreditCostFollowsThePositionsPurchases(t *testing.T) {
	f := newCreditFixture(t, cashDay)

	f.mustTrade(t, tradeInput(t, Sell, "5", "120", true, cashDay.AddDate(0, 0, 1)))
	aboutAmount(t, "cost after the sale", f.summary(t).TotalCostBase, "1000")

	// Ten more at 130: the average is (1000 + 1300) / 20 = 115, the five sold
	// carried 575 out, and the fifteen held cost 1725.
	buy := f.mustTrade(t, tradeInput(t, Buy, "10", "130", false, cashDay.AddDate(0, 0, 2)))
	aboutAmount(t, "cost after a purchase", f.summary(t).TotalCostBase, "2300")

	// And back when the purchase goes.
	if err := f.repo.DeleteTransaction(context.Background(), f.userID, buy.ID); err != nil {
		t.Fatalf("DeleteTransaction: %v", err)
	}
	aboutAmount(t, "cost after deleting it", f.summary(t).TotalCostBase, "1000")
	sameAmount(t, "balance", f.balanceIn(t, money.USD), "600")
}

func TestSaleEditRewritesItsCredit(t *testing.T) {
	f := newCreditFixture(t, cashDay)
	ctx := context.Background()

	txn := f.mustTrade(t, tradeInput(t, Sell, "2", "110", true, cashDay))
	first := f.mustCredit(t)
	sameAmount(t, "proceeds", first.Amount, "220")

	if _, err := f.repo.UpdateTransaction(ctx, f.userID, txn.ID, tradeInput(t, Sell, "3", "110", true, cashDay)); err != nil {
		t.Fatalf("resize: %v", err)
	}
	resized := f.mustCredit(t)
	if resized.ID != first.ID {
		t.Errorf("credit %v was replaced by %v, want it rewritten in place", first.ID, resized.ID)
	}
	sameAmount(t, "proceeds after resize", resized.Amount, "330")

	// A sale edited into a dividend changes what its money is: the credit is
	// replaced by one of the other kind.
	if _, err := f.repo.UpdateTransaction(ctx, f.userID, txn.ID, dividendInput(t, "12", true, cashDay)); err != nil {
		t.Fatalf("turn into a dividend: %v", err)
	}
	dividend := f.mustCredit(t)
	if dividend.Kind != CashKindDividend {
		t.Errorf("credit kind = %s, want dividend", dividend.Kind)
	}
	sameAmount(t, "balance", f.balanceIn(t, money.USD), "12")
}

// The pre-existing half of 000052's price fix: interest-only balances had the
// same trap as credit-only ones. Deleting the only interest left the balance at
// the walk's price of zero with nothing in it, the writers disowned it, and the
// account refused every deposit after.
func TestCashBalanceEmptiedOfItsInterestTakesTheNextDeposit(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	interest := f.mustMove(t, CashKindInterest, "5", cashDay)
	if err := f.repo.DeleteCashMovement(ctx, f.userID, interest.ID); err != nil {
		t.Fatalf("DeleteCashMovement: %v", err)
	}

	deposit, err := f.move(t, CashKindDeposit, "100", "", cashDay.AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("deposit after the interest went: %v", err)
	}
	if deposit.EntryID != interest.EntryID {
		t.Errorf("deposit landed on %v, want the balance %v", deposit.EntryID, interest.EntryID)
	}

	sameAmount(t, "balance", f.balance(t), "100")
	aboutAmount(t, "portfolio gain", f.summary(t).TotalGainLoss, "0")
}

// The importer reported every file as zero rows imported: the count lived in a
// variable the transaction closure shadowed.
func TestImportEntryTransactionsReportsWhatItWrote(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	ticker := "IMP" + uuid.New().String()[:6]
	row := ImportTransactionRow{
		RowNumber:    2,
		Ticker:       ticker,
		AssetName:    "import probe",
		AssetType:    "stock",
		Type:         Buy,
		Quantity:     mustDecimal(t, "3"),
		Price:        mustUSD(t, "10"),
		Fees:         mustUSD(t, "0"),
		Currency:     money.USD,
		FXRate:       decimal.One,
		CostCurrency: money.USD,
		Date:         cashDay,
	}
	second := row
	second.RowNumber = 3

	n, err := f.repo.ImportEntryTransactions(ctx, f.userID, f.portfolioID, f.sourceID, []ImportTransactionRow{row, second})

	// The import created the asset; hand it to a teardown that runs before the
	// fixture's, so the entry pointing at it is gone first.
	var assetID uuid.UUID
	if lookup := f.pool.QueryRow(ctx, `SELECT id FROM assets WHERE ticker = $1`, ticker).Scan(&assetID); lookup == nil {
		dropFixture(t, f.pool, f.userID)(assetID)
	}

	if err != nil {
		t.Fatalf("ImportEntryTransactions: %v", err)
	}
	if n != 2 {
		t.Errorf("imported = %d, want 2", n)
	}
}
