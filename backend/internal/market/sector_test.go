package market

import (
	"errors"
	"testing"

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
	if _, err := normalizeAssetInput("BTC-USD", "Bitcoin", Crypto, "", money.USD, SectorTechnology); !errors.Is(err, errAssetSectorNotApplicable) {
		t.Errorf("a sector on a coin was accepted: err = %v", err)
	}

	// The same coin with no sector is ordinary, and so is a stock with one.
	if _, err := normalizeAssetInput("BTC-USD", "Bitcoin", Crypto, "", money.USD, SectorNone); err != nil {
		t.Errorf("an unclassified coin was rejected: %v", err)
	}

	in, err := normalizeAssetInput("AAPL", "Apple", Stock, "NASDAQ", money.USD, SectorTechnology)
	if err != nil {
		t.Fatalf("a classified stock was rejected: %v", err)
	}
	if in.sector != SectorTechnology {
		t.Errorf("sector = %q, want technology", in.sector)
	}

	// A bucket reaching this far is a caller bypassing NormalizeSector; the
	// validation has to stop it on its own.
	if _, err := normalizeAssetInput("AAPL", "Apple", Stock, "NASDAQ", money.USD, SectorUnclassified); !errors.Is(err, errAssetSectorInvalid) {
		t.Errorf("a derived bucket was stored as a sector: err = %v", err)
	}
}
