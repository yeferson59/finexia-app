package portfolio

import (
	"context"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// The contributions and withdrawals of a fund followed by units, written from
// the funds screen. They are plain purchases and sales at the unit value the
// owner read on the statement — what the generic transaction writers would
// write — so nothing is replayed and no fund_movements row is kept: their
// quantity and price are the fact. What the screen adds is the fund's lock and
// the vocabulary: units in, units out, everything the position holds.

// contributeUnits buys units of a fund followed by units on the position its
// portfolio holds on the platform, opened if there is none. The fund is
// already locked.
func contributeUnits(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID, cur money.Currency, in FundContributionInput) (FundMovement, error) {
	if err := validateFundUnits(in.Units, in.UnitValue); err != nil {
		return FundMovement{}, err
	}

	_, txnID, err := createPortfolioEntryTx(ctx, tx, userID, in.PortfolioID, assetID, in.SourceID, cur, TransactionInput{
		Type:            Buy,
		Quantity:        in.Units,
		Price:           money.NewFromDecimal(in.UnitValue, cur),
		Currency:        cur,
		TransactionDate: cashRateDay(in.Date),
		Notes:           in.Notes,
		PayFromCash:     in.PayFromCash,
		CashPocketID:    in.CashPocketID,
	})
	if err != nil {
		return FundMovement{}, err
	}

	return readFundMovement(ctx, tx, userID, txnID)
}

// withdrawUnits sells units of one position of a fund followed by units. held
// is what the position holds now; All sells all of it, and anything more than
// it is refused. The fund is already locked.
func withdrawUnits(ctx context.Context, tx pgx.Tx, userID uuid.UUID, cur money.Currency, held decimal.Decimal, in FundWithdrawalInput) (FundMovement, error) {
	units := in.Units
	if in.All {
		units = held
	}

	if !held.IsPos() || units.GreaterThan(held) {
		return FundMovement{}, ErrFundNotEnoughUnits
	}

	if err := validateFundUnits(units, in.UnitValue); err != nil {
		return FundMovement{}, err
	}

	if in.Fees.IsPos() && in.Fees.GreaterThanOrEqual(units.Mul(in.UnitValue)) {
		return FundMovement{}, invalidFund("fees must be less than the amount withdrawn")
	}

	txn, err := createTransactionTx(ctx, tx, userID, in.EntryID, TransactionInput{
		Type:            Sell,
		Quantity:        units,
		Price:           money.NewFromDecimal(in.UnitValue, cur),
		Currency:        cur,
		Fees:            money.NewFromDecimal(in.Fees, cur),
		FeesCurrency:    cur,
		TransactionDate: cashRateDay(in.Date),
		Notes:           in.Notes,
		CreditCash:      in.CreditCash,
		CashPocketID:    in.CashPocketID,
	}, false)
	if err != nil {
		return FundMovement{}, err
	}

	return readFundMovement(ctx, tx, userID, txn.ID)
}

// updateUnitsMovement restates a purchase or sale of a fund followed by units:
// its day, units, unit value, fees and note. Its currency, rate and cash side
// stay what they were; the cash row, if it has one, is resized to the new
// figures with the rest of the position's. The fund is already locked.
func updateUnitsMovement(ctx context.Context, tx pgx.Tx, userID, txnID uuid.UUID, locked lockedFundMovement, in FundMovementEdit) (FundMovement, error) {
	if err := validateFundUnits(in.Units, in.UnitValue); err != nil {
		return FundMovement{}, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE transactions
		   SET transaction_date = $2::date, quantity = $3::numeric, price = $4::numeric,
		       fees = $5::numeric, notes = NULLIF($6, ''), updated_at = NOW()
		 WHERE id = $1
	`, txnID, cashRateDay(in.Date).Format(time.DateOnly), in.Units.String(), in.UnitValue.String(),
		in.Fees.String(), in.Notes); err != nil {
		return FundMovement{}, err
	}

	if err := syncEntryCashLinks(ctx, tx, userID, locked.entryID, true); err != nil {
		return FundMovement{}, err
	}

	return readFundMovement(ctx, tx, userID, txnID)
}
