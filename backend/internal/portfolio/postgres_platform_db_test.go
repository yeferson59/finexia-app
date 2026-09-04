package portfolio

import (
	"context"
	"testing"

	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/gofinance/v2/money"
)

// The ordering, the market value and the two-legged conversion all live in SQL,
// so the fake repository cannot see any of them. Skips without TEST_DATABASE_URL
// like the rest of the database-backed tests in this package.

// platformsByWeight plants three platforms of deliberately different sizes, in
// an order that is neither the one they were created in nor alphabetical, so
// the assertion can only pass if the query sorts by what is invested.
func platformsByWeight(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	userID, portfolioID := uuid.New(), uuid.New()

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
	      VALUES ($1, 'platform probe', $2, (SELECT id FROM roles WHERE name = 'customer'), 'USD')`,
		userID, userID.String()+"@probe.test")
	exec(`INSERT INTO portfolios (id, user_id, name, type, risk_id, base_currency)
	      VALUES ($1, $2, 'probe', 'stocks', (SELECT id FROM risks LIMIT 1), 'USD')`,
		portfolioID, userID)

	// Created smallest first and named so that alphabetical order disagrees with
	// both creation order and invested order.
	platforms := []struct {
		name           string
		quantity, cost float64
		currentPrice   any
	}{
		{"zeta", 10, 10, 12.0},  // invested 100, worth 120
		{"alpha", 10, 60, 54.0}, // invested 600, worth 540
		{"mid", 10, 30, 30.0},   // invested 300, worth 300
	}

	for _, p := range platforms {
		sourceID, assetID := uuid.New(), uuid.New()
		exec(`INSERT INTO investment_sources (id, user_id, name, source_type)
		      VALUES ($1, $2, $3, 'broker')`, sourceID, userID, p.name)
		exec(`INSERT INTO assets (id, ticker, name, asset_type, currency, current_price)
		      VALUES ($1, $2, 'probe', 'stock', 'USD', $3)`,
			assetID, uuid.New().String()[:8], p.currentPrice)
		exec(`INSERT INTO portfolio_entries
		        (portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
		      VALUES ($1, $2, $3, $4, $5, 'USD', now())`,
			portfolioID, assetID, sourceID, p.quantity, p.cost)
	}

	return userID
}

func TestPlatformsAreOrderedByWhatIsInvested(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID := platformsByWeight(t, pool)

	platforms, err := repo.GetPlatformsWithStats(context.Background(), userID, money.USD)
	if err != nil {
		t.Fatalf("GetPlatformsWithStats: %v", err)
	}
	if len(platforms) != 3 {
		t.Fatalf("platforms = %d, want 3", len(platforms))
	}

	// Biggest first: neither the creation order (zeta, alpha, mid) nor the
	// alphabetical one (alpha, mid, zeta) produces this.
	wantOrder := []string{"alpha", "mid", "zeta"}
	for i, want := range wantOrder {
		if platforms[i].Name != want {
			t.Errorf("position %d = %q, want %q (order: %v)", i, platforms[i].Name, want, names(platforms))
		}
	}

	dtos := NewPlatformListResponse(platforms)

	// alpha: 600 in, worth 540 — a loss of 60, which is 10% of what went in,
	// and 60% of everything the account has invested.
	alpha := dtos[0]
	if alpha.TotalValue != "600.00000000" || alpha.MarketValue != "540.00000000" {
		t.Errorf("alpha cost/market = %q/%q, want 600/540", alpha.TotalValue, alpha.MarketValue)
	}
	if alpha.GainLoss != "-60" || alpha.GainLossPct != -10 {
		t.Errorf("alpha gain = %q / %v, want -60 / -10", alpha.GainLoss, alpha.GainLossPct)
	}
	if alpha.Percent != 60 {
		t.Errorf("alpha share = %v, want 60", alpha.Percent)
	}

	// One position, one asset, one portfolio — and priced from the operator's
	// column, so the gain above is a real movement and not a position valued at
	// the cost it is being compared against.
	if alpha.Investments != 1 || alpha.Assets != 1 || alpha.Portfolios != 1 {
		t.Errorf("alpha spread = %d positions / %d assets / %d portfolios, want 1/1/1",
			alpha.Investments, alpha.Assets, alpha.Portfolios)
	}
	if alpha.PositionsPricedManual != 1 || alpha.PositionsAtCost != 0 || alpha.PositionsPricedOwn != 0 {
		t.Errorf("alpha pricing = %d own / %d manual / %d at cost, want 0/1/0",
			alpha.PositionsPricedOwn, alpha.PositionsPricedManual, alpha.PositionsAtCost)
	}

	// zeta is the smallest but the only one that gained.
	zeta := dtos[2]
	if zeta.GainLoss != "20" || zeta.GainLossPct != 20 {
		t.Errorf("zeta gain = %q / %v, want 20 / 20", zeta.GainLoss, zeta.GainLossPct)
	}
	if zeta.Percent != 10 {
		t.Errorf("zeta share = %v, want 10", zeta.Percent)
	}
}

// A position sold in full is not something the platform holds, and the three
// counts that describe what it holds have to say so.
//
// It contributed nothing to either total already — quantity zero multiplies
// away — but it was counted as an open position, and its currency was counted
// as unconvertible, so a platform emptied years ago reported positions it no
// longer had and an fx warning over an amount of zero.
func TestSoldOutPositionsDoNotCountAsHeld(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()
	userID := platformsByWeight(t, pool)

	// A second position on the same platform, in a currency this account has no
	// rate for, and sold down to nothing.
	var sourceID, portfolioID uuid.UUID
	if err := pool.QueryRow(ctx,
		`SELECT id, (SELECT id FROM portfolios WHERE user_id = $1 LIMIT 1)
		 FROM investment_sources WHERE user_id = $1 AND name = 'alpha'`,
		userID).Scan(&sourceID, &portfolioID); err != nil {
		t.Fatalf("locate alpha: %v", err)
	}

	assetID := uuid.New()
	if _, err := pool.Exec(ctx,
		`INSERT INTO assets (id, ticker, name, asset_type, currency)
		 VALUES ($1, $2, 'sold out', 'stock', 'JPY')`,
		assetID, uuid.New().String()[:8]); err != nil {
		t.Fatalf("insert asset: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO portfolio_entries
		   (portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
		 VALUES ($1, $2, $3, 0, 900, 'JPY', now())`,
		portfolioID, assetID, sourceID); err != nil {
		t.Fatalf("insert sold-out entry: %v", err)
	}

	platforms, err := repo.GetPlatformsWithStats(ctx, userID, money.USD)
	if err != nil {
		t.Fatalf("GetPlatformsWithStats: %v", err)
	}

	var alpha PlatformStats
	for _, p := range platforms {
		if p.Name == "alpha" {
			alpha = p
		}
	}
	if alpha.Name == "" {
		t.Fatalf("alpha missing from %v", names(platforms))
	}

	if alpha.Investments != 1 || alpha.Assets != 1 {
		t.Errorf("alpha = %d positions / %d assets, want 1/1: the sold-out one is not held",
			alpha.Investments, alpha.Assets)
	}
	if alpha.PositionsUnconverted != 0 {
		t.Errorf("alpha unconverted = %d, want 0: nothing is being added at face value",
			alpha.PositionsUnconverted)
	}
	if alpha.TotalValue != "600.00000000" || alpha.MarketValue != "540.00000000" {
		t.Errorf("alpha cost/market = %q/%q, want 600/540 unchanged",
			alpha.TotalValue, alpha.MarketValue)
	}
}

