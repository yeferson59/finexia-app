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

// A transaction that settled against the account's own cash is two rows: the
// transaction on the holding, and its cash side on the balance of the same
// platform, which is where the money went or came from (000049 to 000054). A
// dividend and a sale pay money in; a purchase takes it out. The cash side
// follows its transaction everywhere — written, rewritten, moved and removed
// with it — and syncCashLink is the one place that does that, so every writer
// that touches a holding leaves its cash rows agreeing with it.

// lockedCashLink is a cash side as a rewrite has to know it: what kind it is,
// what it moved in the balance, which drawer of the account that balance is,
// and what it holds now.
type lockedCashLink struct {
	id       uuid.UUID
	txnType  TransactionType
	amount   decimal.Decimal
	balance  decimal.Decimal
	currency money.Currency
	pocket   *uuid.UUID
}

// cashPlacement is which balance of the account a cash row goes in, as the
// caller states it: nil pocket is the main account, and a caller that states
// nothing leaves the row where it already is.
//
// The distinction is the whole point. A purchase names the drawer its money
// came out of, and an edit that names another moves it. Everything else —
// re-syncing a position's rows after an unrelated write — must not move
// anything, and a plain *uuid.UUID could not tell "the main account" from "I am
// not the one who decides".
type cashPlacement struct {
	stated bool
	pocket *uuid.UUID
}

// statedPocket is the placement a writer builds from what the owner sent, where
// the zero UUID is the main account — the same convention CreateCashMovement
// takes.
func statedPocket(pocketID uuid.UUID) cashPlacement {
	if pocketID == (uuid.UUID{}) {
		return cashPlacement{stated: true}
	}

	return cashPlacement{stated: true, pocket: &pocketID}
}

// samePocket compares two drawers, either of which may be the main account.
func samePocket(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == b
	}

	return *a == *b
}

