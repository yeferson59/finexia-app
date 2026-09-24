package portfolio

import (
	"errors"
	"time"
	"unicode/utf8"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// The contributions and withdrawals of a fund followed by balance. They are the
// fund's purchases and sales, stated in money: the units are the replay's
// business (fund_units.go), so the owner never types one.

// ErrFundMovementNotFound answers for a transaction that does not exist,
// belongs to someone else, or is not a movement of a fund followed by balance.
var ErrFundMovementNotFound = httpx.AsNotFound(errors.New("fund movement not found"))

// FundMovementKind is which way the money went.
type FundMovementKind string

const (
	FundContribution FundMovementKind = "contribution"
	FundWithdrawal   FundMovementKind = "withdrawal"
)

// FundMovement is one contribution or withdrawal, with the units it came to.
type FundMovement struct {
	TxnID         uuid.UUID        `json:"txnId"`
	EntryID       uuid.UUID        `json:"entryId"`
	PortfolioID   uuid.UUID        `json:"portfolioId"`
	PortfolioName string           `json:"portfolioName"`
	SourceName    string           `json:"sourceName"`
	Kind          FundMovementKind `json:"kind"`
	Date          time.Time        `json:"date"`
	// Amount is the money that went in, or came out before Fees. For a fund
	// followed by units it is the units at their price.
	Amount string `json:"amount"`
	Fees   string `json:"fees"`
	// All is a withdrawal of everything its position held.
	All       bool      `json:"all"`
	Units     string    `json:"units"`
	UnitValue string    `json:"unitValue"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"createdAt"`
}

// FundContributionInput is money put into a fund followed by balance.
type FundContributionInput struct {
	// PortfolioID and SourceID are the position it lands on, opened if the
	// portfolio did not hold the fund on that platform yet.
	PortfolioID uuid.UUID
	SourceID    uuid.UUID
	Date        time.Time
	Amount      decimal.Decimal
	// BalanceBefore, when positive, is what the fund held at the close of the
	// day before — the one figure that makes the contribution trade at the
	// exact unit value of its day rather than the last one known. It is written
	// as that day's balance.
	BalanceBefore decimal.Decimal
	PayFromCash   bool
	CashPocketID  uuid.UUID
	Notes         string
}

// Validate checks what the contribution states. today is the server's clock.
func (in FundContributionInput) Validate(today time.Time) error {
	if in.PortfolioID == (uuid.UUID{}) || in.SourceID == (uuid.UUID{}) {
		return invalidFund("portfolioId and sourceId are required")
	}

	if err := validateFundDate(in.Date, today, invalidFund); err != nil {
		return err
	}

	if err := validateFundAmount(in.Amount, "amount", invalidFund); err != nil {
		return err
	}

	if in.BalanceBefore.IsNeg() {
		return invalidFund("balanceBefore cannot be negative")
	}

	if in.BalanceBefore.IsPos() {
		if err := validateFundAmount(in.BalanceBefore, "balanceBefore", invalidFund); err != nil {
			return err
		}
	}

	return validateFundNotes(in.Notes)
}

// FundWithdrawalInput is money taken out of a fund followed by balance.
type FundWithdrawalInput struct {
	// EntryID is the position the units leave: a portfolio can only take out
	// what it holds.
	EntryID uuid.UUID
	Date    time.Time
	// Amount is what the fund paid out, before Fees: what the platform kept —
	// a penalty for leaving early, the tax on the movement.
	Amount decimal.Decimal
	Fees   decimal.Decimal
	// All takes everything the position holds.
	All          bool
	CreditCash   bool
	CashPocketID uuid.UUID
	Notes        string
}

// Validate checks what the withdrawal states. today is the server's clock.
func (in FundWithdrawalInput) Validate(today time.Time) error {
	if in.EntryID == (uuid.UUID{}) {
		return invalidFund("entryId is required")
	}

	return validateFundMovementValues(today, in.Date, in.Amount, in.Fees, true, in.Notes)
}

// FundMovementEdit is a contribution or withdrawal restated: its day, its
// money and its note. Which position it is on and which way it went stay; that
// is a deletion and a new movement.
type FundMovementEdit struct {
	Date   time.Time
	Amount decimal.Decimal
	Fees   decimal.Decimal
	All    bool
	Notes  string
}

// Validate checks the edit against the kind of movement it rewrites.
func (in FundMovementEdit) Validate(today time.Time, kind FundMovementKind) error {
	withdrawal := kind == FundWithdrawal

	if !withdrawal && in.All {
		return invalidFund("only a withdrawal can take everything")
	}

	return validateFundMovementValues(today, in.Date, in.Amount, in.Fees, withdrawal, in.Notes)
}

func validateFundMovementValues(today, date time.Time, amount, fees decimal.Decimal, withdrawal bool, notes string) error {
	if err := validateFundDate(date, today, invalidFund); err != nil {
		return err
	}

	if err := validateFundAmount(amount, "amount", invalidFund); err != nil {
		return err
	}

	switch {
	case fees.IsNeg():
		return invalidFund("fees cannot be negative")
	case !withdrawal && fees.IsPos():
		return invalidFund("a contribution carries no fees: record what went into the fund")
	case fees.GreaterThanOrEqual(amount) && fees.IsPos():
		return invalidFund("fees must be less than the amount withdrawn")
	}

	return validateFundNotes(notes)
}

func validateFundNotes(notes string) error {
	if utf8.RuneCountInString(notes) > maxCashNotesLen {
		return invalidFund("notes cannot exceed %d characters", maxCashNotesLen)
	}

	return nil
}
