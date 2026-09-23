package support

import (
	"context"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type reconciler interface {
	Reconcile(ctx context.Context) (ReconcileCounts, error)
}

// ReconcileJob runs Reconcile on the scheduler.
type ReconcileJob struct {
	svc reconciler
	log logger.Logger
}

func NewReconcileJob(svc reconciler, log logger.Logger) *ReconcileJob {
	if log == nil {
		log = logger.Noop()
	}

	return new(ReconcileJob{svc: svc, log: log.With(logger.Str("job", "support_reconcile"))})
}

func (j *ReconcileJob) Name() string {
	return "support-reconcile"
}

// Run returns the error only when the run as a whole failed — the listing or
// the deletion. Orders Bold could not answer about are in the log and are
// asked about again next time.
func (j *ReconcileJob) Run(ctx context.Context) error {
	counts, err := j.svc.Reconcile(ctx)

	j.log.Info(ctx, "support reconcile completed",
		logger.Int("checked", counts.Checked),
		logger.Int("updated", counts.Updated),
		logger.Int("failed", counts.Failed),
		logger.Int64("abandoned", counts.Abandoned),
		logger.Err(err),
	)

	return err
}
