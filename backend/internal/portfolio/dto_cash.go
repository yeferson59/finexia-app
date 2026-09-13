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
	PortfolioID uuid.UUID       `json:"portfolioId"`
	SourceID    uuid.UUID       `json:"sourceId"`
	Currency    money.Currency  `json:"currency"`
	Kind        string          `json:"kind"`
	Amount      decimal.Decimal `json:"amount"`
	Fees        decimal.Decimal `json:"fees"`
	Date        time.Time       `json:"date"`
	Notes       string          `json:"notes"`
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
