package mcp

// The types below are the tools' wire shapes, and they are deliberately not the
// domain types.
//
// Two reasons, and the first is mechanical: the SDK infers each tool's JSON
// schema by reflecting over these structs, and the domain types are built from
// money.Money, decimal.Decimal and uuid.UUID — types whose Go shape (an opaque
// struct, a [16]byte) says nothing about the string they marshal to. Reflected
// straight, they would advertise a schema no client could satisfy or trust.
//
// The second is editorial. A model reads these fields with no other context, so
// every amount is a decimal string next to the currency it is denominated in,
// every id is a plain uuid string, and the nested rows the REST DTOs carry for
// the frontend are flattened away. What is left is what a question about a
// portfolio actually needs.

// CurrencyInput is the argument shared by the tools that report totals: one
// optional ISO 4217 code.
type CurrencyInput struct {
	Currency string `json:"currency,omitempty" jsonschema:"ISO 4217 code to report every amount in (USD, COP, EUR, GBP, CHF, JPY, CAD, AUD, CNY, MXN, BRL). Omit to use the account's preferred currency."`
}

// EmptyInput is the argument of the tools that take none. The MCP spec requires
// an object schema, so "no arguments" has to be spelled as an empty one.
type EmptyInput struct{}

// TransactionsInput asks for the most recent activity.
type TransactionsInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"how many of the most recent transactions to return, 1-200. Defaults to 20."`
}

// GrowthInput asks for the account-wide value series.
type GrowthInput struct {
	Currency string `json:"currency,omitempty" jsonschema:"ISO 4217 code to report the series in. Omit to use the account's preferred currency."`
	Period   string `json:"period,omitempty" jsonschema:"how far back to start: 1M, 3M, 6M or 1Y. Omit for the whole history."`
}

// AssetSearchInput searches the asset catalog.
type AssetSearchInput struct {
	Query string `json:"query,omitempty" jsonschema:"free text matched against ticker and name. Omit to list the catalog."`
	Limit int    `json:"limit,omitempty" jsonschema:"how many assets to return, 1-100. Defaults to 25."`
}

// PortfoliosOutput is the answer of list_portfolios.
type PortfoliosOutput struct {
	Portfolios []PortfolioSummary `json:"portfolios"`
}

// PortfolioSummary is one portfolio and what it is worth.
type PortfolioSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type" jsonschema:"what the portfolio holds: stocks, etfs, cryptos, bonds, cash, forex, real_estates, commodities, or a combination"`
	Risk        string `json:"risk,omitempty" jsonschema:"the risk profile the owner assigned to this portfolio"`
	// BaseCurrency is the portfolio's own; DisplayCurrency is what the amounts
	// below are actually in. They differ whenever the caller asked for a
	// currency, or the account prefers one.
	BaseCurrency    string `json:"baseCurrency"`
	DisplayCurrency string `json:"displayCurrency" jsonschema:"the currency every amount on this row is expressed in"`
	IsDefault       bool   `json:"isDefault"`
	Positions       int64  `json:"positions" jsonschema:"how many open positions the portfolio holds"`
	CostBasis       string `json:"costBasis" jsonschema:"what the owner paid for what they still hold, at weighted-average cost, excluding commissions"`
	MarketValue     string `json:"marketValue" jsonschema:"what those same positions are worth now"`
	GainLoss        string `json:"gainLoss" jsonschema:"marketValue minus costBasis"`
	GainLossPct     string `json:"gainLossPct" jsonschema:"gainLoss as a percentage of costBasis"`
	// The three counts below partition Positions and are what stops a flat
	// portfolio being confused with an unpriced one: a position with no
	// provider price is carried at its own cost, so it contributes exactly zero
	// to gainLoss no matter what it is really worth.
	PricedWithOwnKey  int64 `json:"pricedWithOwnKey" jsonschema:"positions valued with a price fetched using the user's own market-data key"`
	PricedManually    int64 `json:"pricedManually" jsonschema:"positions valued at an operator-entered reference price, which carries no freshness guarantee"`
	ValuedAtCost      int64 `json:"valuedAtCost" jsonschema:"positions with no price at all, carried at what they cost; their contribution to gainLoss is zero by construction"`
	PositionsUnvalued int64 `json:"positionsUnconverted" jsonschema:"positions added at face value because no exchange rate reached displayCurrency"`
	ConvertedFromBase bool  `json:"fxConverted" jsonschema:"true when the amounts were converted out of baseCurrency"`
}

