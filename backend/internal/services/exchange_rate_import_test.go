package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yeferson59/gofinance/money"

	"github.com/yeferson59/finexia-app/internal/entities"
)

func TestImportExchangeRatesFromFile(t *testing.T) {
	t.Run("valid rows are upserted, invalid rows are skipped and reported", func(t *testing.T) {
		csv := "fromCurrency,toCurrency,rate\n" +
			"EUR,USD,1.0850\n" +
			"GBP,US,1.27\n" +
			"USD,COP,-4000\n"

		var pairs []string
		repo := &fakeRepository{
			upsertExchangeRate: func(_ context.Context, from, to string, rate money.Decimal, _ time.Time) (entities.ExchangeRate, error) {
				pairs = append(pairs, from+"/"+to)
				return entities.ExchangeRate{FromCurrency: from, ToCurrency: to, Rate: rate}, nil
			},
		}
		svc := newTestServices(repo, newMemStorage())

		result, err := svc.ImportExchangeRatesFromFile(context.Background(), []byte(csv), "rates.csv", "")
		if err != nil {
			t.Fatalf("ImportExchangeRatesFromFile: %v", err)
		}
		if result.TotalRows != 3 {
			t.Errorf("TotalRows = %d, want 3", result.TotalRows)
		}
		if result.Imported != 1 {
			t.Errorf("Imported = %d, want 1", result.Imported)
		}
		if result.Skipped != 2 {
			t.Errorf("Skipped = %d, want 2", result.Skipped)
		}
		if len(pairs) != 1 || pairs[0] != "EUR/USD" {
			t.Errorf("pairs = %v, want [EUR/USD]", pairs)
		}
	})

	t.Run("missing required columns fail fast", func(t *testing.T) {
		csv := "par,valor\nEUR/USD,1.08\n"
		svc := newTestServices(&fakeRepository{}, newMemStorage())

		_, err := svc.ImportExchangeRatesFromFile(context.Background(), []byte(csv), "rates.csv", "")
		if err == nil || !strings.Contains(err.Error(), "missing required columns") {
			t.Fatalf("err = %v, want a missing-columns error", err)
		}
	})
}
