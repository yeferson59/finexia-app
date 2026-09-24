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

// Investment funds, migrations 000056 and 000057. Same database contract as
// postgres_cash_db_test.go: they run against TEST_DATABASE_URL and skip without
// it.

type fundFixture struct {
	pool        *pgxpool.Pool
	repo        *PostgresRepository
	plant       func(uuid.UUID)
	userID      uuid.UUID
	portfolioID uuid.UUID
	sourceID    uuid.UUID
}

func newFundFixture(t *testing.T) fundFixture {
	t.Helper()
	pool := growthTestPool(t)

	f := fundFixture{
		pool:        pool,
		repo:        NewPostgresRepository(pool),
		userID:      uuid.New(),
		portfolioID: uuid.New(),
		sourceID:    uuid.New(),
	}

	// The fund's asset is created by the code under test, and handed to the
	// teardown as soon as it exists.
	f.plant = dropFixture(t, pool, f.userID)

	f.exec(t, `INSERT INTO users (id, name, email, role_id, preferred_currency)
	           VALUES ($1, 'fund probe', $2, (SELECT id FROM roles WHERE name = 'customer'), 'USD')`,
		f.userID, f.userID.String()+"@probe.test")
	f.exec(t, `INSERT INTO investment_sources (id, user_id, name, source_type)
	           VALUES ($1, $2, 'fiduciaria', 'mutual_funds')`, f.sourceID, f.userID)
	f.exec(t, `INSERT INTO portfolios (id, user_id, name, type, risk_id, base_currency)
	           VALUES ($1, $2, 'funds', 'diversified', (SELECT id FROM risks LIMIT 1), 'USD')`,
		f.portfolioID, f.userID)

	return f
}

func (f fundFixture) exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	if _, err := f.pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
}

func fundDay(day int) time.Time {
	return time.Date(2026, time.September, day, 0, 0, 0, 0, time.UTC)
}

// create opens a USD fund with units bought at unitValue on the 1st, and a
// current unit value when current is not empty.
func (f fundFixture) create(t *testing.T, units, unitValue, current string, currentOn time.Time) Fund {
	t.Helper()

	in := NewFundInput{
		PortfolioID: f.portfolioID,
		SourceID:    f.sourceID,
		Name:        "FIC Renta Fija",
		Currency:    money.USD,
		Tracking:    FundUnits,
		Date:        fundDay(1),
		Units:       mustDecimal(t, units),
		UnitValue:   mustDecimal(t, unitValue),
	}
	if current != "" {
		in.CurrentUnitValue = mustDecimal(t, current)
		in.CurrentDate = currentOn
	}

	fund, err := f.repo.CreateFund(context.Background(), f.userID, in)
	if err != nil {
		t.Fatalf("CreateFund: %v", err)
	}

	f.plant(fund.AssetID)

	return fund
}

func (f fundFixture) mark(t *testing.T, assetID uuid.UUID, on time.Time, unitValue string) {
	t.Helper()

	if _, err := f.repo.UpsertFundMark(context.Background(), f.userID, assetID, FundMarkInput{
		Date: on, UnitValue: mustDecimal(t, unitValue),
	}); err != nil {
		t.Fatalf("UpsertFundMark(%s @ %s): %v", on.Format(time.DateOnly), unitValue, err)
	}
}

func (f fundFixture) unmark(t *testing.T, assetID uuid.UUID, on time.Time) {
	t.Helper()

	if err := f.repo.DeleteFundMark(context.Background(), f.userID, assetID, on); err != nil {
		t.Fatalf("DeleteFundMark(%s): %v", on.Format(time.DateOnly), err)
	}
}

// ownPrice is what user_asset_prices holds for the fund, "" when nothing.
func (f fundFixture) ownPrice(t *testing.T, assetID uuid.UUID) string {
	t.Helper()

	var price string
	err := f.pool.QueryRow(context.Background(), `
		SELECT COALESCE((SELECT price::text FROM user_asset_prices WHERE user_id = $1 AND asset_id = $2), '')
	`, f.userID, assetID).Scan(&price)
	if err != nil {
		t.Fatalf("read own price: %v", err)
	}

	return price
}

