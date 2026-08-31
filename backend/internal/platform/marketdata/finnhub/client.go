// Package finnhub talks to the Finnhub API with a key the user brought
// themselves.
//
// Unlike Alpha Vantage, Finnhub accepts the key as a request header
// (X-Finnhub-Token), so it is kept out of the URL entirely: a transport error
// quoting the failed URL then has no key to leak. Errors still go through
// marketdata.Errorf as a second line of defence.
package finnhub

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	json "github.com/bytedance/sonic"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
	"github.com/yeferson59/gofinance/v2/money"
)

const baseURL = "https://finnhub.io/api/v1"

// tokenHeader carries the API key, keeping it out of the request URL.
const tokenHeader = "X-Finnhub-Token"

var _ marketdata.Provider = (*Client)(nil)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

// New builds a client for one user's key. Callers should pass the shared
// marketdata.DefaultHTTPClient; a nil client falls back to it rather than
// minting a new one per call.
func New(apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = marketdata.DefaultHTTPClient
	}

	return new(Client{apiKey: apiKey, httpClient: httpClient})
}

func (c *Client) get(ctx context.Context, path string, params url.Values, what string, out any) error {
	endpoint := baseURL + path + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return marketdata.Errorf(marketdata.Finnhub, c.apiKey, nil, "finnhub: build request %s: %v", what, err)
	}

	req.Header.Set(tokenHeader, c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return marketdata.Errorf(marketdata.Finnhub, c.apiKey, nil, "finnhub: http get %s: %v", what, err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusUnauthorized, resp.StatusCode == http.StatusForbidden:
		return marketdata.Errorf(marketdata.Finnhub, c.apiKey, marketdata.ErrUnauthorized, "finnhub: %s: status %d", what, resp.StatusCode)
	case resp.StatusCode == http.StatusTooManyRequests:
		return marketdata.Errorf(marketdata.Finnhub, c.apiKey, marketdata.ErrRateLimited, "finnhub: %s: status %d", what, resp.StatusCode)
	case resp.StatusCode != http.StatusOK:
		return marketdata.Errorf(marketdata.Finnhub, c.apiKey, nil, "finnhub: %s: status %d", what, resp.StatusCode)
	}

	if err := json.ConfigFastest.NewDecoder(resp.Body).Decode(out); err != nil {
		return marketdata.Errorf(marketdata.Finnhub, c.apiKey, nil, "finnhub: decode %s: %v", what, err)
	}

	return nil
}

type resultQuote struct {
	C float64 `json:"c"` // current price
}

// FetchQuote retrieves the current price for a stock, ETF, or bond via the
// Finnhub /quote endpoint. A zero price means the symbol is not covered, which
// lets the fallback chain move on to the next provider.
func (c *Client) FetchQuote(ctx context.Context, symbol string) (marketdata.QuoteResult, error) {
	var result resultQuote

	params := url.Values{}
	params.Set("symbol", symbol)

	if err := c.get(ctx, "/quote", params, symbol, &result); err != nil {
		return marketdata.QuoteResult{}, err
	}

	if result.C == 0 {
		return marketdata.QuoteResult{}, marketdata.Errorf(marketdata.Finnhub, c.apiKey, marketdata.ErrUnsupported, "finnhub: zero price for %s", symbol)
	}

	return marketdata.QuoteResult{
		Price:     strconv.FormatFloat(result.C, 'f', -1, 64),
		Source:    marketdata.Finnhub,
		FetchedAt: time.Now().UTC(),
	}, nil
}

type resultExchange struct {
	Quote map[string]float64 `json:"quote"`
}

// FetchExchangeRate retrieves the rate between two fiat currencies via the
// Finnhub /forex/rates endpoint. Crypto pairs are not served here and return
// ErrUnsupported, allowing the fallback chain to continue.
func (c *Client) FetchExchangeRate(ctx context.Context, from, to money.Currency) (marketdata.ExchangeRateResult, error) {
	what := from.String() + "/" + to.String()

	var result resultExchange

	params := url.Values{}
	params.Set("base", from.String())

	if err := c.get(ctx, "/forex/rates", params, what, &result); err != nil {
		return marketdata.ExchangeRateResult{}, err
	}

	rate, ok := result.Quote[to.String()]
	if !ok || rate == 0 {
		return marketdata.ExchangeRateResult{}, marketdata.Errorf(marketdata.Finnhub, c.apiKey, marketdata.ErrUnsupported, "finnhub: no rate for %s", what)
	}

	return marketdata.ExchangeRateResult{
		Rate:      strconv.FormatFloat(rate, 'f', -1, 64),
		Source:    marketdata.Finnhub,
		FetchedAt: time.Now().UTC(),
	}, nil
}
