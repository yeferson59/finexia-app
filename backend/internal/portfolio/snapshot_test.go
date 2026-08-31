package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/yeferson59/gofinance/v2/money"
)

func TestSyncPortfolioSnapshots(t *testing.T) {
	t.Run("upserts one snapshot per summary row", func(t *testing.T) {
		rows := []SnapshotRow{
			{PortfolioID: uuid.New(), BaseCurrency: money.USD, TotalMarketValue: "1000.00", TotalGainLoss: "100.00", TotalGainLossPct: "11.11"},
			{PortfolioID: uuid.New(), BaseCurrency: money.EUR, TotalMarketValue: "500.00", TotalGainLoss: "-20.00", TotalGainLossPct: "-3.85"},
		}

		type upsertCall struct {
			portfolioID uuid.UUID
			date        time.Time
			totalValue  string
			currency    money.Currency
		}
		var calls []upsertCall

		repo := new(fakeRepository{
			getAllPortfolioSummaryRows: func(context.Context) ([]SnapshotRow, error) {
				return rows, nil
			},
			upsertPortfolioSnapshot: func(_ context.Context, portfolioID uuid.UUID, snapshotDate time.Time, totalValue, totalGainLoss, totalGainLossPct string, currency money.Currency) error {
				calls = append(calls, upsertCall{portfolioID, snapshotDate, totalValue, currency})
				return nil
			},
		})
		storage := newMemStorage()
		svc := newTestServices(repo, storage)

		count, errs := svc.SyncPortfolioSnapshots(context.Background())
		if count != 2 || len(errs) != 0 {
			t.Fatalf("count/errs = %d/%v, want 2/none", count, errs)
		}
		if len(calls) != 2 {
			t.Fatalf("upsert calls = %d, want 2", len(calls))
		}

		wantDate := time.Now().UTC().Truncate(24 * time.Hour)
		for i, call := range calls {
			if call.portfolioID != rows[i].PortfolioID {
				t.Errorf("call %d portfolioID = %s, want %s", i, call.portfolioID, rows[i].PortfolioID)
			}
			if !call.date.Equal(wantDate) {
				t.Errorf("call %d date = %v, want truncated today %v", i, call.date, wantDate)
			}
			if call.totalValue != rows[i].TotalMarketValue || call.currency != rows[i].BaseCurrency {
				t.Errorf("call %d value/currency = %s/%s", i, call.totalValue, call.currency)
			}
		}
	})

	t.Run("a failing row is collected and the rest still sync", func(t *testing.T) {
		badID := uuid.New()
		rows := []SnapshotRow{
			{PortfolioID: badID, TotalMarketValue: "1.00"},
			{PortfolioID: uuid.New(), TotalMarketValue: "2.00"},
		}
		repo := new(fakeRepository{
			getAllPortfolioSummaryRows: func(context.Context) ([]SnapshotRow, error) {
				return rows, nil
			},
			upsertPortfolioSnapshot: func(_ context.Context, portfolioID uuid.UUID, _ time.Time, _, _, _ string, _ money.Currency) error {
				if portfolioID == badID {
					return errors.New("constraint violation")
				}
				return nil
			},
		})
		svc := newTestServices(repo, newMemStorage())

		count, errs := svc.SyncPortfolioSnapshots(context.Background())
		if count != 1 {
			t.Errorf("count = %d, want 1", count)
		}
		if len(errs) != 1 {
			t.Errorf("errs = %v, want exactly one", errs)
		}
	})

	t.Run("summary query failure aborts the sync", func(t *testing.T) {
		repo := new(fakeRepository{
			getAllPortfolioSummaryRows: func(context.Context) ([]SnapshotRow, error) {
				return nil, errors.New("view missing")
			},
		})
		storage := newMemStorage()
		svc := newTestServices(repo, storage)

		count, errs := svc.SyncPortfolioSnapshots(context.Background())
		if count != 0 || len(errs) != 1 {
			t.Fatalf("count/errs = %d/%v, want 0 and one error", count, errs)
		}
	})
}