// HoldingsOutput is the answer of get_holdings.
type HoldingsOutput struct {
	Holdings []Holding `json:"holdings"`
}

// Holding is one asset totalled across every portfolio the user owns — the
// answer to "how much of X do I have?" without asking where it sits.
type Holding struct {
	AssetID   string `json:"assetId"`
	Ticker    string `json:"ticker"`
	Name      string `json:"name"`
	AssetType string `json:"assetType" jsonschema:"stock, etf, crypto, bond, cash, real_estate, commodity or other"`
	Exchange  string `json:"exchange,omitempty"`
	// Quantity is a sum of units and only means something per asset — never
	// across rows. MarketValue is what compares, which is why it alone is
	// converted into DisplayCurrency.
	Quantity        string `json:"quantity" jsonschema:"units held across every portfolio"`
	Currency        string `json:"currency" jsonschema:"the currency the asset is quoted in, which is what marketPrice is in"`
	MarketPrice     string `json:"marketPrice,omitempty" jsonschema:"per-unit price behind marketValue, in currency. Empty when priceSource is cost."`
	MarketValue     string `json:"marketValue" jsonschema:"the whole position, in displayCurrency"`
	DisplayCurrency string `json:"displayCurrency"`
	Portfolios      int64  `json:"portfolios" jsonschema:"how many of the user's portfolios hold this asset"`
	PriceSource     string `json:"priceSource" jsonschema:"own (the user's own market-data key), manual (an operator-entered reference price), or cost (no price available; the position is carried at what it cost and shows no gain)"`
}

// AllocationOutput is the answer of get_allocation.
type AllocationOutput struct {
	Allocation []AllocationSlice `json:"allocation"`
}

// AllocationSlice is one asset category and what the user holds of it. Every
// slice is in the same currency, which is what makes the shares add up.
type AllocationSlice struct {
	Category    string `json:"category" jsonschema:"stock, etf, crypto, bond, cash, real_estate, commodity or other"`
	MarketValue string `json:"marketValue"`
	Currency    string `json:"currency"`
	Unconverted int64  `json:"positionsUnconverted" jsonschema:"positions in this category added at face value because no exchange rate reached currency"`
}

// TransactionsOutput is the answer of list_recent_transactions.
type TransactionsOutput struct {
	Transactions []Transaction `json:"transactions"`
}

// Transaction is one recorded trade.
type Transaction struct {
	ID        string `json:"id"`
	Ticker    string `json:"ticker"`
	AssetName string `json:"assetName,omitempty"`
	Type      string `json:"type" jsonschema:"buy, sell, dividend, split, transfer_in, transfer_out, fee or interest"`
	Quantity  string `json:"quantity"`
	Price     string `json:"price" jsonschema:"the per-unit price, in currency"`
	Currency  string `json:"currency" jsonschema:"the currency the trade was quoted in"`
	// FXRate is historical, not current: it is what one unit of Currency bought
	// in CostCurrency on the day of the trade. Re-deriving it from today's rate
	// is exactly the mistake the column exists to prevent.
	FXRate       string `json:"fxRate" jsonschema:"how much of costCurrency one unit of currency bought on the day of the trade; 1 when the two agree"`
	CostCurrency string `json:"costCurrency,omitempty" jsonschema:"the currency the position is costed in"`
	Fees         string `json:"fees"`
	FeesCurrency string `json:"feesCurrency,omitempty" jsonschema:"which of the two currencies the commission is in"`
	Date         string `json:"date" jsonschema:"trade date, RFC 3339"`
	Notes        string `json:"notes,omitempty"`
}

// GrowthOutput is the answer of get_portfolio_growth.
type GrowthOutput struct {
	Summary GrowthSummary `json:"summary"`
	Points  []GrowthPoint `json:"points"`
}

