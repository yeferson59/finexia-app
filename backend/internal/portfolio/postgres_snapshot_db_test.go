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

var (
	backdatedStart = time.Date(2026, time.June, 28, 0, 0, 0, 0, time.UTC)
	// backdatedAsOf is the day backdatedPortfolio's series closes on: the one
	// after its last snapshot, read live from the positions' catalog prices.
	backdatedAsOf = backdatedStart.AddDate(0, 0, 56)
)

// backdatedPortfolio plants an account that starts on 2026-06-28 and then loads
// two positions bought months earlier, each registered after that day's
// snapshot job had already run — which is how anyone moves an existing
// portfolio into the app.
//
// The market gains 0.25% a day on whatever is held. Snapshots cover days 0–55
// and the catalog prices each asset where day 56 puts it, so the live point at
// backdatedAsOf keeps the pace: the series is worth 1.0025^56 − 1 = +15.01%,
// less the sliver the half-weighted deposits take, and nothing else.
func backdatedPortfolio(t *testing.T, pool *pgxpool.Pool) (userID, portfolioID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	userID, portfolioID = uuid.New(), uuid.New()
	start := backdatedStart
	days := dayGapUTC(backdatedStart, backdatedAsOf)

	// What a position is worth held days after the app first saw it.
	grown := func(cost float64, held int) float64 {
		for range held {
			cost *= 1.0025
		}
		return cost
	}

	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}

	track := dropFixture(t, pool, userID)

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
		exec(`INSERT INTO assets (id, ticker, name, asset_type, currency, current_price)
		      VALUES ($1, $2, 'probe', 'stock', 'USD', $3)`,
			assetID, uuid.New().String()[:8], grown(position.cost, days-position.knownOn))
		track(assetID)
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
		// fees_currency is left unnamed on purpose: since 000030 the schema fills
		// it from the row's own currency, and a fixture that spelled it out would
		// stop exercising that. recorded_market_price is named for the opposite
		// reason: the schema would fill it with the catalog price of today, and
		// each position here was worth its cost on the day it was recorded.
		exec(`INSERT INTO transactions
		        (entry_id, type, quantity, price, currency, fees, transaction_date, created_at,
		         recorded_market_price)
		      VALUES ($1, 'buy', 1, $2, 'USD', 0, $3, $4, $2)`,
			entryID, position.cost, position.tradeDate, recordedAt)
	}

	for day := range days {
		value, cost := 0.0, 0.0
		for _, position := range positions {
			if day < position.knownOn {
				continue
			}
			value += grown(position.cost, day-position.knownOn)
			cost += position.cost
		}

		date := start.AddDate(0, 0, day)
		exec(`INSERT INTO portfolio_snapshots
		        (portfolio_id, snapshot_date, total_value, currency, total_gain_loss,
		         total_gain_loss_pct, created_at)
		      VALUES ($1, $2, $3, 'USD', $4, $5, $6)`,
			portfolioID, date, value, value-cost, (value-cost)/cost*100, date.Add(3*time.Hour))
	}

	return userID, portfolioID
}

// dayGapUTC counts whole days between two UTC midnights.
func dayGapUTC(from, to time.Time) int {
	return int(to.Sub(from) / (24 * time.Hour))
}

