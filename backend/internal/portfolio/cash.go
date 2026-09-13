package portfolio

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
	"github.com/yeferson59/finexia-app/internal/platform/currency"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// A cash balance is not a new kind of record. It is a position — an asset of
// type cash, held on a platform inside a portfolio — whose units are units of
// its own currency, so it is worth exactly its quantity. That is what lets it
// count in every total, allocation and growth series the app already has
// without any of them learning about it.
//
// What this file adds is the vocabulary the cash screens speak (a deposit, a
// withdrawal, the interest the balance earned) and the rules that keep that
// vocabulary from reaching positions it does not describe. The accounting rules
// behind it live in SQL, next to the ones for every other transaction type:
// see migrations 000040 and 000041.

var (
	// ErrInvalidCashMovement rejects a movement that cannot be recorded as
	// stated. It is wrapped with the rule that was broken.
	ErrInvalidCashMovement = httpx.AsBadRequest(errors.New("invalid cash movement"))
	// ErrInsufficientCash refuses a withdrawal — or an edit or deletion — that
	// would leave the balance below zero. It is a conflict rather than a bad
	// request: the same movement is valid against a balance that holds enough,
	// and the recalculation trigger would otherwise clamp the balance to zero in
	// silence while the withdrawal stayed counted as money taken out.
	ErrInsufficientCash = httpx.AsConflict(errors.New("insufficient cash balance"))
	// ErrCashMovementNotFound answers for a transaction that does not exist,
	// belongs to someone else, or is not on a cash position. The three are the
	// same thing to these endpoints.
	ErrCashMovementNotFound = httpx.AsNotFound(errors.New("cash movement not found"))
	// ErrCashMovementNotEditable refuses to rewrite, as a plain cash movement, a
	// row that is not one: a paid-out interest recorded against a balance, or a
	// deposit priced in a currency other than the balance's own. Rewriting it at
	// one unit per unit would change what it cost.
	ErrCashMovementNotEditable = httpx.AsBadRequest(errors.New("cash movement cannot be edited as a cash movement"))
	// ErrCashInterestOutsideCash refuses cash_interest on a position that is not
	// a cash balance. On a share it would add units that cost nothing, and the
	// average cost would spread over them.
	ErrCashInterestOutsideCash = httpx.AsBadRequest(errors.New("cash_interest can only be recorded on a cash position"))
)

// maxCashNotesLen mirrors transactions.notes VARCHAR(500).
const maxCashNotesLen = 500

// AllowedOn reports whether a transaction of this type can be recorded on a
// position holding an asset of type assetType. Only cash_interest is
// restricted: every other type was accepted on every asset before it existed.
func (t TransactionType) AllowedOn(assetType market.AssetType) bool {
	return t != CashInterest || assetType == market.Cash
}

// CashMovementKind is what the owner did with a cash balance, in the words the
// cash screens use. Each kind is recorded as exactly one transaction type.
type CashMovementKind string

const (
	CashKindDeposit    CashMovementKind = "deposit"
	CashKindWithdrawal CashMovementKind = "withdrawal"
	CashKindInterest   CashMovementKind = "interest"
	// CashKindOther is never written. It is how a row that is none of the three
	// reads back: a paid-out interest or a fee recorded against a balance before
	// these screens existed.
	CashKindOther CashMovementKind = "other"
)

// IsValid reports whether the kind can be written. Other cannot.
func (k CashMovementKind) IsValid() bool {
	switch k {
	case CashKindDeposit, CashKindWithdrawal, CashKindInterest:
		return true
	default:
		return false
	}
}

// TransactionType is the type the kind is recorded as.
func (k CashMovementKind) TransactionType() TransactionType {
	switch k {
	case CashKindDeposit:
		return TransferIn
	case CashKindWithdrawal:
		return TransferOut
	case CashKindInterest:
		return CashInterest
	default:
		return ""
	}
}

// cashKindOf reads a transaction on a cash position back as a kind.
//
// A buy and a sell count as a deposit and a withdrawal. They are how a balance
// was recorded before transfer_in and transfer_out were offered for it, and they
// mean the same to the quantity and to the growth series.
func cashKindOf(t TransactionType) CashMovementKind {
	switch t {
	case TransferIn, Buy:
		return CashKindDeposit
	case TransferOut, Sell:
		return CashKindWithdrawal
	case CashInterest:
		return CashKindInterest
	default:
		return CashKindOther
	}
}

// balanceEffect is how far a transaction moves the quantity of its position:
// the Go side of the quantity arms in recalculate_avg_cost (000041). It is what
// lets a write check the balance it will leave before the trigger computes it.
func balanceEffect(t TransactionType, quantity decimal.Decimal) decimal.Decimal {
	switch t {
	case Buy, TransferIn, CashInterest:
		return quantity
	case Sell, TransferOut:
		return quantity.Neg()
	default:
		return decimal.Zero
	}
}

// CashMovementInput is one movement as the owner states it. Currency is only
// read when the movement is created: an edit stays in the currency of the
// balance it is on.
type CashMovementInput struct {
	Kind     CashMovementKind
	Amount   decimal.Decimal
	Fees     decimal.Decimal
	Currency money.Currency
	Date     time.Time
	Notes    string
}

func invalidCash(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidCashMovement, fmt.Sprintf(format, args...))
}

