package portfolio

import (
	"context"
	"errors"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

type PublicFundService interface {
	RefreshPublicFunds(ctx context.Context) (int, error)
	ImportPublicFundValues(ctx context.Context) (int, []error)
}

// PublicFundJob keeps the SFC's catalog of funds current and brings every
// linked fund up to date with the unit values published since its last run.
//
// It is the public exchange rate job's twin, not the market sync's: the feed is
// keyless and public, one read serves every owner who linked the same fund, and
// a retry costs nobody's quota. The SFC publishes with two days of delay at an
// hour it does not promise, so the job runs several times a day and a run that
// finds nothing new writes nothing.
//
// A late value is what the marks were built for: each one revalues the
// snapshots from its day on, so the growth chart shows a fund's move on the day
// it happened, not on the day the SFC told.
type PublicFundJob struct {
	svc PublicFundService
	log logger.Logger
}

func NewPublicFundJob(svc PublicFundService, log logger.Logger) *PublicFundJob {
	return new(PublicFundJob{
		svc: svc,
		log: log.With(logger.Str("scheduler", "public_funds")),
	})
}

func (j *PublicFundJob) Name() string { return "public-fund-values" }

func (j *PublicFundJob) Run(ctx context.Context) error {
	// A catalog that could not be refreshed still names every fund some owner
	// linked, so the import runs either way.
	catalog, catalogErr := j.svc.RefreshPublicFunds(ctx)
	if catalogErr != nil {
		j.log.Error(ctx, "public fund catalog refresh failed", logger.Err(catalogErr))
	}

	written, errs := j.svc.ImportPublicFundValues(ctx)

	j.log.Info(ctx, "public fund values imported",
		logger.Int("catalog", catalog),
		logger.Int("marks", written),
		logger.Int("failed", len(errs)),
	)

	return errors.Join(append([]error{catalogErr}, errs...)...)
}
