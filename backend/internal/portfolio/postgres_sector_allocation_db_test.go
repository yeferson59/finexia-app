package portfolio

import (
	"context"
	"testing"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
)

// The sector breakdown is all SQL — the bucketing, the grouping, the valuation
// and the conversion — so the fake repository the rest of the suite runs on
// cannot see any of it. Same database and same skip as the holdings tests:
//
//	TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/finexia go test ./internal/portfolio/

// heldAcrossSectors plants an account whose positions cover every branch of the
// CASE the query buckets on: two assets in one sector and in two portfolios, a
// second sector, a classifiable asset nobody classified, an asset whose type
// cannot be classified, and a position sold down to nothing.
func heldAcrossSectors(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
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

	// Teardown before the inserts: a t.Fatalf below must still take the rows
	// with it. The assets reach it through track, as they are created.
	track := dropFixture(t, pool, userID)

	exec(`INSERT INTO users (id, name, email, role_id, preferred_currency)
	      VALUES ($1, 'sector probe', $2, (SELECT id FROM roles WHERE name = 'customer'), 'USD')`,
		userID, userID.String()+"@probe.test")

	sourceID := uuid.New()
	exec(`INSERT INTO investment_sources (id, user_id, name, source_type)
	      VALUES ($1, $2, 'probe', 'broker')`, sourceID, userID)

	for _, p := range []uuid.UUID{portfolioA, portfolioB} {
		exec(`INSERT INTO portfolios (id, user_id, name, type, risk_id, base_currency)
		      VALUES ($1, $2, $3, 'stocks', (SELECT id FROM risks LIMIT 1), 'USD')`,
			p, userID, "probe "+p.String()[:8])
	}

	asset := func(ticker, assetType string, sector any, price float64) uuid.UUID {
		t.Helper()
		id := uuid.New()
		exec(`INSERT INTO assets (id, ticker, name, asset_type, currency, current_price, sector)
		      VALUES ($1, $2, $3, $4::asset_type, 'USD', $5, $6)`,
			id, ticker+uuid.New().String()[:6], ticker, assetType, price, sector)
		track(id)

		return id
	}

	entry := func(portfolioID, assetID uuid.UUID, quantity, price float64) {
		t.Helper()
		exec(`INSERT INTO portfolio_entries
		        (portfolio_id, asset_id, source_id, quantity, price, cost_currency, entry_date)
		      VALUES ($1, $2, $3, $4, $5, 'USD', $6)`,
			portfolioID, assetID, sourceID, quantity, price, time.Now())
	}

	// Two different tech assets, one of them split over both portfolios. The
	// sector has to be one row of 3000 regardless of how the positions are
	// arranged, which is the whole reason this view exists beside the holdings.
	chips := asset("CHIPS", "stock", string(market.SectorTechnology), 100)
	entry(portfolioA, chips, 10, 90)
	entry(portfolioB, chips, 10, 95)

	cloud := asset("CLOUD", "stock", string(market.SectorTechnology), 200)
	entry(portfolioA, cloud, 5, 180)

	bank := asset("BANK", "stock", string(market.SectorFinancials), 50)
	entry(portfolioA, bank, 10, 45)

	// Classifiable and unclassified: the bucket that is work to do.
	fund := asset("FUND", "etf", nil, 40)
	entry(portfolioA, fund, 10, 38)

	// Not classifiable at all: the bucket that is nothing to do. It carries no
	// sector in the catalog either, so only the asset type tells it apart from
	// the row above — which is exactly what the CASE is for.
	coin := asset("COIN", "crypto", nil, 250)
	entry(portfolioA, coin, 2, 200)

	soldOut := asset("SOLDOUT", "stock", string(market.SectorEnergy), 500)
	entry(portfolioA, soldOut, 0, 400)

	return userID
}

func TestSectorAllocationBucketsEveryPositionExactlyOnce(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID := heldAcrossSectors(t, pool)

	items, err := repo.GetSectorAllocationByUserID(context.Background(), userID, money.USD)
	if err != nil {
		t.Fatalf("GetSectorAllocationByUserID: %v", err)
	}

	bySector := make(map[market.Sector]SectorAllocationItem, len(items))
	for _, item := range items {
		bySector[item.Sector] = item
	}

	// A sector sold down to nothing is not a sector the user is exposed to.
	if _, listed := bySector[market.SectorEnergy]; listed {
		t.Error("a fully sold position still contributes a sector slice")
	}
	if len(items) != 4 {
		t.Fatalf("slices = %d, want 4 (technology, financials, unclassified, not_applicable): %+v", len(items), items)
	}

	// 20 × 100 across two portfolios, plus 5 × 200: one row, two assets.
	tech := bySector[market.SectorTechnology]
	worth(t, "technology", tech.MarketValue, 3000)
	if tech.Assets != 2 {
		t.Errorf("technology assets = %d, want 2", tech.Assets)
	}

	worth(t, "financials", bySector[market.SectorFinancials].MarketValue, 500)

	// The two kinds of blank must not be added together: one is a prompt to go
	// classify four hundred dollars, the other is a coin that will never have a
	// sector to fill in.
	unclassified := bySector[market.SectorUnclassified]
	worth(t, "unclassified", unclassified.MarketValue, 400)
	if unclassified.Assets != 1 {
		t.Errorf("unclassified assets = %d, want 1", unclassified.Assets)
	}

	worth(t, "not applicable", bySector[market.SectorNotApplicable].MarketValue, 500)

	// Everything the user holds is in exactly one slice, which is what makes
	// the shares a reading of their money rather than of a subset of it.
	total := decimalFromInt(0)
	for _, item := range items {
		total = total.Add(amountOf(item.MarketValue))
	}
	if total.Cmp(decimalFromInt(4400)) != 0 {
		t.Errorf("slices total %s, want 4400 — the whole portfolio", total)
	}
}

// The breakdown and the consolidated list read the same catalog column, so an
// asset's sector has to be the same in both. Two screens disagreeing about one
// position is the failure this app has already had once, over the asset type.
func TestHoldingsCarryTheSameSectorTheBreakdownGroupsOn(t *testing.T) {
	pool := growthTestPool(t)
	repo := NewPostgresRepository(pool)
	userID := heldAcrossSectors(t, pool)
	ctx := context.Background()

	holdings, err := repo.GetAssetHoldingsByUserID(ctx, userID, money.USD)
	if err != nil {
		t.Fatalf("GetAssetHoldingsByUserID: %v", err)
	}

	bySector := make(map[market.Sector]int)
	for _, h := range holdings {
		bySector[h.Sector]++
	}

	if bySector[market.SectorTechnology] != 2 {
		t.Errorf("holdings in technology = %d, want 2", bySector[market.SectorTechnology])
	}

	// Unclassified and unclassifiable both read as empty here, unfolded on
	// purpose: this is a list of assets, and "FUND — " is the row that tells
	// the user which one to go classify.
	if bySector[market.SectorNone] != 2 {
		t.Errorf("holdings with no sector = %d, want 2 (the ETF and the coin)", bySector[market.SectorNone])
	}
	if bySector[market.SectorUnclassified] != 0 || bySector[market.SectorNotApplicable] != 0 {
		t.Error("a holding row carries a derived bucket; those belong to the breakdown alone")
	}
}
