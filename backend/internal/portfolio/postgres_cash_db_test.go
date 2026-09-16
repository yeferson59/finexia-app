package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// A cash balance means what the triggers and the SQL functions of 000041 say
// it means, so these run against a real database with the migrations applied,
// and skip without one:
//
//	TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/finexia go test ./internal/portfolio/

type cashFixture struct {
	pool        *pgxpool.Pool
	repo        *PostgresRepository
	userID      uuid.UUID
	portfolioID uuid.UUID
	sourceID    uuid.UUID
}

// newCashFixture plants a user with one USD portfolio and one platform, and no
// cash yet.
func newCashFixture(t *testing.T) cashFixture {
	t.Helper()
	pool := growthTestPool(t)

	f := cashFixture{
		pool:        pool,
		repo:        NewPostgresRepository(pool),
		userID:      uuid.New(),
		portfolioID: uuid.New(),
		sourceID:    uuid.New(),
	}

	// The cash assets are shared catalog rows every account's balances point
	// at, so none is handed to the teardown.
	dropFixture(t, pool, f.userID)

	f.exec(t, `INSERT INTO users (id, name, email, role_id, preferred_currency)
	           VALUES ($1, 'cash probe', $2, (SELECT id FROM roles WHERE name = 'customer'), 'USD')`,
		f.userID, f.userID.String()+"@probe.test")
	f.exec(t, `INSERT INTO investment_sources (id, user_id, name, source_type)
	           VALUES ($1, $2, 'bank', 'neobank')`, f.sourceID, f.userID)
	f.exec(t, `INSERT INTO portfolios (id, user_id, name, type, risk_id, base_currency)
	           VALUES ($1, $2, 'savings', 'cash', (SELECT id FROM risks LIMIT 1), 'USD')`,
		f.portfolioID, f.userID)

	return f
}

func (f cashFixture) exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	if _, err := f.pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
}

func (f cashFixture) move(t *testing.T, kind CashMovementKind, amount, fees string, date time.Time) (CashMovement, error) {
	t.Helper()

	in := CashMovementInput{Kind: kind, Amount: mustDecimal(t, amount), Currency: money.USD, Date: date}
	if fees != "" {
		in.Fees = mustDecimal(t, fees)
	}

	return f.repo.CreateCashMovement(context.Background(), f.userID, f.portfolioID, f.sourceID, uuid.UUID{}, in)
}

func (f cashFixture) mustMove(t *testing.T, kind CashMovementKind, amount string, date time.Time) CashMovement {
	t.Helper()

	m, err := f.move(t, kind, amount, "", date)
	if err != nil {
		t.Fatalf("CreateCashMovement(%s %s): %v", kind, amount, err)
	}

	return m
}

// balance is what the account's only cash position holds.
func (f cashFixture) balance(t *testing.T) string {
	t.Helper()

	balances, err := f.repo.GetCashBalancesByUserID(context.Background(), f.userID, money.XXX)
	if err != nil {
		t.Fatalf("GetCashBalancesByUserID: %v", err)
	}
	if len(balances) != 1 {
		t.Fatalf("balances = %+v, want exactly one", balances)
	}

	return balances[0].Balance
}

func sameAmount(t *testing.T, label, got, want string) {
	t.Helper()

	g, err := decimal.NewFromString(got)
	if err != nil || !g.Equal(mustDecimal(t, want)) {
		t.Errorf("%s = %q, want %s", label, got, want)
	}
}

var cashDay = time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)

