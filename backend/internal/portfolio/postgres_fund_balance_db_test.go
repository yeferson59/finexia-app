package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// Funds followed by balance, migration 000058. Same database contract as
// postgres_fund_db_test.go.

// createBalance opens a USD fund followed by balance with amount on day, and a
// current balance when current is not empty.
func (f fundFixture) createBalance(t *testing.T, amount string, day time.Time, current string, currentOn time.Time) Fund {
	t.Helper()

	in := NewFundInput{
		PortfolioID: f.portfolioID,
		SourceID:    f.sourceID,
		Name:        "Bolsillo de inversión",
		Currency:    money.USD,
		Tracking:    FundBalance,
		Date:        day,
		Amount:      mustDecimal(t, amount),
	}
	if current != "" {
		in.CurrentBalance = mustDecimal(t, current)
		in.CurrentDate = currentOn
	}

	fund, err := f.repo.CreateFund(context.Background(), f.userID, in)
	if err != nil {
		t.Fatalf("CreateFund: %v", err)
	}

	f.plant(fund.AssetID)

	return fund
}

func (f fundFixture) balanceOn(t *testing.T, assetID uuid.UUID, on time.Time, balance string) {
	t.Helper()

	if _, err := f.repo.UpsertFundMark(context.Background(), f.userID, assetID, FundMarkInput{
		Date: on, Balance: mustDecimal(t, balance),
	}); err != nil {
		t.Fatalf("UpsertFundMark(%s: %s): %v", on.Format(time.DateOnly), balance, err)
	}
}

func (f fundFixture) contribute(t *testing.T, assetID uuid.UUID, on time.Time, amount string) FundMovement {
	t.Helper()

	m, err := f.repo.ContributeToFund(context.Background(), f.userID, assetID, FundContributionInput{
		PortfolioID: f.portfolioID, SourceID: f.sourceID, Date: on, Amount: mustDecimal(t, amount),
	})
	if err != nil {
		t.Fatalf("ContributeToFund(%s: %s): %v", on.Format(time.DateOnly), amount, err)
	}

	return m
}

func (f fundFixture) fund(t *testing.T, assetID uuid.UUID) Fund {
	t.Helper()

	fund, err := f.repo.GetFund(context.Background(), f.userID, assetID)
	if err != nil {
		t.Fatalf("GetFund: %v", err)
	}

	return fund
}

// roundedAmount compares a figure at cents: a balance read back from synthetic
// units is exact to well under one.
func roundedAmount(t *testing.T, label, got, want string) {
	t.Helper()

	g, err := decimal.NewFromString(got)
	if err != nil || !g.RoundHAZ(2).Equal(mustDecimal(t, want)) {
		t.Errorf("%s = %q, want %s at cents", label, got, want)
	}
}

// planFund records the example of §4 of docs/PLAN_FONDOS_INVERSION.md with
// nothing but money and balances.
func (f fundFixture) planFund(t *testing.T) (Fund, FundMovement, FundMovement) {
	t.Helper()
	ctx := context.Background()

	fund := f.createBalance(t, "10000000", planDay(time.July, 1), "", time.Time{})
	f.balanceOn(t, fund.AssetID, planDay(time.July, 31), "10080000")
	august := f.contribute(t, fund.AssetID, planDay(time.August, 15), "5000000")
	f.balanceOn(t, fund.AssetID, planDay(time.August, 31), "15110000")

	withdrawal, err := f.repo.WithdrawFromFund(ctx, f.userID, fund.AssetID, FundWithdrawalInput{
		EntryID: august.EntryID, Date: planDay(time.September, 20), Amount: mustDecimal(t, "2000000"),
	})
	if err != nil {
		t.Fatalf("WithdrawFromFund: %v", err)
	}

	f.balanceOn(t, fund.AssetID, planDay(time.September, 30), "13050000")

	return f.fund(t, fund.AssetID), august, withdrawal
}

