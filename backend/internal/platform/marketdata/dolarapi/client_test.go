package dolarapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/gofinance/v2/money"
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

func TestFetchRates(t *testing.T) {
	t.Run("maps the TRM to USD/COP", func(t *testing.T) {
		var gotURL string
		c := newTestClient(func(r *http.Request) (*http.Response, error) {
			gotURL = r.URL.String()

			return response(http.StatusOK, `{"unidad":"COP","nombre":"TRM","valor":3157.43,"fechaActualizacion":"2026-08-09T18:01:11.985Z"}`), nil
		})

		rates, err := c.FetchRates(context.Background())
		if err != nil {
			t.Fatalf("FetchRates: %v", err)
		}
		if len(rates) != 1 {
			t.Fatalf("got %d rates, want 1", len(rates))
		}

		got := rates[0]
		if got.From != money.USD || got.To != money.COP {
			t.Errorf("pair = %s/%s, want USD/COP", got.From, got.To)
		}
		// The stored column is numeric, so the published decimal must survive as
		// text rather than arriving as a float's shortest-but-different form.
		if got.Rate != "3157.43" {
			t.Errorf("Rate = %q, want 3157.43", got.Rate)
		}
		if got.Source != marketdata.DolarAPI {
			t.Errorf("Source = %q, want dolarapi", got.Source)
		}
		if want := time.Date(2026, 8, 9, 18, 1, 11, 985_000_000, time.UTC); !got.AsOf.Equal(want) {
			t.Errorf("AsOf = %s, want %s", got.AsOf, want)
		}
		if !strings.HasSuffix(gotURL, "/v1/trm") {
			t.Errorf("request URL = %q, want the TRM endpoint", gotURL)
		}
	})

	// The unit is read from the payload rather than hardcoded, so a feed that
	// ever changed it stores pesos under the code it actually published.
	t.Run("takes the target currency from the payload", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return response(http.StatusOK, `{"unidad":"USD","valor":1,"fechaActualizacion":"2026-08-09T18:00:00Z"}`), nil
		})

		rates, err := c.FetchRates(context.Background())
		if err != nil {
			t.Fatalf("FetchRates: %v", err)
		}
		if rates[0].To != money.USD {
			t.Errorf("To = %q, want the payload's unit", rates[0].To)
		}
	})

	t.Run("dates the rate now when the feed publishes no timestamp", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return response(http.StatusOK, `{"unidad":"COP","valor":3157.43}`), nil
		})

		rates, err := c.FetchRates(context.Background())
		if err != nil {
			t.Fatalf("FetchRates: %v", err)
		}
		if rates[0].AsOf.IsZero() {
			t.Error("AsOf is zero; a missing timestamp must fall back to now")
		}
	})

	// A rate of zero or less is one money.Convert would refuse, so it is stopped
	// at the edge rather than stored and hit later by every conversion.
	t.Run("rejects a non-positive rate", func(t *testing.T) {
		for _, body := range []string{`{"unidad":"COP","valor":0}`, `{"unidad":"COP","valor":-1}`} {
			c := newTestClient(func(*http.Request) (*http.Response, error) {
				return response(http.StatusOK, body), nil
			})

			if _, err := c.FetchRates(context.Background()); !errors.Is(err, marketdata.ErrUnsupported) {
				t.Errorf("FetchRates(%s) error = %v, want ErrUnsupported", body, err)
			}
		}
	})

	t.Run("rejects a payload with no currency unit", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return response(http.StatusOK, `{"valor":3157.43}`), nil
		})

		if _, err := c.FetchRates(context.Background()); !errors.Is(err, marketdata.ErrUnsupported) {
			t.Errorf("error = %v, want ErrUnsupported", err)
		}
	})

	t.Run("classifies a throttled feed", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return response(http.StatusTooManyRequests, ``), nil
		})

		if _, err := c.FetchRates(context.Background()); !errors.Is(err, marketdata.ErrRateLimited) {
			t.Errorf("error = %v, want ErrRateLimited", err)
		}
	})

	// An upstream outage says nothing worth classifying: it is neither a bad
	// symbol nor a quota, and the refresh job simply keeps the previous rate.
	t.Run("leaves an upstream failure unclassified", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return response(http.StatusInternalServerError, ``), nil
		})

		_, err := c.FetchRates(context.Background())
		if err == nil {
			t.Fatal("expected an error for a 500")
		}
		for _, sentinel := range []error{marketdata.ErrRateLimited, marketdata.ErrUnsupported, marketdata.ErrUnauthorized} {
			if errors.Is(err, sentinel) {
				t.Errorf("a 500 was classified as %v", sentinel)
			}
		}
	})

	t.Run("reports a transport failure", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("dial tcp: connection refused")
		})

		if _, err := c.FetchRates(context.Background()); err == nil {
			t.Fatal("expected an error when the request never completes")
		}
	})

	// Attribution is what lets a caller tell which source failed when several
	// are consulted. The keyless feed is no exception.
	t.Run("attributes its failures to the feed", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return response(http.StatusBadGateway, ``), nil
		})

		_, err := c.FetchRates(context.Background())

		verdicts := marketdata.Verdicts(err)
		if len(verdicts) != 1 || verdicts[0].Provider != marketdata.DolarAPI {
			t.Errorf("Verdicts = %+v, want one attributed to dolarapi", verdicts)
		}
	})
}

// The feed takes no key, so it must not be offerable as one a user can bring:
// IsValid is what the credential endpoints screen on.
func TestDolarAPIIsNotABYOKeyProvider(t *testing.T) {
	if marketdata.DolarAPI.IsValid() {
		t.Error("dolarapi is accepted as a BYO-key provider; it takes no key")
	}
	for _, p := range marketdata.SupportedProviders {
		if p == marketdata.DolarAPI {
			t.Error("dolarapi appears in SupportedProviders, which is the BYO-key chain order")
		}
	}
}
