package portfolio

import (
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// The cash DTOs carry no validate tags. Every rule is checked by
// CashMovementInput, which answers with one message per broken rule; a tag
// would refuse the same request first, with the validator's wording instead.

type CreateCashMovementRequestDTO struct {
	PortfolioID uuid.UUID      `json:"portfolioId"`
	SourceID    uuid.UUID      `json:"sourceId"`
	Currency    money.Currency `json:"currency"`
	// PocketID is which drawer of the account the movement goes in. Omitted or
	// null is the main account, which is where every movement went before
	// pockets existed (000047).
	PocketID uuid.UUID       `json:"pocketId"`
	Kind     string          `json:"kind"`
	Amount   decimal.Decimal `json:"amount"`
	Fees     decimal.Decimal `json:"fees"`
	Date     time.Time       `json:"date"`
	Notes    string          `json:"notes"`
}

func (d CreateCashMovementRequestDTO) Input() CashMovementInput {
	return CashMovementInput{
		Kind:     CashMovementKind(d.Kind),
		Amount:   d.Amount,
		Fees:     d.Fees,
		Currency: d.Currency,
		Date:     d.Date,
		Notes:    d.Notes,
	}
}

// UpdateCashMovementRequestDTO has no portfolio, platform or currency: an edit
// stays on the balance the movement is already on. Moving it is a deletion and
// a new movement.
type UpdateCashMovementRequestDTO struct {
	Kind   string          `json:"kind"`
	Amount decimal.Decimal `json:"amount"`
	Fees   decimal.Decimal `json:"fees"`
	Date   time.Time       `json:"date"`
	Notes  string          `json:"notes"`
}

func (d UpdateCashMovementRequestDTO) Input() CashMovementInput {
	return CashMovementInput{
		Kind:   CashMovementKind(d.Kind),
		Amount: d.Amount,
		Fees:   d.Fees,
		Date:   d.Date,
		Notes:  d.Notes,
	}
}

type PaginatedCashMovementsDTO struct {
	Data       []CashMovement `json:"data"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"totalPages"`
}

// MoveCashRequestDTO moves money between two cash balances inside one
// portfolio. sourceId and currency are the account it leaves, and fromPocketId
// and toPocketId the drawers on each side; either omitted or null is the main
// account.
//
// toSourceId and toCurrency are the account it arrives at, and both default to
// the one it left — the move between drawers, which is what this was before.
// Naming another platform transfers the money to it, which is what happens
// before a purchase on a broker funded from a savings app.
//
// amount is what leaves and toAmount what arrives; across two currencies both
// are required, and within one, toAmount can be left out because it is the
// same money. The two amounts are stated rather than a rate: they are the
// figures on the statement, and they record exactly what arrived.
type MoveCashRequestDTO struct {
	PortfolioID  uuid.UUID       `json:"portfolioId"`
	SourceID     uuid.UUID       `json:"sourceId"`
	Currency     money.Currency  `json:"currency"`
	FromPocketID uuid.UUID       `json:"fromPocketId"`
	ToSourceID   uuid.UUID       `json:"toSourceId"`
	ToCurrency   money.Currency  `json:"toCurrency"`
	ToPocketID   uuid.UUID       `json:"toPocketId"`
	Amount       decimal.Decimal `json:"amount"`
	ToAmount     decimal.Decimal `json:"toAmount"`
	Date         time.Time       `json:"date"`
	Notes        string          `json:"notes"`
}

func (d MoveCashRequestDTO) Input() CashMoveInput {
	return CashMoveInput{
		Currency:   d.Currency,
		From:       d.FromPocketID,
		To:         d.ToPocketID,
		ToSource:   d.ToSourceID,
		ToCurrency: d.ToCurrency,
		Amount:     d.Amount,
		ToAmount:   d.ToAmount,
		Date:       d.Date,
		Notes:      d.Notes,
	}
}

// CreateCashPocketRequestDTO opens a flexible pocket on an account.
type CreateCashPocketRequestDTO struct {
	SourceID uuid.UUID      `json:"sourceId"`
	Currency money.Currency `json:"currency"`
	Name     string         `json:"name"`
}

func (d CreateCashPocketRequestDTO) Input() NewCashPocketInput {
	return NewCashPocketInput(d)
}

// RenameCashPocketRequestDTO is the one thing a pocket can be told to change.
type RenameCashPocketRequestDTO struct {
	Name string `json:"name"`
}

func (d RenameCashPocketRequestDTO) Input() RenameCashPocketInput {
	return RenameCashPocketInput(d)
}

// OpenFixedDepositRequestDTO opens a deposit that keeps the rate of the day it
// was opened (000048). It states the deposit and its rate in one body, because
// they are one thing: the rate is not given to the pocket afterwards, it is
// what the pocket is.
//
// openedOn may be in the past — "I opened it two weeks ago" — and maturesOn may
// be omitted for a deposit at no term.
type OpenFixedDepositRequestDTO struct {
	PortfolioID uuid.UUID       `json:"portfolioId"`
	SourceID    uuid.UUID       `json:"sourceId"`
	Currency    money.Currency  `json:"currency"`
	Name        string          `json:"name"`
	Amount      decimal.Decimal `json:"amount"`
	OpenedOn    time.Time       `json:"openedOn"`
	MaturesOn   *time.Time      `json:"maturesOn"`

	AnnualRatePct  decimal.Decimal   `json:"annualRatePct"`
	WithholdingPct decimal.Decimal   `json:"withholdingPct"`
	Posting        string            `json:"posting"`
	Tiers          []CashRateTierDTO `json:"tiers"`
}

func (d OpenFixedDepositRequestDTO) Input() NewFixedDepositInput {
	return NewFixedDepositInput{
		PortfolioID:   d.PortfolioID,
		SourceID:      d.SourceID,
		Currency:      d.Currency,
		Name:          d.Name,
		Amount:        d.Amount,
		OpenedOn:      d.OpenedOn,
		MaturesOn:     d.MaturesOn,
		CashRateInput: cashRateValues(d.AnnualRatePct, d.WithholdingPct, d.Posting, d.Tiers, nil),
	}
}

// CloseFixedDepositRequestDTO cancels a deposit before its term: the day the
// money moves, and what the platform keeps for breaking it.
type CloseFixedDepositRequestDTO struct {
	ClosesOn time.Time       `json:"closesOn"`
	Penalty  decimal.Decimal `json:"penalty"`
}

func (d CloseFixedDepositRequestDTO) Input() CloseFixedDepositInput {
	return CloseFixedDepositInput(d)
}