// The example of the plan, end to end: the units, the unit values, the value
// and the price the valuation reads come out as the replay says.
func TestBalanceFundPlanExample(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	fund, august, withdrawal := f.planFund(t)

	if fund.Tracking != FundBalance || fund.Marks != 3 || len(fund.Positions) != 1 {
		t.Fatalf("fund = %+v", fund)
	}

	sameAmount(t, "units", fund.Units, "129801.29841401")
	sameAmount(t, "unit value", *fund.UnitValue, "100.53828551")
	roundedAmount(t, "value", fund.Value, "13050000")
	sameAmount(t, "own price", f.ownPrice(t, fund.AssetID), "100.53828551")

	sameAmount(t, "august units", august.Units, "49603.17460317")
	sameAmount(t, "august unit value", august.UnitValue, "100.8")
	sameAmount(t, "august amount", august.Amount, "5000000")
	sameAmount(t, "withdrawal units", withdrawal.Units, "19801.87618916")

	if august.Kind != FundContribution || withdrawal.Kind != FundWithdrawal {
		t.Errorf("kinds = %s, %s", august.Kind, withdrawal.Kind)
	}

	movements, err := f.repo.GetFundMovements(ctx, f.userID, fund.AssetID)
	if err != nil || len(movements) != 3 {
		t.Fatalf("GetFundMovements = %d, %v; want three", len(movements), err)
	}

	// Newest first, each with the money that was stated.
	sameAmount(t, "the newest", movements[0].Amount, "2000000")
	sameAmount(t, "the oldest", movements[2].Amount, "10000000")

	marks, err := f.repo.GetFundMarks(ctx, f.userID, fund.AssetID)
	if err != nil || len(marks) != 3 || marks[0].Balance == nil {
		t.Fatalf("GetFundMarks = %+v, %v", marks, err)
	}

	sameAmount(t, "the balance of the 30th", *marks[0].Balance, "13050000")
	sameAmount(t, "the unit value of the 31st of July", marks[2].UnitValue, "100.8")
}

// Correcting the contribution of August replays everything after it: its own
// units, the unit value of the balances after it and the withdrawal's units.
// The balance of the 30th still says what the fund is worth.
func TestBalanceFundEditReplaysWhatFollows(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	fund, august, withdrawal := f.planFund(t)

	edited, err := f.repo.UpdateFundMovement(ctx, f.userID, august.TxnID, FundMovementEdit{
		Date: planDay(time.August, 15), Amount: mustDecimal(t, "6000000"),
	})
	if err != nil {
		t.Fatalf("UpdateFundMovement: %v", err)
	}

	// 6.000.000 at 100,8.
	sameAmount(t, "edited units", edited.Units, "59523.80952381")

	after := f.fund(t, fund.AssetID)
	roundedAmount(t, "value", after.Value, "13050000")

	if *after.UnitValue == *fund.UnitValue {
		t.Error("the unit value of the 30th did not move")
	}

	moved, err := f.repo.GetFundMovement(ctx, f.userID, withdrawal.TxnID)
	if err != nil {
		t.Fatalf("GetFundMovement: %v", err)
	}

	if moved.Units == withdrawal.Units {
		t.Error("the withdrawal kept its units after an earlier contribution grew")
	}

	// Taking the withdrawal back leaves every unit in the fund.
	if err := f.repo.DeleteFundMovement(ctx, f.userID, withdrawal.TxnID); err != nil {
		t.Fatalf("DeleteFundMovement: %v", err)
	}

	if _, err := f.repo.GetFundMovement(ctx, f.userID, withdrawal.TxnID); !errors.Is(err, ErrFundMovementNotFound) {
		t.Errorf("the deleted withdrawal reads as %v, want ErrFundMovementNotFound", err)
	}

	roundedAmount(t, "value without the withdrawal", f.fund(t, fund.AssetID).Value, "13050000")
}