// syncCashLink makes the cash row of one transaction what the transaction, as
// it is stored now, says it should be. want is whether the owner wants the
// money to move through cash; a row that settles against nothing — an edit that
// turned a dividend into a fee — keeps no cash row whatever want says.
//
// The row lands on the balance the platform keeps in the position's cost
// currency — the currency the account settled the money in — in the drawer
// `where` names: the main account, or the pocket a purchase was paid from. A
// fixed deposit is refused, like every other write into one.
// Its amount is what the growth series books the transaction at — quantity ×
// price × rate, and the fees in that currency subtracted on the way in or added
// on the way out — so the flow through the holding and the flow through the
// balance cancel exactly.
//
// A sale's credit also carries what its proceeds cost: the sold quantity at the
// position's average cost. That is the capital the shares carried out, and the
// rest of the proceeds is the gain they made. The average moves whenever a
// purchase of the position does, so every writer of a position re-syncs its
// cash rows afterwards (syncEntryCashLinks) and the capital a portfolio shows
// stays what was paid for its purchases.
//
// A row already there is rewritten in place while it stays the same kind in the
// same currency. When either changed — a dividend edited into a sale, a
// position restated in another currency (ChangeEntrySettlement) — it is taken
// out of the old balance and written into the new one. Either way the balance
// it leaves is checked first, as a withdrawal's would be: money the balance
// already spent cannot shrink or go out from under it, and a purchase cannot
// take out money that is not there.
func syncCashLink(ctx context.Context, tx pgx.Tx, userID, txnID uuid.UUID, want bool, where cashPlacement) error {
	var (
		txnType     TransactionType
		portfolioID uuid.UUID
		sourceID    uuid.UUID
		cur         money.Currency
		assetType   market.AssetType
		gross       decimal.Decimal
		fees        decimal.Decimal
		soldCost    decimal.Decimal
		date        time.Time
		notes       string
	)

	// The gross and the fees are read apart rather than netted in SQL, because
	// which way the commission goes is what tells a credit from a debit: the
	// broker keeps it either way, so the money that arrives is short of it and
	// the money that leaves is over it.
	if err := tx.QueryRow(ctx, `
		SELECT t.type, pe.portfolio_id, pe.source_id, pe.cost_currency, a.asset_type,
		       ROUND(t.quantity * t.price * t.fx_rate, 8),
		       ROUND(transaction_fees_in_cost(t.fees, t.fees_currency, t.currency, t.fx_rate), 8),
		       ROUND(t.quantity * pe.price, 8),
		       t.transaction_date, COALESCE(t.notes, '')
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN portfolios p         ON p.id = pe.portfolio_id
		JOIN assets a             ON a.id = pe.asset_id
		WHERE t.id = $1 AND p.user_id = $2
	`, txnID, userID).Scan(&txnType, &portfolioID, &sourceID, &cur, &assetType, &gross, &fees, &soldCost, &date, &notes); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTransactionNotFound
		}

		return err
	}

	current, found, err := lockCashLink(ctx, tx, txnID)
	if err != nil {
		return err
	}

	linkType, linkable := txnType.cashLink()
	if !want || !linkable {
		if found {
			return removeCashLink(ctx, tx, current)
		}

		return nil
	}

	amount := gross.Sub(fees)
	refusal := ErrNotCreditable
	nothing := "its fees leave nothing to credit"

	if linkType == CashPurchase {
		amount = gross.Add(fees)
		refusal = ErrNotPayableFromCash
		nothing = "it cost nothing, so there is nothing to pay"
	}

	// Where the row goes, if it has to be written. What the owner stated wins;
	// otherwise it stays in the drawer it is in. The exception is a restatement:
	// a pocket holds one currency, so the one the row sat in cannot hold the new
	// one and the money goes back to the main account, which every account has
	// in every currency.
	target := current.pocket

	switch {
	case where.stated:
		target = where.pocket
	case found && current.currency != cur:
		target = nil
	}

	if target != nil {
		if _, err := requireWritablePocket(ctx, tx, userID, *target, sourceID, cur); err != nil {
			return err
		}
	}

	switch {
	case assetType == market.Cash:
		return fmt.Errorf("%w: it is recorded on a cash balance, which is cash already", refusal)
	case !currency.IsSupported(cur):
		return fmt.Errorf("%w: no cash balance is kept in %s", refusal, cur)
	case !amount.IsPos():
		return fmt.Errorf("%w: %s", refusal, nothing)
	}

	// Only a sale's proceeds cost anything (000052).
	var costBasis *decimal.Decimal
	if linkType == CashSale {
		costBasis = &soldCost
	}

	if found && current.txnType == linkType && current.currency == cur && samePocket(current.pocket, target) {
		after := current.balance.
			Sub(balanceEffect(linkType, current.amount)).
			Add(balanceEffect(linkType, amount))
		if after.IsNeg() {
			return fmt.Errorf("%w: the change would leave the balance at %s %s", ErrInsufficientCash, after.String(), cur)
		}

		// A write that changes nothing is skipped, so re-syncing every cash row of
		// a position after an unrelated edit fires no trigger.
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
		if err := removeCashLink(ctx, tx, current); err != nil {
			return err
		}
	}

	_, err = writeCashMovement(ctx, tx, userID, portfolioID, sourceID, target, CashMovementInput{
		Kind:      cashKindOf(linkType),
		Amount:    amount,
		Currency:  cur,
		Date:      date,
		Notes:     notes,
		linkedTo:  &txnID,
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

// lockCashLink finds the cash row of a transaction and locks the balance it
// sits on. found is false when the transaction has none.
func lockCashLink(ctx context.Context, tx pgx.Tx, txnID uuid.UUID) (lockedCashLink, bool, error) {
	var c lockedCashLink

	err := tx.QueryRow(ctx, `
		SELECT c.id, c.type, c.quantity, pe.quantity, pe.cost_currency, pe.pocket_id
		FROM transactions c
		JOIN portfolio_entries pe ON pe.id = c.entry_id
		WHERE c.credited_from = $1
		FOR UPDATE OF pe
	`, txnID).Scan(&c.id, &c.txnType, &c.amount, &c.balance, &c.currency, &c.pocket)
	if errors.Is(err, pgx.ErrNoRows) {
		return lockedCashLink{}, false, nil
	}

	if err != nil {
		return lockedCashLink{}, false, err
	}

	return c, true, nil
}

// removeCashLink undoes a cash row: a credit comes back out of its balance,
// unless the balance already spent it, and a debit puts its money back, which
// nothing can refuse.
func removeCashLink(ctx context.Context, tx pgx.Tx, c lockedCashLink) error {
	if after := c.balance.Sub(balanceEffect(c.txnType, c.amount)); after.IsNeg() {
		return fmt.Errorf("%w: taking the credit out would leave the balance at %s %s", ErrInsufficientCash, after.String(), c.currency)
	}

	_, err := tx.Exec(ctx, `DELETE FROM transactions WHERE id = $1`, c.id)

	return err
}

// syncEntryCashLinks runs syncCashLink over every transaction of a position
// that has a cash row, keeping each one (want) or taking each one out.
//
// It is what a write to the position calls once the write is done. A purchase
// added, edited or deleted moves the average cost the credited sales carried
// out; a restatement moves every cash row to the new currency; a deletion of
// the position takes them all back.
func syncEntryCashLinks(ctx context.Context, tx pgx.Tx, userID, entryID uuid.UUID, want bool) error {
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

	linked, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		return err
	}

	for _, id := range linked {
		// No placement: an unrelated write to the position must not move a
		// purchase's money out of the drawer the owner paid it from.
		if err := syncCashLink(ctx, tx, userID, id, want, cashPlacement{}); err != nil {
			return err
		}
	}

	return nil
}
