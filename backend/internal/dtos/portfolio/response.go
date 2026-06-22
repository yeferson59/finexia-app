package portfolio

import (
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/finexia-app/internal/entities"
)

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

// NewPortfolioDetailResponse maps a portfolio entity (with its entries and
// assets populated) into the detail response consumed by the frontend.
func NewPortfolioDetailResponse(p entities.Portfolio) PortfolioDetailResponseDTO {
	holdings := make([]HoldingResponseDTO, 0, len(p.PortfolioEntries))
	for _, entry := range p.PortfolioEntries {
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
