package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/money"
)

type AssetType string

const (
	Stock  AssetType = "stock"
	ETF    AssetType = "etf"
	Crypto AssetType = "crypto"
	Bond   AssetType = "bond"
	Cash   AssetType = "cash"
	Other  AssetType = "other"
)

type SourceType string

const (
	Broker         SourceType = "broker"
	Bank           SourceType = "bank"
	CryptoExchange SourceType = "crypto_exchange"
	Platform       SourceType = "platform"
	Excel          SourceType = "excel"
	Manual         SourceType = "manual"
)

type TransactionType string

const (
	Buy         TransactionType = "buy"
	Sell        TransactionType = "sell"
	Dividend    TransactionType = "dividend"
	Split       TransactionType = "split"
	TransferIn  TransactionType = "transfer_in"
	TransferOut TransactionType = "transfer_out"
	Fee         TransactionType = "fee"
	Interest    TransactionType = "interest"
)

type PortfolioCategory string

const (
	Stocks      PortfolioCategory = "stocks"
	ETFs        PortfolioCategory = "etf"
	Cryptos     PortfolioCategory = "crypto"
	Bonds       PortfolioCategory = "bonds"
	Cashs       PortfolioCategory = "cash"
	RealEstate  PortfolioCategory = "real_estate"
	Commodities PortfolioCategory = "commodities"
	Others      PortfolioCategory = "other"
)

type InvestmentSource struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"userId"`
	Name        string     `json:"name"`
	SourceType  SourceType `json:"sourceType"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type Portfolio struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"userId"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	BaseCurrency string    `json:"baseCurrency"`
	IsDefault    bool      `json:"isDefault"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Asset struct {
	ID        uuid.UUID `json:"id"`
	Ticker    string    `json:"ticker"`
	Name      string    `json:"name"`
	AssetType AssetType `json:"assetType"`
	Exchange  string    `json:"exchange"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PortfolioEntry struct {
	ID           uuid.UUID         `json:"id"`
	PortfolioID  uuid.UUID         `json:"portfolioId"`
	AssetID      uuid.UUID         `json:"assetId"`
	SourceID     uuid.UUID         `json:"sourceId"`
	Quantity     money.Decimal     `json:"quantity"`
	AvgCostPrice money.Money       `json:"avgCostPrice"`
	CostCurrency string            `json:"costCurrency"`
	Category     PortfolioCategory `json:"category"`
	EntryDate    time.Time         `json:"entryDate"`
	Notes        string            `json:"notes"`
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
}

type Transaction struct {
	ID              uuid.UUID       `json:"id"`
	EntryID         uuid.UUID       `json:"entryId"`
	Type            TransactionType `json:"type"`
	Quantity        money.Decimal   `json:"quantity"`
	Price           money.Money     `json:"price"`
	Currency        string          `json:"currency"`
	Fees            money.Money     `json:"fees"`
	TransactionDate time.Time       `json:"transactionDate"`
	Notes           string          `json:"notes"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

type ExchangeRate struct {
	ID           uuid.UUID     `json:"id"`
	FromCurrency string        `json:"fromCurrency"`
	ToCurrency   string        `json:"toCurrency"`
	Rate         money.Decimal `json:"rate"`
	RateDate     time.Time     `json:"rateDate"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
}

type PortfolioSnapshot struct {
	ID               uuid.UUID   `json:"id"`
	PortfolioID      uuid.UUID   `json:"portfolioId"`
	SnapshotDate     time.Time   `json:"snapshotDate"`
	TotalValue       money.Money `json:"totalValue"`
	Currency         string      `json:"currency"`
	Allocation       []byte      `json:"allocation"`
	TotalGainLoss    money.Money `json:"totalGainLoss"`
	TotalGainLossPCT money.Money `json:"totalGainLossPCT"`
	CreatedAt        time.Time   `json:"createdAt"`
}

type PorfolioSummary struct {
	PortfolioID    uuid.UUID     `json:"portfolioId"`
	UserID         uuid.UUID     `json:"userId"`
	PorfolioName   string        `json:"porfolioName"`
	BaseCurrency   string        `json:"baseCurrency"`
	TotalPosicions money.Decimal `json:"totalPosicions"`
	TotalCostBase  money.Money   `json:"totalCostBase"`
	CreatedAt      time.Time     `json:"createdAt"`
}
