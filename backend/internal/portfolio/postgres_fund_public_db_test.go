package portfolio

import (
	"context"
	"errors"
	"maps"
	"math/rand/v2"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata/sfc"
)

// The SFC's catalog and the links to it, migration 000059. Same database
// contract as postgres_fund_db_test.go.

// plantPublicFund writes a fund to the catalog under codes no real fund has,
// and takes it out when the test ends.
func (f fundFixture) plantPublicFund(t *testing.T, name string, valueDate time.Time) PublicFund {
	t.Helper()

	key := sfc.FundKey{EntityType: 900 + rand.IntN(99), Entity: rand.IntN(1_000_000), Fund: rand.IntN(1_000_000), Compartment: 1, Participation: 501}
	public := newPublicFund(sfc.Fund{
		Key: key, EntityName: "Fiduciaria Probe", Name: name, Kind: "FIC DE TIPO GENERAL",
		UnitValue: "10", Date: valueDate, Investors: 10,
	})

	if _, err := f.repo.UpsertPublicFunds(context.Background(), []PublicFund{public}); err != nil {
		t.Fatalf("UpsertPublicFunds: %v", err)
	}

	t.Cleanup(func() {
		_, _ = f.pool.Exec(context.Background(), `DELETE FROM public_funds WHERE id = $1`, public.ID)
	})

	return public
}

func publishedValues(t *testing.T, byDay map[int]string) []PublicFundValue {
	t.Helper()

	values := make([]PublicFundValue, 0, len(byDay))

	for d := 1; d <= 30; d++ {
		if v, ok := byDay[d]; ok {
			values = append(values, PublicFundValue{Date: fundDay(d), UnitValue: mustDecimal(t, v)})
		}
	}

	return values
}

// marksBySource reads the fund's marks as day → "value source".
func (f fundFixture) marksBySource(t *testing.T, assetID uuid.UUID) map[int]string {
	t.Helper()

	marks, err := f.repo.GetFundMarks(context.Background(), f.userID, assetID)
	if err != nil {
		t.Fatalf("GetFundMarks: %v", err)
	}

	out := make(map[int]string, len(marks))

	for _, m := range marks {
		v := mustDecimal(t, m.UnitValue)
		out[m.Date.Day()] = v.String() + " " + string(m.Source)
	}

	return out
}

func TestPublicFundCatalog(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	token := "probe" + uuid.New().String()[:8]
	fresh := f.plantPublicFund(t, "FONDO DE INVERSIÓN "+token+" RENTA", fundDay(22))
	f.plantPublicFund(t, "FONDO "+token+" LIQUIDADO", time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC))

	found, err := f.repo.SearchPublicFunds(ctx, []string{token, "inversion"}, fundDay(1), 25)
	if err != nil || len(found) != 1 || found[0].ID != fresh.ID || found[0].Key() != fresh.Key() {
		t.Fatalf("search = %+v, %v; want only the fund still published", found, err)
	}

	sameAmount(t, "unit value", found[0].UnitValue, "10")

	// An older day does not overwrite a newer one; a newer day does.
	older, newer := fresh, fresh
	older.UnitValue, older.ValueDate = "9", fundDay(20)
	newer.UnitValue, newer.ValueDate = "11", fundDay(23)

	for _, step := range []struct {
		fund PublicFund
		want string
	}{{older, "10"}, {newer, "11"}} {
		if _, err := f.repo.UpsertPublicFunds(ctx, []PublicFund{step.fund}); err != nil {
			t.Fatalf("UpsertPublicFunds: %v", err)
		}

		got, err := f.repo.GetPublicFund(ctx, fresh.ID)
		if err != nil {
			t.Fatalf("GetPublicFund: %v", err)
		}

		sameAmount(t, "catalog value", got.UnitValue, step.want)
	}

	if _, err := f.repo.GetPublicFund(ctx, "0-0-0-0-0"); !errors.Is(err, ErrPublicFundNotFound) {
		t.Errorf("GetPublicFund(unknown) = %v, want ErrPublicFundNotFound", err)
	}
}

