package portfolio

import (
	"context"
	"testing"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/gofinance/v2/money"
)

// The consolidated asset list lives entirely in SQL — the grouping, the price
// pick, the conversion and the units — so the fake repository the rest of the
// suite runs on cannot see any of it. It needs the same database the growth
// tests do, and skips the same way without TEST_DATABASE_URL:
//
//	TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/finexia go test ./internal/portfolio/

// heldAcrossPortfolios plants an account holding the same asset in two
// portfolios plus, one at a time, each case the query has to get right: an
// asset quoted in another currency with a rate, one without a rate, one with no
// price at all, and a position sold down to nothing.
func heldAcrossPortfolios(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	userID := uuid.New()
	portfolioA, portfolioB := uuid.New(), uuid.New()

	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})

	exec(`INSERT INTO users (id, name, email, role_id, preferred_currency)
	      VALUES ($1, 'holdings probe', $2, (SELECT id FROM roles WHERE name = 'customer'), 'USD')`,
		userID, userID.String()+"@probe.test")

	sourceID := uuid.New()
	exec(`INSERT INTO investment_sources (id, user_id, name, source_type)
	      VALUES ($1, $2, 'probe', 'broker')`, sourceID, userID)

	for _, p := range []struct {
		id       uuid.UUID
		name     string
		currency string
	}{{portfolioA, "probe A", "USD"}, {portfolioB, "probe B", "EUR"}} {
		exec(`INSERT INTO portfolios (id, user_id, name, type, risk_id, base_currency)
		      VALUES ($1, $2, $3, 'stocks', (SELECT id FROM risks LIMIT 1), $4)`,
			p.id, userID, p.name, p.currency)
	}

	// EUR converts at 1.20; JPY has no rate at all, in either direction or
	// through USD, which is what makes the last asset unconvertible.
	exec(`INSERT INTO exchange_rates (from_currency, to_currency, rate, rate_date)
	      VALUES ('EUR', 'USD', 1.20, CURRENT_DATE)
	      ON CONFLICT (from_currency, to_currency, rate_date) DO UPDATE SET rate = 1.20`)

	asset := func(ticker, currency string, price any) uuid.UUID {
		t.Helper()
		id := uuid.New()
		exec(`INSERT INTO assets (id, ticker, name, asset_type, currency, current_price)
		      VALUES ($1, $2, $3, 'stock', $4, $5)`, id, ticker+uuid.New().String()[:6], ticker, currency, price)

		return id
	}

	entry := func(portfolioID, assetID uuid.UUID, quantity, price float64) {
		t.Helper()
		exec(`INSERT INTO portfolio_entries
		        (portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
		      VALUES ($1, $2, $3, $4, $5, 'USD', $6)`,
			portfolioID, assetID, sourceID, quantity, price, time.Now())
	}

	priced := asset("PRICED", "USD", 200)
	entry(portfolioA, priced, 10, 150)
	entry(portfolioB, priced, 5, 180)

	inEUR := asset("INEUR", "EUR", 50)
	entry(portfolioA, inEUR, 4, 40)

	atCost := asset("ATCOST", "USD", nil)
	entry(portfolioA, atCost, 2, 100)

	noRate := asset("NORATE", "JPY", 10)
	entry(portfolioA, noRate, 3, 9)

	soldOut := asset("SOLDOUT", "USD", 500)
	entry(portfolioA, soldOut, 0, 400)

	return userID
}

// worth compares an amount against a round figure on the decimal engine: the
// query returns text at eight decimals, and asserting on that string would be
// asserting on Postgres's formatting rather than on the value.
func worth(t *testing.T, label, amount string, want int) {
	t.Helper()
	if amountOf(amount).Cmp(decimalFromInt(want)) != 0 {
		t.Errorf("%s market value = %q, want %d", label, amount, want)
	}
}