func TestGrowthSeriesDoesNotCountABackdatedPurchaseAsAWindfall(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID, _ := backdatedPortfolio(t, pool)

	points, err := repo.GetPortfolioGrowthByUserID(context.Background(), userID, money.USD, false, time.Time{}, backdatedAsOf)
	if err != nil {
		t.Fatalf("GetPortfolioGrowthByUserID: %v", err)
	}
	if len(points) != 57 {
		t.Fatalf("points = %d, want 57 (56 snapshots and the live point)", len(points))
	}

	metrics := BuildGrowthMetrics(points)
	total, err := metrics.TotalReturn.Mul(oneHundred).Float64()
	if err != nil {
		t.Fatalf("total return: %v", err)
	}

	// The market earned about 15%. Attributing the flows to the trade dates
	// instead reported 411%, because the money landed on snapshots written before
	// the positions existed and the days they did appear had nothing to net out.
	if total < 13 || total > 16 {
		t.Errorf("rentabilidad del periodo = %.1f%%, want ~15%%: los aportes se están contando como rentabilidad", total)
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
	userID, _ := backdatedPortfolio(t, pool)

	points, err := repo.GetPortfolioGrowthByUserID(context.Background(), userID, money.XXX, false, time.Time{}, backdatedAsOf)
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

// The series closes on a live point, not on the last snapshot. The job runs
// once a day, and until it did the dashboard printed today's summary above
// yesterday's series: $1,805.35 as net worth and $1,772.40 as the current value
// right below it. The closing point has to carry the summary's own figures, it
// has to replace a row the job already wrote for that day rather than sit next
// to it, and a purchase no closed day reflects has to land on it as a flow, not
// as return.
func TestGrowthSeriesClosesOnTheLiveSummary(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()
	userID, portfolioID := backdatedPortfolio(t, pool)
	track := dropFixture(t, pool, userID)

	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}

	// A stale row for the closing day, as if the job had already run on it.
	exec(`INSERT INTO portfolio_snapshots
	        (portfolio_id, snapshot_date, total_value, currency, total_gain_loss,
	         total_gain_loss_pct, created_at)
	      VALUES ($1, $2, 1, 'USD', 0, 0, $3)`,
		portfolioID, backdatedAsOf, backdatedAsOf.Add(3*time.Hour))

	// And a purchase registered after it: 50 spent, already worth 55.
	assetID, entryID := uuid.New(), uuid.New()
	exec(`INSERT INTO assets (id, ticker, name, asset_type, currency, current_price)
	      VALUES ($1, $2, 'probe', 'stock', 'USD', 55)`, assetID, uuid.New().String()[:8])
	track(assetID)
	exec(`INSERT INTO portfolio_entries
	        (id, portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
	      VALUES ($1, $2, $3, (SELECT id FROM investment_sources WHERE user_id = $4), 1, 50, 'USD', $5)`,
		entryID, portfolioID, assetID, userID, backdatedAsOf)
	exec(`INSERT INTO transactions
	        (entry_id, type, quantity, price, currency, fees, transaction_date, created_at)
	      VALUES ($1, 'buy', 1, 50, 'USD', 0, $2, $3)`,
		entryID, backdatedAsOf, backdatedAsOf.Add(10*time.Hour))

	// What GET /portfolios/summary reads for this portfolio right now.
	var wantValue, wantGain string
	if err := pool.QueryRow(ctx,
		`SELECT total_market_value::text, total_gain_loss::text FROM portfolio_summary WHERE portfolio_id = $1`,
		portfolioID,
	).Scan(&wantValue, &wantGain); err != nil {
		t.Fatalf("leyendo portfolio_summary: %v", err)
	}

	reads := map[string]func() ([]GrowthPoint, error){
		"account": func() ([]GrowthPoint, error) {
			return repo.GetPortfolioGrowthByUserID(ctx, userID, money.USD, false, time.Time{}, backdatedAsOf)
		},
		"portfolio": func() ([]GrowthPoint, error) {
			return repo.GetPortfolioGrowthByPortfolioID(ctx, userID, portfolioID, false, time.Time{}, backdatedAsOf)
		},
	}

	for name, read := range reads {
		t.Run(name, func(t *testing.T) {
			points, err := read()
			if err != nil {
				t.Fatalf("growth series: %v", err)
			}
			if len(points) != 57 {
				t.Fatalf("points = %d, want 57: the 56 closed days and one live point", len(points))
			}

			last := points[len(points)-1]
			if !last.Date.Equal(backdatedAsOf) {
				t.Fatalf("la serie cierra el %s, want %s", last.Date.Format("2006-01-02"), backdatedAsOf.Format("2006-01-02"))
			}
			if !growthDecimal(last.TotalValue).Equal(growthDecimal(wantValue)) {
				t.Errorf("valor de cierre = %s, want %s (portfolio_summary)", last.TotalValue, wantValue)
			}
			if !growthDecimal(last.GainLoss).Equal(growthDecimal(wantGain)) {
				t.Errorf("ganancia de cierre = %s, want %s (portfolio_summary)", last.GainLoss, wantGain)
			}
			if flow, _ := growthDecimal(last.NetFlow).Float64(); flow < 49.99 || flow > 50.01 {
				t.Errorf("flujo de cierre = %s, want 50: la compra que ningún snapshot refleja", last.NetFlow)
			}
		})
	}
}

// A position loaded with history is a holding walking in, not a purchase made
// while the series watched. Most of what an account brings into the app already
// had gains, and netting each one at its cost turned all of them into the return
// of the day it was typed in: +42% of "real" return on an account worth 15% over
// its cost. Bought at 10 in March, partly sold at 12 in April, paying a dividend
// in May, every trade with a commission, and priced at 15 when registered, the
// holding has to flow in at 6 × 15 — exactly what it adds to the value — and
// leave that day growing at the market's pace.
func TestGrowthSeriesValuesLoadedHistoryAtWhatItWasWorth(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()
	userID, portfolioID := backdatedPortfolio(t, pool)
	track := dropFixture(t, pool, userID)

	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}

	assetID, entryID := uuid.New(), uuid.New()
	exec(`INSERT INTO assets (id, ticker, name, asset_type, currency, current_price)
	      VALUES ($1, $2, 'probe', 'stock', 'USD', 15)`, assetID, uuid.New().String()[:8])
	track(assetID)
	exec(`INSERT INTO portfolio_entries
	        (id, portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
	      VALUES ($1, $2, $3, (SELECT id FROM investment_sources WHERE user_id = $4), 0, 10, 'USD', $5)`,
		entryID, portfolioID, assetID, userID, time.Date(2026, time.March, 2, 0, 0, 0, 0, time.UTC))

	// Typed in the morning after the last snapshot, so all three land on the
	// live point. recorded_market_price is left for the schema to fill from the
	// catalog, which is what the application's own inserts rely on.
	recordedAt := backdatedAsOf.Add(-14 * time.Hour)
	history := []struct {
		kind            string
		quantity, price float64
		tradedOn        time.Time
	}{
		{"buy", 10, 10, time.Date(2026, time.March, 2, 0, 0, 0, 0, time.UTC)},
		{"sell", 4, 12, time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)},
		{"dividend", 1, 5, time.Date(2026, time.May, 4, 0, 0, 0, 0, time.UTC)},
	}
	for _, trade := range history {
		exec(`INSERT INTO transactions
		        (entry_id, type, quantity, price, currency, fees, transaction_date, created_at)
		      VALUES ($1, $2, $3, $4, 'USD', 1, $5, $6)`,
			entryID, trade.kind, trade.quantity, trade.price, trade.tradedOn, recordedAt)
	}

	reads := map[string]func() ([]GrowthPoint, error){
		"account": func() ([]GrowthPoint, error) {
			return repo.GetPortfolioGrowthByUserID(ctx, userID, money.USD, false, time.Time{}, backdatedAsOf)
		},
		"portfolio": func() ([]GrowthPoint, error) {
			return repo.GetPortfolioGrowthByPortfolioID(ctx, userID, portfolioID, false, time.Time{}, backdatedAsOf)
		},
	}

	for name, read := range reads {
		t.Run(name, func(t *testing.T) {
			points, err := read()
			if err != nil {
				t.Fatalf("growth series: %v", err)
			}

			last := points[len(points)-1]
			if flow, _ := growthDecimal(last.NetFlow).Float64(); flow < 89.99 || flow > 90.01 {
				t.Errorf("flujo de cierre = %s, want 90 (6 × 15): la historia cargada entra a lo que valía, no a su coste", last.NetFlow)
			}

			// At cost the day would have netted 50 against the 90 that walked in
			// and reported about +2.6%.
			metrics := BuildGrowthMetrics(points)
			closing := metrics.Subperiod[len(metrics.Subperiod)-1]
			rate, err := closing.Rate.Mul(oneHundred).Float64()
			if err != nil {
				t.Fatalf("closing rate: %v", err)
			}
			if rate < 0 || rate > 0.5 {
				t.Errorf("el %s rindió %.2f%%, want ~0.25%%: la ganancia de antes de la app se cuenta como rentabilidad",
					closing.Date.Format("2006-01-02"), rate)
			}
		})
	}
}

