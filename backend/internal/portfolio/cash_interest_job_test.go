package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type fakeCashInterestService struct {
	through time.Time
	errs    []error
}

func (f *fakeCashInterestService) AccrueCashInterest(_ context.Context, through time.Time) (int, []error) {
	f.through = through
	return 3, f.errs
}

// At 05:30 UTC on the 15th — half past midnight in Bogotá — the 14th is over
// everywhere the app's owners are, and it is the last day computed.
func TestCashInterestJobComputesThroughYesterday(t *testing.T) {
	svc := new(fakeCashInterestService)
	job := NewCashInterestJob(svc, logger.Noop())
	job.now = func() time.Time { return time.Date(2026, time.September, 15, 5, 30, 0, 0, time.UTC) }

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if want := time.Date(2026, time.September, 14, 0, 0, 0, 0, time.UTC); !svc.through.Equal(want) {
		t.Errorf("through = %v, want %v", svc.through, want)
	}
	if job.Name() != "accrue-cash-interest" {
		t.Errorf("name = %q", job.Name())
	}
}

// A run caught up at startup before its hour stops where the scheduled one
// would have: at 02:00 UTC on the 15th it is still the 14th in Bogotá.
func TestCashInterestJobBeforeItsHour(t *testing.T) {
	svc := new(fakeCashInterestService)
	job := NewCashInterestJob(svc, logger.Noop())
	job.now = func() time.Time { return time.Date(2026, time.September, 15, 2, 0, 0, 0, time.UTC) }

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if want := time.Date(2026, time.September, 13, 0, 0, 0, 0, time.UTC); !svc.through.Equal(want) {
		t.Errorf("through = %v, want %v", svc.through, want)
	}
}

// The last day closed turns at the nightly run's hour, not at midnight UTC.
func TestLastClosedCashDay(t *testing.T) {
	sep := func(day, hour, minute int) time.Time {
		return time.Date(2026, time.September, day, hour, minute, 0, 0, time.UTC)
	}

	cases := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{"an evening in Bogotá, already the next UTC day", sep(24, 1, 8), sep(22, 0, 0)},
		{"a minute before the run", sep(24, 5, 29), sep(22, 0, 0)},
		{"at the run", sep(24, 5, 30), sep(23, 0, 0)},
		{"the afternoon", sep(24, 18, 0), sep(23, 0, 0)},
		{"another zone, the same instant", time.Date(2026, time.September, 23, 20, 8, 0, 0, time.FixedZone("COT", -5*3600)), sep(22, 0, 0)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := lastClosedCashDay(tc.now); !got.Equal(tc.want) {
				t.Errorf("lastClosedCashDay(%v) = %v, want %v", tc.now, got, tc.want)
			}
		})
	}
}

func TestCashInterestJobReportsFailures(t *testing.T) {
	svc := &fakeCashInterestService{errs: []error{errors.New("first"), errors.New("second")}}
	job := NewCashInterestJob(svc, logger.Noop())

	if err := job.Run(context.Background()); err == nil {
		t.Error("Run = nil, want the failure reported to the scheduler")
	}
}
