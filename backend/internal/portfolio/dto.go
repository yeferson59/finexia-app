package portfolio

import (
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

type CreatePortfolioRequestDTO struct {
	Name        string         `json:"name" validate:"required"`
	Description string         `json:"description"`
	Currency    money.Currency `json:"currency" validate:"required"`
	Type        string         `json:"type" validate:"required"`
	RiskID      uuid.UUID      `json:"riskId"`
	PriceValue  money.Money    `json:"priceValue"`
	IsDefault   bool           `json:"isDefault"`
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
	PortfolioID     uuid.UUID       `json:"portfolioId" validate:"required"`
	AssetID         uuid.UUID       `json:"assetId" validate:"required"`
	SourceID        uuid.UUID       `json:"sourceId" validate:"required"`
	TransactionType string          `json:"transactionType"`
	Quantity        decimal.Decimal `json:"quantity" validate:"required"`
	Price           money.Money     `json:"price" validate:"required"`
	// CostCurrency is the position's: what the account was debited in, and what
	// its cost basis is reported in from here on. Currency is the trade's — the
	// one Price is quoted in, which is the asset's own on a normal fill. They
	// differ only when the broker converted, and then FXRate is what it
	// converted at. Currency is optional and defaults to CostCurrency, so a
	// client that knows nothing about any of this keeps working unchanged.
	CostCurrency money.Currency  `json:"costCurrency" validate:"required"`
	Currency     money.Currency  `json:"currency"`
	FXRate       decimal.Decimal `json:"fxRate"`
	EntryDate    time.Time       `json:"entryDate" validate:"required"`
	Notes        string          `json:"notes"`
}

type CreateTransactionRequestDTO struct {
	Type     string          `json:"type" validate:"required"`
	Quantity decimal.Decimal `json:"quantity" validate:"required"`
	Price    money.Money     `json:"price" validate:"required"`
	Currency money.Currency  `json:"currency" validate:"required"`
	// FXRate converts Currency into the cost currency of the position this is
	// recorded on, at the time of the trade. Omitted means 1, which is only
	// accepted when the two currencies are the same — see
	// TransactionInput.Validate for why a missing rate is refused rather than
	// looked up.
	FXRate decimal.Decimal `json:"fxRate"`
	Fees   money.Money     `json:"fees"`
	// FeesCurrency must be Currency or the position's cost currency. Omitted
	// means Currency, which is where a commission sat on every row the app could
	// write before this field existed.
	FeesCurrency    money.Currency `json:"feesCurrency"`
	TransactionDate time.Time      `json:"transactionDate" validate:"required"`
	Notes           string         `json:"notes"`
}

type UpdateTransactionRequestDTO struct {
	Type            string          `json:"type"`
	Quantity        decimal.Decimal `json:"quantity"`
	Price           money.Money     `json:"price"`
	Currency        money.Currency  `json:"currency"`
	FXRate          decimal.Decimal `json:"fxRate"`
	Fees            money.Money     `json:"fees"`
	FeesCurrency    money.Currency  `json:"feesCurrency"`
	TransactionDate time.Time       `json:"transactionDate"`
	Notes           string          `json:"notes"`
}

// Input folds the three write DTOs into the one shape the service takes. The
// transaction type is left to the caller: the handler has already rejected an
// invalid one by the time this is built.
func (d CreateTransactionRequestDTO) Input(txnType TransactionType) TransactionInput {
	return TransactionInput{
		Type:            txnType,
		Quantity:        d.Quantity,
		Price:           d.Price,
		Currency:        d.Currency,
		FXRate:          d.FXRate,
		Fees:            d.Fees,
		FeesCurrency:    d.FeesCurrency,
		TransactionDate: d.TransactionDate,
		Notes:           d.Notes,
	}
}

func (d UpdateTransactionRequestDTO) Input(txnType TransactionType) TransactionInput {
	return TransactionInput{
		Type:            txnType,
		Quantity:        d.Quantity,
		Price:           d.Price,
		Currency:        d.Currency,
		FXRate:          d.FXRate,
		Fees:            d.Fees,
		FeesCurrency:    d.FeesCurrency,
		TransactionDate: d.TransactionDate,
		Notes:           d.Notes,
	}
}

// Input builds the opening trade of a new position. The entry date doubles as
// the transaction date and the notes are shared, which is what the endpoint has
// always done — a position and the purchase that opened it are one act here.
func (d CreatePortfolioEntryRequestDTO) Input(txnType TransactionType) TransactionInput {
	return TransactionInput{
		Type:            txnType,
		Quantity:        d.Quantity,
		Price:           d.Price,
		Currency:        d.Currency,
		FXRate:          d.FXRate,
		TransactionDate: d.EntryDate,
		Notes:           d.Notes,
	}
}

// DeleteEntryResponseDTO reports what a position's removal took with it.
//
// It is a body on a delete because the number is not derivable by the caller:
// the client knows it asked to remove one position, not that eleven
// transactions cascaded out behind it.
type DeleteEntryResponseDTO struct {
	DeletedTransactions int `json:"deletedTransactions"`
}

type UpdatePortfolioRequestDTO struct {
	Name        string `json:"name,omitzero"`
	Description string `json:"description,omitzero"`
	Type        string `json:"type,omitzero"`
	RiskID      string `json:"riskId,omitzero"`
	IsDefault   bool   `json:"isDefault"`
}

type TransactionResponseDTO struct {
	ID       uuid.UUID `json:"id"`
	EntryID  uuid.UUID `json:"entryId"`
	Type     string    `json:"type"`
	Quantity string    `json:"quantity"`
	Price    string    `json:"price"`
	Currency string    `json:"currency"`
	// FXRate and CostCurrency travel together with Price: on their own, "606.60
	// EUR" does not say what the trade cost an account funded in dollars, and
	// the rate does not say what it converts into. With all three the client can
	// show the line the way the broker's confirmation does.
	FXRate       string `json:"fxRate"`
	CostCurrency string `json:"costCurrency,omitempty"`
	Fees         string `json:"fees"`
	// FeesCurrency is Currency or CostCurrency. It is its own field because a
	// commission is not always billed on the same side as the fill, and a
	// client that assumed it was would misstate it by the whole rate.
	FeesCurrency    string    `json:"feesCurrency,omitempty"`
	TransactionDate time.Time `json:"transactionDate"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"createdAt"`
}

func NewTransactionResponse(t Transaction) TransactionResponseDTO {
	return TransactionResponseDTO{
		ID:              t.ID,
		EntryID:         t.EntryID,
		Type:            string(t.Type),
		Quantity:        t.Quantity.String(),
		Price:           t.Price.String(),
		Currency:        t.Currency.String(),
		FXRate:          transactionRate(t).String(),
		CostCurrency:    costCurrencyCode(t),
		Fees:            t.Fees.String(),
		FeesCurrency:    feesCurrencyCode(t),
		TransactionDate: t.TransactionDate,
		Notes:           t.Notes,
		CreatedAt:       t.CreatedAt,
	}
}

// transactionRate reports 1 rather than 0 for a transaction read from a query
// that does not select fx_rate. A rate of zero is not a rate, and a client
// multiplying by it would erase the price; 1 is both the truthful answer for
// every same-currency row and the harmless one everywhere else.
func transactionRate(t Transaction) decimal.Decimal {
	if t.FXRate.IsZero() {
		return decimal.One
	}

	return t.FXRate
}

// costCurrencyCode omits the field instead of emitting money.XXX when the
// transaction was read without its entry: an absent currency is something a
// client can branch on, whereas "XXX" reads like a real answer.
func costCurrencyCode(t Transaction) string {
	if t.CostCurrency == money.XXX {
		return ""
	}

	return t.CostCurrency.String()
}

// feesCurrencyCode falls back to the trade currency, which is where a fee sat
// on every row written before the column existed and is the default the API
// applies to one that arrives without it. Unlike the cost currency there is no
// reason to leave it absent: the transaction always knows this one.
func feesCurrencyCode(t Transaction) string {
	if t.FeesCurrency == money.XXX {
		return t.Currency.String()
	}

	return t.FeesCurrency.String()
}

func NewTransactionListResponse(txns []Transaction) []TransactionResponseDTO {
	result := make([]TransactionResponseDTO, 0, len(txns))
	for _, t := range txns {
		result = append(result, NewTransactionResponse(t))
	}
	return result
}

type UserTransactionResponseDTO struct {
	ID              uuid.UUID `json:"id"`
	EntryID         uuid.UUID `json:"entryId"`
	Type            string    `json:"type"`
	Quantity        string    `json:"quantity"`
	Price           string    `json:"price"`
	Currency        string    `json:"currency"`
	FXRate          string    `json:"fxRate"`
	CostCurrency    string    `json:"costCurrency,omitempty"`
	Fees            string    `json:"fees"`
	FeesCurrency    string    `json:"feesCurrency,omitempty"`
	TransactionDate time.Time `json:"transactionDate"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"createdAt"`
	AssetTicker     string    `json:"assetTicker"`
	AssetName       string    `json:"assetName"`
}

func NewUserTransactionResponse(t Transaction) UserTransactionResponseDTO {
	return UserTransactionResponseDTO{
		ID:              t.ID,
		EntryID:         t.EntryID,
		Type:            string(t.Type),
		Quantity:        t.Quantity.String(),
		Price:           t.Price.String(),
		Currency:        t.Currency.String(),
		FXRate:          transactionRate(t).String(),
		CostCurrency:    costCurrencyCode(t),
		Fees:            t.Fees.String(),
		FeesCurrency:    feesCurrencyCode(t),
		TransactionDate: t.TransactionDate,
		Notes:           t.Notes,
		CreatedAt:       t.CreatedAt,
		AssetTicker:     t.Entry.Asset.Ticker,
		AssetName:       t.Entry.Asset.Name,
	}
}

func NewUserTransactionListResponse(txns []Transaction) []UserTransactionResponseDTO {
	result := make([]UserTransactionResponseDTO, 0, len(txns))
	for _, t := range txns {
		result = append(result, NewUserTransactionResponse(t))
	}
	return result
}

type AllocationItemDTO struct {
	Category    string  `json:"category"`
	MarketValue string  `json:"marketValue"`
	Percent     float64 `json:"percent"`
	// Currency is what MarketValue is in — the same for every item, which is
	// what makes Percent meaningful.
	Currency string `json:"currency"`
	// PositionsUnconverted counts the positions in this category that had no
	// rate to Currency and are therefore included at face value. Non-zero means
	// this slice, and the share computed from it, mix currencies.
	PositionsUnconverted int64 `json:"positionsUnconverted"`
}

// percentShares turns a list of text amounts into each one's share of their
// total, in percent. Both the sum and the division run on gofinance's decimal
// engine — summing them in float64 drifted against the total the same rows add
// up to in Postgres, and the shares could miss 100% by more than the rounding
// alone accounts for. A malformed amount reads as zero, the same as everywhere
// else these strings are parsed.
func percentShares(amounts []string) []float64 {
	values := make([]decimal.Decimal, len(amounts))
	total := decimal.Zero
	for i, amount := range amounts {
		v, err := decimal.NewFromString(amount)
		if err != nil {
			v = decimal.Zero
		}
		values[i] = v
		total = total.Add(v)
	}

	shares := make([]float64, len(amounts))
	if !total.IsPos() {
		return shares
	}

	for i, value := range values {
		// Scale to a percentage before dividing: Div fixes the result's
		// precision, so dividing first spends those digits on the 0.xx
		// fraction and throws away the ones the percentage is read at.
		if share, err := value.Mul(oneHundred).Div(total); err == nil {
			// RoundHAZ is half-away-from-zero, the mode math.Round used here
			// and the one a reader expects of a displayed percentage.
			shares[i] = share.RoundHAZ(2).InexactFloat64()
		}
	}

	return shares
}

// NewAllocationResponse turns each category's market value into its share of
// the whole.
func NewAllocationResponse(items []AllocationItem) []AllocationItemDTO {
	amounts := make([]string, len(items))
	for i, item := range items {
		amounts[i] = item.MarketValue
	}
	shares := percentShares(amounts)

	result := make([]AllocationItemDTO, 0, len(items))
	for i, item := range items {
		result = append(result, AllocationItemDTO{
			Category:             string(item.Category),
			MarketValue:          item.MarketValue,
			Percent:              shares[i],
			Currency:             item.Currency.String(),
			PositionsUnconverted: item.PositionsUnconverted,
		})
	}
	return result
}

// AssetHoldingDTO is one asset totalled across every portfolio the user owns.
// See AssetHolding for what each amount is denominated in; Percent is this
// asset's share of everything the user holds, computed the same way — and over
// the same positions — as the allocation's, so the two charts agree.
type AssetHoldingDTO struct {
	AssetID     uuid.UUID `json:"assetId"`
	Ticker      string    `json:"ticker"`
	Name        string    `json:"name"`
	AssetType   string    `json:"assetType"`
	Exchange    string    `json:"exchange"`
	Currency    string    `json:"currency"`
	Quantity    string    `json:"quantity"`
	MarketPrice string    `json:"marketPrice"`
	MarketValue string    `json:"marketValue"`
	Percent     float64   `json:"percent"`
	// DisplayCurrency is what MarketValue is in — the same for every row, which
	// is what makes Percent meaningful.
	DisplayCurrency      string `json:"displayCurrency"`
	Portfolios           int64  `json:"portfolios"`
	PriceSource          string `json:"priceSource"`
	PositionsUnconverted int64  `json:"positionsUnconverted"`
}

// NewAssetHoldingsResponse gives each asset its share of the whole.
func NewAssetHoldingsResponse(holdings []AssetHolding) []AssetHoldingDTO {
	amounts := make([]string, len(holdings))
	for i, holding := range holdings {
		amounts[i] = holding.MarketValue
	}
	shares := percentShares(amounts)

	result := make([]AssetHoldingDTO, 0, len(holdings))
	for i, holding := range holdings {
		result = append(result, AssetHoldingDTO{
			AssetID:              holding.AssetID,
			Ticker:               holding.Ticker,
			Name:                 holding.Name,
			AssetType:            string(holding.AssetType),
			Exchange:             holding.Exchange,
			Currency:             holding.Currency.String(),
			Quantity:             holding.Quantity,
			MarketPrice:          holding.MarketPrice,
			MarketValue:          holding.MarketValue,
			Percent:              shares[i],
			DisplayCurrency:      holding.DisplayCurrency.String(),
			Portfolios:           holding.Portfolios,
			PriceSource:          string(holding.PriceSource),
			PositionsUnconverted: holding.PositionsUnconverted,
		})
	}
	return result
}

type PlatformResponseDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SourceType  string    `json:"sourceType"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Investments int64     `json:"investments"`
	// TotalValue is what the platform cost and MarketValue what it is worth;
	// GainLoss is the difference and GainLossPct that difference over the cost.
	// All four are in DisplayCurrency.
	TotalValue  string  `json:"totalValue"`
	MarketValue string  `json:"marketValue"`
	GainLoss    string  `json:"gainLoss"`
	GainLossPct float64 `json:"gainLossPct"`
	// Percent is this platform's share of everything the account has invested,
	// which is what makes the ordering readable: "the biggest" means little
	// until it is "the biggest, and it holds 62% of the money". It is a share of
	// the cost basis, not of market value, so it answers where the money was put
	// rather than where it happens to have grown.
	//
	// Zero when the platform is read on its own: a share needs the whole set,
	// and inventing 100% for a single row would be worse than saying nothing.
	Percent float64 `json:"percent"`
	// The currency the amounts are in, and how many of the entries behind them
	// could not be converted into it. See PlatformStats for why both travel
	// with the amount.
	DisplayCurrency      string `json:"displayCurrency"`
	PositionsUnconverted int64  `json:"positionsUnconverted"`
}

func newPlatformResponse(p PlatformStats) PlatformResponseDTO {
	gain, pct := gainOf(p.TotalValue, p.MarketValue)

	return PlatformResponseDTO{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		SourceType:  string(p.SourceType),
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		Investments: p.Investments,
		TotalValue:  p.TotalValue,
		MarketValue: p.MarketValue,
		GainLoss:    gain,
		GainLossPct: pct,

		DisplayCurrency:      p.DisplayCurrency.String(),
		PositionsUnconverted: p.PositionsUnconverted,
	}
}

// NewPlatformListResponse builds the listing, filling in each platform's share
// of the account's invested total.
//
// The share is computed here and not in SQL because it only means anything
// across the whole set: the same query serves the re-read of a single platform
// after an edit, where a window function would confidently report 100%. The
// shares are computed over exactly the rows being returned, the same way the
// allocation's are, so the column adds to 100 on screen.
func NewPlatformListResponse(platforms []PlatformStats) []PlatformResponseDTO {
	invested := make([]string, len(platforms))
	for i, p := range platforms {
		invested[i] = p.TotalValue
	}
	shares := percentShares(invested)

	result := make([]PlatformResponseDTO, 0, len(platforms))
	for i, p := range platforms {
		dto := newPlatformResponse(p)
		dto.Percent = shares[i]
		result = append(result, dto)
	}

	return result
}

// gainOf turns a cost and a market value into the gain between them and that
// gain as a percentage of the cost.
//
// Both run on the decimal engine rather than through float64, for the reason
// percentShares gives: these amounts are summed in Postgres at eight decimals
// and a float round trip drifts against the totals they came from. An amount
// that will not parse reads as zero, the same as everywhere else these strings
// are read.
//
// A cost of zero yields a zero percentage rather than an infinity: a platform
// that has been fully sold, or that holds only positions valued at cost, has no
// base to express a return against, and the amount beside it already says so.
func gainOf(cost, market string) (string, float64) {
	costDec, err := decimal.NewFromString(cost)
	if err != nil {
		costDec = decimal.Zero
	}

	marketDec, err := decimal.NewFromString(market)
	if err != nil {
		marketDec = decimal.Zero
	}

	gain := marketDec.Sub(costDec)
	if !costDec.IsPos() {
		return gain.String(), 0
	}

	// Scaled to a percentage before dividing, the same way percentShares does
	// it and for the same reason.
	pct, err := gain.Mul(oneHundred).Div(costDec)
	if err != nil {
		return gain.String(), 0
	}

	return gain.String(), pct.RoundHAZ(2).InexactFloat64()
}

// HoldingResponseDTO is a flattened representation of a portfolio entry joined
// with its asset, ready to be consumed by the frontend holdings view.
type HoldingResponseDTO struct {
	ID           uuid.UUID `json:"id"`
	AssetID      uuid.UUID `json:"assetId"`
	Ticker       string    `json:"ticker"`
	Name         string    `json:"name"`
	AssetType    string    `json:"assetType"`
	Exchange     string    `json:"exchange"`
	Currency     string    `json:"currency"`
	Quantity     string    `json:"quantity"`
	Price        string    `json:"price"`
	MarketPrice  string    `json:"marketPrice"`
	CostCurrency string    `json:"costCurrency"`
	Category     string    `json:"category"`
	EntryDate    time.Time `json:"entryDate"`
	Notes        string    `json:"notes"`
	// PriceSource qualifies MarketPrice: "own" (this user's key), "manual"
	// (operator reference price) or "cost" (no price — MarketPrice is empty and
	// the position is valued at Price). A client that shows a return has to
	// read this: at "cost" the return is zero by construction.
	PriceSource string `json:"priceSource"`
	// PriceUpdatedAt is when MarketPrice was fetched, null when there is none.
	// Under BYO-key a sync can take hours and is capped per request, so a
	// price's age is part of reading it.
	PriceUpdatedAt *time.Time `json:"priceUpdatedAt"`
	// CostBasisBase and MarketValueBase are quantity × price and
	// quantity × market price, converted to the portfolio's BaseCurrency. They
	// are the only figures here a client may add across holdings: Price and
	// MarketPrice stay in Currency/CostCurrency, which differ per position.
	CostBasisBase   string `json:"costBasisBase"`
	MarketValueBase string `json:"marketValueBase"`
	// FXConverted is false when no rate connected this position's currency to
	// the base one. The two totals above then hold native amounts, and a client
	// that sums them is adding different currencies — say so instead.
	FXConverted bool `json:"fxConverted"`
}

// PortfolioDetailResponseDTO is the payload returned for a single portfolio,
// including its holdings.
type PortfolioDetailResponseDTO struct {
	ID           uuid.UUID            `json:"id"`
	UserID       uuid.UUID            `json:"userId"`
	Name         string               `json:"name"`
	Description  string               `json:"description"`
	Type         Type                 `json:"type"`
	BaseCurrency string               `json:"baseCurrency"`
	IsDefault    bool                 `json:"isDefault"`
	RiskID       uuid.UUID            `json:"riskId"`
	RiskName     string               `json:"riskName"`
	CreatedAt    time.Time            `json:"createdAt"`
	UpdatedAt    time.Time            `json:"updatedAt"`
	Holdings     []HoldingResponseDTO `json:"holdings"`
}

type GrowthDataPointDTO struct {
	Date          string `json:"date"`
	TotalValue    string `json:"totalValue"`
	TotalCostBase string `json:"totalCostBase"`
	GainLoss      string `json:"gainLoss"`
	GainLossPct   string `json:"gainLossPct"`
	// Portafolios sumados a esta fecha sin tasa con la que convertirlos, y por
	// tanto contados a valor nominal. Normalmente 0.
	PortfoliosUnconverted int64 `json:"portfoliosUnconverted"`
	// Dinero que el dueño metió (positivo) o sacó (negativo) entre el punto
	// anterior y este. Es lo que hay que descontar de la variación de
	// `totalValue` para que quede rentabilidad: un depósito sube el valor sin
	// que nadie haya ganado nada.
	NetFlow string `json:"netFlow"`
}

type GrowthSummaryDTO struct {
	FirstDate      string `json:"firstDate"`
	InitialValue   string `json:"initialValue"`
	CurrentValue   string `json:"currentValue"`
	TotalGrowthPct string `json:"totalGrowthPct"`
	GainLoss       string `json:"gainLoss"`
	GainLossPct    string `json:"gainLossPct"`
	Currency       string `json:"currency"`
}

type GrowthResponseDTO struct {
	Points  []GrowthDataPointDTO `json:"points"`
	Summary GrowthSummaryDTO     `json:"summary"`
}

// netFlowOrZero keeps the payload's amount fields all numeric strings: a point
// with no transactions carries "0", not the empty string a zero-value struct
// would otherwise ship.
func netFlowOrZero(raw string) string {
	if raw == "" {
		return "0"
	}

	return raw
}

func NewGrowthResponse(points []GrowthPoint, summary GrowthSummary) GrowthResponseDTO {
	dtos := make([]GrowthDataPointDTO, 0, len(points))
	for _, p := range points {
		dtos = append(dtos, GrowthDataPointDTO{
			Date:                  p.Date.Format("2006-01-02"),
			TotalValue:            p.TotalValue,
			TotalCostBase:         p.TotalCostBase,
			GainLoss:              p.GainLoss,
			GainLossPct:           p.GainLossPct,
			PortfoliosUnconverted: p.PortfoliosUnconverted,
			NetFlow:               netFlowOrZero(p.NetFlow),
		})
	}
	firstDate := ""
	if !summary.FirstDate.IsZero() {
		firstDate = summary.FirstDate.Format("2006-01-02")
	}
	return GrowthResponseDTO{
		Points: dtos,
		Summary: GrowthSummaryDTO{
			FirstDate:      firstDate,
			InitialValue:   summary.InitialValue,
			CurrentValue:   summary.CurrentValue,
			TotalGrowthPct: summary.TotalGrowthPct,
			GainLoss:       summary.GainLoss,
			GainLossPct:    summary.GainLossPct,
			Currency:       summary.Currency.String(),
		},
	}
}

type TopTransactionDTO struct {
	Value           string    `json:"value"`
	Type            string    `json:"type"`
	Currency        string    `json:"currency"`
	AssetTicker     string    `json:"assetTicker"`
	AssetName       string    `json:"assetName"`
	TransactionDate time.Time `json:"transactionDate"`
}

// NewPortfolioDetailResponse maps a portfolio (with its entries and assets
// populated) into the detail response consumed by the frontend.
func NewPortfolioDetailResponse(p Portfolio) PortfolioDetailResponseDTO {
	holdings := make([]HoldingResponseDTO, 0, len(p.Entries))
	for _, entry := range p.Entries {
		marketPrice := ""
		if entry.Asset.CurrentPrice != nil {
			marketPrice = entry.Asset.CurrentPrice.String()
		}

		holdings = append(holdings, HoldingResponseDTO{
			ID:              entry.ID,
			AssetID:         entry.AssetID,
			Ticker:          entry.Asset.Ticker,
			Name:            entry.Asset.Name,
			AssetType:       string(entry.Asset.AssetType),
			Exchange:        entry.Asset.Exchange,
			Currency:        entry.Asset.Currency.String(),
			Quantity:        entry.Quantity.String(),
			Price:           entry.Price.String(),
			MarketPrice:     marketPrice,
			CostCurrency:    entry.CostCurrency.String(),
			Category:        string(entry.Category),
			EntryDate:       entry.EntryDate,
			Notes:           entry.Notes,
			PriceSource:     string(entry.PriceSource),
			PriceUpdatedAt:  entry.Asset.PriceUpdatedAt,
			CostBasisBase:   entry.CostBasisBase.String(),
			MarketValueBase: entry.MarketValueBase.String(),
			FXConverted:     entry.FXConverted,
		})
	}

	return PortfolioDetailResponseDTO{
		ID:           p.ID,
		UserID:       p.UserID,
		Name:         p.Name,
		Description:  p.Description,
		Type:         p.Type,
		BaseCurrency: p.BaseCurrency.String(),
		IsDefault:    p.IsDefault,
		RiskID:       p.RiskID,
		RiskName:     p.Risk.Name,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		Holdings:     holdings,
	}
}

type PaginatedTransactionsDTO struct {
	Data       []TransactionResponseDTO `json:"data"`
	Total      int                      `json:"total"`
	Page       int                      `json:"page"`
	Limit      int                      `json:"limit"`
	TotalPages int                      `json:"totalPages"`
}

type CurrencyDTO struct {
	Currency money.Currency `query:"currency"`
}
