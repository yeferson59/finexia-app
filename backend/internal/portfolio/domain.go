// Package portfolio is the investments domain module: portfolios, entries,
// transactions, investment platforms, snapshots and bulk import/export.
// Assets (and exchange rates) live here rather than in the market module: the
// market module consumes these types by importing portfolio (see TECH_DEBT #12).
package portfolio

import (
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/identity"
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

func (a AssetType) IsValid() bool {
	switch a {
	case Stock, ETF, Crypto, Bond, Cash, RealEstate, Commodity, Other:
		return true
	default:
		return false
	}
}

func (a AssetType) Transform() EntryCategory {
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

// EntryCategory keeps the legacy PortfolioEntryCategory values; the name drops
// the redundant prefix now that the type lives inside the portfolio package.
type EntryCategory string

const (
	Stocks      EntryCategory = "stocks"
	ETFs        EntryCategory = "etfs"
	Cryptos     EntryCategory = "cryptos"
	Bonds       EntryCategory = "bonds"
	Cashs       EntryCategory = "cash"
	RealEstates EntryCategory = "real_estates"
	Commodities EntryCategory = "commodities"
	Others      EntryCategory = "others"
)

func (t TransactionType) IsValid() bool {
	switch t {
	case Buy, Sell, Dividend, Split, TransferIn, TransferOut, Fee, Interest:
		return true
	default:
		return false
	}
}

func (c EntryCategory) IsValid() bool {
	switch c {
	case Stocks, ETFs, Cryptos, Bonds, Cashs, RealEstates, Commodities, Others:
		return true
	default:
		return false
	}
}

type Type string

// Type values must stay in sync with the `portfolio_type` enum defined in
// 000003_create_portfolio_tables.up.sql.
const (
	TypeStocks      Type = "stocks"
	TypeETFs        Type = "etfs"
	TypeCryptos     Type = "cryptos"
	TypeBonds       Type = "bonds"
	TypeCash        Type = "cash"
	TypeForex       Type = "forex"
	TypeRealEstates Type = "real_estates"
	TypeCommodities Type = "commodities"

	TypeForexStocks      Type = "forex_stocks"
	TypeForexETFs        Type = "forex_etfs"
	TypeForexCryptos     Type = "forex_cryptos"
	TypeForexBonds       Type = "forex_bonds"
	TypeForexCash        Type = "forex_cash"
	TypeForexRealStates  Type = "forex_real_states"
	TypeForexCommodities Type = "forex_commodities"

	TypeStocksETFs        Type = "stocks_etfs"
	TypeStocksCryptos     Type = "stocks_cryptos"
	TypeStocksBonds       Type = "stocks_bonds"
	TypeStocksCash        Type = "stocks_cash"
	TypeStocksRealEstates Type = "stocks_real_estates"
	TypeStocksCommodities Type = "stocks_commodities"

	TypeETFsCryptos     Type = "etfs_cryptos"
	TypeETFsBonds       Type = "etfs_bonds"
	TypeETFsCash        Type = "etfs_cash"
	TypeETFsRealEstates Type = "etfs_real_estates"
	TypeETFsCommodities Type = "etfs_commodities"

	TypeCryptosBonds       Type = "cryptos_bonds"
	TypeCryptosCash        Type = "cryptos_cash"
	TypeCryptosRealEstates Type = "cryptos_real_estates"
	TypeCryptosCommodities Type = "cryptos_commodities"

	TypeBondsCash        Type = "bonds_cash"
	TypeBondsRealEstates Type = "bonds_real_estates"
	TypeBondsCommodities Type = "bonds_commodities"

	TypeCashRealEstates Type = "cash_real_estates"
	TypeCashCommodities Type = "cash_commodities"

	TypeRealEstatesCommodities Type = "real_estates_commodities"

	TypeDiversified Type = "diversified"
)

func (t Type) IsValid() bool {
	switch t {
	case TypeStocks, TypeETFs, TypeCryptos, TypeBonds,
		TypeCash, TypeForex, TypeRealEstates, TypeCommodities,
		TypeForexStocks, TypeForexETFs, TypeForexCryptos, TypeForexBonds,
		TypeForexCash, TypeForexRealStates, TypeForexCommodities,
		TypeStocksETFs, TypeStocksCryptos, TypeStocksBonds, TypeStocksCash,
		TypeStocksRealEstates, TypeStocksCommodities,
		TypeETFsCryptos, TypeETFsBonds, TypeETFsCash, TypeETFsRealEstates,
		TypeETFsCommodities,
		TypeCryptosBonds, TypeCryptosCash, TypeCryptosRealEstates, TypeCryptosCommodities,
		TypeBondsCash, TypeBondsRealEstates, TypeBondsCommodities,
		TypeCashRealEstates, TypeCashCommodities,
		TypeRealEstatesCommodities, TypeDiversified:
		return true
	default:
		return false
	}
}

type InvestmentSource struct {
	ID          uuid.UUID     `json:"id"`
	UserID      uuid.UUID     `json:"userId"`
	Name        string        `json:"name"`
	SourceType  SourceType    `json:"sourceType"`
	Description string        `json:"description"`
	IsActive    bool          `json:"isActive"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
	User        identity.User `json:"user,omitzero"`
	Entries     []Entry       `json:"portfolioEntries,omitempty"`
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
	ID           uuid.UUID     `json:"id"`
	UserID       uuid.UUID     `json:"userId"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Type         Type          `json:"type"`
	RiskID       uuid.UUID     `json:"riskId"`
	BaseCurrency string        `json:"baseCurrency"`
	IsDefault    bool          `json:"isDefault"`
	PriceValue   *money.Money  `json:"priceValue"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
	Risk         Risk          `json:"risk,omitzero"`
	User         identity.User `json:"user,omitzero"`
	Entries      []Entry       `json:"portfolioEntries,omitempty"`
	Snapshots    []Snapshot    `json:"portfolioSnapshots,omitempty"`
}

type Asset struct {
	ID             uuid.UUID    `json:"id"`
	Ticker         string       `json:"ticker"`
	Name           string       `json:"name"`
	AssetType      AssetType    `json:"assetType"`
	Exchange       string       `json:"exchange"`
	Currency       string       `json:"currency"`
	CurrentPrice   *money.Money `json:"currentPrice"`
	PriceUpdatedAt *time.Time   `json:"priceUpdatedAt"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
	Entries        []Entry      `json:"portfolioEntries,omitempty"`
}

type Entry struct {
	ID           uuid.UUID        `json:"id"`
	PortfolioID  uuid.UUID        `json:"portfolioId"`
	AssetID      uuid.UUID        `json:"assetId"`
	SourceID     uuid.UUID        `json:"sourceId"`
	Quantity     money.Decimal    `json:"quantity"`
	Price        money.Money      `json:"price"`
	CostCurrency string           `json:"costCurrency"`
	Category     EntryCategory    `json:"category"`
	EntryDate    time.Time        `json:"entryDate"`
	Notes        string           `json:"notes"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
	Portfolio    Portfolio        `json:"portfolio,omitzero"`
	Asset        Asset            `json:"asset,omitzero"`
	Source       InvestmentSource `json:"source,omitzero"`
	Transactions []Transaction    `json:"transactions,omitempty"`
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
	Entry           Entry           `json:"entry,omitzero"`
}

// ImportTransactionRow is one validated spreadsheet row, ready to be
// persisted as an asset + portfolio entry + transaction.
type ImportTransactionRow struct {
	RowNumber int
	Ticker    string
	AssetName string
	AssetType AssetType
	Category  EntryCategory
	Type      TransactionType
	Quantity  money.Decimal
	Price     money.Money
	Fees      money.Money
	Currency  string
	Date      time.Time
	Notes     string
}

type Snapshot struct {
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

// AllocationItem is the result of grouping portfolio_entries by category.
type AllocationItem struct {
	Category    EntryCategory `json:"category"`
	MarketValue string        `json:"marketValue"`
}

// PlatformStats is the result of joining investment_sources with portfolio_entries stats.
type PlatformStats struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	SourceType  SourceType `json:"sourceType"`
	IsActive    bool       `json:"isActive"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Investments int64      `json:"investments"`
	TotalValue  string     `json:"totalValue"`
}

// SummaryView is the result of joining portfolios + risks + portfolio_summary view.
type SummaryView struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Type             Type      `json:"type"`
	BaseCurrency     string    `json:"baseCurrency"`
	IsDefault        bool      `json:"isDefault"`
	RiskID           uuid.UUID `json:"riskId"`
	RiskName         string    `json:"riskName"`
	TotalPositions   int64     `json:"totalPositions"`
	TotalCostBase    string    `json:"totalCostBase"`
	TotalMarketValue string    `json:"totalMarketValue"`
	TotalGainLoss    string    `json:"totalGainLoss"`
	TotalGainLossPct string    `json:"totalGainLossPct"`
	CreatedAt        time.Time `json:"createdAt"`
	// DisplayCurrency is the currency the totals above are expressed in. It
	// equals BaseCurrency unless the caller requested conversion to another
	// currency (see Service.GetPortfoliosSummaryInCurrency).
	DisplayCurrency string `json:"displayCurrency"`
}

type SnapshotRow struct {
	PortfolioID      uuid.UUID
	BaseCurrency     string
	TotalMarketValue string
	TotalCostBase    string
	TotalGainLoss    string
	TotalGainLossPct string
}

type GrowthPoint struct {
	Date          time.Time
	TotalValue    string
	TotalCostBase string
	GainLoss      string
	GainLossPct   string
}

type GrowthSummary struct {
	FirstDate      time.Time
	InitialValue   string
	CurrentValue   string
	TotalGrowthPct string
}