// snapshot runs the snapshot job's two steps for the fixture's portfolio only.
func (f fundFixture) snapshot(t *testing.T, on time.Time) {
	t.Helper()
	ctx := context.Background()

	rows, err := f.repo.GetAllPortfolioSummaryRows(ctx)
	if err != nil {
		t.Fatalf("GetAllPortfolioSummaryRows: %v", err)
	}

	for _, row := range rows {
		if row.PortfolioID != f.portfolioID {
			continue
		}

		if err := f.repo.UpsertPortfolioSnapshot(ctx, row, on); err != nil {
			t.Fatalf("UpsertPortfolioSnapshot(%s): %v", on.Format(time.DateOnly), err)
		}

		return
	}

	t.Fatalf("no summary row for the fixture's portfolio")
}

type fundSnapshot struct{ value, gain, pct, fundSlice string }

func (f fundFixture) readSnapshot(t *testing.T, on time.Time) fundSnapshot {
	t.Helper()

	var s fundSnapshot
	if err := f.pool.QueryRow(context.Background(), `
		SELECT total_value::text, total_gain_loss::text, total_gain_loss_pct::text, COALESCE(allocation ->> 'fund', '')
		FROM portfolio_snapshots
		WHERE portfolio_id = $1 AND snapshot_date = $2::date
	`, f.portfolioID, on.Format(time.DateOnly)).Scan(&s.value, &s.gain, &s.pct, &s.fundSlice); err != nil {
		t.Fatalf("read snapshot %s: %v", on.Format(time.DateOnly), err)
	}

	return s
}

func (f fundFixture) wantSnapshot(t *testing.T, on time.Time, value, gain, pct string) {
	t.Helper()

	s := f.readSnapshot(t, on)
	day := on.Format(time.DateOnly)

	sameAmount(t, day+" total", s.value, value)
	sameAmount(t, day+" gain", s.gain, gain)
	sameAmount(t, day+" gain %", s.pct, pct)
	sameAmount(t, day+" fund slice", s.fundSlice, value)
}

// A fund opened with what a unit is worth now is priced by its owner from the
// first read, and its purchase walks in at that value, not at its cost: the gain
// it made before anyone here saw it is not the return of the day it was typed
// in (000036).
func TestFundOpensPricedByItsMark(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	fund := f.create(t, "1000", "12345.678901", "12431.22", fundDay(30))

	if fund.Tracking != FundUnits || fund.Currency != money.USD || fund.Name != "FIC Renta Fija" || fund.Marks != 1 {
		t.Fatalf("fund = %+v", fund)
	}

	sameAmount(t, "units", fund.Units, "1000")
	sameAmount(t, "cost", fund.Cost, "12345678.901")
	sameAmount(t, "value", fund.Value, "12431220")

	if fund.PricedAtCost || fund.ValuedOn == nil || !fund.ValuedOn.Equal(fundDay(30)) {
		t.Errorf("priced at cost %v, valued on %v; want the mark of the 30th", fund.PricedAtCost, fund.ValuedOn)
	}

	if len(fund.Positions) != 1 || fund.Positions[0].PortfolioID != f.portfolioID || fund.Positions[0].SourceName != "fiduciaria" {
		t.Errorf("positions = %+v", fund.Positions)
	}

	var (
		marketValue       string
		pricedOwn, atCost int64
		recorded          string
		assetType         string
	)

	if err := f.pool.QueryRow(ctx, `
		SELECT total_market_value::text, positions_priced_own, positions_at_cost
		FROM portfolio_summary WHERE portfolio_id = $1
	`, f.portfolioID).Scan(&marketValue, &pricedOwn, &atCost); err != nil {
		t.Fatalf("read summary: %v", err)
	}

	sameAmount(t, "summary market value", marketValue, "12431220")

	if pricedOwn != 1 || atCost != 0 {
		t.Errorf("priced own %d, at cost %d; want the fund priced by its owner", pricedOwn, atCost)
	}

	if err := f.pool.QueryRow(ctx, `
		SELECT t.recorded_market_price::text, a.asset_type::text
		FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		JOIN assets a ON a.id = pe.asset_id
		WHERE pe.asset_id = $1
	`, fund.AssetID).Scan(&recorded, &assetType); err != nil {
		t.Fatalf("read the purchase: %v", err)
	}

	sameAmount(t, "recorded market price", recorded, "12431.22")

	if assetType != "fund" {
		t.Errorf("asset type = %q, want fund", assetType)
	}

	// The fund is the owner's own catalog row, and nobody else's.
	other := uuid.New()
	if _, err := f.repo.GetFund(ctx, other, fund.AssetID); !errors.Is(err, ErrFundNotFound) {
		t.Errorf("another user's GetFund = %v, want ErrFundNotFound", err)
	}

	if _, err := f.repo.UpsertFundMark(ctx, other, fund.AssetID, FundMarkInput{Date: fundDay(30), UnitValue: mustDecimal(t, "1")}); !errors.Is(err, ErrFundNotFound) {
		t.Errorf("another user's UpsertFundMark = %v, want ErrFundNotFound", err)
	}
}

