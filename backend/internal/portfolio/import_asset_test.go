package portfolio

import (
	"context"
	"errors"
	"strings"
	"testing"

)

func TestImportAssetsFromFile(t *testing.T) {
	t.Run("valid rows are upserted, invalid rows are skipped and reported", func(t *testing.T) {
		csv := "ticker,name,assetType,exchange,currency\n" +
			"AAPL,Apple Inc.,stock,NASDAQ,USD\n" +
			",Missing Ticker,stock,NASDAQ,USD\n" +
			"BTC-USD,Bitcoin,crypto,,USD\n"

		var upserted []string
		repo := &fakeRepository{
			upsertAsset: func(_ context.Context, ticker, name string, assetType AssetType, exchange, currency string) (Asset, error) {
				upserted = append(upserted, ticker)
				return Asset{Ticker: ticker, Name: name, AssetType: assetType, Exchange: exchange, Currency: currency}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		result, err := svc.ImportAssetsFromFile(context.Background(), []byte(csv), "assets.csv", "")
		if err != nil {
			t.Fatalf("ImportAssetsFromFile: %v", err)
		}
		if result.TotalRows != 3 {
			t.Errorf("TotalRows = %d, want 3", result.TotalRows)
		}
		if result.Imported != 2 {
			t.Errorf("Imported = %d, want 2", result.Imported)
		}
		if result.Skipped != 1 {
			t.Errorf("Skipped = %d, want 1", result.Skipped)
		}
		if len(result.Errors) != 1 || result.Errors[0].Row != 3 {
			t.Fatalf("Errors = %+v, want a single error on row 3", result.Errors)
		}
		if len(upserted) != 2 || upserted[0] != "AAPL" || upserted[1] != "BTC-USD" {
			t.Errorf("upserted = %v, want [AAPL BTC-USD]", upserted)
		}
	})

	t.Run("missing required columns fail fast", func(t *testing.T) {
		csv := "symbol,precio\nAAPL,100\n"
		svc := newTestServices(&fakeRepository{}, newMemStorage())

		_, err := svc.ImportAssetsFromFile(context.Background(), []byte(csv), "assets.csv", "")
		if err == nil || !strings.Contains(err.Error(), "missing required columns") {
			t.Fatalf("err = %v, want a missing-columns error", err)
		}
	})

	t.Run("repository failures are reported per row without stopping the import", func(t *testing.T) {
		csv := "ticker,name,assetType,currency\nAAPL,Apple Inc.,stock,USD\n"
		repo := &fakeRepository{
			upsertAsset: func(context.Context, string, string, AssetType, string, string) (Asset, error) {
				return Asset{}, errors.New("db write failed")
			},
		}
		svc := newTestServices(repo, newMemStorage())

		result, err := svc.ImportAssetsFromFile(context.Background(), []byte(csv), "assets.csv", "")
		if err != nil {
			t.Fatalf("ImportAssetsFromFile: %v", err)
		}
		if result.Imported != 0 || result.Skipped != 1 {
			t.Errorf("Imported/Skipped = %d/%d, want 0/1", result.Imported, result.Skipped)
		}
	})
}