// A second run of the snapshot job on the same day refreshes that day's totals,
// and the time the row carries has to move with them. On 2026-08-14 the job ran
// at 00:55 and again at 22:00, the row kept 00:55, and the positions recorded in
// between showed in the 14th's value while their flows went to the 15th: +10%
// one day and −7% the next, on two days the market barely moved.
func TestSnapshotRerunKeepsFlowsOnTheDayItsValuesShow(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()
	userID, portfolioID := backdatedPortfolio(t, pool)
	track := dropFixture(t, pool, userID)

	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}

	// The day after the fixture's last snapshot is closed here, so the series
	// is read one day later.
	day := backdatedAsOf
	asOf := day.AddDate(0, 0, 1)

	upsert := func(readAt time.Time) {
		t.Helper()
		row := SnapshotRow{
			PortfolioID:      portfolioID,
			BaseCurrency:     money.USD,
			TotalMarketValue: "1600",
			TotalGainLoss:    "150",
			TotalGainLossPct: "10.37",
			ReadAt:           readAt,
		}
		if err := repo.UpsertPortfolioSnapshot(ctx, row, day); err != nil {
			t.Fatalf("UpsertPortfolioSnapshot: %v", err)
		}
	}

	// The catch-up of a restart, a little before one in the morning...
	upsert(day.Add(55 * time.Minute))

	// ...a purchase recorded an hour later, traded the evening before...
	assetID, entryID := uuid.New(), uuid.New()
	exec(`INSERT INTO assets (id, ticker, name, asset_type, currency, current_price)
	      VALUES ($1, $2, 'probe', 'stock', 'USD', 55)`, assetID, uuid.New().String()[:8])
	track(assetID)
	exec(`INSERT INTO portfolio_entries
	        (id, portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
	      VALUES ($1, $2, $3, (SELECT id FROM investment_sources WHERE user_id = $4), 0, 50, 'USD', $5)`,
		entryID, portfolioID, assetID, userID, day.AddDate(0, 0, -1))
	exec(`INSERT INTO transactions
	        (entry_id, type, quantity, price, currency, fees, transaction_date, created_at)
	      VALUES ($1, 'buy', 1, 50, 'USD', 0, $2, $3)`,
		entryID, day.AddDate(0, 0, -1), day.Add(101*time.Minute))

	// ...and the regular run at 22:00, whose totals include it.
	upsert(day.Add(22 * time.Hour))

	points, err := repo.GetPortfolioGrowthByUserID(ctx, userID, money.USD, false, time.Time{}, asOf)
	if err != nil {
		t.Fatalf("GetPortfolioGrowthByUserID: %v", err)
	}

	flows := map[string]float64{}
	for _, point := range points {
		flows[point.Date.Format("2006-01-02")], _ = growthDecimal(point.NetFlow).Float64()
	}

	// On the day whose value holds it, and at its cost: it was traded inside
	// the stretch that day closes, not loaded as history.
	if got := flows[day.Format("2006-01-02")]; got < 49.99 || got > 50.01 {
		t.Errorf("flujo del %s = %.2f, want 50: la compra ya está en el valor de ese día", day.Format("2006-01-02"), got)
	}
	if got := flows[asOf.Format("2006-01-02")]; got != 0 {
		t.Errorf("flujo del %s = %.2f, want 0: la compra se contó un día tarde", asOf.Format("2006-01-02"), got)
	}
}

