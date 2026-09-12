package market

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yeferson59/gofinance/v2/money"
)

func TestImportAssetsFromFile(t *testing.T) {
	t.Run("valid rows are upserted, invalid rows are skipped and reported", func(t *testing.T) {
		csv := "ticker,name,assetType,exchange,currency\n" +
			"AAPL,Apple Inc.,stock,NASDAQ,USD\n" +
			",Missing Ticker,stock,NASDAQ,USD\n" +
			"BTC-USD,Bitcoin,crypto,,USD\n"

		var upserted []string
		repo := new(fakeRepository{
			upsertAsset: func(_ context.Context, ticker, name string, assetType AssetType, exchange string, currency money.Currency, _ Sector) (Asset, error) {
				upserted = append(upserted, ticker)
				return Asset{Ticker: ticker, Name: name, AssetType: assetType, Exchange: exchange, Currency: currency}, nil
			},
		})
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

	// The spreadsheet is how a whole catalog gets classified, so the column has
	// to absorb the spellings an operator actually types and refuse the ones
	// nobody can act on rather than importing the row unclassified.
	t.Run("the optional sector column is normalised, and a bad one skips the row", func(t *testing.T) {
		csv := "ticker,name,assetType,currency,industria\n" +
			"AAPL,Apple Inc.,stock,USD,Tecnología\n" +
			"JPM,JPMorgan,stock,USD,Financial Services\n" +
			"VT,Vanguard Total World,etf,USD,\n" +
			"BTC-USD,Bitcoin,crypto,USD,Tecnología\n" +
			"XYZ,Unknown Co,stock,USD,Ganadería\n"

		got := map[string]Sector{}
		repo := new(fakeRepository{
			upsertAsset: func(_ context.Context, ticker, _ string, _ AssetType, _ string, _ money.Currency, sector Sector) (Asset, error) {
				got[ticker] = sector

				return Asset{Ticker: ticker}, nil
			},
		})
		svc := newTestServices(repo, newMemStorage())

		result, err := svc.ImportAssetsFromFile(context.Background(), []byte(csv), "assets.csv", "")
		if err != nil {
			t.Fatalf("ImportAssetsFromFile: %v", err)
		}

		if result.Imported != 3 || result.Skipped != 2 {
			t.Fatalf("imported = %d, skipped = %d, want 3 and 2: %+v", result.Imported, result.Skipped, result.Errors)
		}

		if got["AAPL"] != SectorTechnology {
			t.Errorf("AAPL sector = %q, want technology", got["AAPL"])
		}
		if got["JPM"] != SectorFinancials {
			t.Errorf("JPM sector = %q, want financials", got["JPM"])
		}
		// A blank cell is the ordinary case, not an error: most of a catalog is
		// unclassified and a broad fund has no single industry to give.
		if got["VT"] != SectorNone {
			t.Errorf("VT sector = %q, want none", got["VT"])
		}

		// A coin cannot carry one, and a label nobody recognises is not filed
		// under a guess. Both skip the row rather than import it stripped of
		// the one column the operator opened the file to fill.
		if _, imported := got["BTC-USD"]; imported {
			t.Error("a coin with a sector was imported instead of skipped")
		}
		if _, imported := got["XYZ"]; imported {
			t.Error("an unrecognised sector was imported instead of skipped")
		}
	})

	t.Run("missing required columns fail fast", func(t *testing.T) {
		csv := "symbol,precio\nAAPL,100\n"
		svc := newTestServices(new(fakeRepository{}), newMemStorage())

		_, err := svc.ImportAssetsFromFile(context.Background(), []byte(csv), "assets.csv", "")
		if err == nil || !strings.Contains(err.Error(), "missing required columns") {
			t.Fatalf("err = %v, want a missing-columns error", err)
		}
	})

	t.Run("repository failures are reported per row without stopping the import", func(t *testing.T) {
		csv := "ticker,name,assetType,currency\nAAPL,Apple Inc.,stock,USD\n"
		repo := new(fakeRepository{
			upsertAsset: func(context.Context, string, string, AssetType, string, money.Currency, Sector) (Asset, error) {
				return Asset{}, errors.New("db write failed")
			},
		})
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
