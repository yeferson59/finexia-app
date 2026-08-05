package portfolio

import (
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
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

type UpdateTransactionRequestDTO struct {
	Type            string        `json:"type"`
	Quantity        money.Decimal `json:"quantity"`
	Price           money.Money   `json:"price"`
	Currency        string        `json:"currency"`
	Fees            money.Money   `json:"fees"`
	TransactionDate time.Time     `json:"transactionDate"`
	Notes           string        `json:"notes"`
}

type UpdatePortfolioRequestDTO struct {
	Name        string `json:"name,omitzero"`
	Description string `json:"description,omitzero"`
	Type        string `json:"type,omitzero"`
	RiskID      string `json:"riskId,omitzero"`
	IsDefault   bool   `json:"isDefault"`
}

type TransactionResponseDTO struct {
	ID              uuid.UUID `json:"id"`
	EntryID         uuid.UUID `json:"entryId"`
	Type            string    `json:"type"`
	Quantity        string    `json:"quantity"`
	Price           string    `json:"price"`
	Currency        string    `json:"currency"`
	Fees            string    `json:"fees"`
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
		Currency:        t.Currency,
		Fees:            t.Fees.String(),
		TransactionDate: t.TransactionDate,
		Notes:           t.Notes,
		CreatedAt:       t.CreatedAt,
	}
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
	Fees            string    `json:"fees"`
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
		Currency:        t.Currency,
		Fees:            t.Fees.String(),
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
}

// NewAllocationResponse turns each category's market value into its share of
// the whole. Both the sum and the division run on gofinance's decimal engine —
// summing the categories in float64 drifted against the total the same rows add
// up to in Postgres, and the shares could miss 100% by more than the rounding
// alone accounts for.
func NewAllocationResponse(items []AllocationItem) []AllocationItemDTO {
	values := make([]decimal.Decimal, len(items))
	total := decimal.Zero
	for i, item := range items {
		v, err := decimal.NewFromString(item.MarketValue)
		if err != nil {
			v = decimal.Zero
		}
		values[i] = v
		total = total.Add(v)
	}

	result := make([]AllocationItemDTO, 0, len(items))
	for i, item := range items {
		var pct float64
		if total.IsPos() {
			// Scale to a percentage before dividing: Div fixes the result's
			// precision, so dividing first spends those digits on the 0.xx
			// fraction and throws away the ones the percentage is read at.
			if share, err := values[i].Mul(oneHundred).Div(total); err == nil {
				// RoundHAZ is half-away-from-zero, the mode math.Round used
				// here and the one a reader expects of a displayed percentage.
				pct = share.RoundHAZ(2).InexactFloat64()
			}
		}
		result = append(result, AllocationItemDTO{
			Category:    string(item.Category),
			MarketValue: item.MarketValue,
			Percent:     pct,
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
	TotalValue  string    `json:"totalValue"`
}

func newPlatformResponse(p PlatformStats) PlatformResponseDTO {
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
	}
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
}

type GrowthSummaryDTO struct {
	FirstDate      string `json:"firstDate"`
	InitialValue   string `json:"initialValue"`
	CurrentValue   string `json:"currentValue"`
	TotalGrowthPct string `json:"totalGrowthPct"`
}

type GrowthResponseDTO struct {
	Points  []GrowthDataPointDTO `json:"points"`
	Summary GrowthSummaryDTO     `json:"summary"`
}

func NewGrowthResponse(points []GrowthPoint, summary GrowthSummary) GrowthResponseDTO {
	dtos := make([]GrowthDataPointDTO, 0, len(points))
	for _, p := range points {
		dtos = append(dtos, GrowthDataPointDTO{
			Date:          p.Date.Format("2006-01-02"),
			TotalValue:    p.TotalValue,
			TotalCostBase: p.TotalCostBase,
			GainLoss:      p.GainLoss,
			GainLossPct:   p.GainLossPct,
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
			ID:             entry.ID,
			AssetID:        entry.AssetID,
			Ticker:         entry.Asset.Ticker,
			Name:           entry.Asset.Name,
			AssetType:      string(entry.Asset.AssetType),
			Exchange:       entry.Asset.Exchange,
			Currency:       entry.Asset.Currency,
			Quantity:       entry.Quantity.String(),
			Price:          entry.Price.String(),
			MarketPrice:    marketPrice,
			CostCurrency:   entry.CostCurrency,
			Category:       string(entry.Category),
			EntryDate:      entry.EntryDate,
			Notes:          entry.Notes,
			PriceSource:    string(entry.PriceSource),
			PriceUpdatedAt: entry.Asset.PriceUpdatedAt,
		})
	}

	return PortfolioDetailResponseDTO{
		ID:           p.ID,
		UserID:       p.UserID,
		Name:         p.Name,
		Description:  p.Description,
		Type:         p.Type,
		BaseCurrency: p.BaseCurrency,
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