// seenPurchase plants, on top of backdatedPortfolio, a purchase recorded the
// night after the fixture's last snapshot and held by the next one: one unit
// bought at 50 and worth 55. It hands back what the tests take back — the
// position and its transaction — and the two days they read: the day the
// purchase is in, and the live point after it.
func seenPurchase(t *testing.T, pool *pgxpool.Pool, repo *PostgresRepository) (userID, portfolioID, entryID, txID uuid.UUID, day, asOf time.Time) {
	t.Helper()
	ctx := context.Background()

	userID, portfolioID = backdatedPortfolio(t, pool)
	track := dropFixture(t, pool, userID)

	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}

	day = backdatedAsOf
	asOf = day.AddDate(0, 0, 1)

	assetID := uuid.New()
	entryID, txID = uuid.New(), uuid.New()
	exec(`INSERT INTO assets (id, ticker, name, asset_type, currency, current_price)
	      VALUES ($1, $2, 'probe', 'stock', 'USD', 55)`, assetID, uuid.New().String()[:8])
	track(assetID)
	exec(`INSERT INTO portfolio_entries
	        (id, portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
	      VALUES ($1, $2, $3, (SELECT id FROM investment_sources WHERE user_id = $4), 0, 50, 'USD', $5)`,
		entryID, portfolioID, assetID, userID, day.AddDate(0, 0, -1))
	exec(`INSERT INTO transactions
	        (id, entry_id, type, quantity, price, currency, fees, transaction_date, created_at)
	      VALUES ($1, $2, 'buy', 1, 50, 'USD', 0, $3, $4)`,
		txID, entryID, day.AddDate(0, 0, -1), day.Add(time.Hour))

	row := SnapshotRow{
		PortfolioID:      portfolioID,
		BaseCurrency:     money.USD,
		TotalMarketValue: "1600",
		TotalGainLoss:    "150",
		TotalGainLossPct: "10.37",
		ReadAt:           day.Add(22 * time.Hour),
	}
	if err := repo.UpsertPortfolioSnapshot(ctx, row, day); err != nil {
		t.Fatalf("UpsertPortfolioSnapshot: %v", err)
	}

	return userID, portfolioID, entryID, txID, day, asOf
}