// The price follows the latest mark, whatever order the marks arrive in, and a
// fund whose marks are all deleted falls back to its cost.
func TestFundPriceFollowsTheLatestMark(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	fund := f.create(t, "100", "10", "", time.Time{})

	if !fund.PricedAtCost || fund.UnitValue != nil {
		t.Fatalf("a fund with no mark: priced at cost %v, unit value %v", fund.PricedAtCost, fund.UnitValue)
	}

	sameAmount(t, "value at cost", fund.Value, "1000")

	if got := f.ownPrice(t, fund.AssetID); got != "" {
		t.Fatalf("own price = %q before any mark, want none", got)
	}

	f.mark(t, fund.AssetID, fundDay(10), "10.5")
	f.mark(t, fund.AssetID, fundDay(5), "10.2") // older: does not move the price

	sameAmount(t, "own price", f.ownPrice(t, fund.AssetID), "10.5")

	f.mark(t, fund.AssetID, fundDay(10), "10.6") // same day: replaces it

	sameAmount(t, "own price after the correction", f.ownPrice(t, fund.AssetID), "10.6")

	f.unmark(t, fund.AssetID, fundDay(10))

	sameAmount(t, "own price after deleting the latest", f.ownPrice(t, fund.AssetID), "10.2")

	marks, err := f.repo.GetFundMarks(ctx, f.userID, fund.AssetID)
	if err != nil || len(marks) != 1 || !marks[0].Date.Equal(fundDay(5)) {
		t.Fatalf("GetFundMarks = %+v, %v; want the mark of the 5th", marks, err)
	}

	f.unmark(t, fund.AssetID, fundDay(5))

	if got := f.ownPrice(t, fund.AssetID); got != "" {
		t.Errorf("own price = %q with no mark left, want none", got)
	}

	if err := f.repo.DeleteFundMark(ctx, f.userID, fund.AssetID, fundDay(5)); !errors.Is(err, ErrFundMarkNotFound) {
		t.Errorf("deleting a mark twice = %v, want ErrFundMarkNotFound", err)
	}
}

// A mark that arrives after its date revalues the snapshots from that date on,
// and only those: the gain lands on the day of the statement, not on the day it
// was typed in (D9). Deleting it puts them back.
func TestFundLateMarkRestatesSnapshots(t *testing.T) {
	f := newFundFixture(t)

	fund := f.create(t, "100", "10", "", time.Time{})

	// Three days the job valued the fund at cost: nobody had marked it.
	for _, day := range []int{2, 3, 4} {
		f.snapshot(t, fundDay(day))
		f.wantSnapshot(t, fundDay(day), "1000", "0", "0")
	}

	// The statement of the 3rd, typed in later.
	f.mark(t, fund.AssetID, fundDay(3), "11")

	f.wantSnapshot(t, fundDay(2), "1000", "0", "0")
	f.wantSnapshot(t, fundDay(3), "1100", "100", "10")
	f.wantSnapshot(t, fundDay(4), "1100", "100", "10")

	// And the one of the 4th.
	f.mark(t, fund.AssetID, fundDay(4), "12")

	f.wantSnapshot(t, fundDay(3), "1100", "100", "10")
	f.wantSnapshot(t, fundDay(4), "1200", "200", "20")

	// A snapshot taken now uses the latest mark, and a later correction of it
	// moves it like any other.
	f.snapshot(t, fundDay(5))
	f.wantSnapshot(t, fundDay(5), "1200", "200", "20")

	// Taking back the 3rd: that day returns to cost, the 4th keeps its own.
	f.unmark(t, fund.AssetID, fundDay(3))

	f.wantSnapshot(t, fundDay(2), "1000", "0", "0")
	f.wantSnapshot(t, fundDay(3), "1000", "0", "0")
	f.wantSnapshot(t, fundDay(4), "1200", "200", "20")
	f.wantSnapshot(t, fundDay(5), "1200", "200", "20")

	// Re-running the job on a day keeps its fund rows in step with its totals.
	f.snapshot(t, fundDay(5))
	f.mark(t, fund.AssetID, fundDay(5), "12.5")
	f.wantSnapshot(t, fundDay(5), "1250", "250", "25")
}

