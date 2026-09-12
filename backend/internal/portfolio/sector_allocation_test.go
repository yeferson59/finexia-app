package portfolio

import (
	"testing"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/market"
)

// The share is over every row the breakdown returns, unclassified included.
// Dividing by the classified subtotal instead would report a portfolio as 75 %
// technology when technology is 75 % of the quarter we happen to know about —
// and the user would have no way to tell the two readings apart.
func TestSectorSharesAreTakenOverTheWholePortfolio(t *testing.T) {
	items := []SectorAllocationItem{
		{Sector: market.SectorTechnology, MarketValue: "3000", Currency: money.USD, Assets: 2},
		{Sector: market.SectorUnclassified, MarketValue: "1000", Currency: money.USD, Assets: 3},
	}

	got := NewSectorAllocationResponse(items)
	if len(got) != 2 {
		t.Fatalf("rows = %d, want 2", len(got))
	}

	if got[0].Percent != 75 {
		t.Errorf("technology percent = %v, want 75 (3000 of 4000, not 3000 of 3000)", got[0].Percent)
	}
	if got[1].Percent != 25 {
		t.Errorf("unclassified percent = %v, want 25", got[1].Percent)
	}

	// The count is what makes the unclassified row actionable: "three assets to
	// classify" is a job, a percentage on its own is not.
	if got[1].Assets != 3 {
		t.Errorf("unclassified assets = %d, want 3", got[1].Assets)
	}
	if got[1].Sector != string(market.SectorUnclassified) {
		t.Errorf("unclassified sector = %q, want the bucket spelled out", got[1].Sector)
	}
}

// An account with nothing in it must not divide by zero, and an account whose
// positions are all valued at zero must not either.
func TestSectorSharesSurviveAnEmptyTotal(t *testing.T) {
	if got := NewSectorAllocationResponse(nil); len(got) != 0 {
		t.Errorf("rows = %d, want none", len(got))
	}

	got := NewSectorAllocationResponse([]SectorAllocationItem{
		{Sector: market.SectorTechnology, MarketValue: "0", Currency: money.USD},
	})
	if len(got) != 1 || got[0].Percent != 0 {
		t.Errorf("a zero-valued portfolio reported %+v, want a single row at 0 %%", got)
	}
}
