package ecb

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/yeferson59/gofinance/v2/money"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func response(status int, body string) *http.Response {
	return new(http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	})
}

func newTestClient(fn roundTripFunc) *Client {
	// A dedicated client per test: New's default is process-wide, and mutating
	// its transport here would leak the stub into every other caller.
	return New(new(http.Client{Transport: fn}))
}

// daily is eurofxref-daily.xml trimmed to the currencies these tests read.
const daily = `<?xml version="1.0" encoding="UTF-8"?>
<gesmes:Envelope xmlns:gesmes="http://www.gesmes.org/xml/2002-08-01" xmlns="http://www.ecb.int/vocabulary/2002-08-01/eurofxref">
	<gesmes:subject>Reference rates</gesmes:subject>
	<Cube>
		<Cube time="2026-08-13">
			<Cube currency="USD" rate="1.2500"/>
			<Cube currency="JPY" rate="150.00"/>
			<Cube currency="GBP" rate="0.8000"/>
			<Cube currency="ISK" rate="150.10"/>
		</Cube>
	</Cube>
</gesmes:Envelope>`

func rateFor(rates []marketdata.PublicRate, from, to money.Currency) (marketdata.PublicRate, bool) {
	for _, r := range rates {
		if r.From == from && r.To == to {
			return r, true
		}
	}

	return marketdata.PublicRate{}, false
}

func TestFetchRates(t *testing.T) {
	t.Run("anchors the euro table to the dollar in both directions", func(t *testing.T) {
		var gotURL string
		c := newTestClient(func(r *http.Request) (*http.Response, error) {
			gotURL = r.URL.String()

			return response(http.StatusOK, daily), nil
		})

		rates, err := c.FetchRates(context.Background())
		if err != nil {
			t.Fatalf("FetchRates: %v", err)
		}
		if gotURL != dailyURL {
			t.Errorf("URL = %q, want %q", gotURL, dailyURL)
		}

		// 1 EUR = 1.25 USD, so a dollar buys 0.8 euro and a euro buys 1.25.
		eurUSD, ok := rateFor(rates, money.EUR, money.USD)
		if !ok || eurUSD.Rate != "1.25" {
			t.Errorf("EUR/USD = %q (found=%v), want 1.25", eurUSD.Rate, ok)
		}
		usdEUR, ok := rateFor(rates, money.USD, money.EUR)
		if !ok || usdEUR.Rate != "0.8" {
			t.Errorf("USD/EUR = %q (found=%v), want 0.8", usdEUR.Rate, ok)
		}

		// A cross pair is derived through the euro: 150 JPY per EUR over 1.25
		// USD per EUR is 120 JPY per dollar.
		usdJPY, ok := rateFor(rates, money.USD, money.JPY)
		if !ok || usdJPY.Rate != "120" {
			t.Errorf("USD/JPY = %q (found=%v), want 120", usdJPY.Rate, ok)
		}

		if usdJPY.Source != marketdata.ECB {
			t.Errorf("Source = %q, want ecb", usdJPY.Source)
		}
		// The rate is dated by the day the ECB published it, not by the read.
		if got := usdJPY.AsOf.Format("2006-01-02"); got != "2026-08-13" {
			t.Errorf("AsOf = %s, want the published 2026-08-13", got)
		}
	})

	// The feed quotes about thirty currencies; storing the ones nothing in the
	// app converts would bury the handful that matter in the shared table.
	t.Run("publishes only the majors", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return response(http.StatusOK, daily), nil
		})

		rates, err := c.FetchRates(context.Background())
		if err != nil {
			t.Fatalf("FetchRates: %v", err)
		}

		if _, ok := rateFor(rates, money.USD, money.ISK); ok {
			t.Error("ISK was published; only the majors should be")
		}
		// EUR, JPY and GBP are majors the document quotes: two directions each.
		if len(rates) != 6 {
			t.Errorf("got %d rates, want 6", len(rates))
		}
	})

	// Everything here is anchored to the dollar, so a table without it cannot
	// be used at all — better to say so than to store a euro-only fragment.
	t.Run("fails when the dollar is missing", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return response(http.StatusOK, `<gesmes:Envelope xmlns:gesmes="http://www.gesmes.org/xml/2002-08-01">
				<Cube><Cube time="2026-08-13"><Cube currency="JPY" rate="150.00"/></Cube></Cube>
			</gesmes:Envelope>`), nil
		})

		if _, err := c.FetchRates(context.Background()); !errors.Is(err, marketdata.ErrUnsupported) {
			t.Errorf("err = %v, want ErrUnsupported", err)
		}
	})

	t.Run("fails when no day is published", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return response(http.StatusOK, `<gesmes:Envelope xmlns:gesmes="http://www.gesmes.org/xml/2002-08-01"><Cube/></gesmes:Envelope>`), nil
		})

		if _, err := c.FetchRates(context.Background()); !errors.Is(err, marketdata.ErrUnsupported) {
			t.Errorf("err = %v, want ErrUnsupported", err)
		}
	})

	// A single unparsable quote is not a failed fetch: the currencies that did
	// parse are still worth storing.
	t.Run("skips a malformed quote and keeps the rest", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return response(http.StatusOK, `<gesmes:Envelope xmlns:gesmes="http://www.gesmes.org/xml/2002-08-01">
				<Cube><Cube time="2026-08-13">
					<Cube currency="USD" rate="1.2500"/>
					<Cube currency="GBP" rate="N/A"/>
				</Cube></Cube>
			</gesmes:Envelope>`), nil
		})

		rates, err := c.FetchRates(context.Background())
		if err != nil {
			t.Fatalf("FetchRates: %v", err)
		}
		if _, ok := rateFor(rates, money.USD, money.GBP); ok {
			t.Error("GBP was published from an unparsable quote")
		}
		if _, ok := rateFor(rates, money.EUR, money.USD); !ok {
			t.Error("EUR/USD is missing; a bad GBP quote must not drop it")
		}
	})

	t.Run("classifies a throttled response", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return response(http.StatusTooManyRequests, ""), nil
		})

		if _, err := c.FetchRates(context.Background()); !errors.Is(err, marketdata.ErrRateLimited) {
			t.Errorf("err = %v, want ErrRateLimited", err)
		}
	})

	t.Run("reports a transport failure as its own", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("dial tcp: no route to host")
		})

		_, err := c.FetchRates(context.Background())
		if err == nil || !strings.Contains(err.Error(), "ecb:") {
			t.Errorf("err = %v, want it attributed to ecb", err)
		}
	})
}
