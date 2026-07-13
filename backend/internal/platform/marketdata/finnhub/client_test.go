package finnhub

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func newTestClient(fn roundTripFunc) *Client {
	c := New("test-key")
	c.httpClient.Transport = fn
	return c
}

func TestFetchQuote(t *testing.T) {
	t.Run("formats the current price", func(t *testing.T) {
		var gotURL string
		c := newTestClient(func(r *http.Request) (*http.Response, error) {
			gotURL = r.URL.String()
			return jsonResponse(`{"c":192.53,"h":193.0,"l":190.0}`), nil
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
		if !strings.Contains(gotURL, "/quote") ||
			!strings.Contains(gotURL, "symbol=AAPL") ||
			!strings.Contains(gotURL, "token=test-key") {
			t.Errorf("request URL = %q, missing expected query params", gotURL)
		}
	})

	t.Run("errors on a zero price", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return jsonResponse(`{"c":0}`), nil
		})

		if _, err := c.FetchQuote(context.Background(), "AAPL"); err == nil {
			t.Fatal("expected error for zero price")
		}
	})

	t.Run("errors on malformed JSON", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return jsonResponse(`{`), nil
		})

		if _, err := c.FetchQuote(context.Background(), "AAPL"); err == nil {
			t.Fatal("expected decode error")
		}
	})
}

func TestFetchExchangeRate(t *testing.T) {
	t.Run("reads the target currency from the quote map", func(t *testing.T) {
		var gotURL string
		c := newTestClient(func(r *http.Request) (*http.Response, error) {
			gotURL = r.URL.String()
			return jsonResponse(`{"base":"EUR","quote":{"USD":1.085,"GBP":0.85}}`), nil
		})

		res, err := c.FetchExchangeRate(context.Background(), "EUR", "USD")
		if err != nil {
			t.Fatalf("FetchExchangeRate: %v", err)
		}
		if res.Rate != "1.085" {
			t.Errorf("Rate = %q, want 1.085", res.Rate)
		}
		if !strings.Contains(gotURL, "/forex/rates") ||
			!strings.Contains(gotURL, "base=EUR") {
			t.Errorf("request URL = %q, missing expected query params", gotURL)
		}
	})

	t.Run("errors when the pair is unsupported", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return jsonResponse(`{"base":"BTC","quote":{"GBP":0.85}}`), nil
		})

		if _, err := c.FetchExchangeRate(context.Background(), "BTC", "USD"); err == nil {
			t.Fatal("expected error for missing target currency")
		}
	})
}
