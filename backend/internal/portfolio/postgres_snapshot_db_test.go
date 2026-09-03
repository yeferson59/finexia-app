package portfolio

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// The growth series is the one piece of this package whose correctness lives in
// SQL, and the fake repository the rest of the tests run on cannot see it. The
// bug this file guards against — every metric on the reports page inflated
// four-fold — passed the whole suite without a murmur.
//
// It needs a database with the migrations applied:
//
//	TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/finexia go test ./internal/portfolio/
//
// Without that variable it skips, so `go test ./...` stays a no-setup command.

func growthTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL no está definida: se omite la prueba contra Postgres")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool
}

// backdatedPortfolio plants an account that starts on 2026-06-28 and then loads
// two positions bought months earlier, each registered after that day's
// snapshot job had already run — which is how anyone moves an existing
// portfolio into the app.
//
// The market gains 0.25% a day on whatever is held, so the series is worth
// 1.0025^55 − 1 = +14.72% and nothing else.
func backdatedPortfolio(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	userID, portfolioID := uuid.New(), uuid.New()
	start := time.Date(2026, time.June, 28, 0, 0, 0, 0, time.UTC)

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
	      VALUES ($1, 'growth probe', $2, (SELECT id FROM roles WHERE name = 'customer'), 'USD')`,
		userID, userID.String()+"@probe.test")

	sourceID := uuid.New()
	exec(`INSERT INTO investment_sources (id, user_id, name, source_type)
	      VALUES ($1, $2, 'probe', 'broker')`, sourceID, userID)

	exec(`INSERT INTO portfolios (id, user_id, name, type, risk_id, base_currency)
	      VALUES ($1, $2, 'probe', 'stocks', (SELECT id FROM risks LIMIT 1), 'USD')`,
		portfolioID, userID)

	// cost, the day the app first sees it, and the trade date the user declares.
	positions := []struct {
		cost      float64
		knownOn   int
		tradeDate time.Time
	}{
		{300.00, 0, start},
		{700.00, 13, time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC)},
		{395.23, 39, time.Date(2026, time.May, 2, 0, 0, 0, 0, time.UTC)},
	}

	for i, position := range positions {
		assetID, entryID := uuid.New(), uuid.New()
		exec(`INSERT INTO assets (id, ticker, name, asset_type, currency)
		      VALUES ($1, $2, 'probe', 'stock', 'USD')`, assetID, uuid.New().String()[:8])
		exec(`INSERT INTO portfolio_entries
		        (id, portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
		      VALUES ($1, $2, $3, $4, 1, $5, 'USD', $6)`,
			entryID, portfolioID, assetID, sourceID, position.cost, position.tradeDate)

		// Registered at 10:00, after the 03:00 snapshot of the previous day: the
		// value only moves on the following morning's snapshot.
		recordedAt := start.AddDate(0, 0, position.knownOn-1).Add(10 * time.Hour)
		if i == 0 {
			recordedAt = start.AddDate(0, 0, -1).Add(10 * time.Hour)
		}
		exec(`INSERT INTO transactions
		        (entry_id, type, quantity, price, currency, fees, transaction_date, created_at)
		      VALUES ($1, 'buy', 1, $2, 'USD', 0, $3, $4)`,
			entryID, position.cost, position.tradeDate, recordedAt)
	}

	for day := 0; day <= 55; day++ {
		value, cost := 0.0, 0.0
		for _, position := range positions {
			if day < position.knownOn {
				continue
			}
			held := 1.0
			for range day - position.knownOn {
				held *= 1.0025
			}
			value += position.cost * held
			cost += position.cost
		}

		date := start.AddDate(0, 0, day)
		exec(`INSERT INTO portfolio_snapshots
		        (portfolio_id, snapshot_date, total_value, currency, total_gain_loss,
		         total_gain_loss_pct, created_at)
		      VALUES ($1, $2, $3, 'USD', $4, $5, $6)`,
			portfolioID, date, value, value-cost, (value-cost)/cost*100, date.Add(3*time.Hour))
	}

	return userID
}

func TestGrowthSeriesDoesNotCountABackdatedPurchaseAsAWindfall(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID := backdatedPortfolio(t, pool)

	points, err := repo.GetPortfolioGrowthByUserID(context.Background(), userID, money.USD, false, time.Time{})
	if err != nil {
		t.Fatalf("GetPortfolioGrowthByUserID: %v", err)
	}
	if len(points) != 56 {
		t.Fatalf("points = %d, want 56", len(points))
	}

	metrics := BuildGrowthMetrics(points)
	total, err := metrics.TotalReturn.Mul(oneHundred).Float64()
	if err != nil {
		t.Fatalf("total return: %v", err)
	}

	// The market earned 14.72%. Attributing the flows to the trade dates instead
	// reported 411%, because the money landed on snapshots written before the
	// positions existed and the days they did appear had nothing to net out.
	if total < 13 || total > 16 {
		t.Errorf("rentabilidad del periodo = %.1f%%, want ~14.7%%: los aportes se están contando como rentabilidad", total)
	}

	// And no single day may carry a jump only a deposit could explain.
	for _, sub := range metrics.Subperiod {
		rate, err := sub.Rate.Mul(oneHundred).Float64()
		if err != nil {
			t.Fatalf("subperiod rate: %v", err)
		}
		if rate > 5 || rate < -5 {
			t.Errorf("%s rindió %.1f%% en un día", sub.Date.Format("2006-01-02"), rate)
		}
	}
}

// Omitting the currency means "the account's preferred one" (docs/API.md
// §2.7), and only the SQL can prove it: money.XXX used to reach the query as
// the literal "XXX", which survived the NULLIF and became the target currency
// of the whole series. Nothing converted, every amount kept its nominal value,
// and the reports page labelled the lot with the generic currency sign.
func TestGrowthSeriesWithoutACurrencyUsesTheAccountPreferredOne(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID := backdatedPortfolio(t, pool)

	points, err := repo.GetPortfolioGrowthByUserID(context.Background(), userID, money.XXX, false, time.Time{})
	if err != nil {
		t.Fatalf("GetPortfolioGrowthByUserID: %v", err)
	}
	if len(points) == 0 {
		t.Fatal("the series came back empty")
	}

	for _, point := range points {
		if point.Currency != money.USD {
			t.Fatalf("%s came back in %v, want USD (the account's preferred currency)",
				point.Date.Format("2006-01-02"), point.Currency)
		}
	}
}

func TestGrowthSeriesFlowLandsOnTheDayTheValueMoves(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID := backdatedPortfolio(t, pool)

	points, err := repo.GetPortfolioGrowthByUserID(context.Background(), userID, money.USD, false, time.Time{})
	if err != nil {
		t.Fatalf("GetPortfolioGrowthByUserID: %v", err)
	}

	// The 700 position is registered on the 10th and first snapshotted on the
	// 11th; its flow has to be on the 11th, not on its March trade date.
	for _, point := range points {
		flow := growthDecimal(point.NetFlow)
		if flow.IsZero() {
			continue
		}

		date := point.Date.Format("2006-01-02")
		amount, _ := flow.Float64()
		switch date {
		case "2026-06-28":
			// The opening position, on the point no subperiod measures.
		case "2026-07-11":
			if amount < 699 || amount > 701 {
				t.Errorf("flujo del 11 Jul = %.2f, want 700", amount)
			}
		case "2026-08-06":
			if amount < 394 || amount > 396 {
				t.Errorf("flujo del 6 Ago = %.2f, want 395.23", amount)
			}
		default:
			t.Errorf("flujo inesperado de %.2f el %s", amount, date)
		}
	}
}

// mixedPortfolio plants one portfolio holding three asset types at prices that
// come from all three sources the value expression falls back through: the
// user's own override, the catalog's price, and — for the bond, priced
// nowhere — the entry's cost.
func mixedPortfolio(t *testing.T, pool *pgxpool.Pool) (userID, portfolioID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	userID, portfolioID = uuid.New(), uuid.New()

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
	      VALUES ($1, 'allocation probe', $2, (SELECT id FROM roles WHERE name = 'customer'), 'USD')`,
		userID, userID.String()+"@alloc.test")

	sourceID := uuid.New()
	exec(`INSERT INTO investment_sources (id, user_id, name, source_type)
	      VALUES ($1, $2, 'probe', 'broker')`, sourceID, userID)

	exec(`INSERT INTO portfolios (id, user_id, name, type, risk_id, base_currency)
	      VALUES ($1, $2, 'probe', 'diversified', (SELECT id FROM risks LIMIT 1), 'USD')`,
		portfolioID, userID)

	positions := []struct {
		assetType    string
		quantity     float64
		cost         float64
		currentPrice *float64
		ownPrice     *float64
	}{
		{assetType: "stock", quantity: 10, cost: 100, currentPrice: ptr(151.91)},
		{assetType: "etf", quantity: 4, cost: 60, currentPrice: ptr(70.00), ownPrice: ptr(77.35)},
		{assetType: "bond", quantity: 2, cost: 10.105},
	}

	for _, position := range positions {
		assetID, entryID := uuid.New(), uuid.New()
		exec(`INSERT INTO assets (id, ticker, name, asset_type, currency, current_price)
		      VALUES ($1, $2, 'probe', $3, 'USD', $4)`,
			assetID, uuid.New().String()[:8], position.assetType, position.currentPrice)

		if position.ownPrice != nil {
			exec(`INSERT INTO user_asset_prices (user_id, asset_id, price)
			      VALUES ($1, $2, $3)`, userID, assetID, *position.ownPrice)
		}

		exec(`INSERT INTO portfolio_entries
		        (id, portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
		      VALUES ($1, $2, $3, $4, $5, $6, 'USD', now())`,
			entryID, portfolioID, assetID, sourceID, position.quantity, position.cost)
	}

	return userID, portfolioID
}