func TestCashMovementsLandOnOneBalance(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	first := f.mustMove(t, CashKindDeposit, "1000", cashDay)
	second := f.mustMove(t, CashKindDeposit, "250.5", cashDay)
	interest := f.mustMove(t, CashKindInterest, "4.5", cashDay)

	if first.EntryID != second.EntryID || first.EntryID != interest.EntryID {
		t.Errorf("entries = %v, %v, %v; want one balance", first.EntryID, second.EntryID, interest.EntryID)
	}
	if first.Ticker != "CASH-USD" || first.Type != TransferIn || first.Kind != CashKindDeposit || !first.Editable {
		t.Errorf("deposit = %+v, want an editable transfer_in on CASH-USD", first)
	}
	if interest.Type != CashInterest || interest.Kind != CashKindInterest {
		t.Errorf("interest = %s/%s, want cash_interest/interest", interest.Type, interest.Kind)
	}
	sameAmount(t, "interest amount", interest.Amount, "4.5")

	balances, err := f.repo.GetCashBalancesByUserID(ctx, f.userID, money.XXX)
	if err != nil {
		t.Fatalf("GetCashBalancesByUserID: %v", err)
	}
	if len(balances) != 1 {
		t.Fatalf("balances = %+v, want one", balances)
	}

	b := balances[0]
	// The interest is in the balance: 1000 + 250.5 + 4.5.
	sameAmount(t, "balance", b.Balance, "1255")
	sameAmount(t, "value", b.Value, "1255")
	if b.Currency != money.USD || b.DisplayCurrency != money.USD || !b.FXConverted {
		t.Errorf("currencies = %v shown in %v (converted %v), want USD in USD", b.Currency, b.DisplayCurrency, b.FXConverted)
	}
	if b.Movements != 3 || b.LastMovementDate == nil {
		t.Errorf("movements = %d, last %v; want 3 and a date", b.Movements, b.LastMovementDate)
	}
	if b.PortfolioName != "savings" || b.SourceName != "bank" {
		t.Errorf("names = %q on %q", b.PortfolioName, b.SourceName)
	}

	// The balance counts in its portfolio like any position, at its quantity. Its
	// gain is the interest: those units cost nothing (000042).
	summaries, err := f.repo.GetPortfoliosSummaryByUserID(ctx, f.userID)
	if err != nil {
		t.Fatalf("GetPortfoliosSummaryByUserID: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("summaries = %+v, want one", summaries)
	}

	sameAmount(t, "portfolio value", summaries[0].TotalMarketValue, "1255")
	aboutAmount(t, "portfolio gain", summaries[0].TotalGainLoss, "4.5")
	if summaries[0].TotalPositions != 1 {
		t.Errorf("positions = %d, want 1", summaries[0].TotalPositions)
	}
}

func TestCashWithdrawalsCannotOverdraw(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	if _, err := f.move(t, CashKindWithdrawal, "10", "", cashDay); !errors.Is(err, ErrInsufficientCash) {
		t.Fatalf("withdrawal from nothing = %v, want ErrInsufficientCash", err)
	}

	var entries int
	if err := f.pool.QueryRow(ctx, `SELECT COUNT(*) FROM portfolio_entries WHERE portfolio_id = $1`, f.portfolioID).Scan(&entries); err != nil {
		t.Fatalf("count entries: %v", err)
	}
	if entries != 0 {
		t.Errorf("a refused withdrawal left %d positions behind", entries)
	}

	deposit := f.mustMove(t, CashKindDeposit, "100", cashDay)

	if _, err := f.move(t, CashKindWithdrawal, "100.01", "", cashDay); !errors.Is(err, ErrInsufficientCash) {
		t.Errorf("withdrawal past the balance = %v, want ErrInsufficientCash", err)
	}

	withdrawal, err := f.move(t, CashKindWithdrawal, "60", "1", cashDay)
	if err != nil {
		t.Fatalf("withdrawal within the balance: %v", err)
	}
	sameAmount(t, "balance after withdrawing", f.balance(t), "40")

	// Shrinking the deposit below what the withdrawal already took empties the
	// account as surely as a new withdrawal would.
	shrink := CashMovementInput{Kind: CashKindDeposit, Amount: mustDecimal(t, "50"), Date: cashDay}
	if _, err := f.repo.UpdateCashMovement(ctx, f.userID, deposit.ID, shrink); !errors.Is(err, ErrInsufficientCash) {
		t.Errorf("shrinking the deposit to 50 = %v, want ErrInsufficientCash", err)
	}

	shrink.Amount = mustDecimal(t, "70")
	if _, err := f.repo.UpdateCashMovement(ctx, f.userID, deposit.ID, shrink); err != nil {
		t.Fatalf("shrinking the deposit to 70: %v", err)
	}
	sameAmount(t, "balance after the edit", f.balance(t), "10")

	if err := f.repo.DeleteCashMovement(ctx, f.userID, deposit.ID); !errors.Is(err, ErrInsufficientCash) {
		t.Errorf("deleting the deposit the withdrawal spent = %v, want ErrInsufficientCash", err)
	}

	if err := f.repo.DeleteCashMovement(ctx, f.userID, withdrawal.ID); err != nil {
		t.Fatalf("deleting the withdrawal: %v", err)
	}
	sameAmount(t, "balance after deleting the withdrawal", f.balance(t), "70")

	// Turning the deposit into a withdrawal takes its own money out as well.
	flip := CashMovementInput{Kind: CashKindWithdrawal, Amount: mustDecimal(t, "10"), Date: cashDay}
	if _, err := f.repo.UpdateCashMovement(ctx, f.userID, deposit.ID, flip); !errors.Is(err, ErrInsufficientCash) {
		t.Errorf("turning the only deposit into a withdrawal = %v, want ErrInsufficientCash", err)
	}
}

// The whole reason cash_interest exists: a deposit is money the owner put in and
// has to be netted out of the growth, the interest is money the balance earned
// and must not be.
func TestCashInterestIsReturnAndDepositsAreFlows(t *testing.T) {
	f := newCashFixture(t)
	today := time.Now().UTC().Truncate(24 * time.Hour)

	f.mustMove(t, CashKindDeposit, "1000", today)
	f.mustMove(t, CashKindInterest, "10", today)

	series, err := f.repo.GetPortfolioGrowthByPortfolioID(context.Background(), f.userID, f.portfolioID, false, time.Time{}, today)
	if err != nil {
		t.Fatalf("GetPortfolioGrowthByPortfolioID: %v", err)
	}
	if len(series) == 0 {
		t.Fatal("empty series")
	}

	last := series[len(series)-1]
	sameAmount(t, "value", last.TotalValue, "1010")
	sameAmount(t, "net flow", last.NetFlow, "1000")
}

// A balance loaded with its history walks in whole, interest included. Counting
// that interest as return would book a year of it on the day it was typed.
func TestCashInterestLoadedWithHistoryWalksInWithTheBalance(t *testing.T) {
	f := newCashFixture(t)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	monthAgo := today.AddDate(0, -1, 0)

	f.mustMove(t, CashKindDeposit, "1000", monthAgo)
	f.mustMove(t, CashKindInterest, "10", monthAgo)

	series, err := f.repo.GetPortfolioGrowthByPortfolioID(context.Background(), f.userID, f.portfolioID, false, time.Time{}, today)
	if err != nil {
		t.Fatalf("GetPortfolioGrowthByPortfolioID: %v", err)
	}
	if len(series) == 0 {
		t.Fatal("empty series")
	}

	last := series[len(series)-1]
	sameAmount(t, "value", last.TotalValue, "1010")
	sameAmount(t, "net flow", last.NetFlow, "1010")
}

func TestCashInterestOnlyGoesOnCashPositions(t *testing.T) {
	ctx := context.Background()
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)

	userID, entryID := fxPosition(t, pool)

	_, err := repo.CreateTransaction(ctx, userID, entryID, TransactionInput{
		Type:            CashInterest,
		Quantity:        mustDecimal(t, "1"),
		Price:           mustUSD(t, "5"),
		Currency:        money.USD,
		TransactionDate: cashDay,
	})
	if !errors.Is(err, ErrCashInterestOutsideCash) {
		t.Errorf("cash_interest on a share = %v, want ErrCashInterestOutsideCash", err)
	}

	// On a balance the generic writer takes it, and the balance grows.
	f := newCashFixture(t)
	deposit := f.mustMove(t, CashKindDeposit, "100", cashDay)

	if _, err := f.repo.CreateTransaction(ctx, f.userID, deposit.EntryID, TransactionInput{
		Type:            CashInterest,
		Quantity:        mustDecimal(t, "2"),
		Price:           mustUSD(t, "1"),
		Currency:        money.USD,
		TransactionDate: cashDay,
	}); err != nil {
		t.Fatalf("cash_interest on a balance: %v", err)
	}
	sameAmount(t, "balance", f.balance(t), "102")
}

