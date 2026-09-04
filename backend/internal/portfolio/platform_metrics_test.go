package portfolio

import (
	"testing"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

func TestGainOf(t *testing.T) {
	cases := []struct {
		name         string
		cost, market string
		wantGain     string
		wantPct      float64
	}{
		{"a gain", "1000.00", "1250.00", "250", 25},
		{"a loss", "1000.00", "770.00", "-230", -23},
		{"flat", "1000.00", "1000.00", "0", 0},
		// A platform sold down to nothing, or one holding only positions valued
		// at their own cost, has no base to express a return against. Zero is
		// the answer, not a division by zero dressed up as infinity.
		{"no cost basis", "0", "0", "0", 0},
		// Unparsable reads as zero, the same as everywhere else these strings
		// are read.
		{"garbage", "not-a-number", "500.00", "500", 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gain, pct := gainOf(tc.cost, tc.market)
			if gain != tc.wantGain {
				t.Errorf("gain = %q, want %q", gain, tc.wantGain)
			}
			if pct != tc.wantPct {
				t.Errorf("pct = %v, want %v", pct, tc.wantPct)
			}
		})
	}
}

// The position from the broker screenshot, as a platform would report it:
// 15.55 invested, 11.97 today. The percentage has to come out at the broker's
// own −23.02%, because a return that disagrees with the statement it came from
// is worse than no return at all.
func TestGainOfMatchesTheBrokersOwnFigure(t *testing.T) {
	gain, pct := gainOf("15.55", "11.97")
	if gain != "-3.58" {
		t.Errorf("gain = %q, want -3.58", gain)
	}
	if pct != -23.02 {
		t.Errorf("pct = %v, want -23.02", pct)
	}
}

// The share is what makes the ordering readable, and it has to be computed over
// the rows being returned so the column adds to 100.
func TestPlatformListSharesAddUp(t *testing.T) {
	platforms := []PlatformStats{
		{ID: uuid.New(), Name: "A", TotalValue: "6000.00", MarketValue: "7000.00", DisplayCurrency: money.USD},
		{ID: uuid.New(), Name: "B", TotalValue: "3000.00", MarketValue: "2500.00", DisplayCurrency: money.USD},
		{ID: uuid.New(), Name: "C", TotalValue: "1000.00", MarketValue: "1000.00", DisplayCurrency: money.USD},
	}

	dtos := NewPlatformListResponse(platforms)
	if len(dtos) != 3 {
		t.Fatalf("dtos = %d, want 3", len(dtos))
	}

	wantShares := []float64{60, 30, 10}
	total := 0.0
	for i, dto := range dtos {
		if dto.Percent != wantShares[i] {
			t.Errorf("%s share = %v, want %v", dto.Name, dto.Percent, wantShares[i])
		}
		total += dto.Percent
	}
	if total != 100 {
		t.Errorf("shares add to %v, want 100", total)
	}

	// The gain travels per row, against that row's own cost.
	if dtos[0].GainLoss != "1000" || dtos[0].GainLossPct != 16.67 {
		t.Errorf("A gain = %q / %v, want 1000 / 16.67", dtos[0].GainLoss, dtos[0].GainLossPct)
	}
	if dtos[1].GainLoss != "-500" || dtos[1].GainLossPct != -16.67 {
		t.Errorf("B gain = %q / %v, want -500 / -16.67", dtos[1].GainLoss, dtos[1].GainLossPct)
	}
}

// A single platform read on its own — the re-read after an edit — has no set to
// take a share of, and must say nothing rather than claim the whole account.
func TestSinglePlatformResponseClaimsNoShare(t *testing.T) {
	dto := newPlatformResponse(PlatformStats{
		Name: "solo", TotalValue: "6000.00", MarketValue: "7000.00", DisplayCurrency: money.USD,
	})
	if dto.Percent != 0 {
		t.Errorf("percent = %v, want 0: a share needs the whole set", dto.Percent)
	}
	if dto.GainLoss != "1000" {
		t.Errorf("gainLoss = %q, want 1000: the gain does not need one", dto.GainLoss)
	}
}

// The counts that describe what a platform holds travel with the amounts, and
// the three pricing ones partition the positions exactly.
//
// That partition is the point: a platform whose positions have no market price
// is valued at the cost it is being compared against, so it reports a gain of
// zero — the same figure a genuinely flat platform reports. PositionsAtCost is
// what separates "did not move" from "not priced".
func TestPlatformResponseCarriesWhatItHolds(t *testing.T) {
	dto := newPlatformResponse(PlatformStats{
		Name:            "mixta",
		Investments:     5,
		Assets:          3,
		Portfolios:      2,
		TotalValue:      "1000.00",
		MarketValue:     "1000.00",
		DisplayCurrency: money.USD,

		PositionsPricedOwn:    1,
		PositionsPricedManual: 2,
		PositionsAtCost:       2,
	})

	if dto.Assets != 3 || dto.Portfolios != 2 {
		t.Errorf("spread = %d assets / %d portfolios, want 3/2", dto.Assets, dto.Portfolios)
	}

	sum := dto.PositionsPricedOwn + dto.PositionsPricedManual + dto.PositionsAtCost
	if sum != dto.Investments {
		t.Errorf("pricing counts add to %d, want %d: the three partition the positions",
			sum, dto.Investments)
	}

	// Flat on the face of it, and the counts are the only thing that says two of
	// the five never had a price to move.
	if dto.GainLoss != "0" || dto.GainLossPct != 0 {
		t.Errorf("gain = %q / %v, want 0 / 0", dto.GainLoss, dto.GainLossPct)
	}
	if dto.PositionsAtCost == 0 {
		t.Error("positionsAtCost = 0: a zero gain would then be indistinguishable from an unpriced one")
	}
}
