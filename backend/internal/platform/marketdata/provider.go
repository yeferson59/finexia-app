package marketdata

import (
	"context"
	"time"
)

type QuoteResult struct {
	Price     string
	FetchedAt time.Time
}

type ExchangeRateResult struct {
	Rate      string
	FetchedAt time.Time
}

type Provider interface {
	FetchQuote(ctx context.Context, symbol string) (QuoteResult, error)
	FetchExchangeRate(ctx context.Context, from, to string) (ExchangeRateResult, error)
}
