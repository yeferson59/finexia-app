package market

import (
	"errors"
	"testing"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// The sector is free text on every door it comes through — a request body, a
// spreadsheet column, a seed — so what NormalizeSector accepts is what decides
// whether two users filing the same industry land in the same slice of the
// breakdown or in two.
func TestNormalizeSector(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want Sector
		ok   bool
	}{
		// The spellings the app's own audience types.
		{"Tecnología", SectorTechnology, true},
		{"TECNOLOGIA", SectorTechnology, true},
		{"  tecnologia  ", SectorTechnology, true},
		{"Salud", SectorHealthcare, true},
		{"Servicios financieros", SectorFinancials, true},
		{"Consumo básico", SectorConsumerStaple, true},

		// Provider spellings. These are why sectorKey drops the conjunctions
		// NormKey leaves alone: an ampersand is a word separator in a sector
		// label and nothing else.
		{"Financial Services", SectorFinancials, true},
		{"Health Care", SectorHealthcare, true},
		{"Metals & Mining", SectorMaterials, true},
		{"Oil & Gas", SectorEnergy, true},
		{"Aerospace & Defense", SectorIndustrials, true},
		{"Hotels/Restaurants/Leisure", SectorConsumerDisc, true},

		// The canonical value, which is what this app's own export and API
		// send back. NormKey has already spaced the underscores out by the time
		// the lookup happens, so this is the path that puts them back.
		{"consumer_discretionary", SectorConsumerDisc, true},
		{"communication_services", SectorCommunication, true},
		{"real_estate", SectorRealEstate, true},

		// Blank is the ordinary case, not an error: most of the catalog is
		// unclassified and a row that leaves the column out is not malformed.
		{"", SectorNone, true},
		{"   ", SectorNone, true},

		// Neither bucket is storable. They are derived at read time, and
		// accepting one here would let a catalog row claim a classification the
		// breakdown means to compute.
		{"unclassified", SectorNone, false},
		{"not_applicable", SectorNone, false},

		// A label nobody recognises is refused rather than filed under a guess.
		{"criptomonedas", SectorNone, false},
		{"asdf", SectorNone, false},
	} {
		got, ok := NormalizeSector(tc.raw)
		if got != tc.want || ok != tc.ok {
			t.Errorf("NormalizeSector(%q) = (%q, %t), want (%q, %t)", tc.raw, got, ok, tc.want, tc.ok)
		}
	}
}

// Every storable sector has to survive the round trip its own canonical
// spelling makes through the normaliser, or a value this package wrote could
// come back rejected.
func TestEveryStorableSectorNormalisesToItself(t *testing.T) {
	for _, sector := range Sectors {
		if !sector.IsValid() {
			t.Errorf("%q is in Sectors but IsValid rejects it", sector)
		}

		if got, ok := NormalizeSector(string(sector)); !ok || got != sector {
			t.Errorf("NormalizeSector(%q) = (%q, %t), want it back unchanged", sector, got, ok)
		}
	}

	for _, bucket := range []Sector{SectorUnclassified, SectorNotApplicable} {
		if bucket.IsValid() {
			t.Errorf("%q is a derived bucket and must not be storable", bucket)
		}
	}
}

// The line HasSector draws is what keeps the breakdown's two empty buckets
// meaning different things: one is work to do, the other is nothing to do.
func TestHasSector(t *testing.T) {
	for _, at := range []AssetType{Stock, ETF, Bond, Other} {
		if !at.HasSector() {
			t.Errorf("%q should be classifiable", at)
		}
	}

	for _, at := range []AssetType{Crypto, Cash, RealEstate, Commodity} {
		if at.HasSector() {
			t.Errorf("%q has no industry behind it and should not be classifiable", at)
		}
	}
}