// flowsByDay reads the account-wide series closing on asOf and indexes its net
// flows by date.
func flowsByDay(t *testing.T, repo *PostgresRepository, userID uuid.UUID, asOf time.Time) map[time.Time]float64 {
	t.Helper()

	points, err := repo.GetPortfolioGrowthByUserID(context.Background(), userID, money.USD, false, time.Time{}, asOf)
	if err != nil {
		t.Fatalf("GetPortfolioGrowthByUserID: %v", err)
	}

	flows := make(map[time.Time]float64, len(points))
	for _, point := range points {
		flows[point.Date.UTC()], _ = growthDecimal(point.NetFlow).Float64()
	}

	return flows
}

func retiredCount(t *testing.T, pool *pgxpool.Pool, portfolioID uuid.UUID) int {
	t.Helper()

	var n int
	if err := pool.QueryRow(context.Background(),
		`SELECT count(*) FROM retired_transactions WHERE portfolio_id = $1`, portfolioID,
	).Scan(&n); err != nil {
		t.Fatalf("counting retired transactions: %v", err)
	}

	return n
}

func assertFlow(t *testing.T, flows map[time.Time]float64, day time.Time, want float64, why string) {
	t.Helper()

	if got := flows[day]; got < want-0.01 || got > want+0.01 {
		t.Errorf("flujo del %s = %.2f, want %.2f: %s", day.Format("2006-01-02"), got, want, why)
	}
}