// A fund still held cannot be dropped; once its position is gone it can, and
// its asset goes with it.
func TestDeleteFundWaitsForItsPositions(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	fund := f.create(t, "100", "10", "10.5", fundDay(10))

	if err := f.repo.DeleteFund(ctx, f.userID, fund.AssetID); !errors.Is(err, ErrFundHasPositions) {
		t.Fatalf("DeleteFund while held = %v, want ErrFundHasPositions", err)
	}

	if _, err := f.repo.DeletePortfolioEntry(ctx, f.userID, fund.Positions[0].EntryID); err != nil {
		t.Fatalf("DeletePortfolioEntry: %v", err)
	}

	if err := f.repo.DeleteFund(ctx, f.userID, fund.AssetID); err != nil {
		t.Fatalf("DeleteFund: %v", err)
	}

	if _, err := f.repo.GetFund(ctx, f.userID, fund.AssetID); !errors.Is(err, ErrFundNotFound) {
		t.Errorf("GetFund after deleting = %v, want ErrFundNotFound", err)
	}

	if got := f.ownPrice(t, fund.AssetID); got != "" {
		t.Errorf("own price = %q after deleting the fund, want none", got)
	}

	var exists bool
	if err := f.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM assets WHERE id = $1)`, fund.AssetID).Scan(&exists); err != nil {
		t.Fatalf("read asset: %v", err)
	}

	if exists {
		t.Error("the fund's asset is still in the catalog")
	}
}

// A fund that reached the portfolio without being created here — a file whose
// category said "fondo" — is listed and can be marked like any other: the
// owner holds it, so they follow it by units.
func TestHeldFundIsAdopted(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	var assetID uuid.UUID
	if err := f.pool.QueryRow(ctx, `
		INSERT INTO assets (ticker, name, asset_type, currency, created_by, is_curated)
		VALUES ($1, 'Fondo importado', 'fund', 'USD', $2, FALSE)
		RETURNING id
	`, newFundTicker(), f.userID).Scan(&assetID); err != nil {
		t.Fatalf("insert asset: %v", err)
	}

	f.plant(assetID)

	if _, err := f.repo.CreatePortfolioEntry(ctx, f.userID, f.portfolioID, assetID, f.sourceID, money.USD, TransactionInput{
		Type:            Buy,
		Quantity:        mustDecimal(t, "50"),
		Price:           money.NewFromDecimal(mustDecimal(t, "20"), money.USD),
		Currency:        money.USD,
		TransactionDate: fundDay(1),
	}); err != nil {
		t.Fatalf("CreatePortfolioEntry: %v", err)
	}

	funds, err := f.repo.GetFundsByUserID(ctx, f.userID)
	if err != nil || len(funds) != 1 || funds[0].AssetID != assetID || funds[0].Tracking != FundUnits {
		t.Fatalf("GetFundsByUserID = %+v, %v; want the held fund, by units", funds, err)
	}

	sameAmount(t, "units", funds[0].Units, "50")

	if marks, err := f.repo.GetFundMarks(ctx, f.userID, assetID); err != nil || len(marks) != 0 {
		t.Fatalf("GetFundMarks = %+v, %v; want none yet", marks, err)
	}

	f.mark(t, assetID, fundDay(20), "21")

	fund, err := f.repo.GetFund(ctx, f.userID, assetID)
	if err != nil {
		t.Fatalf("GetFund: %v", err)
	}

	sameAmount(t, "value", fund.Value, "1050")

	// Something that is not a fund is not adopted.
	if _, err := f.repo.UpsertFundMark(ctx, f.userID, uuid.New(), FundMarkInput{Date: fundDay(20), UnitValue: mustDecimal(t, "1")}); !errors.Is(err, ErrFundNotFound) {
		t.Errorf("marking an unknown asset = %v, want ErrFundNotFound", err)
	}
}
