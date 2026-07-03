package yahoo

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
	c := New()
	c.httpClient.Transport = fn
	return c
}

const chartBody = `{"chart":{"result":[{"meta":{"regularMarketPrice":192.53}}]}}`

func TestFetchQuote(t *testing.T) {
	t.Run("returns the regular market price", func(t *testing.T) {
		var gotURL string
		var gotUA string
		c := newTestClient(func(r *http.Request) (*http.Response, error) {
			gotURL = r.URL.String()
			gotUA = r.Header.Get("User-Agent")
			return jsonResponse(chartBody), nil
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
		if !strings.HasSuffix(gotURL, "/AAPL") {
			t.Errorf("request URL = %q, want it to end with /AAPL", gotURL)
		}
		if gotUA == "" {
			t.Error("expected a User-Agent header to be set")
		}
	})

	t.Run("errors when there is no result", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return jsonResponse(`{"chart":{"result":[]}}`), nil
		})

		if _, err := c.FetchQuote(context.Background(), "NOPE"); err == nil {
			t.Fatal("expected error for empty result")
		}
	})

	t.Run("errors on a zero price", func(t *testing.T) {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return jsonResponse(`{"chart":{"result":[{"meta":{"regularMarketPrice":0}}]}}`), nil
		})

		if _, err := c.FetchQuote(context.Background(), "AAPL"); err == nil {
			t.Fatal("expected error for zero price")
		}
	})
}

func TestFetchExchangeRate(t *testing.T) {
	t.Run("builds the =X symbol and returns the rate", func(t *testing.T) {
		var gotURL string
		c := newTestClient(func(r *http.Request) (*http.Response, error) {
			gotURL = r.URL.String()
			return jsonResponse(`{"chart":{"result":[{"meta":{"regularMarketPrice":1.085}}]}}`), nil
		})

		res, err := c.FetchExchangeRate(context.Background(), "EUR", "USD")
		if err != nil {
			t.Fatalf("FetchExchangeRate: %v", err)
		}
		if res.Rate != "1.085" {
			t.Errorf("Rate = %q, want 1.085", res.Rate)
		}
		if !strings.HasSuffix(gotURL, "/EURUSD=X") {
			t.Errorf("request URL = %q, want it to end with /EURUSD=X", gotURL)
		}
	})
}
