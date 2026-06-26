package portfolio

import (
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/entities"
)

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

func NewTransactionResponse(t entities.Transaction) TransactionResponseDTO {
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

func NewTransactionListResponse(txns []entities.Transaction) []TransactionResponseDTO {
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

func NewUserTransactionResponse(t entities.Transaction) UserTransactionResponseDTO {
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

func NewUserTransactionListResponse(txns []entities.Transaction) []UserTransactionResponseDTO {
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

func NewAllocationResponse(items []entities.AllocationItem) []AllocationItemDTO {
	var total float64
	for _, item := range items {
		v, _ := strconv.ParseFloat(item.MarketValue, 64)
		total += v
	}

	result := make([]AllocationItemDTO, 0, len(items))
	for _, item := range items {
		v, _ := strconv.ParseFloat(item.MarketValue, 64)
		var pct float64
		if total > 0 {
			pct = math.Round((v/total)*10000) / 100
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
}

// PortfolioDetailResponseDTO is the payload returned for a single portfolio,
// including its holdings.
type PortfolioDetailResponseDTO struct {
	ID           uuid.UUID              `json:"id"`
	UserID       uuid.UUID              `json:"userId"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Type         entities.PortfolioType `json:"type"`
	BaseCurrency string                 `json:"baseCurrency"`
	IsDefault    bool                   `json:"isDefault"`
	RiskID       uuid.UUID              `json:"riskId"`
	RiskName     string                 `json:"riskName"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
	Holdings     []HoldingResponseDTO   `json:"holdings"`
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

func NewGrowthResponse(points []entities.PortfolioGrowthPoint, summary entities.PortfolioGrowthSummary) GrowthResponseDTO {
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

// NewPortfolioDetailResponse maps a portfolio entity (with its entries and
// assets populated) into the detail response consumed by the frontend.
func NewPortfolioDetailResponse(p entities.Portfolio) PortfolioDetailResponseDTO {
	holdings := make([]HoldingResponseDTO, 0, len(p.PortfolioEntries))
	for _, entry := range p.PortfolioEntries {
		marketPrice := ""
		if entry.Asset.CurrentPrice != nil {
			marketPrice = entry.Asset.CurrentPrice.String()
		}

		holdings = append(holdings, HoldingResponseDTO{
			ID:           entry.ID,
			AssetID:      entry.AssetID,
			Ticker:       entry.Asset.Ticker,
			Name:         entry.Asset.Name,
			AssetType:    string(entry.Asset.AssetType),
			Exchange:     entry.Asset.Exchange,
			Currency:     entry.Asset.Currency,
			Quantity:     entry.Quantity.String(),
			Price:        entry.Price.String(),
			MarketPrice:  marketPrice,
			CostCurrency: entry.CostCurrency,
			Category:     string(entry.Category),
			EntryDate:    entry.EntryDate,
			Notes:        entry.Notes,
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