// Its quantities are derived, so the generic writers keep off it (D12).
func TestBalanceFundRefusesGenericWrites(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	fund := f.createBalance(t, "1000", planDay(time.July, 1), "", time.Time{})
	entryID := fund.Positions[0].EntryID

	buy := TransactionInput{
		Type:            Buy,
		Quantity:        mustDecimal(t, "1"),
		Price:           money.NewFromDecimal(mustDecimal(t, "100"), money.USD),
		Currency:        money.USD,
		TransactionDate: planDay(time.July, 2),
	}

	if _, err := f.repo.CreateTransaction(ctx, f.userID, entryID, buy); !errors.Is(err, ErrFundBalanceManaged) {
		t.Errorf("CreateTransaction = %v, want ErrFundBalanceManaged", err)
	}

	if _, err := f.repo.CreatePortfolioEntry(ctx, f.userID, f.portfolioID, fund.AssetID, f.sourceID, money.USD, buy); !errors.Is(err, ErrFundBalanceManaged) {
		t.Errorf("CreatePortfolioEntry = %v, want ErrFundBalanceManaged", err)
	}

	movements, err := f.repo.GetFundMovements(ctx, f.userID, fund.AssetID)
	if err != nil || len(movements) != 1 {
		t.Fatalf("GetFundMovements = %d, %v; want the opening only", len(movements), err)
	}

	if _, err := f.repo.UpdateTransaction(ctx, f.userID, movements[0].TxnID, buy); !errors.Is(err, ErrFundBalanceManaged) {
		t.Errorf("UpdateTransaction = %v, want ErrFundBalanceManaged", err)
	}

	if err := f.repo.DeleteTransaction(ctx, f.userID, movements[0].TxnID); !errors.Is(err, ErrFundBalanceManaged) {
		t.Errorf("DeleteTransaction = %v, want ErrFundBalanceManaged", err)
	}

	// And a fund followed by units is not told money: it takes units.
	units := f.create(t, "10", "10", "", time.Time{})
	if _, err := f.repo.ContributeToFund(ctx, f.userID, units.AssetID, FundContributionInput{
		PortfolioID: f.portfolioID, SourceID: f.sourceID, Date: fundDay(2), Amount: mustDecimal(t, "5"),
	}); !errors.Is(err, ErrInvalidFund) {
		t.Errorf("ContributeToFund of an amount on a fund by units = %v, want ErrInvalidFund", err)
	}
}

// What cannot be valued is refused whole: a withdrawal of more than there is,
// and a balance on a day the fund held nothing.
func TestBalanceFundRefusals(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	fund := f.createBalance(t, "1000", planDay(time.July, 10), "", time.Time{})

	if _, err := f.repo.WithdrawFromFund(ctx, f.userID, fund.AssetID, FundWithdrawalInput{
		EntryID: fund.Positions[0].EntryID, Date: planDay(time.July, 20), Amount: mustDecimal(t, "1500"),
	}); !errors.Is(err, ErrFundNotEnoughUnits) {
		t.Fatalf("withdrawing too much = %v, want ErrFundNotEnoughUnits", err)
	}

	if _, err := f.repo.UpsertFundMark(ctx, f.userID, fund.AssetID, FundMarkInput{
		Date: planDay(time.July, 5), Balance: mustDecimal(t, "900"),
	}); !errors.Is(err, ErrFundNoUnits) {
		t.Fatalf("a balance before the first contribution = %v, want ErrFundNoUnits", err)
	}

	// Neither left anything behind.
	if movements, _ := f.repo.GetFundMovements(ctx, f.userID, fund.AssetID); len(movements) != 1 {
		t.Errorf("movements = %d, want the opening only", len(movements))
	}

	if marks, _ := f.repo.GetFundMarks(ctx, f.userID, fund.AssetID); len(marks) != 0 {
		t.Errorf("marks = %d, want none", len(marks))
	}

	// A withdrawal of everything, stated as money, leaves nothing.
	f.balanceOn(t, fund.AssetID, planDay(time.July, 31), "1010")

	if _, err := f.repo.WithdrawFromFund(ctx, f.userID, fund.AssetID, FundWithdrawalInput{
		EntryID: fund.Positions[0].EntryID, Date: planDay(time.August, 5), Amount: mustDecimal(t, "1012"), All: true,
	}); err != nil {
		t.Fatalf("withdrawing everything: %v", err)
	}

	sameAmount(t, "units left", f.fund(t, fund.AssetID).Units, "0")
}

