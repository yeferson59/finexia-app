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

	"github.com/yeferson59/finexia-app/internal/platform/currency"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// A pocket is a subaccount of a cash account, not a platform of its own: the
// "cajita" inside the savings account, the sub-wallet inside the broker. It
// belongs to a platform and a currency, and its money counts inside that
// platform, so every figure by platform keeps adding it in. The model is
// described in migration 000047.
//
// What a pocket owns is its rate. Where a rate, its tiers and the ledger looked
// up a platform and a currency, they look up a platform, a currency and a
// pocket: the account pays 8 % and the pocket pays 10 %, each on what its own
// balances hold.
//
// The main account is the pocket that is not there — pocket_id NULL — so a
// platform without pockets behaves exactly as it did before they existed.

var (
	// ErrCashPocketNotFound answers for a pocket that does not exist or belongs
	// to someone else. The two are the same thing to these endpoints.
	ErrCashPocketNotFound = httpx.AsNotFound(errors.New("cash pocket not found"))
	// ErrCashPocketNotEmpty refuses to delete a pocket that still holds money or
	// carries movements. It is a conflict rather than a bad request: the same
	// deletion is fine once the money has been moved out.
	ErrCashPocketNotEmpty = httpx.AsConflict(errors.New("cash pocket still holds money"))
	// ErrCashPocketNameTaken refuses two pockets of one account under one name.
	// Two accounts can each have a "Viajes"; one account cannot have two.
	ErrCashPocketNameTaken = httpx.AsConflict(errors.New("this account already has a pocket with that name"))
	// ErrInvalidCashPocket rejects a pocket that cannot be recorded as stated.
	// It is wrapped with the rule that was broken.
	ErrInvalidCashPocket = httpx.AsBadRequest(errors.New("invalid cash pocket"))
	// ErrInvalidCashMove rejects a transfer between two balances that cannot be
	// made as stated: to itself, across platforms, across currencies.
	ErrInvalidCashMove = httpx.AsBadRequest(errors.New("invalid cash move"))
)

// CashPocketKind is what a pocket is for.
type CashPocketKind string

const (
	// PocketFlexible takes deposits and withdrawals like the main account does,
	// and earns whatever rate its versions say.
	PocketFlexible CashPocketKind = "flexible"
	// PocketFixed is a deposit that keeps the rate of the day it was opened,
	// with a maturity. Opening one comes with the phase after this one; the kind
	// is read here so a pocket that is one is never written to by hand.
	PocketFixed CashPocketKind = "fixed"
)

// IsValid reports whether the kind is one the app writes.
func (k CashPocketKind) IsValid() bool {
	return k == PocketFlexible || k == PocketFixed
}

// maxCashPocketNameLen mirrors cash_pockets.name VARCHAR(100).
const maxCashPocketNameLen = 100