func TestAssetHoldingsTotalTheSameAssetAcrossPortfolios(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID := heldAcrossPortfolios(t, pool)

	holdings, err := repo.GetAssetHoldingsByUserID(context.Background(), userID, money.USD)
	if err != nil {
		t.Fatalf("GetAssetHoldingsByUserID: %v", err)
	}

	byName := make(map[string]AssetHolding, len(holdings))
	for _, h := range holdings {
		byName[h.Name] = h
	}

	// A position sold in full is not something the user holds; it must not
	// reach the list as a zero-unit row.
	if _, listed := byName["SOLDOUT"]; listed {
		t.Error("a fully sold position is listed as a holding")
	}
	if len(holdings) != 4 {
		t.Fatalf("holdings = %d, want 4: %+v", len(holdings), holdings)
	}

	// The whole point of the view: 10 units in one portfolio and 5 in another
	// are one row of 15, valued at the asset's price.
	priced := byName["PRICED"]
	if amountOf(priced.Quantity).Cmp(decimalFromInt(15)) != 0 {
		t.Errorf("PRICED quantity = %q, want 15 (10 in one portfolio, 5 in the other)", priced.Quantity)
	}
	if priced.Portfolios != 2 {
		t.Errorf("PRICED portfolios = %d, want 2", priced.Portfolios)
	}
	// 15 × 200.
	worth(t, "PRICED", priced.MarketValue, 3000)
	if priced.PriceSource != PriceSourceManual {
		t.Errorf("PRICED price source = %q, want manual", priced.PriceSource)
	}

	// Priced in euros, reported in dollars: the price stays in the asset's own
	// currency and only the total is converted.
	inEUR := byName["INEUR"]
	// 4 × 50 × 1.20.
	worth(t, "INEUR", inEUR.MarketValue, 240)
	if inEUR.Currency != money.EUR {
		t.Errorf("INEUR currency = %v, want EUR (the asset's own)", inEUR.Currency)
	}
	if inEUR.DisplayCurrency != money.USD {
		t.Errorf("INEUR display currency = %v, want USD", inEUR.DisplayCurrency)
	}
	if inEUR.PositionsUnconverted != 0 {
		t.Errorf("INEUR unconverted = %d, want 0: the rate exists", inEUR.PositionsUnconverted)
	}

	// No price anywhere: the position is carried at what it cost, and no single
	// number stands for the asset's price.
	atCost := byName["ATCOST"]
	if atCost.PriceSource != PriceSourceCost {
		t.Errorf("ATCOST price source = %q, want cost", atCost.PriceSource)
	}
	if atCost.MarketPrice != "" {
		t.Errorf("ATCOST market price = %q, want empty", atCost.MarketPrice)
	}
	// 2 × 100, its cost.
	worth(t, "ATCOST", atCost.MarketValue, 200)

	// No rate in any direction: counted at face value and flagged, the same
	// choice the summary and the allocation make.
	noRate := byName["NORATE"]
	if noRate.PositionsUnconverted != 1 {
		t.Errorf("NORATE unconverted = %d, want 1", noRate.PositionsUnconverted)
	}
	// 3 × 10, unconverted.
	worth(t, "NORATE", noRate.MarketValue, 30)

	// Ordered by what they are worth, so the list opens on what matters most
	// and the pie chart's first slices are its first rows.
	for i := 1; i < len(holdings); i++ {
		if amountOf(holdings[i-1].MarketValue).Cmp(amountOf(holdings[i].MarketValue)) < 0 {
			t.Errorf("holding %d (%s) is worth less than the one after it", i-1, holdings[i-1].Name)
		}
	}
}

// Omitting the currency means "the account's preferred one", the same contract
// the summary and the allocation keep. Only the SQL can prove it: money.XXX
// reaching the query as the literal "XXX" survives the NULLIF and becomes the
// target currency, converting nothing and labelling everything ¤.
func TestAssetHoldingsWithoutACurrencyUseTheAccountPreferredOne(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID := heldAcrossPortfolios(t, pool)

	holdings, err := repo.GetAssetHoldingsByUserID(context.Background(), userID, money.XXX)
	if err != nil {
		t.Fatalf("GetAssetHoldingsByUserID: %v", err)
	}
	if len(holdings) == 0 {
		t.Fatal("the list came back empty")
	}

	for _, h := range holdings {
		if h.DisplayCurrency != money.USD {
			t.Fatalf("%s came back in %v, want USD (the account's preferred currency)", h.Name, h.DisplayCurrency)
		}
	}
}