// The quick start with what the fund is worth now: the price is in place before
// the contribution, so it walks in at its value (D13), and a balance before a
// contribution fixes the unit value it trades at.
func TestBalanceFundOpeningAndBalanceBefore(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	fund := f.createBalance(t, "10000000", planDay(time.July, 1), "10500000", fundDay(1))

	sameAmount(t, "unit value", *fund.UnitValue, "105")
	roundedAmount(t, "value", fund.Value, "10500000")

	var recorded string
	if err := f.pool.QueryRow(ctx, `
		SELECT t.recorded_market_price::text FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		WHERE pe.asset_id = $1
	`, fund.AssetID).Scan(&recorded); err != nil {
		t.Fatalf("read the contribution: %v", err)
	}

	sameAmount(t, "recorded market price", recorded, "105")

	m, err := f.repo.ContributeToFund(ctx, f.userID, fund.AssetID, FundContributionInput{
		PortfolioID:   f.portfolioID,
		SourceID:      f.sourceID,
		Date:          fundDay(10),
		Amount:        mustDecimal(t, "1060000"),
		BalanceBefore: mustDecimal(t, "10600000"),
	})
	if err != nil {
		t.Fatalf("ContributeToFund with a balance before: %v", err)
	}

	// The 9th closed at 106; the contribution buys 10.000 units at it.
	sameAmount(t, "unit value", m.UnitValue, "106")
	sameAmount(t, "units", m.Units, "10000")
}

// Paid from the platform's cash, a contribution takes its amount out of the
// balance, and keeps doing so after the replay moves its units.
func TestBalanceFundContributionPaidFromCash(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	fund := f.createBalance(t, "1000", planDay(time.July, 1), "", time.Time{})
	f.balanceOn(t, fund.AssetID, planDay(time.July, 31), "1030")

	if _, err := f.repo.CreateCashMovement(ctx, f.userID, f.portfolioID, f.sourceID, uuid.UUID{}, CashMovementInput{
		Kind: CashKindDeposit, Amount: mustDecimal(t, "5000"), Currency: money.USD, Date: planDay(time.July, 1),
	}); err != nil {
		t.Fatalf("CreateCashMovement: %v", err)
	}

	if _, err := f.repo.ContributeToFund(ctx, f.userID, fund.AssetID, FundContributionInput{
		PortfolioID: f.portfolioID, SourceID: f.sourceID, Date: planDay(time.August, 3),
		Amount: mustDecimal(t, "2060"), PayFromCash: true,
	}); err != nil {
		t.Fatalf("ContributeToFund paid from cash: %v", err)
	}

	balances, err := f.repo.GetCashBalancesByUserID(ctx, f.userID, money.XXX)
	if err != nil || len(balances) != 1 {
		t.Fatalf("GetCashBalancesByUserID = %+v, %v", balances, err)
	}

	sameAmount(t, "cash left", balances[0].Balance, "2940")
}

// A balance typed in late revalues the snapshots from its day, like a unit
// value does.
func TestBalanceFundLateBalanceRestatesSnapshots(t *testing.T) {
	f := newFundFixture(t)

	fund := f.createBalance(t, "1000", fundDay(1), "", time.Time{})

	for _, day := range []int{2, 3, 4} {
		f.snapshot(t, fundDay(day))
	}

	f.balanceOn(t, fund.AssetID, fundDay(3), "1100")

	f.wantSnapshot(t, fundDay(2), "1000", "0", "0")
	f.wantSnapshot(t, fundDay(3), "1100", "100", "10")
	f.wantSnapshot(t, fundDay(4), "1100", "100", "10")
}

