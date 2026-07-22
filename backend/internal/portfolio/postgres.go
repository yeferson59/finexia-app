package portfolio

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
)

// PostgresRepository implements Repository over the shared pgx pool. Its
// methods are split by sub-area across the postgres_*.go files.
type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return new(PostgresRepository{db: db})
}

// scanAssetCurrentPrice populates asset.CurrentPrice from a nullable numeric string
// using the asset's own currency. money.Money.Scan only stores the value and leaves
// the currency at the zero value (XXX), which serializes to "¤" in the browser.
func scanAssetCurrentPrice(asset *market.Asset, priceStr *string) {
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
