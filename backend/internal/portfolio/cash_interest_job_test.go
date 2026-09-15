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

func TestCashInterestJobReportsFailures(t *testing.T) {
	svc := &fakeCashInterestService{errs: []error{errors.New("first"), errors.New("second")}}
	job := NewCashInterestJob(svc, logger.Noop())

	if err := job.Run(context.Background()); err == nil {
		t.Error("Run = nil, want the failure reported to the scheduler")
	}
}