// Deleting a position replays the fund — its balances valued those units too —
// and a fund nobody holds can then be dropped.
func TestBalanceFundPositionDeleted(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	fund := f.createBalance(t, "1000", planDay(time.July, 1), "", time.Time{})
	f.balanceOn(t, fund.AssetID, planDay(time.July, 31), "1010")

	if _, err := f.repo.DeletePortfolioEntry(ctx, f.userID, fund.Positions[0].EntryID); err != nil {
		t.Fatalf("DeletePortfolioEntry: %v", err)
	}

	if err := f.repo.DeleteFund(ctx, f.userID, fund.AssetID, false); err != nil {
		t.Fatalf("DeleteFund: %v", err)
	}
}

// The plan's example read back through the service: the returns of each period
// and the money, from what the database keeps.
func TestFundPerformanceFromTheDatabase(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	fund, _, _ := f.planFund(t)
	svc := newService(f.repo, testConfig(), nil, nil, nil, logger.Noop())

	perf, err := svc.GetFundPerformance(ctx, f.userID, fund.AssetID)
	if err != nil {
		t.Fatalf("GetFundPerformance: %v", err)
	}

	wantPeriod(t, periodByKey(t, perf.Periods, "30d"), planDay(time.August, 31), 30, "-0.4577", "-5.43")
	wantPeriod(t, periodByKey(t, perf.Periods, "inception"), planDay(time.July, 1), 91, "0.5383", "2.18")
	wantEmptyPeriod(t, periodByKey(t, perf.Periods, "ytd"))

	sameAmount(t, "invested", perf.Invested, "15000000")
	sameAmount(t, "withdrawn", perf.Withdrawn, "2000000")
	sameAmount(t, "realized", perf.RealizedGain, "14559.89")

	if perf.UnrealizedGain == nil {
		t.Fatal("no unrealized gain")
	}

	sameAmount(t, "unrealized", *perf.UnrealizedGain, "35440.11")

	if len(perf.Series) != 4 {
		t.Errorf("series = %d points, want the opening and three balances", len(perf.Series))
	}
}

// A statement's table in one write: every day lands, the price is the latest,
// and on a fund followed by balance a day with nothing held refuses the whole
// table.
func TestFundMarksInBulk(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	units := f.create(t, "10", "100", "", time.Time{})

	saved, err := f.repo.UpsertFundMarks(ctx, f.userID, units.AssetID, []FundMarkInput{
		{Date: fundDay(10), UnitValue: mustDecimal(t, "101")},
		{Date: fundDay(20), UnitValue: mustDecimal(t, "103")},
		{Date: fundDay(15), UnitValue: mustDecimal(t, "102")},
	})
	if err != nil || saved != 3 {
		t.Fatalf("UpsertFundMarks = %d, %v", saved, err)
	}

	sameAmount(t, "own price", f.ownPrice(t, units.AssetID), "103")

	balance := f.createBalance(t, "1000", fundDay(5), "", time.Time{})

	if _, err := f.repo.UpsertFundMarks(ctx, f.userID, balance.AssetID, []FundMarkInput{
		{Date: fundDay(10), Balance: mustDecimal(t, "1010")},
		{Date: fundDay(2), Balance: mustDecimal(t, "990")},
	}); !errors.Is(err, ErrFundNoUnits) {
		t.Fatalf("a table with a day before the first contribution = %v, want ErrFundNoUnits", err)
	}

	if marks, _ := f.repo.GetFundMarks(ctx, f.userID, balance.AssetID); len(marks) != 0 {
		t.Errorf("marks = %d after a refused table, want none", len(marks))
	}
}
