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
	// ErrCashPocketFixed refuses a write a fixed deposit does not take: a deposit,
	// a withdrawal or an interest recorded by hand in it, a new version of its
	// rate, or pausing the one it has. A deposit is one amount at one rate for
	// one term, and every one of those would make it something else. It is a
	// conflict rather than a bad request: the write is fine, the pocket is not
	// the place for it, and cancelling the deposit is what opens that place.
	ErrCashPocketFixed = httpx.AsConflict(errors.New("a fixed deposit takes no movements or rate versions of its own"))
	// ErrCashPocketClosed refuses to cancel a deposit that has already ended,
	// whether it came due or was cancelled before. It is what makes settling one
	// idempotent: the job catching up after a day down finds it closed and moves
	// nothing a second time.
	ErrCashPocketClosed = httpx.AsConflict(errors.New("the deposit is already closed"))
)

// CashPocketKind is what a pocket is for.
type CashPocketKind string

const (
	// PocketFlexible takes deposits and withdrawals like the main account does,
	// and earns whatever rate its versions say.
	PocketFlexible CashPocketKind = "flexible"
	// PocketFixed is a deposit that keeps the rate of the day it was opened, with
	// an optional maturity: a CDT, a term pocket, a promotional rate locked for
	// ninety days. It holds one deposit and one version of its rate, and takes no
	// writes by hand — see NewFixedDepositInput and ErrCashPocketFixed.
	PocketFixed CashPocketKind = "fixed"
)

// IsValid reports whether the kind is one the app writes.
func (k CashPocketKind) IsValid() bool {
	return k == PocketFlexible || k == PocketFixed
}

const (
	// maxCashPocketNameLen mirrors cash_pockets.name VARCHAR(100).
	maxCashPocketNameLen = 100
	// maxFixedDepositYears is how far back a deposit can be opened. Recording one
	// costs a day of computation per day since it opened, and a term nobody
	// quotes is a mistyped year: 2016 for 2026.
	maxFixedDepositYears = 5
)

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

// CashMoveInput moves money between two cash balances inside one portfolio:
// two drawers of one account, or two accounts — the savings app and the broker
// the money is transferred to before a purchase.
//
// It is not a deposit and a withdrawal recorded by hand. The two legs go in one
// transaction, so the money is never in both places or in neither, and because
// they offset each other the portfolio's net flow — and its return — does not
// move. Across currencies they offset at the rate the mover states rather than
// nominally: what leaves is Amount and what arrives is Amount × FXRate, so the
// value is the same on both sides and only a wrong rate books a gain.
type CashMoveInput struct {
	Currency money.Currency
	// From and To are the pockets the money leaves and arrives at. The zero UUID
	// is the main account, on either side.
	From uuid.UUID
	To   uuid.UUID
	// ToSource is the platform the money arrives at and ToCurrency the currency
	// it arrives in. Both default to the ones it left, which is the move within
	// one account that pockets were built for; naming another platform is the
	// transfer between them, and naming another currency is that transfer with
	// the conversion the platform applied.
	ToSource   uuid.UUID
	ToCurrency money.Currency
	// Amount is what leaves, in Currency, and ToAmount what arrives, in
	// ToCurrency. Within one currency they are the same number and ToAmount can
	// be left out; between two it is required.
	//
	// The two amounts are stated rather than derived from a rate because the two
	// amounts are what the mover has in front of them — the pesos that left the
	// app and the dollars that reached the broker are both printed on the
	// statement, and the rate between them is not. Stating them also records
	// exactly what happened: an amount computed from a rate lands a few cents
	// away from what really arrived, and those cents would read as a gain.
	Amount   decimal.Decimal
	ToAmount decimal.Decimal
	Date     time.Time
	Notes    string
}

// withDefaults fills the destination a move left unsaid: the account the money
// is already in. It is applied before Validate, so every caller — the HTTP
// request and anything else that builds the input — states the same move.
func (in CashMoveInput) withDefaults(sourceID uuid.UUID) CashMoveInput {
	if in.ToSource == (uuid.UUID{}) {
		in.ToSource = sourceID
	}

	if in.ToCurrency == money.XXX {
		in.ToCurrency = in.Currency
	}

	// Within one currency what arrives is what left, so it need not be said.
	// Across two, left at zero, the move is one that forgot to say how much
	// reached the other side, and Validate refuses it rather than carrying the
	// departing amount over unchanged.
	if in.ToAmount.IsZero() && in.ToCurrency == in.Currency {
		in.ToAmount = in.Amount
	}

	return in
}