func ptr(v float64) *float64 { return &v }

// The allocation a snapshot stores has to be a breakdown *of the total stored
// beside it*. It is aggregated by a second expression, so nothing but a test
// against real rows stops the two from drifting into a row that contradicts
// itself — which is what the column did for its whole life as a literal '{}'.
func TestSnapshotAllocationAddsUpToTheStoredTotal(t *testing.T) {
	pool := growthTestPool(t)
	ctx := context.Background()

	_, portfolioID := mixedPortfolio(t, pool)
	repo := &PostgresRepository{db: pool}

	rows, err := repo.GetAllPortfolioSummaryRows(ctx)
	if err != nil {
		t.Fatalf("GetAllPortfolioSummaryRows: %v", err)
	}

	var row SnapshotRow
	for _, candidate := range rows {
		if candidate.PortfolioID == portfolioID {
			row = candidate
			break
		}
	}
	if row.PortfolioID != portfolioID {
		t.Fatalf("el portafolio de prueba no aparece en las filas del job")
	}

	allocation := map[string]string{}
	if err := json.Unmarshal([]byte(row.Allocation), &allocation); err != nil {
		t.Fatalf("allocation no es un objeto JSON (%q): %v", row.Allocation, err)
	}

	// Los tres precios salen de las tres ramas del COALESCE: override propio,
	// precio del catálogo y coste de la entrada.
	want := map[string]string{
		"stock": "1519.11", // 10 × 151.91, precio del catálogo
		"etf":   "309.40",  // 4 × 77.35, precio propio del usuario
		"bond":  "20.21",   // 2 × 10.105, sin precio: se lleva a coste
	}
	if len(allocation) != len(want) {
		t.Fatalf("allocation = %v, quería %d clases", allocation, len(want))
	}

	total := decimal.Zero
	for assetType, wantAmount := range want {
		got, ok := allocation[assetType]
		if !ok {
			t.Fatalf("falta la clase %q en %v", assetType, allocation)
		}
		gotDec, err := decimal.NewFromString(got)
		if err != nil {
			t.Fatalf("importe de %q no parseable (%q): %v", assetType, got, err)
		}
		wantDec, _ := decimal.NewFromString(wantAmount)
		if !gotDec.Equal(wantDec) {
			t.Errorf("%s = %s, quería %s", assetType, gotDec, wantDec)
		}
		total = total.Add(gotDec)
	}

	// La propiedad que importa: las partes son el todo que se guarda al lado.
	storedTotal, err := decimal.NewFromString(row.TotalMarketValue)
	if err != nil {
		t.Fatalf("total_market_value no parseable (%q): %v", row.TotalMarketValue, err)
	}
	if !total.Equal(storedTotal) {
		t.Errorf("las porciones suman %s pero el snapshot guarda %s", total, storedTotal)
	}
}

