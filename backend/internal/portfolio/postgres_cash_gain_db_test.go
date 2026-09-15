package portfolio

import (
	"context"
	"errors"
	"testing"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// What 000042 makes of the interest a balance earns: units that cost nothing,
// so the balance's gain is the interest still in it. Same database contract as
// postgres_cash_db_test.go.

// aboutAmount compares to a hundredth of a cent. A position's price is kept to
// eight decimals, so a cost rebuilt as quantity × price can miss the exact
// figure by a few millionths.
func aboutAmount(t *testing.T, label, got, want string) {
	t.Helper()

	g, err := decimal.NewFromString(got)
	if err != nil {
		t.Errorf("%s = %q, not a number", label, got)
		return
	}

	diff := g.Add(mustDecimal(t, want).Neg())
	if diff.IsNeg() {
		diff = diff.Neg()
	}

	if diff.GreaterThan(mustDecimal(t, "0.0001")) {
		t.Errorf("%s = %s, want %s", label, got, want)
	}
}

// summary is the fixture's only portfolio as portfolio_summary reads it.
func (f cashFixture) summary(t *testing.T) SummaryView {
	t.Helper()

	summaries, err := f.repo.GetPortfoliosSummaryByUserID(context.Background(), f.userID)
	if err != nil {
		t.Fatalf("GetPortfoliosSummaryByUserID: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("summaries = %+v, want one", summaries)
	}

	return summaries[0]
}

func TestCashInterestIsGainInThePortfolioAndThePlatform(t *testing.T) {
	f := newCashFixture(t)

	f.mustMove(t, CashKindDeposit, "1000", cashDay)
	f.mustMove(t, CashKindInterest, "10", cashDay.AddDate(0, 0, 1))

	s := f.summary(t)
	aboutAmount(t, "portfolio value", s.TotalMarketValue, "1010")
	aboutAmount(t, "portfolio gain", s.TotalGainLoss, "10")

	platforms, err := f.repo.GetPlatformsWithStats(context.Background(), f.userID, money.USD)
	if err != nil {
		t.Fatalf("GetPlatformsWithStats: %v", err)
	}
	if len(platforms) != 1 {
		t.Fatalf("platforms = %+v, want one", platforms)
	}

	// TotalValue is what the platform cost.
	aboutAmount(t, "platform cost", platforms[0].TotalValue, "1000")
	aboutAmount(t, "platform value", platforms[0].MarketValue, "1010")
}

// A withdrawal takes its share of the cost, and with it its share of the
// interest: half the balance out leaves half the gain.
func TestCashWithdrawalTakesItsShareOfTheInterest(t *testing.T) {
	f := newCashFixture(t)

	f.mustMove(t, CashKindDeposit, "1000", cashDay)
	f.mustMove(t, CashKindInterest, "10", cashDay.AddDate(0, 0, 1))
	// The balance's price is no longer one, and the withdrawal still finds it.
	f.mustMove(t, CashKindWithdrawal, "505", cashDay.AddDate(0, 0, 2))

	sameAmount(t, "balance", f.balance(t), "505")

	s := f.summary(t)
	aboutAmount(t, "portfolio value", s.TotalMarketValue, "505")
	aboutAmount(t, "portfolio gain", s.TotalGainLoss, "5")
}

func TestCashDepositAfterInterestLandsOnTheSameBalance(t *testing.T) {
	f := newCashFixture(t)

	first := f.mustMove(t, CashKindDeposit, "1000", cashDay)
	f.mustMove(t, CashKindInterest, "10", cashDay.AddDate(0, 0, 1))
	later := f.mustMove(t, CashKindDeposit, "500", cashDay.AddDate(0, 0, 2))

	if later.EntryID != first.EntryID {
		t.Errorf("deposit after interest landed on %v, want the balance %v", later.EntryID, first.EntryID)
	}

	sameAmount(t, "balance", f.balance(t), "1510")
	aboutAmount(t, "portfolio gain", f.summary(t).TotalGainLoss, "10")
}

// An account emptied and filled again starts over. Averaging every deposit it
// ever had would carry the first round's interest into the cost of the second.
func TestCashEmptiedBalanceStartsOver(t *testing.T) {
	f := newCashFixture(t)

	f.mustMove(t, CashKindDeposit, "100", cashDay)
	f.mustMove(t, CashKindInterest, "1", cashDay.AddDate(0, 0, 1))
	f.mustMove(t, CashKindWithdrawal, "101", cashDay.AddDate(0, 0, 2))
	f.mustMove(t, CashKindDeposit, "50", cashDay.AddDate(0, 0, 3))

	sameAmount(t, "balance", f.balance(t), "50")

	s := f.summary(t)
	aboutAmount(t, "portfolio value", s.TotalMarketValue, "50")
	aboutAmount(t, "portfolio gain", s.TotalGainLoss, "0")
}

// Interest recorded on a balance that held nothing yet opens it at a price of
// zero — all of it is gain — and the next deposit still finds it.
func TestCashInterestOnlyBalanceTakesTheNextDeposit(t *testing.T) {
	f := newCashFixture(t)

	interest := f.mustMove(t, CashKindInterest, "5", cashDay)
	deposit := f.mustMove(t, CashKindDeposit, "100", cashDay.AddDate(0, 0, 1))

	if deposit.EntryID != interest.EntryID {
		t.Errorf("deposit landed on %v, want the balance %v", deposit.EntryID, interest.EntryID)
	}

	sameAmount(t, "balance", f.balance(t), "105")
	aboutAmount(t, "portfolio gain", f.summary(t).TotalGainLoss, "5")
}

// Deleting the last interest puts the balance back on 000041's formula, and
// back at a price of one.
func TestCashGainLeavesWithTheDeletedInterest(t *testing.T) {
	f := newCashFixture(t)

	f.mustMove(t, CashKindDeposit, "100", cashDay)
	interest := f.mustMove(t, CashKindInterest, "3", cashDay.AddDate(0, 0, 1))

	if err := f.repo.DeleteCashMovement(context.Background(), f.userID, interest.ID); err != nil {
		t.Fatalf("DeleteCashMovement: %v", err)
	}

	var price string
	if err := f.pool.QueryRow(context.Background(),
		`SELECT price::text FROM portfolio_entries WHERE id = $1`, interest.EntryID,
	).Scan(&price); err != nil {
		t.Fatalf("read price: %v", err)
	}

	sameAmount(t, "price", price, "1")
	sameAmount(t, "portfolio gain", f.summary(t).TotalGainLoss, "0")
}

// The price test moved to the transactions, and a position opened by hand at
// another price with none yet is still not a balance the cash writers keep.
func TestCashMovementsLeaveAPositionAtAnotherPriceAlone(t *testing.T) {
	f := newCashFixture(t)
	ctx := context.Background()

	// The first deposit is also what puts CASH-USD in the catalog.
	f.mustMove(t, CashKindDeposit, "100", cashDay)

	broker := uuid.New()
	f.exec(t, `INSERT INTO investment_sources (id, user_id, name, source_type) VALUES ($1, $2, 'broker', 'broker')`, broker, f.userID)
	f.exec(t, `INSERT INTO portfolio_entries (portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
	           VALUES ($1, (SELECT id FROM assets WHERE ticker = 'CASH-USD' AND exchange IS NULL), $2, 0, 0.9, 'USD', '2026-09-01')`,
		f.portfolioID, broker)

	_, err := f.repo.CreateCashMovement(ctx, f.userID, f.portfolioID, broker, CashMovementInput{
		Kind: CashKindDeposit, Amount: mustDecimal(t, "10"), Currency: money.USD, Date: cashDay,
	})
	if !errors.Is(err, ErrInvalidCashMovement) {
		t.Errorf("deposit into a dollar position priced at 0.9 = %v, want ErrInvalidCashMovement", err)
	}
}
