package market

import "testing"

// A bare "fondo" is a FIC, and a fund that trades on an exchange says so. The
// labels are matched accent- and case-insensitively.
func TestNormalizeAssetTypeFunds(t *testing.T) {
	cases := map[string]AssetType{
		"Fondo":                          Fund,
		"fondos":                         Fund,
		"FIC":                            Fund,
		"Fondo de Inversión Colectiva":   Fund,
		"fondo de pensiones voluntarias": Fund,
		"FPV":                            Fund,
		"money market":                   Fund,
		"mutual fund":                    Fund,
		"ETF":                            ETF,
		"fondo indexado":                 ETF,
		"index fund":                     ETF,
	}

	for label, want := range cases {
		got, ok := NormalizeAssetType(label)
		if !ok || got != want {
			t.Errorf("NormalizeAssetType(%q) = %q, %v; want %q", label, got, ok, want)
		}
	}

	if !Fund.IsValid() {
		t.Error("fund is not a valid asset type")
	}
}
