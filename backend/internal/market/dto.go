// Package market owns the asset catalog and exchange-rate domains: the request
// DTOs, entities, persistence, services and HTTP handlers for both.
package market

import (
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

type CreateAssetRequestDTO struct {
	Ticker    string         `json:"ticker"    validate:"required"`
	Name      string         `json:"name"      validate:"required"`
	AssetType string         `json:"assetType" validate:"required"`
	Exchange  string         `json:"exchange"`
	Currency  money.Currency `json:"currency"  validate:"required"`
	// Sector is free text on the way in — "Tecnología", "Financial Services",
	// "technology" all arrive here — and is normalised before it is stored. It
	// is optional, and only honoured for an admin: see CreateAsset in the
	// handler for why a contribution cannot classify a shared row.
	Sector string `json:"sector"`
	// SectorWeights is the alternative to the field above, for the asset one
	// industry cannot describe: a whole-market ETF. Sending both is a 400 — see
	// errAssetSectorBoth — and sending neither leaves the asset unclassified,
	// which is the ordinary case.
	SectorWeights []SectorWeightRequestDTO `json:"sectorWeights"`
}

// SectorWeightRequestDTO is one line of a breakdown as a client sends it.
//
// The sector half is as free as the single sector column beside it, and goes
// through the same normaliser, so a form built from this app's own vocabulary
// and a script pasting "Information Technology" both land on the same value.
//
// The weight is a percentage — 33.1, not 0.331 — and is a decimal rather than a
// float64 because it is a number somebody typed off a fact sheet, and the one
// thing a float64 cannot do is hand it back unchanged. It accepts both the JSON
// number and the quoted form, so a client that keeps its decimals in strings
// needs no special case.
type SectorWeightRequestDTO struct {
	Sector string          `json:"sector"`
	Weight decimal.Decimal `json:"weight"`
}

// normalizeSectorWeights maps a request's breakdown onto the domain's, resolving
// each label the way the single sector field is resolved.
//
// It reports false for a label nobody recognises and for one that is blank: a
// weight with no industry to attach it to is not an omission the way an empty
// sector column is, it is a line of the breakdown that says a number and not
// what the number is about. Arithmetic — the duplicates, the range, the total —
// belongs to SectorBreakdown.Validate and is not repeated here.
func normalizeSectorWeights(rows []SectorWeightRequestDTO) (SectorBreakdown, bool) {
	if len(rows) == 0 {
		return nil, true
	}

	breakdown := make(SectorBreakdown, 0, len(rows))

	for _, row := range rows {
		sector, ok := NormalizeSector(row.Sector)
		if !ok || sector == SectorNone {
			return nil, false
		}

		breakdown = append(breakdown, SectorWeight{Sector: sector, Weight: row.Weight})
	}

	return breakdown, true
}

// UpdateAssetRequestDTO is a catalog row as the operator wants it to read from
// now on: the identifying fields travel whole, so an edit that only changes the
// name still sends the ticker it is keeping.
//
// The last two are pointers because they are the fields whose zero value is a
// legitimate instruction. A body with no isCurated leaves the audience alone,
// and one with `false` hides the row from everybody but its contributors; a
// body with no price leaves the manual price alone, which is not the same as
// clearing it.
type UpdateAssetRequestDTO struct {
	Ticker    string         `json:"ticker"    validate:"required"`
	Name      string         `json:"name"      validate:"required"`
	AssetType string         `json:"assetType" validate:"required"`
	Exchange  string         `json:"exchange"`
	Currency  money.Currency `json:"currency"  validate:"required"`
	// Sector travels whole like the fields above it: an edit that sends it
	// empty is clearing the classification, not omitting it.
	Sector string `json:"sector"`
	// SectorWeights travels whole for the same reason: an edit with no weights
	// removes the breakdown the asset had. It is the field that turns a fund
	// filed under one industry into the eleven it is actually made of, and back.
	SectorWeights []SectorWeightRequestDTO `json:"sectorWeights"`
	IsCurated     *bool                    `json:"isCurated"`
	Price         *money.Money             `json:"price"`
}

type CreateExchangeRateRequestDTO struct {
	FromCurrency money.Currency  `json:"fromCurrency" validate:"required"`
	ToCurrency   money.Currency  `json:"toCurrency"   validate:"required"`
	Rate         decimal.Decimal `json:"rate"         validate:"required"`
}

type UpdateExchangeRateRequestDTO struct {
	Rate decimal.Decimal `json:"rate" validate:"required"`
}

// SaveCredentialRequestDTO carries a user's own provider API key.
//
// It has no response counterpart on purpose: a stored key is never sent back,
// so nothing that leaves this module can carry one.
type SaveCredentialRequestDTO struct {
	APIKey string `json:"apiKey" validate:"required,min=8,max=256"`
}

// RefreshAssetPriceRequestDTO names the key one asset must be re-quoted with.
//
// The provider is required and has no default on purpose: the whole point of
// the request is that the caller chose one, and a body that left it out would
// quietly become a fallback run whose result is attributed to whichever
// provider happened to answer.
type RefreshAssetPriceRequestDTO struct {
	Provider string `json:"provider" validate:"required"`
}

// SyncResultDTO is what POST /market/sync returns: the prices and the rates the
// caller's own keys fetched.
//
// Both halves are named because a sync is not just prices. A holding quoted in
// a currency other than its portfolio's needs a rate to be worth anything, and
// under BYO-key nobody else's rate may be used, so a run that refreshed prices
// but no rates has done half the job — and the response should say which half.
type SyncResultDTO struct {
	Prices []UserAssetPrice   `json:"prices"`
	Rates  []UserExchangeRate `json:"rates"`
}