// CashPocket is one subaccount of a cash account.
type CashPocket struct {
	ID         uuid.UUID      `json:"id"`
	SourceID   uuid.UUID      `json:"sourceId"`
	SourceName string         `json:"sourceName"`
	Currency   money.Currency `json:"currency"`
	Name       string         `json:"name"`
	Kind       CashPocketKind `json:"kind"`
	// OpenedOn is the day it started holding money, and what a fixed deposit
	// earns from.
	OpenedOn time.Time `json:"openedOn"`
	// MaturesOn is when a fixed deposit ends, nil for a flexible pocket.
	MaturesOn *time.Time `json:"maturesOn"`
	// ClosedOn is the day it stopped, nil while it is open.
	ClosedOn *time.Time `json:"closedOn"`
	// Balance is what its balances hold together, in Currency, and Balances how
	// many portfolios hold part of it. They are what says whether it can be
	// deleted, so the screen can ask before the answer is a 409.
	Balance  string `json:"balance"`
	Balances int64  `json:"balances"`
	// Movements is how many movements were ever recorded in it, so an emptied
	// pocket reads differently from one that was never used.
	Movements int64     `json:"movements"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func invalidCashPocket(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidCashPocket, fmt.Sprintf(format, args...))
}

// NewCashPocketInput is a flexible pocket as the owner states it: which account
// it belongs to, and what it is called.
type NewCashPocketInput struct {
	SourceID uuid.UUID
	Currency money.Currency
	Name     string
}

// Validate checks everything that does not depend on the account it lands on.
// The account itself — that it is the owner's, and active — is checked where it
// can be locked, in the repository.
func (in NewCashPocketInput) Validate() error {
	if in.SourceID == (uuid.UUID{}) {
		return invalidCashPocket("sourceId is required")
	}

	// The list is the one every ?currency= accepts, for the reason a balance
	// gives: one in a currency with no rate would sit in every total at face
	// value.
	if !currency.IsSupported(in.Currency) {
		return invalidCashPocket("currency must be one of: %s", currency.List())
	}

	return validateCashPocketName(in.Name)
}

// CleanName is the name as it is stored: trimmed, so " Viajes " and "Viajes"
// are the same pocket to the unique key.
func (in NewCashPocketInput) CleanName() string {
	return strings.TrimSpace(in.Name)
}

// RenameCashPocketInput is the one thing a flexible pocket can be told to
// change. What it holds moves with movements, and when it earns moves with its
// rate.
type RenameCashPocketInput struct {
	Name string
}

func (in RenameCashPocketInput) Validate() error {
	return validateCashPocketName(in.Name)
}

// CleanName is the name as it is stored.
func (in RenameCashPocketInput) CleanName() string {
	return strings.TrimSpace(in.Name)
}

func validateCashPocketName(name string) error {
	clean := strings.TrimSpace(name)

	if clean == "" {
		return invalidCashPocket("name is required")
	}

	if utf8.RuneCountInString(clean) > maxCashPocketNameLen {
		return invalidCashPocket("name cannot exceed %d characters", maxCashPocketNameLen)
	}

	return nil
}

func invalidCashMove(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidCashMove, fmt.Sprintf(format, args...))
}

// CashMoveInput moves money between two balances of one account inside one
// portfolio: the main account and a pocket, or two pockets.
//
// It is not a deposit and a withdrawal recorded by hand. The two legs go in one
// transaction, so the money is never in both places or in neither, and because
// they offset each other the portfolio's net flow — and its return — does not
// move.
type CashMoveInput struct {
	Currency money.Currency
	// From and To are the pockets the money leaves and arrives at. The zero UUID
	// is the main account, on either side.
	From uuid.UUID
	To   uuid.UUID
	// Amount is what moves, in Currency.
	Amount decimal.Decimal
	Date   time.Time
	Notes  string
}

// Validate checks what the move states. Whether the balance it leaves holds
// enough, and whether the pockets are the owner's and flexible, is checked
// where they can be locked.
func (in CashMoveInput) Validate() error {
	if !currency.IsSupported(in.Currency) {
		return invalidCashMove("currency must be one of: %s", currency.List())
	}

	if in.From == in.To {
		return invalidCashMove("from and to must be different: moving money to where it already is does nothing")
	}

	if !in.Amount.IsPos() {
		return invalidCashMove("amount must be greater than zero")
	}

	if in.Date.IsZero() {
		return invalidCashMove("date is required")
	}

	if utf8.RuneCountInString(in.Notes) > maxCashNotesLen {
		return invalidCashMove("notes cannot exceed %d characters", maxCashNotesLen)
	}

	return nil
}

// legs is the move as the two movements that record it: the withdrawal that
// empties one side and the deposit that fills the other, both on the same day
// and with the same note, and neither carrying a fee. A fee would make the two
// stop cancelling out, and the money would read as a loss rather than as money
// that changed drawer.
func (in CashMoveInput) legs() (out, into CashMovementInput) {
	out = CashMovementInput{
		Kind:     CashKindWithdrawal,
		Amount:   in.Amount,
		Currency: in.Currency,
		Date:     in.Date,
		Notes:    in.Notes,
	}

	into = out
	into.Kind = CashKindDeposit

	return out, into
}

// CashMove is what a move did: the two movements it recorded.
type CashMove struct {
	From CashMovement `json:"from"`
	To   CashMovement `json:"to"`
}
