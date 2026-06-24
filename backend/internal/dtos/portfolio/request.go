package portfolio

import (
	"time"

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

type CreatePlatformRequestDTO struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Type        string `json:"type" validate:"required"`
}

type UpdatePlatformRequestDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	IsActive    bool   `json:"isActive"`
}

type UpdateAssetPriceRequestDTO struct {
	Price money.Money `json:"price" validate:"required"`
}

type CreatePortfolioEntryRequestDTO struct {
	PortfolioID     uuid.UUID     `json:"portfolioId" validate:"required"`
	AssetID         uuid.UUID     `json:"assetId" validate:"required"`
	SourceID        uuid.UUID     `json:"sourceId" validate:"required"`
	TransactionType string        `json:"transactionType"`
	Quantity        money.Decimal `json:"quantity" validate:"required"`
	Price           money.Money   `json:"price" validate:"required"`
	CostCurrency    string        `json:"costCurrency" validate:"required"`
	Category        string        `json:"category"`
	EntryDate       time.Time     `json:"entryDate" validate:"required"`
	Notes           string        `json:"notes"`
}

type CreateTransactionRequestDTO struct {
	Type            string        `json:"type" validate:"required"`
	Quantity        money.Decimal `json:"quantity" validate:"required"`
	Price           money.Money   `json:"price" validate:"required"`
	Currency        string        `json:"currency" validate:"required"`
	Fees            money.Money   `json:"fees"`
	TransactionDate time.Time     `json:"transactionDate" validate:"required"`
	Notes           string        `json:"notes"`
}