// GrowthSummary carries two readings of the same series that must not be
// confused: how the value moved (which counts deposits as growth) and what was
// actually earned.
type GrowthSummary struct {
	FirstDate    string `json:"firstDate,omitempty" jsonschema:"date of the first point, RFC 3339"`
	InitialValue string `json:"initialValue"`
	CurrentValue string `json:"currentValue"`
	// GrowthPct includes money paid in. It is not a return.
	GrowthPct   string `json:"growthPct" jsonschema:"how the total value moved between the first point and the last, which counts deposits as growth"`
	GainLoss    string `json:"gainLoss" jsonschema:"the actual profit at the latest point: market value minus invested capital"`
	GainLossPct string `json:"gainLossPct" jsonschema:"gainLoss relative to invested capital — this is the return, growthPct is not"`
	Currency    string `json:"currency"`
}

// GrowthPoint is one day of the series.
type GrowthPoint struct {
	Date        string `json:"date" jsonschema:"RFC 3339"`
	TotalValue  string `json:"totalValue"`
	CostBasis   string `json:"costBasis"`
	GainLoss    string `json:"gainLoss"`
	GainLossPct string `json:"gainLossPct"`
	Currency    string `json:"currency"`
	// NetFlow is what has to be netted out of the change in TotalValue to leave
	// a return: a deposit raises the value without anyone having earned it.
	NetFlow     string `json:"netFlow" jsonschema:"money paid in (positive) or taken out (negative) since the previous point"`
	Unconverted int64  `json:"portfoliosUnconverted" jsonschema:"portfolios added at face value on this date because no exchange rate reached currency"`
}

// PlatformsOutput is the answer of list_platforms.
type PlatformsOutput struct {
	Platforms []Platform `json:"platforms"`
}

// Platform is one broker, bank or wallet the user holds positions through.
type Platform struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type" jsonschema:"broker, investment_bank, trading_platform, neobank, de_fi, crypto_wallet, mutual_funds, brokerage_house or other"`
	IsActive    bool   `json:"isActive"`
	Positions   int64  `json:"positions" jsonschema:"open positions held through this platform"`
	Assets      int64  `json:"assets" jsonschema:"distinct assets those positions are spread over"`
	Portfolios  int64  `json:"portfolios" jsonschema:"portfolios those positions belong to"`
	// CostBasis is the field the REST API still calls totalValue; the name here
	// says what it is, since nothing outside this package reads it.
	CostBasis        string `json:"costBasis" jsonschema:"what the owner paid for what is still held, at weighted-average cost"`
	MarketValue      string `json:"marketValue"`
	DisplayCurrency  string `json:"displayCurrency"`
	PricedWithOwnKey int64  `json:"pricedWithOwnKey"`
	PricedManually   int64  `json:"pricedManually"`
	ValuedAtCost     int64  `json:"valuedAtCost" jsonschema:"positions carried at cost, contributing zero gain"`
}

// AssetsOutput is the answer of search_assets.
type AssetsOutput struct {
	Assets []Asset `json:"assets"`
}

// Asset is one row of the catalog. It says nothing about whether the caller
// holds it — get_holdings answers that.
type Asset struct {
	ID        string `json:"id"`
	Ticker    string `json:"ticker"`
	Name      string `json:"name"`
	AssetType string `json:"assetType" jsonschema:"stock, etf, crypto, bond, cash, real_estate, commodity or other"`
	Exchange  string `json:"exchange,omitempty"`
	Currency  string `json:"currency" jsonschema:"the currency the asset is quoted in"`
	// Price is the catalog's shared reference price, not a per-user valuation:
	// it is what an operator entered or the last sync stored. A holding's own
	// price comes from get_holdings.
	Price          string `json:"price,omitempty" jsonschema:"shared reference price in currency, if the catalog has one"`
	PriceUpdatedAt string `json:"priceUpdatedAt,omitempty" jsonschema:"when that reference price was last written, RFC 3339"`
	IsCurated      bool   `json:"isCurated" jsonschema:"false for an asset a user contributed, which only they can see"`
}

// RatesOutput is the answer of list_exchange_rates.
type RatesOutput struct {
	Rates []ExchangeRate `json:"rates"`
}

// ExchangeRate is one shared conversion rate: how much of ToCurrency one unit
// of FromCurrency buys.
type ExchangeRate struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Rate   string `json:"rate" jsonschema:"how much of the to currency one unit of the from currency buys"`
	Date   string `json:"date" jsonschema:"the day the rate is for, RFC 3339"`
	Source string `json:"source" jsonschema:"who published it: a provider id, or manual for an operator-entered rate"`
}