// A position deleted after a snapshot held it leaves its value behind in that
// snapshot. Dropping its flow with the row turned the day it came in into a
// gain nobody made and the day it went into a loss nobody took: +$30.61 on
// 2026-08-22 and −$29.58 on 2026-09-03 on the account that found it.
func TestGrowthSeriesKeepsADeletedPositionOnTheDaysItWasIn(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()
	userID, portfolioID, entryID, _, day, asOf := seenPurchase(t, pool, repo)

	if _, err := repo.DeletePortfolioEntry(ctx, userID, entryID); err != nil {
		t.Fatalf("DeletePortfolioEntry: %v", err)
	}

	// Kept once: the position's trigger keeps it, and the transaction's own
	// trigger, reached by the cascade, must not keep it again.
	if n := retiredCount(t, pool, portfolioID); n != 1 {
		t.Fatalf("retired versions = %d, want 1", n)
	}

	flows := flowsByDay(t, repo, userID, asOf)
	assertFlow(t, flows, day, 50, "la compra sigue en el valor de ese día")
	assertFlow(t, flows, asOf, -55, "la posición sale a lo que valía")

	// A portfolio deleted with its positions goes as it always did.
	if _, err := pool.Exec(ctx, `DELETE FROM portfolios WHERE id = $1`, portfolioID); err != nil {
		t.Fatalf("deleting the portfolio: %v", err)
	}
}

// A quantity edit is the same problem with the difference: the snapshot held
// one unit, and the row now says two.
func TestGrowthSeriesNetsAQuantityEditOnTheDayItWasMade(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID, portfolioID, _, txID, day, asOf := seenPurchase(t, pool, repo)

	if _, err := pool.Exec(context.Background(),
		`UPDATE transactions SET quantity = 2 WHERE id = $1`, txID); err != nil {
		t.Fatalf("editing the quantity: %v", err)
	}

	if n := retiredCount(t, pool, portfolioID); n != 1 {
		t.Fatalf("retired versions = %d, want 1", n)
	}

	flows := flowsByDay(t, repo, userID, asOf)
	assertFlow(t, flows, day, 50, "la versión de una unidad es la que ese día tenía")
	// One unit out at 55 and two back in at 55.
	assertFlow(t, flows, asOf, 55, "la diferencia entra el día de la edición")
}

// Correcting a price, like correcting a date, never changed any snapshot's
// value, so it is fixed where it happened and nothing is kept.
func TestGrowthSeriesCorrectsAPriceInPlace(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID, portfolioID, _, txID, day, asOf := seenPurchase(t, pool, repo)

	if _, err := pool.Exec(context.Background(),
		`UPDATE transactions SET price = 48 WHERE id = $1`, txID); err != nil {
		t.Fatalf("editing the price: %v", err)
	}

	if n := retiredCount(t, pool, portfolioID); n != 0 {
		t.Fatalf("retired versions = %d, want 0", n)
	}

	flows := flowsByDay(t, repo, userID, asOf)
	assertFlow(t, flows, day, 48, "el precio corregido vale desde el principio")
	assertFlow(t, flows, asOf, 0, "no hubo nada que entrara o saliera")
}

func TestGrowthSeriesFlowLandsOnTheDayTheValueMoves(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID, _ := backdatedPortfolio(t, pool)

	points, err := repo.GetPortfolioGrowthByUserID(context.Background(), userID, money.USD, false, time.Time{}, backdatedAsOf)
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

	track := dropFixture(t, pool, userID)

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
		track(assetID)

		if position.ownPrice != nil {
			// currency and source are NOT NULL since 000018; the fixture had
			// never named them, so this insert failed on every run against a
			// migrated database and the test could only ever pass by skipping.
			exec(`INSERT INTO user_asset_prices (user_id, asset_id, price, currency, source)
			      VALUES ($1, $2, $3, 'USD', 'probe')`, userID, assetID, *position.ownPrice)
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
		"stock": "1519.10", // 10 × 151.91, precio del catálogo
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
	// No entries and no catalog rows here, so this one's user delete already
	// worked; it goes through the same helper so the package has one teardown
	// and not two spellings of it.
	dropFixture(t, pool, userID)

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
