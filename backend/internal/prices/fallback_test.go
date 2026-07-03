package prices

import (
	"context"
	"errors"
	"testing"
)

// stubProvider is a configurable Provider used to drive the fallback chain.
type stubProvider struct {
	quote    QuoteResult
	quoteErr error
	rate     ExchangeRateResult
	rateErr  error
	calls    int
}

func (s *stubProvider) FetchQuote(context.Context, string) (QuoteResult, error) {
	s.calls++
	return s.quote, s.quoteErr
}

func (s *stubProvider) FetchExchangeRate(context.Context, string, string) (ExchangeRateResult, error) {
	s.calls++
	return s.rate, s.rateErr
}

func TestFallbackFetchQuote(t *testing.T) {
	t.Run("returns the first successful provider and stops", func(t *testing.T) {
		first := &stubProvider{quoteErr: errors.New("first down")}
		second := &stubProvider{quote: QuoteResult{Price: "10.00"}}
		third := &stubProvider{quote: QuoteResult{Price: "99.00"}}

		res, err := NewFallback(first, second, third).FetchQuote(context.Background(), "AAPL")
		if err != nil {
			t.Fatalf("FetchQuote: %v", err)
		}
		if res.Price != "10.00" {
			t.Errorf("Price = %q, want 10.00", res.Price)
		}
		if first.calls != 1 || second.calls != 1 {
			t.Errorf("calls = first:%d second:%d, want 1 each", first.calls, second.calls)
		}
		if third.calls != 0 {
			t.Errorf("third.calls = %d, want 0 (should short-circuit)", third.calls)
		}
	})

	t.Run("joins every error when all providers fail", func(t *testing.T) {
		errA := errors.New("provider A failed")
		errB := errors.New("provider B failed")
		f := NewFallback(
			&stubProvider{quoteErr: errA},
			&stubProvider{quoteErr: errB},
		)

		_, err := f.FetchQuote(context.Background(), "AAPL")
		if err == nil {
			t.Fatal("expected an error when all providers fail")
		}
		if !errors.Is(err, errA) || !errors.Is(err, errB) {
			t.Errorf("joined error = %v, want it to wrap both provider errors", err)
		}
	})

	t.Run("returns zero value and no error with no providers", func(t *testing.T) {
		res, err := NewFallback().FetchQuote(context.Background(), "AAPL")
		if err != nil {
			t.Errorf("err = %v, want nil", err)
		}
		if res != (QuoteResult{}) {
			t.Errorf("res = %+v, want zero value", res)
		}
	})
}

func TestFallbackFetchExchangeRate(t *testing.T) {
	t.Run("returns the first successful provider and stops", func(t *testing.T) {
		first := &stubProvider{rateErr: errors.New("first down")}
		second := &stubProvider{rate: ExchangeRateResult{Rate: "1.08"}}
		third := &stubProvider{rate: ExchangeRateResult{Rate: "9.99"}}

		res, err := NewFallback(first, second, third).FetchExchangeRate(context.Background(), "EUR", "USD")
		if err != nil {
			t.Fatalf("FetchExchangeRate: %v", err)
		}
		if res.Rate != "1.08" {
			t.Errorf("Rate = %q, want 1.08", res.Rate)
		}
		if third.calls != 0 {
			t.Errorf("third.calls = %d, want 0 (should short-circuit)", third.calls)
		}
	})

	t.Run("joins every error when all providers fail", func(t *testing.T) {
		errA := errors.New("rate A failed")
		errB := errors.New("rate B failed")
		f := NewFallback(
			&stubProvider{rateErr: errA},
			&stubProvider{rateErr: errB},
		)

		_, err := f.FetchExchangeRate(context.Background(), "EUR", "USD")
		if err == nil {
			t.Fatal("expected an error when all providers fail")
		}
		if !errors.Is(err, errA) || !errors.Is(err, errB) {
			t.Errorf("joined error = %v, want it to wrap both provider errors", err)
		}
	})
}
