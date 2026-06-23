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

func (t TransactionType) IsValid() bool {
	switch t {
	case Buy, Sell, Dividend, Split, TransferIn, TransferOut, Fee, Interest:
		return true
	default:
		return false
	}
}

func (c PortfolioEntryCategory) IsValid() bool {
	switch c {
	case Stocks, ETFs, Cryptos, Bonds, Cashs, RealEstates, Commodities, Others:
		return true
	default:
		return false
	}
}

type PortfolioType string

// PortfolioType values must stay in sync with the `portfolio_type` enum
// defined in 000003_create_portfolio_tables.up.sql.
const (
	PortfolioTypeStocks      PortfolioType = "stocks"
	PortfolioTypeETFs        PortfolioType = "etfs"
	PortfolioTypeCryptos     PortfolioType = "cryptos"
	PortfolioTypeBonds       PortfolioType = "bonds"
	PortfolioTypeCash        PortfolioType = "cash"
	PortfolioTypeForex       PortfolioType = "forex"
	PortfolioTypeRealEstates PortfolioType = "real_estates"
	PortfolioTypeCommodities PortfolioType = "commodities"

	PortfolioTypeForexStocks      PortfolioType = "forex_stocks"
	PortfolioTypeForexETFs        PortfolioType = "forex_etfs"
	PortfolioTypeForexCryptos     PortfolioType = "forex_cryptos"
	PortfolioTypeForexBonds       PortfolioType = "forex_bonds"
	PortfolioTypeForexCash        PortfolioType = "forex_cash"
	PortfolioTypeForexRealStates  PortfolioType = "forex_real_states"
	PortfolioTypeForexCommodities PortfolioType = "forex_commodities"

	PortfolioTypeStocksETFs        PortfolioType = "stocks_etfs"
	PortfolioTypeStocksCryptos     PortfolioType = "stocks_cryptos"
	PortfolioTypeStocksBonds       PortfolioType = "stocks_bonds"
	PortfolioTypeStocksCash        PortfolioType = "stocks_cash"
	PortfolioTypeStocksRealEstates PortfolioType = "stocks_real_estates"
	PortfolioTypeStocksCommodities PortfolioType = "stocks_commodities"

	PortfolioTypeETFsCryptos     PortfolioType = "etfs_cryptos"
	PortfolioTypeETFsBonds       PortfolioType = "etfs_bonds"
	PortfolioTypeETFsCash        PortfolioType = "etfs_cash"
	PortfolioTypeETFsRealEstates PortfolioType = "etfs_real_estates"
	PortfolioTypeETFsCommodities PortfolioType = "etfs_commodities"

	PortfolioTypeCryptosBonds       PortfolioType = "cryptos_bonds"
	PortfolioTypeCryptosCash        PortfolioType = "cryptos_cash"
	PortfolioTypeCryptosRealEstates PortfolioType = "cryptos_real_estates"
	PortfolioTypeCryptosCommodities PortfolioType = "cryptos_commodities"

	PortfolioTypeBondsCash        PortfolioType = "bonds_cash"
	PortfolioTypeBondsRealEstates PortfolioType = "bonds_real_estates"
	PortfolioTypeBondsCommodities PortfolioType = "bonds_commodities"

	PortfolioTypeCashRealEstates PortfolioType = "cash_real_estates"
	PortfolioTypeCashCommodities PortfolioType = "cash_commodities"

	PortfolioTypeRealEstatesCommodities PortfolioType = "real_estates_commodities"

	PortfolioTypeDiversified PortfolioType = "diversified"
)

func (t PortfolioType) IsValid() bool {
	switch t {
	case PortfolioTypeStocks, PortfolioTypeETFs, PortfolioTypeCryptos, PortfolioTypeBonds,
		PortfolioTypeCash, PortfolioTypeForex, PortfolioTypeRealEstates, PortfolioTypeCommodities,
		PortfolioTypeForexStocks, PortfolioTypeForexETFs, PortfolioTypeForexCryptos, PortfolioTypeForexBonds,
		PortfolioTypeForexCash, PortfolioTypeForexRealStates, PortfolioTypeForexCommodities,
		PortfolioTypeStocksETFs, PortfolioTypeStocksCryptos, PortfolioTypeStocksBonds, PortfolioTypeStocksCash,
		PortfolioTypeStocksRealEstates, PortfolioTypeStocksCommodities,
		PortfolioTypeETFsCryptos, PortfolioTypeETFsBonds, PortfolioTypeETFsCash, PortfolioTypeETFsRealEstates,
		PortfolioTypeETFsCommodities,
		PortfolioTypeCryptosBonds, PortfolioTypeCryptosCash, PortfolioTypeCryptosRealEstates, PortfolioTypeCryptosCommodities,
		PortfolioTypeBondsCash, PortfolioTypeBondsRealEstates, PortfolioTypeBondsCommodities,
		PortfolioTypeCashRealEstates, PortfolioTypeCashCommodities,
		PortfolioTypeRealEstatesCommodities, PortfolioTypeDiversified:
		return true
	default:
		return false
	}
}

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
	CurrentPrice     *money.Money     `json:"currentPrice"`
	PriceUpdatedAt   *time.Time       `json:"priceUpdatedAt"`
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

// PortfolioSummaryView is the result of joining portfolios + risks + portfolio_summary view.
type PortfolioSummaryView struct {
	ID               uuid.UUID     `json:"id"`
	Name             string        `json:"name"`
	Description      string        `json:"description"`
	Type             PortfolioType `json:"type"`
	BaseCurrency     string        `json:"baseCurrency"`
	IsDefault        bool          `json:"isDefault"`
	RiskID           uuid.UUID     `json:"riskId"`
	RiskName         string        `json:"riskName"`
	TotalPositions   int64         `json:"totalPositions"`
	TotalCostBase    string        `json:"totalCostBase"`
	TotalMarketValue string        `json:"totalMarketValue"`
	TotalGainLoss    string        `json:"totalGainLoss"`
	TotalGainLossPct string        `json:"totalGainLossPct"`
	CreatedAt        time.Time     `json:"createdAt"`
}

type PorfolioSummary struct {
	PortfolioID      uuid.UUID     `json:"portfolioId"`
	UserID           uuid.UUID     `json:"userId"`
	PorfolioName     string        `json:"porfolioName"`
	BaseCurrency     string        `json:"baseCurrency"`
	TotalPosicions   money.Decimal `json:"totalPosicions"`
	TotalCostBase    money.Money   `json:"totalCostBase"`
	TotalMarketValue money.Money   `json:"totalMarketValue"`
	TotalGainLoss    money.Money   `json:"totalGainLoss"`
	TotalGainLossPct money.Decimal `json:"totalGainLossPct"`
	CreatedAt        time.Time     `json:"createdAt"`
}
