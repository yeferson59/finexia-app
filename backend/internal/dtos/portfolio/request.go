package portfolio

import (
	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/money"
)

type CreatePortfolioRequestDTO struct {
	Name        string      `json:"name" validate:"required"`
	Description string      `json:"description"`
	Currency    string      `json:"currency" validate:"required"`
	Type        string      `json:"type" validate:"required"`
	RiskID      uuid.UUID   `json:"riskId"`
	PriceValue  money.Money `json:"priceValue"`
	IsDefault   bool        `json:"isDefault"`
}