// A platform with nothing in it still appears — it exists, the owner created it
// — and reports zeroes rather than dropping out of the list or dividing by one.
func TestPlatformWithNoPositionsReportsZeroes(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID := platformsByWeight(t, pool)

	sourceID := uuid.New()
	if _, err := pool.Exec(context.Background(),
		`INSERT INTO investment_sources (id, user_id, name, source_type)
		 VALUES ($1, $2, 'vacía', 'broker')`, sourceID, userID); err != nil {
		t.Fatalf("insert empty platform: %v", err)
	}

	platforms, err := repo.GetPlatformsWithStats(context.Background(), userID, money.USD)
	if err != nil {
		t.Fatalf("GetPlatformsWithStats: %v", err)
	}

	dtos := NewPlatformListResponse(platforms)
	empty := dtos[len(dtos)-1]
	if empty.Name != "vacía" {
		t.Fatalf("last platform = %q, want the empty one (order: %v)", empty.Name, names(platforms))
	}
	if empty.Investments != 0 || empty.GainLossPct != 0 || empty.Percent != 0 {
		t.Errorf("empty platform = %d positions, %v%% gain, %v%% share; want zeroes",
			empty.Investments, empty.GainLossPct, empty.Percent)
	}
}

func names(platforms []PlatformStats) []string {
	out := make([]string, len(platforms))
	for i, p := range platforms {
		out[i] = p.Name
	}

	return out
}