// Linking writes the published values as marks: the price follows them and the
// snapshots from their day on are revalued. A mark the owner wrote is never
// overwritten, and unlinking takes back only what the link brought.
func TestLinkedFundIsPricedByThePublishedValues(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	public := f.plantPublicFund(t, "FIC PROBE", fundDay(22))
	fund := f.create(t, "100", "10", "", time.Time{})

	for _, d := range []int{2, 3, 4} {
		f.snapshot(t, fundDay(d))
	}

	linked, err := f.repo.LinkFund(ctx, f.userID, fund.AssetID, public.ID, publishedValues(t, map[int]string{3: "11", 4: "12"}))
	if err != nil {
		t.Fatalf("LinkFund: %v", err)
	}

	if linked.PublicFund == nil || linked.PublicFund.ID != public.ID || linked.PublicFund.FundName != "FIC PROBE" {
		t.Fatalf("linked fund = %+v", linked.PublicFund)
	}

	sameAmount(t, "value", linked.Value, "1200")
	sameAmount(t, "own price", f.ownPrice(t, fund.AssetID), "12")
	f.wantSnapshot(t, fundDay(2), "1000", "0", "0")
	f.wantSnapshot(t, fundDay(3), "1100", "100", "10")
	f.wantSnapshot(t, fundDay(4), "1200", "200", "20")

	// The owner corrects the 4th; the next import brings the 4th and the 5th.
	f.mark(t, fund.AssetID, fundDay(4), "12.5")

	n, err := f.repo.ImportPublicMarks(ctx, f.userID, fund.AssetID, public.ID, publishedValues(t, map[int]string{3: "11", 4: "13", 5: "14"}))
	if err != nil {
		t.Fatalf("ImportPublicMarks: %v", err)
	}

	// The 3rd was already there and the 4th is the owner's: only the 5th.
	if n != 1 {
		t.Errorf("imported %d marks, want 1", n)
	}

	want := map[int]string{3: "11 public", 4: "12.5 user", 5: "14 public"}
	if got := f.marksBySource(t, fund.AssetID); !maps.Equal(got, want) {
		t.Errorf("marks = %v, want %v", got, want)
	}

	sameAmount(t, "own price", f.ownPrice(t, fund.AssetID), "14")

	linkedFunds, err := f.repo.GetLinkedFunds(ctx)
	if err != nil {
		t.Fatalf("GetLinkedFunds: %v", err)
	}

	var mine *LinkedFund

	for i := range linkedFunds {
		if linkedFunds[i].AssetID == fund.AssetID {
			mine = &linkedFunds[i]
		}
	}

	if mine == nil || mine.PublicFundID != public.ID || !mine.Since.Equal(fundDay(6)) {
		t.Fatalf("linked = %+v, want it to start from the 6th", mine)
	}

	// An import for a link that is gone writes nothing.
	if n, err := f.repo.ImportPublicMarks(ctx, f.userID, fund.AssetID, "1-2-3-4-5", publishedValues(t, map[int]string{6: "15"})); err != nil || n != 0 {
		t.Errorf("import for another fund = %d, %v; want nothing written", n, err)
	}

	unlinked, err := f.repo.UnlinkFund(ctx, f.userID, fund.AssetID)
	if err != nil {
		t.Fatalf("UnlinkFund: %v", err)
	}

	if unlinked.PublicFund != nil {
		t.Errorf("still linked to %+v", unlinked.PublicFund)
	}

	if got := f.marksBySource(t, fund.AssetID); !maps.Equal(got, map[int]string{4: "12.5 user"}) {
		t.Errorf("marks after unlinking = %v, want only the owner's", got)
	}

	sameAmount(t, "own price after unlinking", f.ownPrice(t, fund.AssetID), "12.5")
	f.wantSnapshot(t, fundDay(3), "1000", "0", "0")
	f.wantSnapshot(t, fundDay(4), "1250", "250", "25")
}

// A fund created linked walks in at the latest published value, like one
// created with its current unit value (D13): the purchase records it, and the
// gain it made before is not the return of the day it was typed in.
func TestCreateLinkedFundWalksInAtThePublishedValue(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	public := f.plantPublicFund(t, "FIC PROBE", fundDay(22))

	in := NewFundInput{
		PortfolioID: f.portfolioID, SourceID: f.sourceID, Name: "Mi FIC", Currency: money.USD, Tracking: FundUnits,
		Date: fundDay(1), Units: mustDecimal(t, "100"), UnitValue: mustDecimal(t, "10"),
		PublicFundID: public.ID, publicValues: publishedValues(t, map[int]string{1: "10", 10: "10.4", 22: "10.8"}),
	}

	fund, err := f.repo.CreateFund(ctx, f.userID, in)
	if err != nil {
		t.Fatalf("CreateFund: %v", err)
	}

	f.plant(fund.AssetID)

	if fund.PublicFund == nil || fund.PublicFund.ID != public.ID || fund.Marks != 3 {
		t.Fatalf("fund = %+v (link %+v)", fund, fund.PublicFund)
	}

	sameAmount(t, "value", fund.Value, "1080")

	var recorded string
	if err := f.pool.QueryRow(ctx, `
		SELECT t.recorded_market_price::text FROM transactions t
		JOIN portfolio_entries pe ON pe.id = t.entry_id
		WHERE pe.asset_id = $1
	`, fund.AssetID).Scan(&recorded); err != nil {
		t.Fatalf("read the purchase: %v", err)
	}

	sameAmount(t, "recorded market price", recorded, "10.8")
}

// A fund followed by balance has synthetic units: it cannot take a published
// unit value, whether the code or the schema is asked.
func TestBalanceFundCannotBeLinked(t *testing.T) {
	f := newFundFixture(t)
	ctx := context.Background()

	public := f.plantPublicFund(t, "FIC PROBE", fundDay(22))

	in := validBalanceFund(t)
	in.PortfolioID, in.SourceID, in.Currency, in.Date = f.portfolioID, f.sourceID, money.USD, fundDay(1)

	fund, err := f.repo.CreateFund(ctx, f.userID, in)
	if err != nil {
		t.Fatalf("CreateFund: %v", err)
	}

	f.plant(fund.AssetID)

	if _, err := f.repo.LinkFund(ctx, f.userID, fund.AssetID, public.ID, nil); !errors.Is(err, ErrFundNotLinkable) {
		t.Errorf("LinkFund = %v, want ErrFundNotLinkable", err)
	}

	if _, err := f.pool.Exec(ctx, `UPDATE user_funds SET public_fund_id = $3 WHERE user_id = $1 AND asset_id = $2`,
		f.userID, fund.AssetID, public.ID); err == nil {
		t.Error("the schema let a fund followed by balance be linked")
	}
}
