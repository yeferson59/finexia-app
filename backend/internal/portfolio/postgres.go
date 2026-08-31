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
	return new(PostgresRepository{db})
}

// currencyOf resolves an ISO 4217 code stored on a row into the money package's
// Currency, so amounts read back from Postgres carry the currency they were
// actually settled in. Codes gofinance does not know — and rows written before
// the currency columns existed — fall back to USD, which is what every amount
// was tagged with unconditionally before.
func currencyOf(code string) money.Currency {
	cur, err := money.GetCurrencyFromISOCode(code)
	if err != nil {
		return money.USD
	}

	return cur
}

// moneyOf parses a numeric column into Money tagged with the row's own
// currency. It panics on a malformed value, like the direct
// money.MustMoneyFromString calls it replaces: a numeric column that does not
// parse is a broken database, not a user error.
func moneyOf(value, code string) money.Money {
	return money.MustMoneyFromString(value, currencyOf(code))
}

// retagCurrency rewrites m in place so it carries currency `code` instead of
// the zero value money.Money.Scan leaves behind. Scan only decodes the numeric
// value; without this the amount serializes with currency "XXX", which the
// browser renders as "¤".
func retagCurrency(m *money.Money, code string) {
	if m == nil {
		return
	}

	m.SetCurrency(currencyOf(code))
}

// scanAssetCurrentPrice populates asset.CurrentPrice from a nullable numeric string
// using the asset's own currency. money.Money.Scan only stores the value and leaves
// the currency at the zero value (XXX), which serializes to "¤" in the browser.
func scanAssetCurrentPrice(asset *market.Asset, priceStr *string) {
	if priceStr == nil {
		return
	}

	m, err := money.NewMoneyFromString(*priceStr, asset.Currency)
	if err != nil {
		return
	}

	*asset.CurrentPrice = m
}
