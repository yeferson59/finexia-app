package marketdata

import (
	"context"
	"errors"
)

var _ Provider = (*FallbackProvider)(nil)

// FallbackProvider tries each provider in order and returns the first
// successful result. If all providers fail, it returns a joined error
// containing all individual failures.
type FallbackProvider struct {
	providers []Provider
}

func NewFallback(providers ...Provider) *FallbackProvider {
	return &FallbackProvider{providers: providers}
}

func (f *FallbackProvider) FetchQuote(ctx context.Context, symbol string) (QuoteResult, error) {
	var errs []error
	for _, p := range f.providers {
		result, err := p.FetchQuote(ctx, symbol)
		if err == nil {
			return result, nil
		}
		errs = append(errs, err)
	}
	return QuoteResult{}, errors.Join(errs...)
}

func (f *FallbackProvider) FetchExchangeRate(ctx context.Context, from, to string) (ExchangeRateResult, error) {
	var errs []error
	for _, p := range f.providers {
		result, err := p.FetchExchangeRate(ctx, from, to)
		if err == nil {
			return result, nil
		}
		errs = append(errs, err)
	}
	return ExchangeRateResult{}, errors.Join(errs...)
}
