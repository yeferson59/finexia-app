package alphavantage

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// roundTripFunc lets a test stub the HTTP transport, capturing the outgoing
// request and returning a canned response without touching the network.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

// newTestClient builds a Client whose HTTP transport is replaced by fn.
func newTestClient(fn roundTripFunc) *Client {
	c := New("test-key")
	c.httpClient.Transport = fn
	return c
}

func TestFetchQuote(t *testing.T) {
	t.Run("returns price from a valid response", func(t *testing.T) {
		var gotURL string
		c := newTestClient(func(r *http.Request) (*http.Response, error) {
			gotURL = r.URL.String()
			return jsonResponse(`{"Global Quote":{"01. symbol":"AAPL","05. price":"192.53"}}`), nil
		})

		res, err := c.FetchQuote(context.Background(), "AAPL")
		if err != nil {
			t.Fatalf("FetchQuote: %v", err)
		}
		if res.Price != "192.53" {
			t.Errorf("Price = %q, want 192.53", res.Price)
		}
		if res.FetchedAt.IsZero() {
			t.Error("FetchedAt should be set")
		}
		if !strings.Contains(gotURL, "function=GLOBAL_QUOTE") ||
			!strings.Contains(gotURL, "symbol=AAPL") ||
			!strings.Contains(gotURL, "apikey=test-key") {
			t.Errorf("request URL = %q, missing expected query params", gotURL)
		}
	})

	t.Run("errors when the price is missing", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return jsonResponse(`{"Global Quote":{}}`), nil
		})

		if _, err := c.FetchQuote(context.Background(), "AAPL"); err == nil {
			t.Fatal("expected error for missing price")
		}
	})

	t.Run("errors on malformed JSON", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return jsonResponse(`not json`), nil
		})

		if _, err := c.FetchQuote(context.Background(), "AAPL"); err == nil {
			t.Fatal("expected decode error")
		}
	})
}

func TestFetchExchangeRate(t *testing.T) {
	t.Run("returns rate from a valid response", func(t *testing.T) {
		var gotURL string
		c := newTestClient(func(r *http.Request) (*http.Response, error) {
			gotURL = r.URL.String()
			return jsonResponse(`{"Realtime Currency Exchange Rate":{"5. Exchange Rate":"1.0850"}}`), nil
		})

		res, err := c.FetchExchangeRate(context.Background(), "EUR", "USD")
		if err != nil {
			t.Fatalf("FetchExchangeRate: %v", err)
		}
		if res.Rate != "1.0850" {
			t.Errorf("Rate = %q, want 1.0850", res.Rate)
		}
		if res.FetchedAt.IsZero() {
			t.Error("FetchedAt should be set")
		}
		if !strings.Contains(gotURL, "function=CURRENCY_EXCHANGE_RATE") ||
			!strings.Contains(gotURL, "from_currency=EUR") ||
			!strings.Contains(gotURL, "to_currency=USD") {
			t.Errorf("request URL = %q, missing expected query params", gotURL)
		}
	})

	t.Run("errors when the rate is missing", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return jsonResponse(`{"Realtime Currency Exchange Rate":{"5. Exchange Rate":""}}`), nil
		})

		if _, err := c.FetchExchangeRate(context.Background(), "EUR", "USD"); err == nil {
			t.Fatal("expected error for missing rate")
		}
	})
}