// A sector on an asset that cannot have one is refused rather than stored: the
// breakdown files those by their type, so the value would be one no screen
// reads and every later reader has to explain.
func TestAssetInputRejectsASectorOnAnUnclassifiableType(t *testing.T) {
	coin := func(sector Sector) AssetSpec {
		return AssetSpec{Ticker: "BTC-USD", Name: "Bitcoin", AssetType: Crypto, Currency: money.USD, Sector: sector}
	}
	stock := func(sector Sector) AssetSpec {
		return AssetSpec{Ticker: "AAPL", Name: "Apple", AssetType: Stock, Exchange: "NASDAQ", Currency: money.USD, Sector: sector}
	}

	if _, err := normalizeAssetSpec(coin(SectorTechnology)); !errors.Is(err, errAssetSectorNotApplicable) {
		t.Errorf("a sector on a coin was accepted: err = %v", err)
	}

	// The same coin with no sector is ordinary, and so is a stock with one.
	if _, err := normalizeAssetSpec(coin(SectorNone)); err != nil {
		t.Errorf("an unclassified coin was rejected: %v", err)
	}

	in, err := normalizeAssetSpec(stock(SectorTechnology))
	if err != nil {
		t.Fatalf("a classified stock was rejected: %v", err)
	}
	if in.Sector != SectorTechnology {
		t.Errorf("sector = %q, want technology", in.Sector)
	}

	// A bucket reaching this far is a caller bypassing NormalizeSector; the
	// validation has to stop it on its own.
	if _, err := normalizeAssetSpec(stock(SectorUnclassified)); !errors.Is(err, errAssetSectorInvalid) {
		t.Errorf("a derived bucket was stored as a sector: err = %v", err)
	}
}

// dec is the weights these tests are written in. A helper rather than
// decimal.MustFromString at every line because a breakdown is eleven of them and
// the numbers are the part worth reading.
func dec(t *testing.T, raw string) decimal.Decimal {
	t.Helper()

	d, err := decimal.NewFromString(raw)
	if err != nil {
		t.Fatalf("decimal %q: %v", raw, err)
	}

	return d
}