// Deleting interest a snapshot already holds would read as a loss on the day it
// went, unless its version is kept the way a deleted buy's is (000038).
func TestCashInterestASnapshotSawIsRetiredWhenDeleted(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	f.mustMove(t, CashKindDeposit, "100", cashDay)
	interest := f.mustMove(t, CashKindInterest, "3", cashDay)

	f.exec(t, `INSERT INTO portfolio_snapshots (portfolio_id, snapshot_date, total_value, currency, created_at)
	           VALUES ($1, CURRENT_DATE - 1, 103, 'USD', NOW() + INTERVAL '1 minute')`, f.portfolioID)

	if err := f.repo.DeleteCashMovement(ctx, f.userID, interest.ID); err != nil {
		t.Fatalf("DeleteCashMovement: %v", err)
	}

	var retired int
	if err := f.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM retired_transactions WHERE portfolio_id = $1 AND type = 'cash_interest'
	`, f.portfolioID).Scan(&retired); err != nil {
		t.Fatalf("count retired: %v", err)
	}
	if retired != 1 {
		t.Errorf("retired cash_interest versions = %d, want 1", retired)
	}

	sameAmount(t, "balance", f.balance(t), "100")
}

// Dollars bought with pesos, opened by hand, cost what the pesos were. A
// deposit at one dollar a dollar would be averaged into that cost as one peso.
func TestCashMovementsLeaveAPositionInAnotherCurrencyAlone(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	// The first deposit is also what puts CASH-USD in the catalog.
	f.mustMove(t, CashKindDeposit, "100", cashDay)

	broker := uuid.New()
	f.exec(t, `INSERT INTO investment_sources (id, user_id, name, source_type) VALUES ($1, $2, 'broker', 'broker')`, broker, f.userID)
	f.exec(t, `INSERT INTO portfolio_entries (portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
	           VALUES ($1, (SELECT id FROM assets WHERE ticker = 'CASH-USD' AND exchange IS NULL), $2, 0, 4000, 'COP', '2026-09-01')`,
		f.portfolioID, broker)

	_, err := f.repo.CreateCashMovement(ctx, f.userID, f.portfolioID, broker, uuid.UUID{}, CashMovementInput{
		Kind: CashKindDeposit, Amount: mustDecimal(t, "10"), Currency: money.USD, Date: cashDay,
	})
	if !errors.Is(err, ErrInvalidCashMovement) {
		t.Errorf("deposit into a peso-funded dollar position = %v, want ErrInvalidCashMovement", err)
	}
}

func TestCashMovementsListAndReadOnlyRows(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	deposit := f.mustMove(t, CashKindDeposit, "100", cashDay)

	// A paid-out interest recorded against the balance through the generic
	// writer. It did nothing to the balance, and a cash movement it is not.
	paid, err := f.repo.CreateTransaction(ctx, f.userID, deposit.EntryID, TransactionInput{
		Type:            Interest,
		Quantity:        mustDecimal(t, "1"),
		Price:           mustUSD(t, "5"),
		Currency:        money.USD,
		TransactionDate: cashDay.AddDate(0, 0, 1),
	})
	if err != nil {
		t.Fatalf("CreateTransaction(interest): %v", err)
	}

	total, err := f.repo.CountCashMovements(ctx, f.userID)
	if err != nil || total != 2 {
		t.Fatalf("CountCashMovements = %d, %v; want 2", total, err)
	}

	page, err := f.repo.GetCashMovementsPaginated(ctx, f.userID, 10, 0)
	if err != nil {
		t.Fatalf("GetCashMovementsPaginated: %v", err)
	}
	if len(page) != 2 {
		t.Fatalf("page = %+v, want two", page)
	}

	// Most recent first.
	if page[0].ID != paid.ID || page[0].Kind != CashKindOther || page[0].Editable {
		t.Errorf("first row = %+v, want the paid-out interest, read-only", page[0])
	}
	if page[1].ID != deposit.ID || !page[1].Editable {
		t.Errorf("second row = %+v, want the deposit, editable", page[1])
	}

	rewrite := CashMovementInput{Kind: CashKindDeposit, Amount: mustDecimal(t, "5"), Date: cashDay}
	if _, err := f.repo.UpdateCashMovement(ctx, f.userID, paid.ID, rewrite); !errors.Is(err, ErrCashMovementNotEditable) {
		t.Errorf("rewriting the paid-out interest = %v, want ErrCashMovementNotEditable", err)
	}

	if _, err := f.repo.UpdateCashMovement(ctx, uuid.New(), deposit.ID, rewrite); !errors.Is(err, ErrCashMovementNotFound) {
		t.Errorf("someone else's movement = %v, want ErrCashMovementNotFound", err)
	}

	// It can still be removed: that does not reprice anything.
	if err := f.repo.DeleteCashMovement(ctx, f.userID, paid.ID); err != nil {
		t.Errorf("deleting the paid-out interest: %v", err)
	}
}
