package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Contributions and withdrawals of a fund followed by units, written from the
// funds screen. Same database contract as postgres_fund_db_test.go.

func TestUnitsFundMovements(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	fund := f.create(t, "100", "10", "", time.Time{})

	units := func(t *testing.T, want string) {
		t.Helper()

		got, err := f.repo.GetFund(ctx, f.userID, fund.AssetID)
		if err != nil {
			t.Fatalf("GetFund: %v", err)
		}

		sameAmount(t, "fund units", got.Units, want)
	}

	in, err := f.repo.ContributeToFund(ctx, f.userID, fund.AssetID, FundContributionInput{
		PortfolioID: f.portfolioID, SourceID: f.sourceID, Date: fundDay(5),
		Units: mustDecimal(t, "50"), UnitValue: mustDecimal(t, "11"), Notes: "aporte",
	})
	if err != nil {
		t.Fatalf("ContributeToFund: %v", err)
	}

	if in.Kind != FundContribution || in.AssetID != fund.AssetID {
		t.Fatalf("contribution = %+v, want a contribution of the fund", in)
	}

	sameAmount(t, "contribution amount", in.Amount, "550")
	units(t, "150")

	entryID := in.EntryID

	if _, err := f.repo.WithdrawFromFund(ctx, f.userID, fund.AssetID, FundWithdrawalInput{
		EntryID: entryID, Date: fundDay(6), Units: mustDecimal(t, "151"), UnitValue: mustDecimal(t, "12"),
	}); !errors.Is(err, ErrFundNotEnoughUnits) {
		t.Fatalf("WithdrawFromFund of more than held = %v, want ErrFundNotEnoughUnits", err)
	}

	out, err := f.repo.WithdrawFromFund(ctx, f.userID, fund.AssetID, FundWithdrawalInput{
		EntryID: entryID, Date: fundDay(6), Units: mustDecimal(t, "30"), UnitValue: mustDecimal(t, "12"),
		Fees: mustDecimal(t, "5"),
	})
	if err != nil {
		t.Fatalf("WithdrawFromFund: %v", err)
	}

	if out.Kind != FundWithdrawal {
		t.Fatalf("withdrawal kind = %s", out.Kind)
	}

	sameAmount(t, "withdrawal amount", out.Amount, "360")
	sameAmount(t, "withdrawal fees", out.Fees, "5")
	units(t, "120")

	edited, err := f.repo.UpdateFundMovement(ctx, f.userID, out.TxnID, FundMovementEdit{
		Date: fundDay(7), Units: mustDecimal(t, "20"), UnitValue: mustDecimal(t, "12.5"), Notes: "corregido",
	})
	if err != nil {
		t.Fatalf("UpdateFundMovement: %v", err)
	}

	sameAmount(t, "edited amount", edited.Amount, "250")

	if edited.Notes != "corregido" || !cashRateDay(edited.Date).Equal(fundDay(7)) {
		t.Fatalf("edited = %+v, want the new note and day", edited)
	}

	units(t, "130")

	all, err := f.repo.WithdrawFromFund(ctx, f.userID, fund.AssetID, FundWithdrawalInput{
		EntryID: entryID, Date: fundDay(8), UnitValue: mustDecimal(t, "13"), All: true,
	})
	if err != nil {
		t.Fatalf("WithdrawFromFund of everything: %v", err)
	}

	sameAmount(t, "everything withdrawn", all.Units, "130")
	units(t, "0")

	if err := f.repo.DeleteFundMovement(ctx, f.userID, all.TxnID); err != nil {
		t.Fatalf("DeleteFundMovement: %v", err)
	}

	units(t, "130")

	movements, err := f.repo.GetFundMovements(ctx, f.userID, fund.AssetID)
	if err != nil || len(movements) != 3 {
		t.Fatalf("GetFundMovements = %d, %v; want the opening, the contribution and the edited withdrawal", len(movements), err)
	}
}

// A fund still held goes whole when told to take its positions: every
// position, its transactions and its marks, in one transaction.
func TestDeleteFundWithPositions(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	fund := f.create(t, "100", "10", "10.5", fundDay(10))

	if _, err := f.repo.WithdrawFromFund(ctx, f.userID, fund.AssetID, FundWithdrawalInput{
		EntryID: fund.Positions[0].EntryID, Date: fundDay(11), Units: mustDecimal(t, "10"), UnitValue: mustDecimal(t, "10.5"),
	}); err != nil {
		t.Fatalf("WithdrawFromFund: %v", err)
	}

	if err := f.repo.DeleteFund(ctx, f.userID, fund.AssetID, true); err != nil {
		t.Fatalf("DeleteFund with its positions: %v", err)
	}

	if _, err := f.repo.GetFund(ctx, f.userID, fund.AssetID); !errors.Is(err, ErrFundNotFound) {
		t.Errorf("GetFund after deleting = %v, want ErrFundNotFound", err)
	}

	var left int
	if err := f.pool.QueryRow(ctx, `SELECT COUNT(*) FROM portfolio_entries WHERE asset_id = $1`, fund.AssetID).Scan(&left); err != nil || left != 0 {
		t.Errorf("positions left = %d, %v; want none", left, err)
	}
}
