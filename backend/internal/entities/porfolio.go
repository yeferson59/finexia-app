package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/money"
)

type AssetType string

const (
	Stock      AssetType = "stock"
	ETF        AssetType = "etf"
	Crypto     AssetType = "crypto"
	Bond       AssetType = "bond"
	Cash       AssetType = "cash"
	RealEstate AssetType = "real_estate"
	Commodity  AssetType = "commodity"
	Other      AssetType = "other"
)

func (a AssetType) Transform() PortfolioEntryCategory {
	switch a {
	case Stock:
		return Stocks
	case ETF:
		return ETFs
	case Crypto:
		return Cryptos
	case Bond:
		return Bonds
	case Cash:
		return Cashs
	case RealEstate:
		return RealEstates
	case Commodity:
		return Commodities
	default:
		return Others
	}
}

type SourceType string

const (
	Broker          SourceType = "broker"
	Bank            SourceType = "investment_bank"
	TradingPlatform SourceType = "trading_platform"
	Neobank         SourceType = "neobank"
	DeFi            SourceType = "de_fi"
	CryptoWallet    SourceType = "crypto_wallet"
	MutualFunds     SourceType = "mutual_funds"
	BrokerageHouse  SourceType = "brokerage_house"
	OtherType       SourceType = "other"
)

func (s SourceType) IsValid() bool {
	switch s {
	case Broker, Bank, TradingPlatform, Neobank, DeFi, CryptoWallet, MutualFunds, BrokerageHouse, OtherType:
		return true
	default:
		return false
	}
}

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

type PortfolioEntryCategory string

const (
	Stocks      PortfolioEntryCategory = "stocks"
	ETFs        PortfolioEntryCategory = "etfs"
	Cryptos     PortfolioEntryCategory = "cryptos"
	Bonds       PortfolioEntryCategory = "bonds"
	Cashs       PortfolioEntryCategory = "cash"
	RealEstates PortfolioEntryCategory = "real_estates"
	Commodities PortfolioEntryCategory = "commodities"
	Others      PortfolioEntryCategory = "others"
)

func (c PortfolioEntryCategory) IsValid() bool {
	switch c {
	case Stocks, ETFs, Cryptos, Bonds, Cashs, RealEstates, Commodities, Others:
		return true
	default:
		return false
	}
}

type PortfolioType string

const (
	PortfolioTypeOnlyStock       PortfolioType = "stock"
	PortfolioTypeOnlyETF         PortfolioType = "etf"
	PortfolioTypeOnlyCrypto      PortfolioType = "crypto"
	PortfolioTypeOnlyBond        PortfolioType = "bond"
	PortfolioTypeOnlyCash        PortfolioType = "cash"
	PortfolioTypeOnlyRealEstate  PortfolioType = "real_estate"
	PortfolioTypeOnlyCommodities PortfolioType = "commodities"

	PortfolioTypeStockAndETF         PortfolioType = "stock_etf"
	PortfolioTypeStockAndCrypto      PortfolioType = "stock_crypto"
	PortfolioTypeStockAndBond        PortfolioType = "stock_bond"
	PortfolioTypeStockAndCash        PortfolioType = "stock_cash"
	PortfolioTypeStockAndRealEstate  PortfolioType = "stock_real_estate"
	PortfolioTypeStockAndCommodities PortfolioType = "stock_commodities"

	PortfolioTypeETFAndCrypto      PortfolioType = "etf_crypto"
	PortfolioTypeETFAndBond        PortfolioType = "etf_bond"
	PortfolioTypeETFAndCash        PortfolioType = "etf_cash"
	PortfolioTypeETFAndRealEstate  PortfolioType = "etf_real_estate"
	PortfolioTypeETFAndCommodities PortfolioType = "etf_commodities"

	PortfolioTypeCryptoAndBond        PortfolioType = "crypto_bond"
	PortfolioTypeCryptoAndCash        PortfolioType = "crypto_cash"
	PortfolioTypeCryptoAndRealEstate  PortfolioType = "crypto_real_estate"
	PortfolioTypeCryptoAndCommodities PortfolioType = "crypto_commodities"

	PortfolioTypeBondAndCash        PortfolioType = "bond_cash"
	PortfolioTypeBondAndRealEstate  PortfolioType = "bond_real_estate"
	PortfolioTypeBondAndCommodities PortfolioType = "bond_commodities"

	PortfolioTypeCashAndRealEstate  PortfolioType = "cash_real_estate"
	PortfolioTypeCashAndCommodities PortfolioType = "cash_commodities"

	PortfolioTypeRealEstateAndCommodities PortfolioType = "real_estate_commodities"

	PortfolioTypeMultiple PortfolioType = "multiple"
)

type InvestmentSource struct {
	ID               uuid.UUID        `json:"id"`
	UserID           uuid.UUID        `json:"userId"`
	Name             string           `json:"name"`
	SourceType       SourceType       `json:"sourceType"`
	Description      string           `json:"description"`
	IsActive         bool             `json:"isActive"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
	User             User             `json:"user,omitzero"`
	PortfolioEntries []PortfolioEntry `json:"portfolioEntries,omitempty"`
}

type Risk struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
	Portfolios  []Portfolio `json:"portfolios,omitempty"`
}

type Portfolio struct {
	ID                 uuid.UUID           `json:"id"`
	UserID             uuid.UUID           `json:"userId"`
	Name               string              `json:"name"`
	Description        string              `json:"description"`
	Type               PortfolioType       `json:"type"`
	RiskID             uuid.UUID           `json:"riskId"`
	BaseCurrency       string              `json:"baseCurrency"`
	IsDefault          bool                `json:"isDefault"`
	PriceValue         *money.Money        `json:"priceValue"`
	CreatedAt          time.Time           `json:"createdAt"`
	UpdatedAt          time.Time           `json:"updatedAt"`
	Risk               Risk                `json:"risk,omitzero"`
	User               User                `json:"user,omitzero"`
	PortfolioEntries   []PortfolioEntry    `json:"portfolioEntries,omitempty"`
	PortfolioSnapshots []PortfolioSnapshot `json:"portfolioSnapshots,omitempty"`
}

type Asset struct {
	ID               uuid.UUID        `json:"id"`
	Ticker           string           `json:"ticker"`
	Name             string           `json:"name"`
	AssetType        AssetType        `json:"assetType"`
	Exchange         string           `json:"exchange"`
	Currency         string           `json:"currency"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
	PortfolioEntries []PortfolioEntry `json:"portfolioEntries,omitempty"`
}

type PortfolioEntry struct {
	ID           uuid.UUID              `json:"id"`
	PortfolioID  uuid.UUID              `json:"portfolioId"`
	AssetID      uuid.UUID              `json:"assetId"`
	SourceID     uuid.UUID              `json:"sourceId"`
	Quantity     money.Decimal          `json:"quantity"`
	Price        money.Money            `json:"price"`
	CostCurrency string                 `json:"costCurrency"`
	Category     PortfolioEntryCategory `json:"category"`
	EntryDate    time.Time              `json:"entryDate"`
	Notes        string                 `json:"notes"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
	Portfolio    Portfolio              `json:"portfolio,omitzero"`
	Asset        Asset                  `json:"asset,omitzero"`
	Source       InvestmentSource       `json:"source,omitzero"`
	Transactions []Transaction          `json:"transactions,omitempty"`
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
	Entry           PortfolioEntry  `json:"entry,omitzero"`
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
	Portfolio        Portfolio   `json:"portfolio,omitzero"`
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