// Un portafolio sin posiciones guarda un objeto vacío, no NULL: la columna es
// NOT NULL y quien lo lea espera un objeto en todos los casos.
func TestSnapshotAllocationOfAnEmptyPortfolioIsAnEmptyObject(t *testing.T) {
	pool := growthTestPool(t)
	ctx := context.Background()

	userID := uuid.New()
	portfolioID := uuid.New()
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})

	mustExec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}
	mustExec(`INSERT INTO users (id, name, email, role_id, preferred_currency)
	          VALUES ($1, 'empty probe', $2, (SELECT id FROM roles WHERE name = 'customer'), 'USD')`,
		userID, userID.String()+"@empty.test")
	mustExec(`INSERT INTO portfolios (id, user_id, name, type, risk_id, base_currency)
	          VALUES ($1, $2, 'vacío', 'stocks', (SELECT id FROM risks LIMIT 1), 'USD')`,
		portfolioID, userID)

	repo := &PostgresRepository{db: pool}
	rows, err := repo.GetAllPortfolioSummaryRows(ctx)
	if err != nil {
		t.Fatalf("GetAllPortfolioSummaryRows: %v", err)
	}

	for _, row := range rows {
		if row.PortfolioID != portfolioID {
			continue
		}
		if row.Allocation != "{}" {
			t.Fatalf("allocation de un portafolio vacío = %q, quería {}", row.Allocation)
		}

		// Y sobrevive el viaje de ida y vuelta a la columna jsonb.
		if err := repo.UpsertPortfolioSnapshot(ctx, row, time.Now().UTC().Truncate(24*time.Hour)); err != nil {
			t.Fatalf("UpsertPortfolioSnapshot: %v", err)
		}

		var stored string
		if err := pool.QueryRow(ctx,
			`SELECT allocation::text FROM portfolio_snapshots WHERE portfolio_id = $1`,
			portfolioID,
		).Scan(&stored); err != nil {
			t.Fatalf("releyendo el snapshot: %v", err)
		}
		if stored != "{}" {
			t.Fatalf("allocation persistida = %q, quería {}", stored)
		}

		return
	}

	t.Fatalf("el portafolio vacío no aparece en las filas del job")
}
