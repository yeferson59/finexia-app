package portfolio

import (
	"context"
	"fmt"

	"uuid"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// SettledRate is what one transaction becomes when its position is restated in
// another cost currency: the rate it converted at, and the side its commission
// was billed on.
type SettledRate struct {
	TransactionID uuid.UUID
	FXRate        decimal.Decimal
	FeesCurrency  money.Currency
}

// PlanSettlement restates every transaction of a position in the currency its
// account really settled in.
//
// A position's cost currency is chosen once, when it is opened, and every
// transaction on it is validated against it. Leaving "my account settled in
// another currency" unticked for a EUR ETF bought from a USD account had one way
// out: delete the position and load it again, which rewrote its history. This
// keeps the history and changes the two facts that were wrong — the currency
// and the rates.
//
// rates holds the rate of each transaction quoted in a currency other than
// costCurrency, the one its broker confirmation shows for that day. A
// transaction quoted in costCurrency settles at 1 whatever it is given, and a
// split, which moves no money, settles at 1 when it is given nothing. Any other
// transaction without a rate is refused by the same Validate that refuses it on
// entry: the only fallback would be today's rate, which is the error this fixes.
//
// A commission billed to the account moves with the account, so one recorded
// outside the trade's currency is now in costCurrency. One billed on the fill
// stays in the trade's currency.
func PlanSettlement(txns []Transaction, costCurrency money.Currency, rates map[uuid.UUID]decimal.Decimal) ([]SettledRate, error) {
	if len(txns) == 0 {
		return nil, ErrSettlementWithoutTransactions
	}

	known := make(map[uuid.UUID]bool, len(txns))
	for _, t := range txns {
		known[t.ID] = true
	}
	for id := range rates {
		if !known[id] {
			return nil, fmt.Errorf("%w: %s is not a transaction of this position", ErrTransactionNotFound, id)
		}
	}

	plan := make([]SettledRate, 0, len(txns))
	for _, t := range txns {
		rate := rates[t.ID]
		switch {
		case t.Currency == costCurrency:
			rate = decimal.One
		case rate.IsZero() && t.Type == Split:
			rate = decimal.One
		}

		feesCurrency := t.FeesCurrency
		if feesCurrency != t.Currency {
			feesCurrency = costCurrency
		}

		settled, err := TransactionInput{
			Type:            t.Type,
			Quantity:        t.Quantity,
			Price:           t.Price,
			Currency:        t.Currency,
			FXRate:          rate,
			Fees:            t.Fees,
			FeesCurrency:    feesCurrency,
			TransactionDate: t.TransactionDate,
			Notes:           t.Notes,
		}.Validate(costCurrency)
		if err != nil {
			return nil, fmt.Errorf("transaction of %s: %w", t.TransactionDate.Format("2006-01-02"), err)
		}

		plan = append(plan, SettledRate{
			TransactionID: t.ID,
			FXRate:        settled.FXRate,
			FeesCurrency:  settled.FeesCurrency,
		})
	}

	return plan, nil
}

// ChangeEntrySettlement restates a position the user owns in another cost
// currency. The rules are PlanSettlement's; the repository applies them inside
// one database transaction, so a refused rate leaves the position as it was.
func (s *service) ChangeEntrySettlement(ctx context.Context, userID, entryID uuid.UUID, costCurrency money.Currency, rates map[uuid.UUID]decimal.Decimal) (int, error) {
	// XXX is money.Currency's zero value, so it is what an omitted field decodes
	// to — and Valid accepts it, being the ISO code for "no currency".
	if costCurrency == money.XXX || !costCurrency.Valid() {
		return 0, httpx.AsBadRequest(fmt.Errorf("unknown settlement currency %q", costCurrency))
	}

	return s.repo.ChangeEntrySettlement(ctx, userID, entryID, costCurrency, rates)
}
