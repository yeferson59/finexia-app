package marketdata

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yeferson59/gofinance/v2/money"
)

type stubSource struct {
	rates []PublicRate
	err   error
}

func (s stubSource) FetchRates(context.Context) ([]PublicRate, error) { return s.rates, s.err }

// The keyless feeds are additive, not a fallback chain: no single one covers
// every pair the app converts, so what one publishes never substitutes for
// another and one being down must not silence the rest.
func TestPublicRateSources(t *testing.T) {
	trm := PublicRate{From: money.USD, To: money.COP, Rate: "3157.43", Source: DolarAPI}
	eur := PublicRate{From: money.EUR, To: money.USD, Rate: "1.16", Source: ECB}

	t.Run("returns every feed's rates together", func(t *testing.T) {
		sources := PublicRateSources{stubSource{rates: []PublicRate{trm}}, stubSource{rates: []PublicRate{eur}}}

		rates, err := sources.FetchRates(context.Background())
		if err != nil {
			t.Fatalf("FetchRates: %v", err)
		}
		if len(rates) != 2 || rates[0].To != money.COP || rates[1].From != money.EUR {
			t.Errorf("rates = %+v, want the TRM followed by the euro pair", rates)
		}
	})

	t.Run("keeps the working feeds when one fails", func(t *testing.T) {
		sources := PublicRateSources{
			stubSource{err: errors.New("ecb: status 503")},
			stubSource{rates: []PublicRate{trm}},
		}

		rates, err := sources.FetchRates(context.Background())
		if len(rates) != 1 || rates[0].To != money.COP {
			t.Errorf("rates = %+v, want the pair the working feed published", rates)
		}
		// Still reported: a silent partial refresh is how a feed stays broken
		// for weeks while the table quietly goes stale.
		if err == nil || !strings.Contains(err.Error(), "503") {
			t.Errorf("err = %v, want the failing feed reported", err)
		}
	})

	t.Run("fails when every feed fails", func(t *testing.T) {
		sources := PublicRateSources{
			stubSource{err: errors.New("ecb: status 503")},
			stubSource{err: errors.New("dolarapi: status 500")},
		}

		rates, err := sources.FetchRates(context.Background())
		if rates != nil {
			t.Errorf("rates = %+v, want none", rates)
		}
		if err == nil {
			t.Fatal("expected an error")
		}
		for _, want := range []string{"ecb", "dolarapi"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("err = %v, want it to name %s", err, want)
			}
		}
	})
}
