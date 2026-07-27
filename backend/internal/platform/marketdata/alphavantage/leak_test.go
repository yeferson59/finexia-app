package alphavantage

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

// Alpha Vantage accepts the key only as a URL query parameter, so every request
// URL carries it. Go's transport errors quote the URL they failed on, which
// makes an unscrubbed error a direct path from a user's key to our logs and to
// HTTP response bodies. These tests pin the scrubbing shut.

const apiKey = "test-key"

func TestTransportErrorsNeverCarryTheKey(t *testing.T) {
	// The failure mode that matters: the transport itself errors, and the
	// error it returns embeds the full request URL.
	c := newTestClient(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("dial tcp 1.2.3.4:443: connect: connection refused")
	})

	t.Run("FetchQuote", func(t *testing.T) {
		_, err := c.FetchQuote(context.Background(), "AAPL")
		assertScrubbed(t, err)
	})

	t.Run("FetchExchangeRate", func(t *testing.T) {
		_, err := c.FetchExchangeRate(context.Background(), "EUR", "USD")
		assertScrubbed(t, err)
	})
}

// The provider's own reply is echoed into the error message; it must be
// scrubbed too, in case the provider quotes the key back at us.
func TestProviderMessagesNeverCarryTheKey(t *testing.T) {
	c := newTestClient(func(*http.Request) (*http.Response, error) {
		return jsonResponse(`{"Error Message":"Invalid API call. The apikey test-key is not valid."}`), nil
	})

	_, err := c.FetchQuote(context.Background(), "AAPL")
	assertScrubbed(t, err)

	if !errors.Is(err, marketdata.ErrUnauthorized) {
		t.Errorf("err = %v, want ErrUnauthorized so the key gets marked invalid", err)
	}
}

func TestRateLimitIsNotMistakenForABadKey(t *testing.T) {
	// A spent free-tier quota must leave the key usable, or one busy day would
	// permanently disable it.
	c := newTestClient(func(*http.Request) (*http.Response, error) {
		return jsonResponse(`{"Note":"Thank you for using Alpha Vantage! Our standard API rate limit is 25 requests per day."}`), nil
	})

	_, err := c.FetchQuote(context.Background(), "AAPL")

	if !errors.Is(err, marketdata.ErrRateLimited) {
		t.Fatalf("err = %v, want ErrRateLimited", err)
	}
	if errors.Is(err, marketdata.ErrUnauthorized) {
		t.Error("an exhausted quota must not be classified as a bad key")
	}
}

func TestNonOKStatusIsClassified(t *testing.T) {
	tests := map[int]error{
		http.StatusUnauthorized:    marketdata.ErrUnauthorized,
		http.StatusForbidden:       marketdata.ErrUnauthorized,
		http.StatusTooManyRequests: marketdata.ErrRateLimited,
	}

	for status, want := range tests {
		c := newTestClient(func(*http.Request) (*http.Response, error) {
			return new(http.Response{
				StatusCode: status,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}), nil
		})

		_, err := c.FetchQuote(context.Background(), "AAPL")
		if !errors.Is(err, want) {
			t.Errorf("status %d gave %v, want %v", status, err, want)
		}
		assertScrubbed(t, err)
	}
}

func assertScrubbed(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), apiKey) {
		t.Fatalf("the error carries the API key: %s", err)
	}
}
