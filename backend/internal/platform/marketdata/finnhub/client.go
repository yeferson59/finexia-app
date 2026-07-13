package finnhub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/marketdata"
)

const baseURL = "https://finnhub.io/api/v1"

var _ marketdata.Provider = (*Client)(nil)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func New(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// FetchQuote retrieves the current price for a stock, ETF, or bond via the
// Finnhub /quote endpoint. Returns an error if the price is zero (symbol not
// found or API key limit reached).
func (c *Client) FetchQuote(ctx context.Context, symbol string) (marketdata.QuoteResult, error) {
	url := fmt.Sprintf("%s/quote?symbol=%s&token=%s", baseURL, symbol, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return marketdata.QuoteResult{}, fmt.Errorf("finnhub: build request %s: %w", symbol, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return marketdata.QuoteResult{}, fmt.Errorf("finnhub: http get %s: %w", symbol, err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		C float64 `json:"c"` // current price
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return marketdata.QuoteResult{}, fmt.Errorf("finnhub: decode %s: %w", symbol, err)
	}

	if result.C == 0 {
		return marketdata.QuoteResult{}, fmt.Errorf("finnhub: zero price for %s (API limit reached or invalid symbol)", symbol)
	}

	return marketdata.QuoteResult{
		Price:     strconv.FormatFloat(result.C, 'f', -1, 64),
		FetchedAt: time.Now().UTC(),
	}, nil
}

// FetchExchangeRate retrieves the rate between two fiat currencies via the
// Finnhub /forex/rates endpoint. Crypto pairs are not supported by this
// endpoint and will return an error, allowing the fallback chain to continue.
func (c *Client) FetchExchangeRate(ctx context.Context, from, to string) (marketdata.ExchangeRateResult, error) {
	url := fmt.Sprintf("%s/forex/rates?base=%s&token=%s", baseURL, from, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return marketdata.ExchangeRateResult{}, fmt.Errorf("finnhub: build request %s/%s: %w", from, to, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return marketdata.ExchangeRateResult{}, fmt.Errorf("finnhub: http get %s/%s: %w", from, to, err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Quote map[string]float64 `json:"quote"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return marketdata.ExchangeRateResult{}, fmt.Errorf("finnhub: decode %s/%s: %w", from, to, err)
	}

	rate, ok := result.Quote[to]
	if !ok || rate == 0 {
		return marketdata.ExchangeRateResult{}, fmt.Errorf("finnhub: missing rate for %s/%s (unsupported pair or API limit)", from, to)
	}

	return marketdata.ExchangeRateResult{
		Rate:      strconv.FormatFloat(rate, 'f', -1, 64),
		FetchedAt: time.Now().UTC(),
	}, nil
}