// What Validate refuses is what separates a breakdown from a pile of numbers.
// The case it must not refuse is the important one: a real fact sheet does not
// add up to a hundred, and a rule that demanded it would reject every fund the
// feature exists for.
func TestSectorBreakdownValidate(t *testing.T) {
	weight := func(s Sector, w string) SectorWeight {
		return SectorWeight{Sector: s, Weight: dec(t, w)}
	}

	t.Run("a fact sheet's own weights are accepted short of 100", func(t *testing.T) {
		// VOO, rounded off a published breakdown: the missing 0.4 % is the
		// fund's cash and futures, and it is not the operator's to invent.
		voo := SectorBreakdown{
			weight(SectorTechnology, "33.1"),
			weight(SectorFinancials, "13.8"),
			weight(SectorConsumerDisc, "10.4"),
			weight(SectorHealthcare, "10.2"),
			weight(SectorCommunication, "9.4"),
			weight(SectorIndustrials, "8.1"),
			weight(SectorConsumerStaple, "5.6"),
			weight(SectorEnergy, "3.2"),
			weight(SectorUtilities, "2.4"),
			weight(SectorRealEstate, "2.1"),
			weight(SectorMaterials, "1.3"),
		}

		if err := voo.Validate(); err != nil {
			t.Fatalf("a real fund's breakdown was rejected: %v", err)
		}
		if got := voo.Total().String(); got != "99.6" {
			t.Errorf("total = %s, want 99.6", got)
		}
	})

	t.Run("an empty breakdown is the ordinary asset, not an error", func(t *testing.T) {
		if err := SectorBreakdown(nil).Validate(); err != nil {
			t.Errorf("no breakdown was treated as a bad one: %v", err)
		}
	})

	for _, tc := range []struct {
		name      string
		breakdown SectorBreakdown
		want      error
	}{
		{
			"a repeated industry",
			SectorBreakdown{weight(SectorTechnology, "20"), weight(SectorTechnology, "30")},
			errAssetSectorWeightDuplicate,
		},
		{
			"a weight of zero says nothing a missing row does not",
			SectorBreakdown{weight(SectorTechnology, "0")},
			errAssetSectorWeightRange,
		},
		{
			"a negative weight",
			SectorBreakdown{weight(SectorTechnology, "-5")},
			errAssetSectorWeightRange,
		},
		{
			"a weight above a hundred",
			SectorBreakdown{weight(SectorTechnology, "120")},
			errAssetSectorWeightRange,
		},
		{
			"a total above a hundred is a transcription error, not an omission",
			SectorBreakdown{weight(SectorTechnology, "60"), weight(SectorFinancials, "50")},
			errAssetSectorWeightsTotal,
		},
		{
			"a derived bucket cannot take a share of an asset",
			SectorBreakdown{weight(SectorUnclassified, "50")},
			errAssetSectorInvalid,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.breakdown.Validate(); !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

// The spreadsheet cell is the door the eleven weights of a fund realistically
// come through, so what it accepts decides whether loading a catalog is a paste
// or an afternoon.
func TestParseSectorBreakdown(t *testing.T) {
	t.Run("the spelling a person types", func(t *testing.T) {
		got, ok := ParseSectorBreakdown("technology: 33.1; financials: 13.8")
		if !ok {
			t.Fatal("a plain breakdown was rejected")
		}
		if len(got) != 2 {
			t.Fatalf("rows = %d, want 2: %+v", len(got), got)
		}
		if got[0].Sector != SectorTechnology || got[0].Weight.String() != "33.1" {
			t.Errorf("first row = %+v, want technology 33.1", got[0])
		}
		if got[1].Sector != SectorFinancials {
			t.Errorf("second sector = %q, want financials", got[1].Sector)
		}
	})

	t.Run("the spelling a paste produces", func(t *testing.T) {
		// A Spanish label, a decimal comma, a percent sign, an equals instead of
		// a colon and a line break instead of a semicolon: every one of these
		// comes off a real fact sheet or a real keyboard.
		got, ok := ParseSectorBreakdown("Tecnología = 33,1%\nSalud = 10,2 %")
		if !ok {
			t.Fatal("a pasted breakdown was rejected")
		}
		if len(got) != 2 || got[0].Sector != SectorTechnology || got[1].Sector != SectorHealthcare {
			t.Fatalf("got %+v, want technology and healthcare", got)
		}
		if got[0].Weight.String() != "33.1" {
			t.Errorf("weight = %s, want 33.1", got[0].Weight)
		}
	})

	t.Run("an empty cell is the ordinary row", func(t *testing.T) {
		got, ok := ParseSectorBreakdown("   ")
		if !ok || !got.IsEmpty() {
			t.Errorf("got %+v (ok=%t), want an empty breakdown and no error", got, ok)
		}
	})

	for _, raw := range []string{
		"technology",             // a weight with no number
		"ganadería: 30",          // an industry nobody recognises
		"technology: mucho",      // a number that is not one
		": 30",                   // a weight with no industry
		"technology: 30; ; : 10", // a pair that is neither
	} {
		t.Run("unreadable: "+raw, func(t *testing.T) {
			if got, ok := ParseSectorBreakdown(raw); ok {
				t.Errorf("%q was read as %+v; the importer would store a fund it could not read", raw, got)
			}
		})
	}
}

// The two classifications are alternatives, and the rule that keeps them from
// sitting side by side is in the service rather than in a column: the database
// can hold both, and one asset answering the industry question twice is what
// no chart could resolve.
func TestAssetSpecRejectsBothClassificationsAtOnce(t *testing.T) {
	fund := func(sector Sector, weights SectorBreakdown) AssetSpec {
		return AssetSpec{
			Ticker:        "VOO",
			Name:          "Vanguard S&P 500",
			AssetType:     ETF,
			Currency:      money.USD,
			Sector:        sector,
			SectorWeights: weights,
		}
	}

	breakdown := SectorBreakdown{
		{Sector: SectorTechnology, Weight: dec(t, "60")},
		{Sector: SectorFinancials, Weight: dec(t, "40")},
	}

	if _, err := normalizeAssetSpec(fund(SectorTechnology, breakdown)); !errors.Is(err, errAssetSectorBoth) {
		t.Errorf("an asset was filed under one industry and eleven at once: err = %v", err)
	}

	if _, err := normalizeAssetSpec(fund(SectorNone, breakdown)); err != nil {
		t.Errorf("a fund with only a breakdown was rejected: %v", err)
	}

	// The type rule applies to the breakdown exactly as it does to the single
	// sector: there is no business behind a coin to split into industries.
	coin := AssetSpec{Ticker: "BTC-USD", Name: "Bitcoin", AssetType: Crypto, Currency: money.USD, SectorWeights: breakdown}
	if _, err := normalizeAssetSpec(coin); !errors.Is(err, errAssetSectorNotApplicable) {
		t.Errorf("a coin was split across industries: err = %v", err)
	}

	// And the arithmetic reaches the caller through the same door, so a bad
	// breakdown is a 400 from the API rather than a constraint violation.
	over := SectorBreakdown{{Sector: SectorTechnology, Weight: dec(t, "80")}, {Sector: SectorEnergy, Weight: dec(t, "80")}}
	if _, err := normalizeAssetSpec(fund(SectorNone, over)); !errors.Is(err, errAssetSectorWeightsTotal) {
		t.Errorf("a breakdown adding to 160 %% was accepted: err = %v", err)
	}
}
