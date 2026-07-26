// Package market owns the asset catalog and exchange-rate domains: the request
// DTOs, entities, persistence, services and HTTP handlers for both.
package market

import (
	"github.com/yeferson59/gofinance/v2/money"
)

type CreateAssetRequestDTO struct {
	Ticker    string `json:"ticker"    validate:"required"`
	Name      string `json:"name"      validate:"required"`
	AssetType string `json:"assetType" validate:"required"`
	Exchange  string `json:"exchange"`
	Currency  string `json:"currency"  validate:"required"`
}

type CreateExchangeRateRequestDTO struct {
	FromCurrency string        `json:"fromCurrency" validate:"required"`
	ToCurrency   string        `json:"toCurrency"   validate:"required"`
	Rate         money.Decimal `json:"rate"         validate:"required"`
}

type UpdateExchangeRateRequestDTO struct {
	Rate money.Decimal `json:"rate" validate:"required"`
}

// SaveCredentialRequestDTO carries a user's own provider API key.
//
// It has no response counterpart on purpose: a stored key is never sent back,
// so nothing that leaves this module can carry one.
type SaveCredentialRequestDTO struct {
	APIKey string `json:"apiKey" validate:"required,min=8,max=256"`
}
