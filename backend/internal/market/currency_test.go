package market

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

func TestNormalizeAssetInputCurrency(t *testing.T) {
	t.Run("a three-letter non-currency is rejected", func(t *testing.T) {
		_, err := normalizeAssetInput("AAPL", "Apple", Stock, "NASDAQ", money.Currency(255))
		if !errors.Is(err, errAssetCurrencyInvalid) {
			t.Errorf("err = %v, want errAssetCurrencyInvalid", err)
		}
	})
}

func TestCreateExchangeRateValidatesItsInput(t *testing.T) {
	upsertCalled := false
	newSvc := func() *Service {
		upsertCalled = false
		repo := new(fakeRepository{
			upsertExchangeRate: func(_ context.Context, from, to money.Currency, rate decimal.Decimal, _ time.Time) (ExchangeRate, error) {
				upsertCalled = true
				return ExchangeRate{FromCurrency: from, ToCurrency: to, Rate: rate}, nil
			},
		})
		return newTestServices(repo, newMemStorage())
	}

	t.Run("an unknown source currency is rejected", func(t *testing.T) {
		svc := newSvc()
		_, err := svc.CreateExchangeRate(context.Background(), money.Currency(255), money.USD, decimal.MustFromString("4000"))
		if !errors.Is(err, errExchangeRateCurrencyInvalid) {
			t.Errorf("err = %v, want errExchangeRateCurrencyInvalid", err)
		}
		if upsertCalled {
			t.Error("a rejected pair must not reach the repository")
		}
	})

	t.Run("an unknown target currency is rejected", func(t *testing.T) {
		svc := newSvc()
		_, err := svc.CreateExchangeRate(context.Background(), money.USD, money.Currency(255), decimal.MustFromString("4000"))
		if !errors.Is(err, errExchangeRateCurrencyInvalid) {
			t.Errorf("err = %v, want errExchangeRateCurrencyInvalid", err)
		}
	})

	for _, rate := range []string{"0", "-1", "-4000.5"} {
		t.Run("rate "+rate+" is rejected", func(t *testing.T) {
			svc := newSvc()
			_, err := svc.CreateExchangeRate(context.Background(), money.USD, money.COP, decimal.MustFromString(rate))
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
		got, err := svc.CreateExchangeRate(context.Background(), money.USD, money.COP, decimal.MustFromString("4123.45"))
		if err != nil {
			t.Fatalf("CreateExchangeRate: %v", err)
		}
		if !upsertCalled {
			t.Fatal("a valid pair should reach the repository")
		}
		if got.FromCurrency != money.USD || got.ToCurrency != money.COP {
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
		upsertAsset: func(_ context.Context, ticker, name string, assetType AssetType, exchange string, currency money.Currency) (Asset, error) {
			stored = append(stored, ticker+"/"+currency.String())
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
