package portfolio

import (
	"context"
	"errors"
	"fmt"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
	"github.com/yeferson59/finexia-app/internal/platform/currency"
)

// A dividend or a sale paid into cash is two rows: the transaction on the
// holding, and its credit on the cash balance of the same platform, which is
// where the money went (000049 to 000052). The credit follows its transaction
// everywhere — written, rewritten, moved and removed with it — and
// syncCashCredit is the one place that does that, so every writer that touches
// a holding leaves its credits agreeing with it.

// lockedCashCredit is a credit as a rewrite has to know it: what kind it is,
// what it put in the balance, and what the balance holds now.
type lockedCashCredit struct {
	id       uuid.UUID
	txnType  TransactionType
	amount   decimal.Decimal
	balance  decimal.Decimal
	currency money.Currency
}

// syncCashCredit makes the cash credit of one transaction what the transaction,
// as it is stored now, says it should be. want is whether the owner wants the
// money in cash; a row that is neither a dividend nor a sale — an edit that
// turned one into a fee — keeps no credit whatever want says.
//
// The credit lands on the main account of the platform that holds the share, in
// the position's cost currency: the currency the account settled the money in.
// Its amount is what the growth series books the transaction at — quantity ×
// price × rate, less the fees in that currency, read by the same SQL — so the
// flow out of the holding and the flow into the balance cancel exactly.
//
// A sale's credit also carries what its proceeds cost: the sold quantity at the
// position's average cost. That is the capital the shares carried out, and the
// rest of the proceeds is the gain they made. The average moves whenever a
// purchase of the position does, so every writer of a position re-syncs its
// credits afterwards (syncEntryCashCredits) and the capital a portfolio shows
// stays what was paid for its purchases.
//
// A credit already there is rewritten in place while it stays the same kind in
// the same currency. When either changed — a dividend edited into a sale, a
// position restated in another currency (ChangeEntrySettlement) — it is taken
// out of the old balance and written into the new one. Either way the balance
// it leaves is checked first, as a withdrawal's would be: money the balance
// already spent cannot shrink or go out from under it.
func syncCashCredit(ctx context.Context, tx pgx.Tx, userID, txnID uuid.UUID, want bool) error {
	var (
		txnType     TransactionType
		portfolioID uuid.UUID
		sourceID    uuid.UUID
		cur         money.Currency
		assetType   market.AssetType
		amount      decimal.Decimal
		soldCost    decimal.Decimal
		date        time.Time
		notes       string
	)

	if err := tx.QueryRow(ctx, `
		SELECT t.type, pe.portfolio_id, pe.source_id, pe.cost_currency, a.asset_type,
		       ROUND(t.quantity * t.price * t.fx_rate
		             - transaction_fees_in_cost(t.fees, t.fees_currency, t.currency, t.fx_rate), 8),
		       ROUND(t.quantity * pe.price, 8),
		       t.transaction_date, COALESCE(t.notes, '')
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN portfolios p         ON p.id = pe.portfolio_id
		JOIN assets a             ON a.id = pe.asset_id
		WHERE t.id = $1 AND p.user_id = $2
	`, txnID, userID).Scan(&txnType, &portfolioID, &sourceID, &cur, &assetType, &amount, &soldCost, &date, &notes); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTransactionNotFound
		}

		return err
	}

	current, found, err := lockCashCredit(ctx, tx, txnID)
	if err != nil {
		return err
	}

	creditType, creditable := txnType.cashCredit()
	if !want || !creditable {
		if found {
			return removeCashCredit(ctx, tx, current)
		}

		return nil
	}

	switch {
	case assetType == market.Cash:
		return fmt.Errorf("%w: it is recorded on a cash balance, which is cash already", ErrNotCreditable)
	case !currency.IsSupported(cur):
		return fmt.Errorf("%w: no cash balance is kept in %s", ErrNotCreditable, cur)
	case !amount.IsPos():
		return fmt.Errorf("%w: its fees leave nothing to credit", ErrNotCreditable)
	}

	// Only a sale's proceeds cost anything (000052).
	var costBasis *decimal.Decimal
	if creditType == CashSale {
		costBasis = &soldCost
	}

	if found && current.txnType == creditType && current.currency == cur {
		after := current.balance.Sub(current.amount).Add(amount)
		if after.IsNeg() {
			return fmt.Errorf("%w: the change would leave the balance at %s %s", ErrInsufficientCash, after.String(), cur)
		}

		// A write that changes nothing is skipped, so re-syncing every credit of a
		// position after an unrelated edit fires no trigger.
		_, err := tx.Exec(ctx, `
			UPDATE transactions SET
				quantity         = $2::numeric,
				transaction_date = $3::date,
				notes            = $4,
				cost_basis       = $5::numeric,
				updated_at       = NOW()
			WHERE id = $1
			  AND (quantity, transaction_date, COALESCE(notes, ''), cost_basis)
			      IS DISTINCT FROM ($2::numeric, $3::date, $4, $5::numeric)
		`, current.id, amount.String(), date, notes, decimalParam(costBasis))

		return err
	}

	if found {
		if err := removeCashCredit(ctx, tx, current); err != nil {
			return err
		}
	}

	kind := CashKindDividend
	if creditType == CashSale {
		kind = CashKindSale
	}

	_, err = writeCashMovement(ctx, tx, userID, portfolioID, sourceID, nil, CashMovementInput{
		Kind:      kind,
		Amount:    amount,
		Currency:  cur,
		Date:      date,
		Notes:     notes,
		creditOf:  &txnID,
		costBasis: costBasis,
	})

	return err
}

