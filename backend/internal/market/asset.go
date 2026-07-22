package market

import (
	"time"

	"github.com/google/uuid"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/spreadsheet"
)

// Column limits mirror the assets table (ticker VARCHAR(20), name VARCHAR(255)).
const (
	maxTickerLen    = 20
	maxAssetNameLen = 255
)

// AssetType is the kind of a tradable asset.
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

// Asset is a tradable instrument in the catalog. Owned by the market module;
// the portfolio module references it (entries hold an Asset) but does not own
// its lifecycle.
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
}

// scanAssetCurrentPrice attaches the current price (stored as a string) to an
// asset, ignoring malformed values.
func scanAssetCurrentPrice(asset *Asset, priceStr *string) {
	if priceStr == nil {
		return
	}
	cur, err := money.CurrencyFromISOCode(asset.Currency)
	if err != nil {
		return
	}
	m, err := money.NewMoneyFromString(*priceStr, cur)
	if err != nil {
		return
	}
	asset.CurrentPrice = &m
}

var categorySynonyms = map[string]AssetType{
	"stock": Stock, "stocks": Stock, "accion": Stock, "acciones": Stock,
	"equity": Stock, "equities": Stock,
	"etf": ETF, "etfs": ETF, "fondo": ETF, "fondos": ETF,
	"fund": ETF, "funds": ETF, "fondo indexado": ETF, "index fund": ETF,
	"crypto": Crypto, "cripto": Crypto, "criptomoneda": Crypto,
	"criptomonedas": Crypto, "cryptocurrency": Crypto, "criptos": Crypto,
	"bond": Bond, "bonds": Bond, "bono": Bond, "bonos": Bond,
	"renta fija": Bond, "cdt": Bond, "fixed income": Bond,
	"cash": Cash, "efectivo": Cash, "liquidez": Cash, "dinero": Cash,
	"real estate": RealEstate, "real_estate": RealEstate, "inmueble": RealEstate,
	"inmuebles": RealEstate, "bienes raices": RealEstate, "reit": RealEstate, "fibra": RealEstate,
	"commodity": Commodity, "commodities": Commodity, "materia prima": Commodity,
	"materias primas": Commodity, "oro": Commodity, "gold": Commodity, "plata": Commodity,
	"other": Other, "otro": Other, "otros": Other,
}

// NormalizeAssetType maps a free-form category label (accent/case-insensitive)
// to a known AssetType. It is shared with the portfolio importer, which needs
// the same mapping when a transaction file carries a category column.
func NormalizeAssetType(raw string) (AssetType, bool) {
	if c, ok := categorySynonyms[spreadsheet.NormKey(raw)]; ok {
		return c, true
	}
	return "", false
}
