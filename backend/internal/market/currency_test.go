package market

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yeferson59/gofinance/v2/decimal"
)

// Currency codes and exchange rates are now validated against the same
// gofinance rules that later have to accept them: the ISO 4217 table
// money.CurrencyFromISOCode reads, and money.Convert's positive-rate rule.

func TestNormalizeCurrencyCode(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{"canonical code", "USD", "USD", true},
		{"lower case", "cop", "COP", true},
		{"padded", "  eur  ", "EUR", true},
		{"three letters that are not a currency", "ABC", "", false},
		{"a word", "DOLAR", "", false},
		{"two letters", "US", "", false},
		{"empty", "", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := NormalizeCurrencyCode(tc.raw)
			if ok != tc.ok || got != tc.want {
				t.Errorf("NormalizeCurrencyCode(%q) = %q/%v, want %q/%v", tc.raw, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestNormalizeAssetInputCurrency(t *testing.T) {
	t.Run("a three-letter non-currency is rejected", func(t *testing.T) {
		// "ABC" passed the old length check, reached the database, and then
		// cost the asset its price: scanAssetCurrentPrice drops a price whose
		// currency gofinance cannot resolve.
		_, err := normalizeAssetInput("AAPL", "Apple", Stock, "NASDAQ", "ABC")
		if !errors.Is(err, errAssetCurrencyInvalid) {
			t.Errorf("err = %v, want errAssetCurrencyInvalid", err)
		}
	})

	t.Run("a valid currency is stored in its canonical spelling", func(t *testing.T) {
		in, err := normalizeAssetInput("AAPL", "Apple", Stock, "NASDAQ", " cop ")
		if err != nil {
			t.Fatalf("normalizeAssetInput: %v", err)
		}
		if in.currency != "COP" {
			t.Errorf("currency = %q, want COP", in.currency)
		}
	})
}

func TestCreateExchangeRateValidatesItsInput(t *testing.T) {
	upsertCalled := false
	newSvc := func() *Service {
		upsertCalled = false
		repo := new(fakeRepository{
			upsertExchangeRate: func(_ context.Context, from, to string, rate decimal.Decimal, _ time.Time) (ExchangeRate, error) {
				upsertCalled = true
				return ExchangeRate{FromCurrency: from, ToCurrency: to, Rate: rate}, nil
			},
		})
		return newTestServices(repo, newMemStorage())
	}

	t.Run("an unknown source currency is rejected", func(t *testing.T) {
		svc := newSvc()
		_, err := svc.CreateExchangeRate(context.Background(), "ABC", "USD", decimal.MustFromString("4000"))
		if !errors.Is(err, errExchangeRateCurrencyInvalid) {
			t.Errorf("err = %v, want errExchangeRateCurrencyInvalid", err)
		}
		if upsertCalled {
			t.Error("a rejected pair must not reach the repository")
		}
	})

	t.Run("an unknown target currency is rejected", func(t *testing.T) {
		svc := newSvc()
		_, err := svc.CreateExchangeRate(context.Background(), "USD", "ZZZ", decimal.MustFromString("4000"))
		if !errors.Is(err, errExchangeRateCurrencyInvalid) {
			t.Errorf("err = %v, want errExchangeRateCurrencyInvalid", err)
		}
	})

	for _, rate := range []string{"0", "-1", "-4000.5"} {
		t.Run("rate "+rate+" is rejected", func(t *testing.T) {
			svc := newSvc()
			_, err := svc.CreateExchangeRate(context.Background(), "USD", "COP", decimal.MustFromString(rate))
			if !errors.Is(err, errExchangeRateInvalid) {
				t.Errorf("err = %v, want errExchangeRateInvalid", err)
			}
			if upsertCalled {
				t.Error("an unusable rate must not reach the repository")
			}
		})
	}

	t.Run("a valid pair is stored in its canonical spelling", func(t *testing.T) {
		svc := newSvc()
		got, err := svc.CreateExchangeRate(context.Background(), " usd ", "cop", decimal.MustFromString("4123.45"))
		if err != nil {
			t.Fatalf("CreateExchangeRate: %v", err)
		}
		if !upsertCalled {
			t.Fatal("a valid pair should reach the repository")
		}
		if got.FromCurrency != "USD" || got.ToCurrency != "COP" {
			t.Errorf("pair = %s/%s, want USD/COP", got.FromCurrency, got.ToCurrency)
		}
	})
}

func TestUpdateExchangeRateRejectsUnusableRates(t *testing.T) {
	svc := newTestServices(new(fakeRepository{}), newMemStorage())

	// The fake leaves UpdateExchangeRateByID unstubbed, so reaching it would
	// panic — the assertion is that validation returns first.
	_, err := svc.UpdateExchangeRate(context.Background(), [16]byte{}, decimal.MustFromString("0"))
	if !errors.Is(err, errExchangeRateInvalid) {
		t.Errorf("err = %v, want errExchangeRateInvalid", err)
	}
}

func TestImportAssetsRejectsNonISOCurrencies(t *testing.T) {
	csv := "ticker,name,assetType,currency\n" +
		"AAPL,Apple,stock,USD\n" +
		"XYZ,Example,stock,ABC\n" +
		"MSFT,Microsoft,stock,usd\n"

	var stored []string
	repo := new(fakeRepository{
		upsertAsset: func(_ context.Context, ticker, name string, assetType AssetType, exchange, currency string) (Asset, error) {
			stored = append(stored, ticker+"/"+currency)
			return Asset{Ticker: ticker, Currency: currency}, nil
		},
	})
	svc := newTestServices(repo, newMemStorage())

	result, err := svc.ImportAssetsFromFile(context.Background(), []byte(csv), "assets.csv", "")
	if err != nil {
		t.Fatalf("ImportAssetsFromFile: %v", err)
	}
	if result.Imported != 2 || result.Skipped != 1 {
		t.Errorf("imported/skipped = %d/%d, want 2/1", result.Imported, result.Skipped)
	}
	if len(stored) != 2 || stored[0] != "AAPL/USD" || stored[1] != "MSFT/USD" {
		t.Errorf("stored = %v, want [AAPL/USD MSFT/USD]", stored)
	}
}
