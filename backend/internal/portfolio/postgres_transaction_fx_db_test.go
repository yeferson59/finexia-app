package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/gofinance/v2/money"
)

// The average-cost rule lives in a trigger (migration 000029), so the fake
// repository the rest of this package tests against cannot see it. These run
// against a real database with the migrations applied, and skip without one:
//
//	TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/finexia go test ./internal/portfolio/

// fxPosition plants a user, a portfolio funded in USD and one empty position on
// an asset quoted in EUR, and returns what is needed to write trades onto it.
func fxPosition(t *testing.T, pool *pgxpool.Pool) (userID, entryID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	userID, portfolioID, sourceID := uuid.New(), uuid.New(), uuid.New()
	assetID, entryID := uuid.New(), uuid.New()

	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM assets WHERE id = $1`, assetID)
	})

	exec(`INSERT INTO users (id, name, email, role_id, preferred_currency)
	      VALUES ($1, 'fx probe', $2, (SELECT id FROM roles WHERE name = 'customer'), 'USD')`,
		userID, userID.String()+"@probe.test")
	exec(`INSERT INTO investment_sources (id, user_id, name, source_type)
	      VALUES ($1, $2, 'probe', 'broker')`, sourceID, userID)
	exec(`INSERT INTO portfolios (id, user_id, name, type, risk_id, base_currency)
	      VALUES ($1, $2, 'probe', 'stocks', (SELECT id FROM risks LIMIT 1), 'USD')`,
		portfolioID, userID)

	// MC.FR: quoted in EUR, held in an account that settles in USD. The two
	// currencies on the position are the whole point of the fixture.
	exec(`INSERT INTO assets (id, ticker, name, asset_type, exchange, currency)
	      VALUES ($1, $2, 'LVMH Moet Hennessy Louis Vuitton SE', 'stock', 'MC.FR', 'EUR')`,
		assetID, uuid.New().String()[:8])
	exec(`INSERT INTO portfolio_entries
	        (id, portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
	      VALUES ($1, $2, $3, $4, 0, 0, 'USD', '2024-12-05')`,
		entryID, portfolioID, assetID, sourceID)

	return userID, entryID
}

// mustFXEntry is fxPosition reduced to the entry id, for the tests that only
// need something to write against.
func mustFXEntry(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	_, entryID := fxPosition(t, pool)

	return entryID
}

func entryCost(t *testing.T, pool *pgxpool.Pool, entryID uuid.UUID) (quantity, price string) {
	t.Helper()
	if err := pool.QueryRow(context.Background(),
		`SELECT quantity::text, price::text FROM portfolio_entries WHERE id = $1`, entryID,
	).Scan(&quantity, &price); err != nil {
		t.Fatalf("read entry: %v", err)
	}

	return quantity, price
}

// The position from the broker screenshot that prompted the column: 0.0241
// shares filled at 606.60 EUR, converted at 1.0638, debited as 15.55 USD.
//
// What is being checked is that the entry's price — the column every cost basis
// in the app is derived from, and which is labelled cost_currency everywhere —
// comes out in USD rather than in the EUR the price was quoted in.
func TestAvgCostConvertsAtTheTradesOwnRate(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID, entryID := fxPosition(t, pool)

	tradeDate := time.Date(2024, time.December, 5, 0, 0, 0, 0, time.UTC)
	txn, err := repo.CreateTransaction(context.Background(), userID, entryID, TransactionInput{
		Type:            Buy,
		Quantity:        mustDecimal(t, "0.0241"),
		Price:           mustEUR(t, "606.60"),
		Currency:        money.EUR,
		FXRate:          mustDecimal(t, "1.0638"),
		Fees:            mustEUR(t, "0"),
		TransactionDate: tradeDate,
	})
	if err != nil {
		t.Fatalf("CreateTransaction: %v", err)
	}

	// The trade is stored as the broker states it, not pre-multiplied.
	if got := txn.Price.String(); got != "606.60" {
		t.Errorf("stored price = %s, want 606.60 (the quoted price, unconverted)", got)
	}
	if txn.Currency != money.EUR || txn.CostCurrency != money.USD {
		t.Errorf("currencies = %q into %q, want EUR into USD", txn.Currency, txn.CostCurrency)
	}

	quantity, price := entryCost(t, pool, entryID)
	if quantity != "0.02410000" {
		t.Errorf("entry quantity = %s, want 0.02410000", quantity)
	}
	// 606.60 × 1.0638, which is what one share cost the dollar account.
	if price != "645.30108000" {
		t.Errorf("entry price = %s, want 645.30108000 USD", price)
	}
}

// A second buy at a different rate has to average the two converted costs, not
// the two quoted prices: the position was funded in dollars and the dollars are
// what it cost.
func TestAvgCostAveragesConvertedCostsAcrossRates(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID, entryID := fxPosition(t, pool)

	buy := func(qty, unitPrice, rate string, date time.Time) {
		t.Helper()
		if _, err := repo.CreateTransaction(context.Background(), userID, entryID, TransactionInput{
			Type:            Buy,
			Quantity:        mustDecimal(t, qty),
			Price:           mustEUR(t, unitPrice),
			Currency:        money.EUR,
			FXRate:          mustDecimal(t, rate),
			Fees:            mustEUR(t, "0"),
			TransactionDate: date,
		}); err != nil {
			t.Fatalf("CreateTransaction: %v", err)
		}
	}

	// One unit at 100 EUR when a euro bought 1.00 USD, one at 100 EUR when it
	// bought 1.50. Same quoted price both times; the position cost 250 USD.
	buy("1", "100.00", "1.00", time.Date(2025, time.January, 10, 0, 0, 0, 0, time.UTC))
	buy("1", "100.00", "1.50", time.Date(2025, time.June, 10, 0, 0, 0, 0, time.UTC))

	quantity, price := entryCost(t, pool, entryID)
	if quantity != "2.00000000" {
		t.Errorf("entry quantity = %s, want 2.00000000", quantity)
	}
	if price != "125.00000000" {
		t.Errorf("entry price = %s, want 125.00000000 USD (250 / 2), not 100 (the quoted average)", price)
	}
}

// The rule that cannot be a CHECK constraint, enforced where the row is written.
func TestCreateTransactionRefusesAContradictoryRate(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID, entryID := fxPosition(t, pool)

	date := time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC)

	t.Run("a rate other than one between a currency and itself", func(t *testing.T) {
		_, err := repo.CreateTransaction(context.Background(), userID, entryID, TransactionInput{
			Type:            Buy,
			Quantity:        mustDecimal(t, "1"),
			Price:           mustUSD(t, "100.00"),
			Currency:        money.USD,
			FXRate:          mustDecimal(t, "1.0638"),
			Fees:            mustUSD(t, "0"),
			TransactionDate: date,
		})
		if !errors.Is(err, ErrTransactionFXRate) {
			t.Fatalf("err = %v, want ErrTransactionFXRate", err)
		}
	})

	t.Run("a cross-currency trade with no rate at all", func(t *testing.T) {
		_, err := repo.CreateTransaction(context.Background(), userID, entryID, TransactionInput{
			Type:            Buy,
			Quantity:        mustDecimal(t, "1"),
			Price:           mustEUR(t, "606.60"),
			Currency:        money.EUR,
			Fees:            mustEUR(t, "0"),
			TransactionDate: date,
		})
		if !errors.Is(err, ErrTransactionFXRate) {
			t.Fatalf("err = %v, want ErrTransactionFXRate", err)
		}
	})

	// Neither attempt may have left a row behind.
	quantity, _ := entryCost(t, pool, entryID)
	if quantity != "0.00000000" {
		t.Errorf("entry quantity = %s, want 0: a refused transaction must not persist", quantity)
	}
}

// CreatePortfolioEntry upserts: a second call for the same portfolio/asset/source
// adds a trade to the position already there, and that position keeps the
// currency it was opened with. The rate therefore has to be judged against the
// currency the entry actually has, not the one the request asked for — judging
// it against the request would approve a conversion into a currency the column
// does not hold, and the trigger would apply it anyway.
func TestCreatePortfolioEntryValidatesAgainstTheExistingPosition(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	var userID, portfolioID, assetID, sourceID uuid.UUID
	if err := pool.QueryRow(ctx, `
		SELECT p.user_id, pe.portfolio_id, pe.asset_id, pe.source_id
		FROM portfolio_entries pe
		JOIN portfolios p ON p.id = pe.portfolio_id
		WHERE pe.id = $1
	`, mustFXEntry(t, pool)).Scan(&userID, &portfolioID, &assetID, &sourceID); err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	// The fixture's position costs in USD. Asking to open it in EUR at a rate of
	// 1 is a request the existing row contradicts: the trade is in EUR, the
	// position is in USD, and no rate was given for that pair.
	_, err := repo.CreatePortfolioEntry(ctx, userID, portfolioID, assetID, sourceID, money.EUR, TransactionInput{
		Type:            Buy,
		Quantity:        mustDecimal(t, "1"),
		Price:           mustEUR(t, "606.60"),
		Currency:        money.EUR,
		TransactionDate: time.Date(2025, time.February, 2, 0, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, ErrTransactionFXRate) {
		t.Fatalf("err = %v, want ErrTransactionFXRate", err)
	}
}

// Every row written before 000029 has fx_rate = 1 and both currencies equal, so
// the rewritten trigger has to reproduce the old numbers exactly. This is the
// same single-currency position the app has always supported.
func TestSingleCurrencyPositionIsUnchanged(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID, entryID := fxPosition(t, pool)

	if _, err := repo.CreateTransaction(context.Background(), userID, entryID, TransactionInput{
		Type:            Buy,
		Quantity:        mustDecimal(t, "10"),
		Price:           mustUSD(t, "150.50"),
		Currency:        money.USD,
		Fees:            mustUSD(t, "0"),
		TransactionDate: time.Date(2025, time.April, 1, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("CreateTransaction: %v", err)
	}

	quantity, price := entryCost(t, pool, entryID)
	if quantity != "10.00000000" || price != "150.50000000" {
		t.Errorf("quantity/price = %s/%s, want 10.00000000/150.50000000", quantity, price)
	}
}
