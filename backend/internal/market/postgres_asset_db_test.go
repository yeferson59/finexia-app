package market

import (
	"context"
	"errors"
	"os"
	"testing"

	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// What an operator's edit does to a catalog row lives entirely in SQL: which
// fields fall back to the ones already stored, and what happens to the manual
// price when the currency underneath it changes. The fake repository the rest of
// this package's tests run on cannot see any of it, so these need the same
// database the other *_db_test.go files do, and skip the same way without
// TEST_DATABASE_URL:
//
//	TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/finexia go test ./internal/market/

func assetTestPool(t *testing.T) *pgxpool.Pool {
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

// plantAsset inserts one curated row priced in USD, which is the state every
// case below edits away from.
func plantAsset(t *testing.T, pool *pgxpool.Pool, ticker, exchange string) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	id := uuid.New()

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM assets WHERE id = $1`, id)
	})

	if _, err := pool.Exec(ctx, `
		INSERT INTO assets (id, ticker, name, asset_type, exchange, currency, current_price, price_updated_at, is_curated)
		VALUES ($1, $2, 'Probe Inc.', 'stock', NULLIF($3, ''), 'USD', 190.50, NOW() - INTERVAL '3 days', TRUE)
	`, id, ticker, exchange); err != nil {
		t.Fatalf("plant asset: %v", err)
	}

	return id
}

func TestPostgresUpdateAsset(t *testing.T) {
	pool := assetTestPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	// A ticker per subtest: the unique index is on (ticker, exchange), and the
	// duplicate case below needs the collision to be the one it planted.
	base := func(ticker string) AssetUpdate {
		return AssetUpdate{
			Ticker:    ticker,
			Name:      "Probe Inc.",
			AssetType: Stock,
			Exchange:  "NASDAQ",
			Currency:  money.USD,
		}
	}

	price := func(t *testing.T, value string, currency money.Currency) *money.Money {
		t.Helper()
		m := mustMoney(t, value, currency)

		return &m
	}

	t.Run("an edit that leaves the currency alone keeps the stored price", func(t *testing.T) {
		id := plantAsset(t, pool, "PRB1", "NASDAQ")

		upd := base("PRB1")
		upd.Name = "Probe Renamed"

		asset, err := repo.UpdateAsset(ctx, id, upd)
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}

		if asset.Name != "Probe Renamed" {
			t.Errorf("name = %q, want the edited one", asset.Name)
		}
		if asset.CurrentPrice == nil || asset.CurrentPrice.String() != "190.5" {
			t.Errorf("price = %v, want the stored 190.50 untouched", asset.CurrentPrice)
		}
		// Untouched means untouched: the table is ordered by how old the price
		// is, so bumping the timestamp on a name change would hide a stale one.
		if asset.PriceUpdatedAt == nil {
			t.Fatal("priceUpdatedAt = nil, want the original timestamp")
		}
		if hours := asset.UpdatedAt.Sub(*asset.PriceUpdatedAt).Hours(); hours < 71 {
			t.Errorf("priceUpdatedAt is %.0fh before updatedAt, want the original ~72h", hours)
		}
	})

	t.Run("a new price is written with today's timestamp", func(t *testing.T) {
		id := plantAsset(t, pool, "PRB2", "NASDAQ")

		upd := base("PRB2")
		upd.Price = price(t, "205.25", money.USD)

		asset, err := repo.UpdateAsset(ctx, id, upd)
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}

		if asset.CurrentPrice == nil || asset.CurrentPrice.String() != "205.25" {
			t.Errorf("price = %v, want 205.25", asset.CurrentPrice)
		}
		if asset.PriceUpdatedAt == nil || asset.UpdatedAt.Sub(*asset.PriceUpdatedAt).Hours() > 1 {
			t.Errorf("priceUpdatedAt = %v, want it sealed now", asset.PriceUpdatedAt)
		}
	})

	t.Run("re-denominating without a price drops the one that no longer means anything", func(t *testing.T) {
		id := plantAsset(t, pool, "PRB3", "NASDAQ")

		upd := base("PRB3")
		upd.Currency = money.COP

		asset, err := repo.UpdateAsset(ctx, id, upd)
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}

		if asset.Currency != money.COP {
			t.Errorf("currency = %s, want COP", asset.Currency)
		}
		if asset.CurrentPrice != nil || asset.PriceUpdatedAt != nil {
			t.Errorf("price = %v (%v), want both cleared: 190.50 USD is not 190.50 COP",
				asset.CurrentPrice, asset.PriceUpdatedAt)
		}
	})

	t.Run("re-denominating with a price keeps the one that was sent", func(t *testing.T) {
		id := plantAsset(t, pool, "PRB4", "NASDAQ")

		upd := base("PRB4")
		upd.Currency = money.COP
		upd.Price = price(t, "812340.75", money.COP)

		asset, err := repo.UpdateAsset(ctx, id, upd)
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}

		if asset.CurrentPrice == nil || asset.CurrentPrice.String() != "812340.75" {
			t.Errorf("price = %v, want the new one in COP", asset.CurrentPrice)
		}
		if asset.CurrentPrice.GetCurrency() != money.COP {
			t.Errorf("price currency = %s, want COP", asset.CurrentPrice.GetCurrency())
		}
	})

	t.Run("the audience only changes when the edit says so", func(t *testing.T) {
		id := plantAsset(t, pool, "PRB5", "NASDAQ")

		kept, err := repo.UpdateAsset(ctx, id, base("PRB5"))
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}
		if !kept.IsCurated {
			t.Error("isCurated = false after an edit that did not mention it")
		}

		upd := base("PRB5")
		hidden := false
		upd.IsCurated = &hidden

		unpublished, err := repo.UpdateAsset(ctx, id, upd)
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}
		if unpublished.IsCurated {
			t.Error("isCurated = true, want the row off the shared catalog")
		}
	})

	t.Run("an empty exchange is stored as NULL, so the unique index still pairs", func(t *testing.T) {
		id := plantAsset(t, pool, "PRB6", "NASDAQ")

		upd := base("PRB6")
		upd.Exchange = ""

		asset, err := repo.UpdateAsset(ctx, id, upd)
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}
		if asset.Exchange != "" {
			t.Errorf("exchange = %q, want it cleared", asset.Exchange)
		}
	})

	t.Run("a rename onto an existing pair is a conflict, not a driver error", func(t *testing.T) {
		taken := plantAsset(t, pool, "PRB7", "NASDAQ")
		id := plantAsset(t, pool, "PRB8", "NASDAQ")

		if taken == id {
			t.Fatal("the fixture planted one row for two")
		}

		_, err := repo.UpdateAsset(ctx, id, base("PRB7"))
		if !errors.Is(err, errAssetDuplicate) {
			t.Fatalf("err = %v, want errAssetDuplicate", err)
		}
	})

	t.Run("an asset that is gone answers not found", func(t *testing.T) {
		_, err := repo.UpdateAsset(ctx, uuid.New(), base("PRB9"))
		if !errors.Is(err, ErrAssetNotFound) {
			t.Fatalf("err = %v, want ErrAssetNotFound", err)
		}
	})
}

// The sector is the one field the two write paths treat differently, and the
// difference lives entirely in SQL, so it needs the database to be checked.
//
// An upsert that carries no sector keeps the one the row has, because the
// operator's create and the asset spreadsheet do not mean "unclassify it" by
// leaving a column out. An edit does replace it with whatever it sends,
// blank included, because there the whole row travels and an empty field is an
// instruction.
func TestPostgresAssetSector(t *testing.T) {
	pool := assetTestPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	t.Run("an upsert writes the sector and a later one without it keeps it", func(t *testing.T) {
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `DELETE FROM assets WHERE ticker = 'PRBSEC1'`)
		})

		asset, err := repo.UpsertAsset(ctx, AssetSpec{
			Ticker:    "PRBSEC1",
			Name:      "Probe Inc.",
			AssetType: Stock,
			Exchange:  "NASDAQ",
			Currency:  money.USD,
			Sector:    SectorTechnology,
		})
		if err != nil {
			t.Fatalf("UpsertAsset: %v", err)
		}
		if asset.Sector != SectorTechnology {
			t.Fatalf("sector = %q, want technology", asset.Sector)
		}

		// The same ticker re-imported from a file with no sector column. The
		// classification an operator already did must survive it.
		again, err := repo.UpsertAsset(ctx, AssetSpec{
			Ticker:    "PRBSEC1",
			Name:      "Probe Renamed",
			AssetType: Stock,
			Exchange:  "NASDAQ",
			Currency:  money.USD,
		})
		if err != nil {
			t.Fatalf("UpsertAsset again: %v", err)
		}
		if again.Name != "Probe Renamed" {
			t.Errorf("name = %q, want the upsert to have rewritten it", again.Name)
		}
		if again.Sector != SectorTechnology {
			t.Errorf("sector = %q, want technology kept", again.Sector)
		}
	})

	t.Run("an edit replaces the sector, and an empty one clears it", func(t *testing.T) {
		id := plantAsset(t, pool, "PRBSEC2", "NASDAQ")

		upd := AssetUpdate{
			Ticker:    "PRBSEC2",
			Name:      "Probe Inc.",
			AssetType: Stock,
			Exchange:  "NASDAQ",
			Currency:  money.USD,
			Sector:    SectorFinancials,
		}

		asset, err := repo.UpdateAsset(ctx, id, upd)
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}
		if asset.Sector != SectorFinancials {
			t.Fatalf("sector = %q, want financials", asset.Sector)
		}

		upd.Sector = SectorNone
		cleared, err := repo.UpdateAsset(ctx, id, upd)
		if err != nil {
			t.Fatalf("UpdateAsset clearing: %v", err)
		}
		if cleared.Sector != SectorNone {
			t.Errorf("sector = %q, want it cleared", cleared.Sector)
		}

		// Cleared means NULL in the column, not an empty string: the partial
		// index and the breakdown's CASE both test for NULL.
		var stored *string
		if err := pool.QueryRow(ctx, `SELECT sector FROM assets WHERE id = $1`, id).Scan(&stored); err != nil {
			t.Fatalf("read back: %v", err)
		}
		if stored != nil {
			t.Errorf("stored sector = %q, want NULL", *stored)
		}
	})

	t.Run("a contribution leaves the asset unclassified", func(t *testing.T) {
		userID := uuid.New()

		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
			_, _ = pool.Exec(context.Background(), `DELETE FROM assets WHERE ticker = 'PRBSEC3'`)
		})

		if _, err := pool.Exec(ctx, `
			INSERT INTO users (id, name, email, role_id, preferred_currency)
			VALUES ($1, 'sector probe', $2, (SELECT id FROM roles WHERE name = 'customer'), 'USD')
		`, userID, userID.String()+"@probe.test"); err != nil {
			t.Fatalf("plant user: %v", err)
		}

		asset, err := repo.CreateAssetIfAbsent(ctx, userID, "PRBSEC3", "Probe Local", Stock, "BVC", money.COP)
		if err != nil {
			t.Fatalf("CreateAssetIfAbsent: %v", err)
		}
		if asset.Sector != SectorNone {
			t.Errorf("sector = %q, want none: a contribution classifies nothing", asset.Sector)
		}
	})
}

// weightsOf spells a breakdown out in the order the caller wrote it, which is
// deliberately not the order it comes back in: the projection sorts by weight,
// and a test that planted them sorted could not tell the two apart.
func weightsOf(t *testing.T, pairs ...any) SectorBreakdown {
	t.Helper()

	breakdown := make(SectorBreakdown, 0, len(pairs)/2)

	for i := 0; i < len(pairs); i += 2 {
		sector, ok := pairs[i].(Sector)
		if !ok {
			t.Fatalf("weightsOf: %v is not a Sector", pairs[i])
		}

		raw, ok := pairs[i+1].(string)
		if !ok {
			t.Fatalf("weightsOf: %v is not a weight", pairs[i+1])
		}

		breakdown = append(breakdown, SectorWeight{Sector: sector, Weight: decimal.MustFromString(raw)})
	}

	return breakdown
}

// The breakdown is what a whole-market ETF is classified with, and it lives in
// a table of its own — so unlike the single sector column, every one of these
// rules is a write to two tables that has to agree with itself.
func TestPostgresAssetSectorWeights(t *testing.T) {
	pool := assetTestPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	fund := func(ticker string, sector Sector, weights SectorBreakdown) AssetSpec {
		return AssetSpec{
			Ticker:        ticker,
			Name:          "Probe Fund",
			AssetType:     ETF,
			Exchange:      "NYSEARCA",
			Currency:      money.USD,
			Sector:        sector,
			SectorWeights: weights,
		}
	}

	t.Run("an upsert writes the breakdown and reads it back heaviest first", func(t *testing.T) {
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `DELETE FROM assets WHERE ticker = 'PRBW1'`)
		})

		asset, err := repo.UpsertAsset(ctx, fund("PRBW1", SectorNone, weightsOf(t, SectorFinancials, "20.5", SectorTechnology, "30")))
		if err != nil {
			t.Fatalf("UpsertAsset: %v", err)
		}

		if len(asset.SectorWeights) != 2 {
			t.Fatalf("weights = %+v, want two rows", asset.SectorWeights)
		}
		if asset.SectorWeights[0].Sector != SectorTechnology || asset.SectorWeights[1].Sector != SectorFinancials {
			t.Errorf("weights = %+v, want technology before financials", asset.SectorWeights)
		}
		// The column is NUMERIC(7,4), so what comes back is what the column
		// holds — the digits are the ones that were written, padded.
		if asset.SectorWeights[1].Weight.String() != "20.5000" {
			t.Errorf("financials weight = %s, want 20.5", asset.SectorWeights[1].Weight)
		}

		// The same read through the other door, since both go through one
		// projection and this is what proves it.
		again, err := repo.GetAssetByID(ctx, asset.ID)
		if err != nil {
			t.Fatalf("GetAssetByID: %v", err)
		}
		if len(again.SectorWeights) != 2 {
			t.Errorf("a fetch by id lost the breakdown: %+v", again.SectorWeights)
		}
	})

	// The two classifications are alternatives, so writing either has to take
	// the other one with it. Nothing in the schema says so — this is the write
	// path keeping that promise.
	t.Run("a breakdown clears the single sector, and a single sector clears the breakdown", func(t *testing.T) {
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `DELETE FROM assets WHERE ticker = 'PRBW2'`)
		})

		if _, err := repo.UpsertAsset(ctx, fund("PRBW2", SectorTechnology, nil)); err != nil {
			t.Fatalf("UpsertAsset with a sector: %v", err)
		}

		weighted, err := repo.UpsertAsset(ctx, fund("PRBW2", SectorNone, weightsOf(t, SectorTechnology, "60", SectorEnergy, "40")))
		if err != nil {
			t.Fatalf("UpsertAsset with weights: %v", err)
		}
		if weighted.Sector != SectorNone {
			t.Errorf("sector = %q, want it cleared by the breakdown", weighted.Sector)
		}
		if len(weighted.SectorWeights) != 2 {
			t.Fatalf("weights = %+v, want two rows", weighted.SectorWeights)
		}

		narrowed, err := repo.UpsertAsset(ctx, fund("PRBW2", SectorEnergy, nil))
		if err != nil {
			t.Fatalf("UpsertAsset back to one sector: %v", err)
		}
		if narrowed.Sector != SectorEnergy {
			t.Errorf("sector = %q, want energy", narrowed.Sector)
		}
		if !narrowed.SectorWeights.IsEmpty() {
			t.Errorf("weights = %+v, want the breakdown gone with the single sector written", narrowed.SectorWeights)
		}

		// An upsert carrying neither — a re-import from a file with no
		// classification columns — still leaves what is there alone.
		kept, err := repo.UpsertAsset(ctx, fund("PRBW2", SectorNone, nil))
		if err != nil {
			t.Fatalf("UpsertAsset with no classification: %v", err)
		}
		if kept.Sector != SectorEnergy {
			t.Errorf("sector = %q, want energy kept by an upsert that said nothing about it", kept.Sector)
		}
	})

	t.Run("an edit replaces the breakdown, and an empty one removes it", func(t *testing.T) {
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `DELETE FROM assets WHERE ticker = 'PRBW3'`)
		})

		created, err := repo.UpsertAsset(ctx, fund("PRBW3", SectorNone, weightsOf(t, SectorTechnology, "60", SectorEnergy, "40")))
		if err != nil {
			t.Fatalf("UpsertAsset: %v", err)
		}

		upd := AssetUpdate{
			Ticker:        "PRBW3",
			Name:          "Probe Fund",
			AssetType:     ETF,
			Exchange:      "NYSEARCA",
			Currency:      money.USD,
			SectorWeights: weightsOf(t, SectorHealthcare, "55"),
		}

		rewritten, err := repo.UpdateAsset(ctx, created.ID, upd)
		if err != nil {
			t.Fatalf("UpdateAsset: %v", err)
		}
		// The industry the fund no longer holds has to be gone, not left behind
		// at its old weight: that row would keep putting the user's money in a
		// sector the fund exited.
		if len(rewritten.SectorWeights) != 1 || rewritten.SectorWeights[0].Sector != SectorHealthcare {
			t.Fatalf("weights = %+v, want only healthcare", rewritten.SectorWeights)
		}

		upd.SectorWeights = nil
		cleared, err := repo.UpdateAsset(ctx, created.ID, upd)
		if err != nil {
			t.Fatalf("UpdateAsset clearing: %v", err)
		}
		if !cleared.SectorWeights.IsEmpty() {
			t.Errorf("weights = %+v, want an edit with none to clear them", cleared.SectorWeights)
		}
	})
}