// decimalParam is a nullable decimal as a query parameter: NULL for nil, its
// text otherwise, the form every numeric here is written in.
func decimalParam(d *decimal.Decimal) *string {
	if d == nil {
		return nil
	}

	s := d.String()

	return &s
}

// lockCashCredit finds the credit of a transaction and locks the balance it
// sits on. found is false when the transaction has none.
func lockCashCredit(ctx context.Context, tx pgx.Tx, txnID uuid.UUID) (lockedCashCredit, bool, error) {
	var c lockedCashCredit

	err := tx.QueryRow(ctx, `
		SELECT c.id, c.type, c.quantity, pe.quantity, pe.cost_currency
		FROM transactions c
		JOIN portfolio_entries pe ON pe.id = c.entry_id
		WHERE c.credited_from = $1
		FOR UPDATE OF pe
	`, txnID).Scan(&c.id, &c.txnType, &c.amount, &c.balance, &c.currency)
	if errors.Is(err, pgx.ErrNoRows) {
		return lockedCashCredit{}, false, nil
	}

	if err != nil {
		return lockedCashCredit{}, false, err
	}

	return c, true, nil
}

// removeCashCredit takes a credit out of its balance, unless the balance
// already spent it.
func removeCashCredit(ctx context.Context, tx pgx.Tx, c lockedCashCredit) error {
	if after := c.balance.Sub(c.amount); after.IsNeg() {
		return fmt.Errorf("%w: taking the credit out would leave the balance at %s %s", ErrInsufficientCash, after.String(), c.currency)
	}

	_, err := tx.Exec(ctx, `DELETE FROM transactions WHERE id = $1`, c.id)

	return err
}

// syncEntryCashCredits runs syncCashCredit over every transaction of a position
// that has a credit, keeping each one (want) or taking each one out.
//
// It is what a write to the position calls once the write is done. A purchase
// added, edited or deleted moves the average cost the credited sales carried
// out; a restatement moves every credit to the new currency; a deletion of the
// position takes them all back.
func syncEntryCashCredits(ctx context.Context, tx pgx.Tx, userID, entryID uuid.UUID, want bool) error {
	rows, err := tx.Query(ctx, `
		SELECT t.id
		FROM transactions t
		WHERE t.entry_id = $1
		  AND EXISTS (SELECT 1 FROM transactions c WHERE c.credited_from = t.id)
		ORDER BY t.transaction_date, t.created_at
	`, entryID)
	if err != nil {
		return err
	}

	credited, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		return err
	}

	for _, id := range credited {
		if err := syncCashCredit(ctx, tx, userID, id, want); err != nil {
			return err
		}
	}

	return nil
}
