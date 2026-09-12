package portfolio

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// Restating a position in another cost currency rewrites what it cost and
// nothing it held: the rate changes, the average cost follows it through
// trg_recalculate_avg_cost, a refused request changes nothing, and the growth
// series retires nothing, because no quantity moved. IBCZ.DE, a EUR ETF recorded
// as costing in EUR after being bought from a USD account, is the case.
func TestChangeEntrySettlementRepricesThePositionInPlace(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	userID, portfolioID, sourceID := uuid.New(), uuid.New(), uuid.New()
	assetID, entryID, txnID := uuid.New(), uuid.New(), uuid.New()
	track := dropFixture(t, pool, userID)

	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}

	exec(`INSERT INTO users (id, name, email, role_id, preferred_currency)
	      VALUES ($1, 'settlement probe', $2, (SELECT id FROM roles WHERE name = 'customer'), 'USD')`,
		userID, userID.String()+"@settlement.test")
	exec(`INSERT INTO investment_sources (id, user_id, name, source_type)
	      VALUES ($1, $2, 'probe', 'broker')`, sourceID, userID)
	exec(`INSERT INTO portfolios (id, user_id, name, type, risk_id, base_currency)
	      VALUES ($1, $2, 'probe', 'etfs', (SELECT id FROM risks LIMIT 1), 'USD')`,
		portfolioID, userID)
	exec(`INSERT INTO assets (id, ticker, name, asset_type, currency, current_price)
	      VALUES ($1, $2, 'probe', 'etf', 'EUR', 14.5)`, assetID, uuid.New().String()[:8])
	track(assetID)
	exec(`INSERT INTO portfolio_entries
	        (id, portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
	      VALUES ($1, $2, $3, $4, 0, 14.323, 'EUR', '2026-08-18')`,
		entryID, portfolioID, assetID, sourceID)
	exec(`INSERT INTO transactions
	        (id, entry_id, type, quantity, price, currency, fees, transaction_date, created_at)
	      VALUES ($1, $2, 'buy', 1.7986, 14.323, 'EUR', 0, '2026-08-18', $3)`,
		txnID, entryID, time.Date(2026, time.August, 22, 0, 9, 0, 0, time.UTC))
	// Already in a snapshot, so a rewrite that moved a quantity would retire it.
	exec(`INSERT INTO portfolio_snapshots
	        (portfolio_id, snapshot_date, total_value, currency, total_gain_loss, total_gain_loss_pct, created_at)
	      VALUES ($1, '2026-08-22', 30, 'USD', 0, 0, $2)`,
		portfolioID, time.Date(2026, time.August, 22, 22, 0, 0, 0, time.UTC))

	position := func() (costCurrency string, price, rate float64) {
		t.Helper()

		var rawPrice, rawRate string
		if err := pool.QueryRow(ctx, `
			SELECT pe.cost_currency, pe.price::text, t.fx_rate::text
			FROM portfolio_entries pe
			JOIN transactions t ON t.entry_id = pe.id
			WHERE pe.id = $1
		`, entryID).Scan(&costCurrency, &rawPrice, &rawRate); err != nil {
			t.Fatalf("reading the position: %v", err)
		}
		price, _ = strconv.ParseFloat(rawPrice, 64)
		rate, _ = strconv.ParseFloat(rawRate, 64)

		return strings.TrimSpace(costCurrency), price, rate
	}

	// Without the rate of the day nothing is written.
	if _, err := repo.ChangeEntrySettlement(ctx, userID, entryID, money.USD, nil); !errors.Is(err, ErrTransactionFXRate) {
		t.Fatalf("missing rate: err = %v, want ErrTransactionFXRate", err)
	}
	if currency, _, _ := position(); currency != "EUR" {
		t.Fatalf("a refused change left the position in %s, want EUR", currency)
	}

	// Somebody else's position does not exist.
	if _, err := repo.ChangeEntrySettlement(ctx, uuid.New(), entryID, money.USD, nil); !errors.Is(err, ErrEntryNotFound) {
		t.Fatalf("foreign position: err = %v, want ErrEntryNotFound", err)
	}

	rate, err := decimal.NewFromString("1.1698")
	if err != nil {
		t.Fatalf("rate: %v", err)
	}
	changed, err := repo.ChangeEntrySettlement(ctx, userID, entryID, money.USD, map[uuid.UUID]decimal.Decimal{txnID: rate})
	if err != nil {
		t.Fatalf("ChangeEntrySettlement: %v", err)
	}
	if changed != 1 {
		t.Errorf("changed = %d, want 1", changed)
	}

	currency, price, gotRate := position()
	if currency != "USD" {
		t.Errorf("cost currency = %s, want USD", currency)
	}
	if gotRate < 1.16979 || gotRate > 1.16981 {
		t.Errorf("rate = %v, want 1.1698", gotRate)
	}
	// 14.323 EUR at 1.1698 is what a unit cost the USD account.
	if price < 16.75504 || price > 16.75505 {
		t.Errorf("average cost = %v, want 16.7550454", price)
	}

	if n := retiredCount(t, pool, portfolioID); n != 0 {
		t.Errorf("retired versions = %d, want 0: changing a rate moves no holding", n)
	}
}
