package market

import (
	"context"

	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// PublicRatesJob keeps the shared exchange rates current.
//
// It is deliberately not a step inside SyncJob. That one walks the users who
// configured a key and spends each one's personal quota, which is why it runs
// once at the market open and never retries. This fetch is keyless, costs
// nobody's quota and produces one result for everybody, so it runs on its own
// short cadence and a retry is free.
type PublicRatesJob struct {
	service *Service
	log     logger.Logger
}

func NewPublicRatesJob(service *Service, log logger.Logger) *PublicRatesJob {
	return new(PublicRatesJob{
		service: service,
		log:     log.With(logger.Str("scheduler", "public_rates")),
	})
}

func (j *PublicRatesJob) Name() string { return "public-exchange-rates" }

func (j *PublicRatesJob) Run(ctx context.Context) error {
	// A partial refresh is still a refresh: RefreshPublicRates returns the rates
	// it stored alongside the failures, so what landed is logged either way.
	rates, err := j.service.RefreshPublicRates(ctx)

	for _, rate := range rates {
		j.log.Info(ctx, "public exchange rate updated",
			logger.Str("pair", rate.FromCurrency.String()+"/"+rate.ToCurrency.String()),
			logger.Str("rate", rate.Rate.String()),
			logger.Str("source", string(rate.Source)),
		)
	}

	return err
}
