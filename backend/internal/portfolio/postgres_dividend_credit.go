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

// A dividend paid into cash is two rows: the dividend on the holding, which is
// the income, and its credit on the cash balance of the same platform, which is
// where the money went (000049, 000050). The credit follows the dividend
// everywhere — written, rewritten, moved and removed with it — and
// syncDividendCredit is the one place that does that, so every writer that
// touches a dividend leaves its credit agreeing with it.

// lockedDividendCredit is a credit as a rewrite has to know it: what it put in
// the balance, and what the balance holds now.
type lockedDividendCredit struct {
	id       uuid.UUID
	entryID  uuid.UUID
	amount   decimal.Decimal
	balance  decimal.Decimal
	currency money.Currency
}

// syncDividendCredit makes the cash credit of one dividend what the dividend,
// as it is stored now, says it should be. want is whether the owner wants the
// money in cash; a row that is not a dividend — an edit that turned one into a
// fee — keeps no credit whatever want says.
//
// The credit lands on the main account of the platform that holds the share, in
// the position's cost currency: the currency the account settled the dividend
// in. Its amount is what the growth series books the dividend at — quantity ×
// price × rate, less the fees in that currency, read by the same SQL — so the
// flow out of the holding and the flow into the balance cancel exactly.
//
// A credit already there is rewritten in place while it stays in the same
// currency. When the currency changed — the position was restated in another
// (ChangeEntrySettlement) — it is taken out of the old balance and written into
// the new one. Either way the balance it leaves is checked first, as a
// withdrawal's would be: a dividend whose money was spent cannot shrink or go
// out from under the balance that spent it.
func syncDividendCredit(ctx context.Context, tx pgx.Tx, userID, dividendID uuid.UUID, want bool) error {
	var (
		txnType     TransactionType
		portfolioID uuid.UUID
		sourceID    uuid.UUID
		cur         money.Currency
		assetType   market.AssetType
		amount      decimal.Decimal
		date        time.Time
		notes       string
	)

	if err := tx.QueryRow(ctx, `
		SELECT t.type, pe.portfolio_id, pe.source_id, pe.cost_currency, a.asset_type,
		       ROUND(t.quantity * t.price * t.fx_rate
		             - transaction_fees_in_cost(t.fees, t.fees_currency, t.currency, t.fx_rate), 8),
		       t.transaction_date, COALESCE(t.notes, '')
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN portfolios p         ON p.id = pe.portfolio_id
		JOIN assets a             ON a.id = pe.asset_id
		WHERE t.id = $1 AND p.user_id = $2
	`, dividendID, userID).Scan(&txnType, &portfolioID, &sourceID, &cur, &assetType, &amount, &date, &notes); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTransactionNotFound
		}

		return err
	}

	current, found, err := lockDividendCredit(ctx, tx, dividendID)
	if err != nil {
		return err
	}

	if !want || txnType != Dividend {
		if found {
			return removeDividendCredit(ctx, tx, current)
		}

		return nil
	}

	switch {
	case assetType == market.Cash:
		return fmt.Errorf("%w: it is recorded on a cash balance, which is cash already", ErrDividendNotCreditable)
	case !currency.IsSupported(cur):
		return fmt.Errorf("%w: no cash balance is kept in %s", ErrDividendNotCreditable, cur)
	case !amount.IsPos():
		return fmt.Errorf("%w: its fees leave nothing to credit", ErrDividendNotCreditable)
	}

	if found && current.currency == cur {
		after := current.balance.Sub(current.amount).Add(amount)
		if after.IsNeg() {
			return fmt.Errorf("%w: the change would leave the balance at %s %s", ErrInsufficientCash, after.String(), cur)
		}

		_, err := tx.Exec(ctx, `
			UPDATE transactions SET
				quantity         = $2::numeric,
				transaction_date = $3::date,
				notes            = $4,
				updated_at       = NOW()
			WHERE id = $1
		`, current.id, amount.String(), date, notes)

		return err
	}

	if found {
		if err := removeDividendCredit(ctx, tx, current); err != nil {
			return err
		}
	}

	_, err = writeCashMovement(ctx, tx, userID, portfolioID, sourceID, nil, CashMovementInput{
		Kind:     CashKindDividend,
		Amount:   amount,
		Currency: cur,
		Date:     date,
		Notes:    notes,
		dividend: &dividendID,
	})

	return err
}

// lockDividendCredit finds the credit of a dividend and locks the balance it
// sits on. found is false when the dividend has none.
func lockDividendCredit(ctx context.Context, tx pgx.Tx, dividendID uuid.UUID) (lockedDividendCredit, bool, error) {
	var c lockedDividendCredit

	err := tx.QueryRow(ctx, `
		SELECT c.id, c.entry_id, c.quantity, pe.quantity, pe.cost_currency
		FROM transactions c
		JOIN portfolio_entries pe ON pe.id = c.entry_id
		WHERE c.dividend_id = $1
		FOR UPDATE OF pe
	`, dividendID).Scan(&c.id, &c.entryID, &c.amount, &c.balance, &c.currency)
	if errors.Is(err, pgx.ErrNoRows) {
		return lockedDividendCredit{}, false, nil
	}

	if err != nil {
		return lockedDividendCredit{}, false, err
	}

	return c, true, nil
}

// removeDividendCredit takes a credit out of its balance, unless the balance
// already spent it.
//
// A credit costs nothing, so a balance that held only credits is priced at zero
// (000042's walk). Once its last row is gone that price describes no unit, and
// cash_entry_at_par would read the empty balance as one opened by hand at
// another price, which the cash writers leave alone: the next deposit or credit
// would open a second balance beside it. A credit is only ever written on a
// balance the app keeps at par, so the empty one goes back to the price it was
// opened at.
func removeDividendCredit(ctx context.Context, tx pgx.Tx, c lockedDividendCredit) error {
	if after := c.balance.Sub(c.amount); after.IsNeg() {
		return fmt.Errorf("%w: taking the dividend out would leave the balance at %s %s", ErrInsufficientCash, after.String(), c.currency)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM transactions WHERE id = $1`, c.id); err != nil {
		return err
	}

	_, err := tx.Exec(ctx, `
		UPDATE portfolio_entries SET price = 1
		WHERE id = $1 AND NOT EXISTS (SELECT 1 FROM transactions WHERE entry_id = $1)
	`, c.entryID)

	return err
}

// syncEntryDividendCredits runs syncDividendCredit over every dividend of a
// position that has a credit, keeping each one (want) or taking each one out.
// It is what a write to the whole position calls: a restatement moves every
// credit to the new currency, a deletion takes them all back.
func syncEntryDividendCredits(ctx context.Context, tx pgx.Tx, userID, entryID uuid.UUID, want bool) error {
	rows, err := tx.Query(ctx, `
		SELECT t.id
		FROM transactions t
		WHERE t.entry_id = $1
		  AND EXISTS (SELECT 1 FROM transactions c WHERE c.dividend_id = t.id)
		ORDER BY t.transaction_date, t.created_at
	`, entryID)
	if err != nil {
		return err
	}

	dividends, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		return err
	}

	for _, id := range dividends {
		if err := syncDividendCredit(ctx, tx, userID, id, want); err != nil {
			return err
		}
	}

	return nil
}