// crosses reports whether the move leaves the account it started in, which is
// what makes it a transfer rather than a change of drawer.
func (in CashMoveInput) crosses(sourceID uuid.UUID) bool {
	return in.ToSource != sourceID || in.ToCurrency != in.Currency
}

// Validate checks what the move states, against the account it starts from.
// Whether the balance it leaves holds enough, and whether the pockets are the
// owner's and flexible, is checked where they can be locked.
//
// sourceID is the platform the money leaves, which is what says whether the
// destination is somewhere else: two drawers are only the same drawer when the
// account is the same too, and the same drawer number of two platforms is two
// different places.
func (in CashMoveInput) Validate(sourceID uuid.UUID) error {
	if !currency.IsSupported(in.Currency) {
		return invalidCashMove("currency must be one of: %s", currency.List())
	}

	if !currency.IsSupported(in.ToCurrency) {
		return invalidCashMove("the currency it arrives in must be one of: %s", currency.List())
	}

	if !in.crosses(sourceID) && in.From == in.To {
		return invalidCashMove("from and to must be different: moving money to where it already is does nothing")
	}

	if !in.Amount.IsPos() {
		return invalidCashMove("amount must be greater than zero")
	}

	// What arrives is checked against what leaves. Within one currency they are
	// the same money and any other figure is a conversion that did not happen;
	// across two, a missing one would carry the departing amount over with a
	// different label, which is the difference between four hundred thousand
	// pesos and four hundred thousand dollars.
	if in.ToCurrency == in.Currency {
		if in.ToAmount.Cmp(in.Amount) != 0 {
			return invalidCashMove("a move that stays in %s arrives at what it left: leave out what arrives, or state the same amount", in.Currency)
		}
	} else if !in.ToAmount.IsPos() {
		return invalidCashMove("a move from %s to %s has to say how much %s arrived", in.Currency, in.ToCurrency, in.ToCurrency)
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
// that changed hands.
//
// The deposit arrives in the destination's currency, for the amount the move
// states, rounded to the eight decimals the balances are kept at. Within one
// currency that is the amount that left, which is the move this was before it
// could cross accounts.
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
	into.Currency = in.ToCurrency
	into.Amount = in.ToAmount.RoundHAZ(8)

	return out, into
}

// CashMove is what a move did: the two movements it recorded.
type CashMove struct {
	From CashMovement `json:"from"`
	To   CashMovement `json:"to"`
}

// A fixed deposit is a pocket of kind fixed (000048). It is opened whole — the
// money, the rate and the term in one write — and from then on only two things
// happen to it: the days it earns, and the day it ends.
//
// It is the one cash account that can be recorded late. Nothing is written in
// it by hand, so there is nothing to count twice: a deposit opened two weeks
// ago and registered today is computed from the day it opened, and the days it
// already earned are credited at once. The main account and a flexible pocket
// stay the other way round — what the platform already paid is recorded as an
// interest movement, because the statement's figure is exact and a computation
// is an estimate.

// NewFixedDepositInput is a deposit as the owner states it: what went in, where,
// when, and at what rate.
//
// It carries a portfolio because a deposit is one lot of money. An account is
// shared by every portfolio that keeps money on the platform; a CDT is bought
// once, by one of them.
type NewFixedDepositInput struct {
	PortfolioID uuid.UUID
	SourceID    uuid.UUID
	Currency    money.Currency
	Name        string
	// Amount is what was deposited, in Currency.
	Amount decimal.Decimal
	// OpenedOn is the day the money went in, which is the first day it earns.
	OpenedOn time.Time
	// MaturesOn is the day it comes due, nil for a deposit at no term. The rate
	// runs through the day before, and on the day itself the money goes back to
	// the main account.
	MaturesOn *time.Time
	// CashRateInput is the rate it keeps: the version written with it, which
	// never gets another.
	CashRateInput
}

// CleanName is the name as it is stored, trimmed like any other pocket's.
func (in NewFixedDepositInput) CleanName() string {
	return strings.TrimSpace(in.Name)
}

// Validate checks what the deposit states. today is the server's clock.
//
// The account, the portfolio and the money are checked where they can be
// locked, in the repository.
func (in NewFixedDepositInput) Validate(today time.Time) error {
	if in.PortfolioID == (uuid.UUID{}) || in.SourceID == (uuid.UUID{}) {
		return invalidCashPocket("portfolioId and sourceId are required")
	}

	if !currency.IsSupported(in.Currency) {
		return invalidCashPocket("currency must be one of: %s", currency.List())
	}

	if err := validateCashPocketName(in.Name); err != nil {
		return err
	}

	if !in.Amount.IsPos() || !in.Amount.LessThan(maxCashRateBalance) {
		return invalidCashPocket("amount must be greater than 0 and less than %s", maxCashRateBalance)
	}

	day := cashRateDay(today)

	switch {
	case in.OpenedOn.IsZero():
		return invalidCashPocket("openedOn is required")
	case cashRateDay(in.OpenedOn).After(day):
		return invalidCashPocket("openedOn cannot be in the future")
	case cashRateDay(in.OpenedOn).Before(day.AddDate(-maxFixedDepositYears, 0, 0)):
		return invalidCashPocket("openedOn cannot be more than %d years ago", maxFixedDepositYears)
	}

	if in.MaturesOn != nil {
		matures := cashRateDay(*in.MaturesOn)

		// Its term has to have some of itself left. One that came due already is
		// a deposit that has been paid, and what that leaves is a deposit and an
		// interest in the main account, recorded as the movements they were.
		switch {
		case !matures.After(cashRateDay(in.OpenedOn)):
			return invalidCashPocket("maturesOn must be after openedOn")
		case matures.Before(day.AddDate(0, 0, -cashRateDateGrace)):
			return invalidCashPocket("maturesOn cannot be before %s: a deposit that already came due is recorded as the movements it paid", day.AddDate(0, 0, -cashRateDateGrace).Format(time.DateOnly))
		}
	}

	// Crediting at maturity needs a maturity to credit on.
	if in.Posting == PostingAtMaturity && in.MaturesOn == nil {
		return invalidCashPocket("posting at_maturity needs a maturesOn to credit on")
	}

	return in.validateValues(true)
}

// CloseFixedDepositInput cancels a deposit before its term, or ends one that
// has none.
//
// The rate stops the day before, what it has earned and not credited is paid,
// and the balance goes back to the main account of its portfolio. What the
// platform charges for breaking the term rides on that withdrawal as its fee,
// so it counts as a loss and not as money the owner took out.
type CloseFixedDepositInput struct {
	// ClosesOn is the day the money moves, and the first day the deposit no
	// longer earns.
	ClosesOn time.Time
	// Penalty is what the platform keeps for cancelling early, in the deposit's
	// currency. Zero when there is none.
	Penalty decimal.Decimal
}

// Validate checks what the cancellation states. today is the server's clock.
func (in CloseFixedDepositInput) Validate(today time.Time) error {
	if in.ClosesOn.IsZero() {
		return invalidCashPocket("closesOn is required")
	}

	// Past days are not recomputed, so a deposit cannot stop earning in one.
	earliest := cashRateDay(today).AddDate(0, 0, -cashRateDateGrace)
	if cashRateDay(in.ClosesOn).Before(earliest) {
		return invalidCashPocket("closesOn cannot be before %s: past days are not recomputed", earliest.Format(time.DateOnly))
	}

	// And it cannot stop in a day that has not happened: the money moves when
	// the cancellation is recorded, and the days up to then are computed with
	// it — days the deposit has not lived yet.
	if cashRateDay(in.ClosesOn).After(cashRateDay(today)) {
		return invalidCashPocket("closesOn cannot be in the future")
	}

	if in.Penalty.IsNeg() {
		return invalidCashPocket("penalty cannot be negative")
	}

	return nil
}
