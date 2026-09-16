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

// MoveCashRequestDTO moves money between two balances of one account inside one
// portfolio. fromPocketId and toPocketId are the drawers it leaves and arrives
// at; either omitted or null is the main account.
type MoveCashRequestDTO struct {
	PortfolioID  uuid.UUID       `json:"portfolioId"`
	SourceID     uuid.UUID       `json:"sourceId"`
	Currency     money.Currency  `json:"currency"`
	FromPocketID uuid.UUID       `json:"fromPocketId"`
	ToPocketID   uuid.UUID       `json:"toPocketId"`
	Amount       decimal.Decimal `json:"amount"`
	Date         time.Time       `json:"date"`
	Notes        string          `json:"notes"`
}

func (d MoveCashRequestDTO) Input() CashMoveInput {
	return CashMoveInput{
		Currency: d.Currency,
		From:     d.FromPocketID,
		To:       d.ToPocketID,
		Amount:   d.Amount,
		Date:     d.Date,
		Notes:    d.Notes,
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
