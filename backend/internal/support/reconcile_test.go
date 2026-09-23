package support

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/bold"
)

func TestReconcile(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)

	// One order of each kind the reconciliation has to tell apart.
	seed := func() *memRepository {
		repo := newMemRepository()
		repo.plant(Contribution{OrderID: "paid-no-webhook", Amount: 20_000, Status: StatusCreated, CreatedAt: now.Add(-2 * time.Hour)})
		repo.plant(Contribution{OrderID: "fresh", Amount: 20_000, Status: StatusCreated, CreatedAt: now.Add(-5 * time.Minute)})
		repo.plant(Contribution{OrderID: "pse-pending", Amount: 20_000, Status: StatusPending, CreatedAt: now.Add(-30 * 24 * time.Hour)})
		repo.plant(Contribution{OrderID: "abandoned", Amount: 20_000, Status: StatusCreated, CreatedAt: now.Add(-8 * 24 * time.Hour)})
		repo.plant(Contribution{OrderID: "old-approved", Amount: 20_000, Status: StatusApproved, CreatedAt: now.Add(-8 * 24 * time.Hour)})
		return repo
	}

	newReconciler := func(repo *memRepository, gw *fakeGateway, cfg Config) *service {
		svc := newService(repo, gw, cfg, nil)
		svc.now = func() time.Time { return now }
		return svc
	}

	t.Run("records what a lost webhook would have said and deletes the unpaid", func(t *testing.T) {
		repo := seed()
		gw := new(fakeGateway{
			payment: bold.Payment{Status: bold.StatusNoTransaction},
			byOrder: map[string]bold.Payment{
				"paid-no-webhook": {Status: bold.StatusApproved, Total: new(int64(20_000)), TransactionID: "TX1"},
				"pse-pending":     {Status: bold.StatusRejected},
			},
		})

		counts, err := newReconciler(repo, gw, enabledConfig).Reconcile(ctx)
		if err != nil {
			t.Fatalf("Reconcile: %v", err)
		}

		// fresh is left to its webhook; abandoned is past the window; old-approved
		// is settled. Only the other two are asked about.
		if counts.Checked != 2 || gw.calls != 2 {
			t.Errorf("checked = %d, calls = %d, want 2 and 2", counts.Checked, gw.calls)
		}
		if counts.Updated != 2 || counts.Failed != 0 || counts.Abandoned != 1 {
			t.Errorf("counts = %+v", counts)
		}

		if c, _ := repo.GetContribution(ctx, "paid-no-webhook"); c.Status != StatusApproved || c.PaymentID != "TX1" {
			t.Errorf("paid-no-webhook = %+v", c)
		}
		if c, _ := repo.GetContribution(ctx, "pse-pending"); c.Status != StatusRejected {
			t.Errorf("pse-pending = %s", c.Status)
		}
		if _, err := repo.GetContribution(ctx, "abandoned"); !errors.Is(err, ErrContributionNotFound) {
			t.Error("the abandoned order is still there")
		}
		for _, kept := range []string{"fresh", "old-approved"} {
			if _, err := repo.GetContribution(ctx, kept); err != nil {
				t.Errorf("%s was deleted", kept)
			}
		}
		if want := now.Add(-abandonAfter); !repo.deletedBefore.Equal(want) {
			t.Errorf("deleted before %v, want %v", repo.deletedBefore, want)
		}
	})

	t.Run("deletes nothing when Bold did not answer", func(t *testing.T) {
		repo := seed()
		gw := new(fakeGateway{err: errors.New("timeout")})

		counts, err := newReconciler(repo, gw, enabledConfig).Reconcile(ctx)
		if err != nil {
			t.Fatalf("Reconcile: %v", err)
		}
		if counts.Failed != 2 || counts.Abandoned != 0 {
			t.Errorf("counts = %+v", counts)
		}
		if _, err := repo.GetContribution(ctx, "abandoned"); err != nil {
			t.Error("an order was deleted without Bold confirming it was never paid")
		}
	})

	t.Run("does nothing with contributions off", func(t *testing.T) {
		repo := seed()
		gw := new(fakeGateway{})

		counts, err := newReconciler(repo, gw, Config{}).Reconcile(ctx)
		if err != nil || counts != (ReconcileCounts{}) || gw.calls != 0 {
			t.Fatalf("counts = %+v, calls = %d, err = %v", counts, gw.calls, err)
		}
		if !repo.deletedBefore.IsZero() {
			t.Error("DeleteAbandoned ran with contributions off")
		}
	})

	t.Run("the job reports a failed run", func(t *testing.T) {
		job := NewReconcileJob(failingReconciler{}, nil)
		if job.Name() != "support-reconcile" {
			t.Errorf("Name = %q", job.Name())
		}
		if err := job.Run(ctx); err == nil {
			t.Error("Run swallowed the error")
		}
	})
}

type failingReconciler struct{}

func (failingReconciler) Reconcile(context.Context) (ReconcileCounts, error) {
	return ReconcileCounts{}, errors.New("database down")
}