// Validate checks everything about a movement that does not depend on the
// balance it lands on. The balance itself is checked where it can be locked,
// in the repository.
//
// Two refusals are about the growth series rather than about form:
//
//   - interest carries no fees. cash_interest has no flow at all, so a fee on it
//     would be subtracted from nothing and vanish; the amount to record is the
//     net interest credited.
//   - a withdrawal's fees cannot exceed the amount. The flow of a transfer_out is
//     the amount minus its fees, and past that point it turns into money the
//     owner put in.
func (in CashMovementInput) Validate() error {
	if !in.Kind.IsValid() {
		return invalidCash("kind must be one of: deposit, withdrawal, interest")
	}

	if !in.Amount.IsPos() {
		return invalidCash("amount must be greater than zero")
	}

	if in.Fees.IsNeg() {
		return invalidCash("fees cannot be negative")
	}

	if in.Kind == CashKindInterest && !in.Fees.IsZero() {
		return invalidCash("interest is recorded net of fees, so fees must be zero")
	}

	if in.Kind == CashKindWithdrawal && in.Fees.GreaterThan(in.Amount) {
		return invalidCash("the fees of a withdrawal cannot exceed its amount")
	}

	if in.Date.IsZero() {
		return invalidCash("date is required")
	}

	if utf8.RuneCountInString(in.Notes) > maxCashNotesLen {
		return invalidCash("notes cannot exceed %d characters", maxCashNotesLen)
	}

	return nil
}

// ValidateNew is Validate plus what only a new movement states: which currency
// the balance is in. The list is the one every ?currency= accepts, because a
// balance in a currency with no rate would sit in every total at face value.
func (in CashMovementInput) ValidateNew() error {
	if !currency.IsSupported(in.Currency) {
		return invalidCash("currency must be one of: %s", currency.List())
	}

	return in.Validate()
}

// transactionInput is the movement as the transaction that records it. The
// amount is the quantity and the price is one unit of the balance's currency,
// which is what makes the quantity the balance; the fees ride in the same
// currency, so TransactionInput.Validate has no conversion to judge.
func (in CashMovementInput) transactionInput(cur money.Currency) TransactionInput {
	return TransactionInput{
		Type:            in.Kind.TransactionType(),
		Quantity:        in.Amount,
		Price:           money.NewFromDecimal(decimal.One, cur),
		Currency:        cur,
		FXRate:          decimal.One,
		Fees:            money.NewFromDecimal(in.Fees, cur),
		FeesCurrency:    cur,
		TransactionDate: in.Date,
		Notes:           strings.TrimSpace(in.Notes),
	}
}

// cashTicker is the catalog ticker of the balance the app opens in cur. The
// prefix keeps it clear of real tickers: USD is also an ETF.
func cashTicker(cur money.Currency) string {
	return "CASH-" + cur.String()
}

var cashCurrencyNames = map[money.Currency]string{
	money.USD: "dólares estadounidenses",
	money.COP: "pesos colombianos",
	money.EUR: "euros",
	money.GBP: "libras esterlinas",
	money.CHF: "francos suizos",
	money.JPY: "yenes",
	money.CAD: "dólares canadienses",
	money.AUD: "dólares australianos", //nolint:misspell // Spanish, not a misspelling of "australians".
	money.CNY: "yuanes",
	money.MXN: "pesos mexicanos",
	money.BRL: "reales",
}

// cashAssetName is the catalog name of that balance, in the language of the
// screens that list it.
func cashAssetName(cur money.Currency) string {
	if name, ok := cashCurrencyNames[cur]; ok {
		return fmt.Sprintf("Efectivo en %s (%s)", name, cur)
	}

	return fmt.Sprintf("Efectivo (%s)", cur)
}

// CashBalance is one cash position: what a platform holds in one currency for
// one portfolio.
type CashBalance struct {
	EntryID       uuid.UUID `json:"entryId"`
	PortfolioID   uuid.UUID `json:"portfolioId"`
	PortfolioName string    `json:"portfolioName"`
	SourceID      uuid.UUID `json:"sourceId"`
	SourceName    string    `json:"sourceName"`
	AssetID       uuid.UUID `json:"assetId"`
	Ticker        string    `json:"ticker"`
	Name          string    `json:"name"`
	// Balance is what the position holds, in Currency. For every balance these
	// screens open that is its quantity; one recorded by hand at another price
	// is valued the way the rest of the app values it.
	Balance  string         `json:"balance"`
	Currency money.Currency `json:"currency"`
	// Value is Balance in DisplayCurrency, so the rows add up. FXConverted is
	// false when there was no rate and Value is Balance at face value — the same
	// contract as a holding's.
	Value           string         `json:"value"`
	DisplayCurrency money.Currency `json:"displayCurrency"`
	FXConverted     bool           `json:"fxConverted"`
	// Movements and LastMovementDate say whether an empty balance is an account
	// that was emptied or one that was never used.
	Movements        int64      `json:"movements"`
	LastMovementDate *time.Time `json:"lastMovementDate"`
}

// CashMovement is one transaction on a cash position, read as a movement.
type CashMovement struct {
	ID      uuid.UUID        `json:"id"`
	EntryID uuid.UUID        `json:"entryId"`
	Type    TransactionType  `json:"type"`
	Kind    CashMovementKind `json:"kind"`
	// Amount is what moved, in Currency — the balance's currency — with the
	// transaction's own rate applied.
	Amount       string         `json:"amount"`
	Currency     money.Currency `json:"currency"`
	Fees         string         `json:"fees"`
	FeesCurrency money.Currency `json:"feesCurrency"`
	Date         time.Time      `json:"date"`
	Notes        string         `json:"notes"`
	// Editable is whether PUT /portfolios/cash/movements/:id can rewrite it: a
	// known kind, at one unit per unit, in the balance's own currency.
	Editable      bool      `json:"editable"`
	PortfolioID   uuid.UUID `json:"portfolioId"`
	PortfolioName string    `json:"portfolioName"`
	SourceID      uuid.UUID `json:"sourceId"`
	SourceName    string    `json:"sourceName"`
	Ticker        string    `json:"ticker"`
	CreatedAt     time.Time `json:"createdAt"`
}
