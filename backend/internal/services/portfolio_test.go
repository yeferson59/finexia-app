package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/entities"
)

func TestParsePeriod(t *testing.T) {
	now := time.Now().UTC()

	cases := []struct {
		period    string
		wantSince bool
		wantFrom  time.Time
	}{
		{"1M", true, now.AddDate(0, -1, 0)},
		{"3M", true, now.AddDate(0, -3, 0)},
		{"6M", true, now.AddDate(0, -6, 0)},
		{"1Y", true, now.AddDate(-1, 0, 0)},
		{"ALL", false, time.Time{}},
		{"", false, time.Time{}},
		{"garbage", false, time.Time{}},
	}

	for _, tc := range cases {
		t.Run("period "+tc.period, func(t *testing.T) {
			hasSince, since := parsePeriod(tc.period)
			if hasSince != tc.wantSince {
				t.Fatalf("parsePeriod(%q) hasSince = %v, want %v", tc.period, hasSince, tc.wantSince)
			}
			if !tc.wantSince {
				return
			}
			if diff := since.Sub(tc.wantFrom); diff < -time.Minute || diff > time.Minute {
				t.Errorf("parsePeriod(%q) since = %v, want ~%v", tc.period, since, tc.wantFrom)
			}
		})
	}
}

func TestBuildGrowthSummary(t *testing.T) {
	t.Run("empty input returns zero summary", func(t *testing.T) {
		got := buildGrowthSummary(nil)
		if got != (entities.PortfolioGrowthSummary{}) {
			t.Errorf("buildGrowthSummary(nil) = %+v, want zero value", got)
		}
	})

	t.Run("computes growth between first and last point", func(t *testing.T) {
		points := []entities.PortfolioGrowthPoint{
			{TotalValue: "1000.00"},
			{TotalValue: "900.00"},
			{TotalValue: "1250.00"},
		}

		got := buildGrowthSummary(points)
		if got.InitialValue != "1000.00" {
			t.Errorf("InitialValue = %q, want 1000.00", got.InitialValue)
		}
		if got.CurrentValue != "1250.00" {
			t.Errorf("CurrentValue = %q, want 1250.00", got.CurrentValue)
		}
		if got.TotalGrowthPct != "25.00" {
			t.Errorf("TotalGrowthPct = %q, want 25.00", got.TotalGrowthPct)
		}
	})

	t.Run("negative growth", func(t *testing.T) {
		points := []entities.PortfolioGrowthPoint{
			{TotalValue: "200.00"},
			{TotalValue: "150.00"},
		}

		got := buildGrowthSummary(points)
		if got.TotalGrowthPct != "-25.00" {
			t.Errorf("TotalGrowthPct = %q, want -25.00", got.TotalGrowthPct)
		}
	})

	t.Run("zero initial value avoids division by zero", func(t *testing.T) {
		points := []entities.PortfolioGrowthPoint{
			{TotalValue: "0"},
			{TotalValue: "500.00"},
		}

		got := buildGrowthSummary(points)
		if got.TotalGrowthPct != "0.00" {
			t.Errorf("TotalGrowthPct = %q, want 0.00 when starting from zero", got.TotalGrowthPct)
		}
	})
}

func TestGetAssetTransactionsPaginated(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()

	var gotLimit, gotOffset int
	repo := &fakeRepository{
		countAssetTransactions: func(context.Context, uuid.UUID, uuid.UUID, string) (int, error) {
			return 42, nil
		},
		getAssetTransactionsPaginated: func(_ context.Context, _, _ uuid.UUID, _ string, limit, offset int) ([]entities.Transaction, error) {
			gotLimit, gotOffset = limit, offset
			return []entities.Transaction{{ID: uuid.New()}}, nil
		},
	}
	svc := newTestServices(repo, newMemStorage())

	txns, total, err := svc.GetAssetTransactionsPaginated(context.Background(), userID, portfolioID, "AAPL", 3, 10)
	if err != nil {
		t.Fatalf("GetAssetTransactionsPaginated: %v", err)
	}
	if total != 42 {
		t.Errorf("total = %d, want 42", total)
	}
	if len(txns) != 1 {
		t.Errorf("len(txns) = %d, want 1", len(txns))
	}
	if gotLimit != 10 || gotOffset != 20 {
		t.Errorf("limit/offset = %d/%d, want 10/20 (page 3 with limit 10)", gotLimit, gotOffset)
	}
}

func TestGetPortfolioGrowthUsesPeriodFilter(t *testing.T) {
	userID := uuid.New()
	var gotHasSince bool
	var gotSince time.Time

	repo := &fakeRepository{
		getPortfolioGrowthByUserID: func(_ context.Context, uid uuid.UUID, hasSince bool, since time.Time) ([]entities.PortfolioGrowthPoint, error) {
			if uid != userID {
				t.Errorf("userID = %s, want %s", uid, userID)
			}
			gotHasSince, gotSince = hasSince, since
			return []entities.PortfolioGrowthPoint{
				{TotalValue: "100.00"},
				{TotalValue: "110.00"},
			}, nil
		},
	}
	svc := newTestServices(repo, newMemStorage())

	points, summary, err := svc.GetPortfolioGrowth(context.Background(), userID, "1M")
	if err != nil {
		t.Fatalf("GetPortfolioGrowth: %v", err)
	}
	if !gotHasSince {
		t.Error("expected the 1M period to enable the since filter")
	}
	if gotSince.After(time.Now().UTC()) {
		t.Error("since must be in the past")
	}
	if len(points) != 2 {
		t.Errorf("len(points) = %d, want 2", len(points))
	}
	if summary.TotalGrowthPct != "10.00" {
		t.Errorf("TotalGrowthPct = %q, want 10.00", summary.TotalGrowthPct)
	}
}
